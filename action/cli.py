"""Customer CLI preset. All processes use argv; target code runs only in proof containers."""
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tempfile
import uuid

import delivery

ROOT = Path(__file__).resolve().parent.parent
BACKENDS = ("claude", "copilot", "codex", "cursor", "qwen")
LABEL = "io.proven.cli-run"


def run(argv, *, capture=False, **kwargs):
    return subprocess.run(argv, check=True, text=True,
                          stdout=subprocess.PIPE if capture else None, **kwargs)


def output(key, value):
    if "\n" in value or "\r" in value:
        raise ValueError("invalid multiline Action output")
    with open(os.environ["GITHUB_OUTPUT"], "a") as f:
        f.write(f"{key}={value}\n")


def input_value(name, default=""):
    return os.environ.get("INPUT_" + name.upper(), "") or default


def local_image(image):
    # Resolve the immutable local identity once; later tag changes cannot switch it.
    inspected = subprocess.run(["docker", "image", "inspect", image], text=True,
                               stdout=subprocess.PIPE, stderr=subprocess.DEVNULL)
    if inspected.returncode:
        run(["docker", "pull", image])
        inspected = run(["docker", "image", "inspect", image], capture=True)
    item = json.loads(inspected.stdout)[0]
    if item["Os"] != "linux":
        raise ValueError("the CLI image must target Linux")
    return item["Id"]


def prepare_tests(state, target, baseline):
    # Only trusted base manifests enter the build. Never run a PR's Dockerfile,
    # package scripts or build script on the host or during dependency preparation.
    top = Path(run(["git", "-C", str(target), "rev-parse", "--show-toplevel"], capture=True).stdout.strip())
    prefix = target.relative_to(top).as_posix()
    prefix = "" if prefix == "." else prefix + "/"
    context = state / "test-build"
    context.mkdir()
    manifests = {}
    for name in ("go.mod", "go.sum", "package.json", "package-lock.json", "requirements.txt", "pyproject.toml"):
        data = subprocess.run(["git", "-C", str(target), "show", baseline + ":" + prefix + name],
                              stdout=subprocess.PIPE, stderr=subprocess.DEVNULL)
        if data.returncode == 0:
            if len(data.stdout) > 1024 * 1024:
                raise ValueError("Base dependency manifest exceeds 1 MiB; supply a prepared test-image")
            manifests[name] = data.stdout
            (context / name).write_bytes(data.stdout)
    languages = []
    if "go.mod" in manifests: languages.append("go")
    if "package.json" in manifests: languages.append("node")
    if "requirements.txt" in manifests or "pyproject.toml" in manifests: languages.append("python")
    if len(languages) != 1:
        raise ValueError("Automatic test preparation needs one Go, Node/TypeScript or Python module at target. Set target to that module, or supply a prepared test-image for a mixed project.")
    if languages[0] == "go":
        (context / "go.sum").touch(exist_ok=True)
        recipe = ("FROM golang:1.26.10\nWORKDIR /deps\nENV GOTOOLCHAIN=local\n"
                  "COPY go.mod go.sum ./\nRUN go mod download\n"
                  "ENV GOPROXY=off GOSUMDB=off CGO_ENABLED=0\n")
    elif languages[0] == "node":
        if "package-lock.json" not in manifests:
            raise ValueError("Automatic Node preparation needs package-lock.json at base; otherwise supply test-image")
        recipe = ("FROM node:24-alpine3.24\nWORKDIR /deps\nCOPY package.json package-lock.json ./\n"
                  "RUN npm ci --ignore-scripts --no-audit --no-fund\n"
                  "RUN test -f node_modules/vitest/vitest.mjs || npm install --ignore-scripts --no-audit --no-fund --no-save vitest@5.1.0\n")
    else:
        if "requirements.txt" not in manifests:
            raise ValueError("Automatic Python preparation needs requirements.txt at base; use test-image for pyproject/Poetry or private dependencies")
        recipe = ("FROM python:3.14-alpine3.24\nWORKDIR /deps\nCOPY requirements.txt ./\n"
                  "RUN pip install --no-cache-dir --only-binary=:all: -r requirements.txt pytest==9.2.0\n"
                  "ENV PYTHONDONTWRITEBYTECODE=1\n")
    (context / "Dockerfile").write_text(recipe)
    tag = "proven-tests:" + state.name
    run(["docker", "build", "-t", tag, str(context)])
    return local_image(tag)


def check_environment():
    if os.environ.get("PROVEN_NETWORK", "public") not in ("", "public"):
        raise ValueError("The external CLI preset requires PROVEN_NETWORK=public; it cannot enforce local HTTP policy inside a vendor CLI")
    if sys.platform not in ("linux", "darwin"):
        raise ValueError("This Action needs a Linux or macOS runner with Docker, Git and Python 3. See docs/START.md.")


def prepare():
    check_environment()
    context_token = os.environ.pop("INPUT_CONTEXT_TOKEN", "")
    backend = input_value("backend")
    if backend not in BACKENDS:
        raise ValueError("Choose backend: " + ", ".join(BACKENDS))
    seconds = input_value("call_timeout", "300")
    if not seconds.isascii() or not seconds.isdigit() or not 15 <= int(seconds) <= 86400:
        raise ValueError("call-timeout must be an integer from 15 to 86400 seconds")
    oauth = input_value("has_oauth_token") == "true"
    if oauth and backend != "claude":
        raise ValueError("claude-oauth-token applies only to Claude; use the other CLI's logged-in home-volume")
    if oauth and input_value("home_volume"):
        raise ValueError("Choose a logged-in home-volume or claude-oauth-token, not both")
    home = "" if oauth else input_value("home_volume", "proven-" + backend + "-home")
    if home:
        if not re.fullmatch(r"[a-zA-Z0-9][a-zA-Z0-9_.-]*", home):
            raise ValueError("home-volume must be a Docker volume name")
        if subprocess.run(["docker", "volume", "inspect", home], stdout=subprocess.DEVNULL,
                          stderr=subprocess.DEVNULL).returncode:
            if backend == "claude" and input_value("runner_environment") == "github-hosted":
                raise ValueError("Claude needs authentication. For GitHub-hosted runners, create an OAuth credential with claude setup-token, save it as the CLAUDE_CODE_OAUTH_TOKEN repository secret and pass claude-oauth-token: ${{ secrets.CLAUDE_CODE_OAUTH_TOKEN }}. See docs/START.md.")
            raise ValueError(f"No login volume {home}. On this runner, run: sh scripts/proven-cli.sh build {backend}, then login {backend}. See docs/START.md.")
    workspace = Path(os.environ["GITHUB_WORKSPACE"]).resolve()
    target = (workspace / input_value("target", ".")).resolve()
    baseline = input_value("baseline_commit")
    if not re.fullmatch(r"[a-fA-F0-9]{40}|[a-fA-F0-9]{64}", baseline):
        raise ValueError("baseline-commit must be the full base commit SHA present in the checkout")
    state = Path(tempfile.mkdtemp(prefix="proven-cli-", dir=os.environ["RUNNER_TEMP"]))
    os.chmod(state, 0o700)
    output("state", str(state))  # Cleanup also runs when preparation later fails.
    image = input_value("image")
    if not image:
        image = "proven-" + backend + ":" + state.name
        recipe = ROOT / "examples/cli" / ("Dockerfile.cursor" if backend == "cursor" else "Dockerfile")
        argv = ["docker", "build", "-f", str(recipe), "-t", image]
        if backend != "cursor":
            argv += ["--target", backend]
        run(argv + [str(ROOT)])
    image = local_image(image)
    # The controller comes from the pinned Action revision, independent of
    # the customer's CLI image age/OS. Only trusted tooling runs on the host.
    build_env = dict(os.environ, CGO_ENABLED="0", GOWORK="off", GOTOOLCHAIN="local")
    run(["go", "build", "-trimpath", "-mod=readonly", "-o", str(state / "proven"), "./cmd/proven"], cwd=ROOT, env=build_env)
    (state / "proven").chmod(0o700)
    run([str(state / "proven"), "--version"])
    model = input_value("model", {"claude": "sonnet", "copilot": "auto"}.get(backend, "cli-default"))
    settings = dict(backend=backend, image=image, home=home, oauth=oauth,
                    seconds=int(seconds), model=model)
    second = input_value("second_backend")
    if second:
        if second not in BACKENDS:
            raise ValueError("second-backend must be one of: " + ", ".join(BACKENDS))
        second_home = input_value("second_home_volume", "proven-" + second + "-home")
        if not re.fullmatch(r"[a-zA-Z0-9][a-zA-Z0-9_.-]*", second_home):
            raise ValueError("second-home-volume must be a Docker volume name")
        if subprocess.run(["docker", "volume", "inspect", second_home], stdout=subprocess.DEVNULL,
                          stderr=subprocess.DEVNULL).returncode:
            raise ValueError("Log the second CLI in on this runner before adding it to the workflow")
        second_image = input_value("second_image")
        if not second_image:
            second_image = "proven-" + second + ":" + state.name
            recipe = ROOT / "examples/cli" / ("Dockerfile.cursor" if second == "cursor" else "Dockerfile")
            argv = ["docker", "build", "-f", str(recipe), "-t", second_image]
            if second != "cursor":
                argv += ["--target", second]
            run(argv + [str(ROOT)])
        settings["secondary"] = dict(backend=second, image=local_image(second_image), home=second_home,
            model=input_value("second_model", {"claude": "sonnet", "copilot": "auto"}.get(second, "cli-default")),
            seconds=int(seconds), oauth=False)
    check_delivery = input_value("check_delivery", "false")
    if check_delivery not in ("true", "false"):
        raise ValueError("check-delivery takes true or false")
    settings["delivery"] = check_delivery == "true"
    if settings["delivery"]:
        metadata = delivery.collect(target, baseline, os.environ, context_token)
        (state / "pr-context.json").write_bytes(delivery.encoded(metadata))
    settings["test_image"] = (local_image(input_value("test_image")) if input_value("test_image")
                              else prepare_tests(state, target, baseline))
    (state / "settings.json").write_text(json.dumps(settings))


def load(state):
    return json.loads((state / "settings.json").read_text())


def model(state, selected_model, slot="primary"):
    cfg = load(state)
    if slot == "secondary":
        cfg = cfg["secondary"]
    elif slot != "primary":
        raise ValueError("invalid model slot")
    name = state.name + "-" + uuid.uuid4().hex[:12]
    argv = ["docker", "run", "--rm", "-i", "--name", name,
            "--label", LABEL + "=" + state.name, "--init", "--read-only",
            "--cap-drop", "ALL", "--security-opt", "no-new-privileges",
            "--pids-limit", "128", "--memory", "1g", "--cpus", "2",
            "--tmpfs", "/tmp:rw,nosuid,nodev,size=512m,mode=1777"]
    if cfg["oauth"]:
        uid, gid = os.getuid() or 10001, os.getgid() or 10001
        argv += ["--user", f"{uid}:{gid}", "-e", "HOME=/tmp/proven-home",
                 "--mount", f"type=bind,src={state / 'oauth-token'},dst=/run/proven-oauth-token,readonly"]
    else:
        argv += ["--mount", "type=volume,src=" + cfg["home"] + ",dst=/opt/proven"]
    argv += ["--entrypoint", "timeout", cfg["image"], "--kill-after=2s",
             str(cfg["seconds"] - 5) + "s"]
    if cfg["oauth"]:
        # Fixed shell program; neither the token nor prompt is interpolated.
        argv += ["sh", "-c", 'CLAUDE_CODE_OAUTH_TOKEN=$(cat /run/proven-oauth-token); '
                 'export CLAUDE_CODE_OAUTH_TOKEN; exec proven-claude "$1"', "proven", selected_model]
    else:
        argv += ["proven-" + cfg["backend"], selected_model]
    try:
        return subprocess.call(argv)
    finally:
        subprocess.run(["docker", "rm", "-f", name], stdout=subprocess.DEVNULL,
                       stderr=subprocess.DEVNULL, timeout=10)


def remove_containers(state):
    containers = run(["docker", "ps", "-aq", "--filter", "label=" + LABEL + "=" + state.name], capture=True).stdout.split()
    if containers:
        try:
            run(["docker", "rm", "-f", *containers], capture=True, timeout=15)
        except subprocess.SubprocessError as exc:
            raise ValueError("Model containers may remain: " + ", ".join(containers) + ". Remove containers with label " + LABEL + "=" + state.name) from exc


def review(state):
    cfg = load(state)
    env = dict(os.environ)
    token = env.pop("INPUT_CLAUDE_OAUTH_TOKEN", "")
    if cfg["oauth"]:
        if len(token.strip()) < 8 or "\n" in token or "\r" in token:
            raise ValueError("claude-oauth-token is missing or malformed; use the OAuth token printed by claude setup-token")
        fd = os.open(state / "oauth-token", os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
        with os.fdopen(fd, "w") as f:
            f.write(token)
        if os.getuid() == 0:
            os.chown(state / "oauth-token", 10001, os.getgid() or 10001)
        # Redaction only: Command's environment allowlist does not forward it.
        env["CLAUDE_CODE_OAUTH_TOKEN"] = token
    env.update(PROVEN_BIN=str(state / "proven"), PROVEN_LOG_DIR=str(state / "logs"),
               PROVEN_ACTION_HOST="true",
               PROVEN_EVIDENCE_STORE="sqlite", PROVEN_EVIDENCE_FILE=str(state / "evidence.sqlite"),
               PROVEN_GENERATOR_MODE="command", PROVEN_GENERATOR_MODEL=cfg["model"],
               PROVEN_GENERATOR_COMMAND=json.dumps([sys.executable, str(Path(__file__).resolve()), "model", str(state), "{model}"]),
               PROVEN_GENERATOR_TIMEOUT=str(cfg["seconds"]) + "s",
               PROVEN_ADAPTER_TIMEOUT=str(cfg["seconds"] + 10) + "s", PROVEN_TEST_IMAGE=cfg["test_image"])
    if cfg.get("delivery"):
        env["PROVEN_PR_CONTEXT"] = str(state / "pr-context.json")
    if cfg.get("secondary"):
        secondary = cfg["secondary"]
        env.update(PROVEN_SECOND_REVIEWER_MODE="command", PROVEN_SECOND_REVIEWER_MODEL=secondary["model"],
            PROVEN_SECOND_REVIEWER_COMMAND=json.dumps([sys.executable, str(Path(__file__).resolve()), "model", str(state), "{model}", "secondary"]),
            PROVEN_SECOND_REVIEWER_TIMEOUT=str(secondary["seconds"]) + "s")
    try:
        print("Checking the CLI login with one short request (uses your account quota).", flush=True)
        probe_env = {k: v for k, v in env.items() if k not in ("INPUT_GITHUB_TOKEN", "GITHUB_TOKEN", "ACTIONS_RUNTIME_TOKEN", "ACTIONS_ID_TOKEN_REQUEST_TOKEN", "ACTIONS_ID_TOKEN_REQUEST_URL")}
        probe_env["PROVEN_GENERATOR_TIMEOUT"] = str(min(cfg["seconds"], 60)) + "s"
        checked = subprocess.call([str(state / "proven"), "model", "check"], env=probe_env, cwd=state)
        if checked:
            print("CLI check failed. Re-run the login/check steps on this runner; verify model access and quota. No PR review was started.", file=sys.stderr)
            return checked
        if cfg.get("secondary"):
            print("Checking the second CLI with one short request.", flush=True)
            probe_env["PROVEN_GENERATOR_COMMAND"] = env["PROVEN_SECOND_REVIEWER_COMMAND"]
            probe_env["PROVEN_GENERATOR_MODEL"] = env["PROVEN_SECOND_REVIEWER_MODEL"]
            checked = subprocess.call([str(state / "proven"), "model", "check"], env=probe_env, cwd=state)
            if checked:
                print("Second CLI check failed; refresh its own login or model selection. No PR review was started.", file=sys.stderr)
                return checked
        return subprocess.call(["sh", str(ROOT / "action/entrypoint.sh")], env=env, cwd=state)
    finally:
        (state / "oauth-token").unlink(missing_ok=True)
        remove_containers(state)


def state_path(raw):
    state = Path(raw).resolve()
    parent = Path(os.environ["RUNNER_TEMP"]).resolve()
    if state.parent != parent or not state.name.startswith("proven-cli-"):
        raise ValueError("invalid private Action state directory")
    return state


def main():
    command = sys.argv[1]
    if command == "check-environment":
        check_environment()
        version = re.search(r"^go ([0-9]+\.[0-9]+(?:\.[0-9]+)?)$", (ROOT / "go.mod").read_text(), re.MULTILINE)
        if not version:
            raise ValueError("the Action source has no supported Go version declaration")
        output("go-version", version.group(1))
        return 0
    if command == "prepare":
        prepare()
        return 0
    if command == "model":
        # The command child receives only the environment allowlist, so its
        # validated state path is an argv value rather than RUNNER_TEMP.
        return model(Path(sys.argv[2]), sys.argv[3], sys.argv[4] if len(sys.argv) > 4 else "primary")
    state = state_path(os.environ["PROVEN_CLI_STATE"])
    if command == "review":
        return review(state)
    if command == "cleanup":
        try:
            remove_containers(state)
        finally:
            shutil.rmtree(state, ignore_errors=True)
        return 0
    raise ValueError("unknown Action operation")


if __name__ == "__main__":
    try:
        sys.exit(main())
    except (ValueError, OSError, subprocess.SubprocessError) as exc:
        # Do not print a CLI's captured output, login material or the env.
        print("proven cli action: " + (str(exc) if isinstance(exc, ValueError) else type(exc).__name__), file=sys.stderr)
        sys.exit(2)
