"""Read-only GitHub scope collector. Never imports or executes target code."""
import argparse
import json
import os
from pathlib import Path
import re
import subprocess
import threading
import time
from urllib.error import HTTPError, URLError
from urllib.parse import urlsplit
from urllib.request import HTTPRedirectHandler, Request, build_opener

MAX_BYTES = 64 * 1024
SHA = re.compile(r"[a-fA-F0-9]{40}|[a-fA-F0-9]{64}")


class NoRedirect(HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        raise HTTPError(req.full_url, code, "redirect refused", headers, fp)


def git(target, *args):
    # Bound the pipe while reading, before buffering untrusted commit messages.
    # The timer also ends a Git process that never closes its stdout.
    with subprocess.Popen(["git", "-C", str(target), *args], stdout=subprocess.PIPE,
                          stderr=subprocess.DEVNULL) as process:
        timer = threading.Timer(20, process.kill)
        timer.daemon = True
        timer.start()
        try:
            content = process.stdout.read(MAX_BYTES + 1)
            if len(content) > MAX_BYTES:
                process.kill()
                raise ValueError("Git scope context exceeds 64 KiB")
            if process.wait() != 0:
                raise ValueError("Git scope context could not be read within its 20-second limit")
            return content.decode("utf-8")
        finally:
            timer.cancel()


def encoded(metadata):
    return json.dumps(metadata, ensure_ascii=False).encode("utf-8")


def collect(target, baseline, env, token="", fetch=None):
    if not SHA.fullmatch(baseline):
        raise ValueError("delivery context needs the full base commit SHA")
    with Path(env["GITHUB_EVENT_PATH"]).open("rb") as event_file:
        event = event_file.read(2 * 1024 * 1024 + 1)
    if len(event) > 2 * 1024 * 1024:
        raise ValueError("PR event exceeds 2 MiB")
    event = json.loads(event)
    pr = event.get("pull_request")
    if not isinstance(pr, dict):
        raise ValueError("scope checking needs a pull_request event")
    repository = env.get("GITHUB_REPOSITORY", "")
    number = pr.get("number", event.get("number"))
    if not re.fullmatch(r"[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+", repository) or type(number) is not int or number < 1:
        raise ValueError("scope checking needs a repository and PR number")
    server = env.get("GITHUB_SERVER_URL", "https://github.com").rstrip("/")
    api = env.get("GITHUB_API_URL", "https://api.github.com").rstrip("/")
    for address in (server, api):
        parts = urlsplit(address)
        if parts.scheme != "https" or not parts.hostname or parts.username or parts.password or parts.query or parts.fragment:
            raise ValueError("GitHub context URLs must use HTTPS without embedded credentials")
    url = server + "/" + repository + "/pull/" + str(number)
    head = git(target, "rev-parse", "HEAD").strip()
    expected = pr.get("head", {}).get("sha")
    if expected and head != expected:
        raise ValueError("scope checking needs the exact PR head checkout, not a merge commit")
    commits = []
    records = git(target, "log", "--reverse", "--format=%H%x00%B%x00", "--end-of-options", baseline + "..HEAD").split("\0")
    for index in range(0, len(records) - 1, 2):
        sha = records[index].strip()
        if not SHA.fullmatch(sha):
            raise ValueError("invalid commit identity in scope context")
        commits.append(dict(sha=sha, message=records[index + 1].strip(),
                            url=server + "/" + repository + "/commit/" + sha))
        if len(encoded(commits)) > MAX_BYTES - 512:
            raise ValueError("Commit scope context exceeds 64 KiB; supply a smaller explicit PROVEN_PR_CONTEXT")
    metadata = dict(context="pull_request", url=url, title=pr.get("title") or "", description=pr.get("body") or "", commits=commits, comments=[], issues=[])
    if len(encoded(metadata)) > MAX_BYTES - 512:
        raise ValueError("PR and commit context exceed 64 KiB; supply a smaller explicit PROVEN_PR_CONTEXT")
    deadline = time.monotonic() + 60
    if env.get("PROVEN_NETWORK", "public") not in ("", "public"):
        metadata["issues"].append("GitHub comment collection is disabled by local network policy; supply an external context file containing those comments.")
        return metadata
    opener = build_opener(NoRedirect())

    def request(path, timeout):
        headers = {"Accept": "application/vnd.github+json", "User-Agent": "Proven-scope-check"}
        if token:
            headers["Authorization"] = "Bearer " + token
        req = Request(api + path, headers=headers, method="GET")
        with opener.open(req, timeout=timeout) as response:
            body = response.read(2 * 1024 * 1024 + 1)
        if len(body) > 2 * 1024 * 1024:
            raise ValueError("comment response exceeds 2 MiB")
        return json.loads(body)

    fetch = fetch or request
    root = "/repos/" + repository
    feeds = [("pr_comment", root + f"/issues/{number}/comments", ""),
             ("review", root + f"/pulls/{number}/reviews", ""),
             ("review_comment", root + f"/pulls/{number}/comments", "")]
    feeds += [("commit_comment", root + f"/commits/{c['sha']}/comments", c["sha"]) for c in commits]
    account = env.get("PROVEN_COMMENT_ACCOUNT") or "github-actions[bot]"
    if token and not env.get("PROVEN_COMMENT_ACCOUNT"):
        try:
            me = fetch("/user", min(20, max(0.1, deadline - time.monotonic())))
            if isinstance(me, dict) and isinstance(me.get("login"), str):
                account = me["login"]
        except (HTTPError, URLError, OSError, TimeoutError, ValueError):
            pass  # Installation tokens can lack /user; the job account is the fallback.
    full = False
    for kind, path, commit_sha in feeds:
        if full:
            break
        seen = set()
        try:
            for page in range(1, 11):
                left = deadline - time.monotonic()
                if left <= 0:
                    raise TimeoutError()
                rows = fetch(path + f"?per_page=100&page={page}", min(20, left))
                if not isinstance(rows, list):
                    raise ValueError("comment response is not an array")
                for item in rows:
                    if not isinstance(item, dict) or type(item.get("id")) is not int or item["id"] < 1:
                        raise ValueError("comment has no stable numeric identity")
                    if item["id"] in seen:
                        raise ValueError("comment pagination repeated an identity")
                    seen.add(item["id"])
                    body = item.get("body") or ""
                    if not isinstance(body, str):
                        raise ValueError("comment body is not text")
                    author = (item.get("user") or {}).get("login") or ""
                    # Do not review Proven's own summary on the next run. A
                    # marker copied by another account cannot hide its comment.
                    if author == account and body.lstrip().startswith("<!-- proven-review"):
                        continue
                    if not body.strip():
                        continue  # An approval without text declares no scope.
                    anchor = {"pr_comment": "issuecomment-", "review": "pullrequestreview-", "review_comment": "discussion_r", "commit_comment": "commitcomment-"}[kind]
                    link = (server + "/" + repository + "/commit/" + commit_sha if commit_sha else url) + "#" + anchor + str(item["id"])
                    sha = commit_sha or item.get("commit_id") or ""
                    if sha and not SHA.fullmatch(sha):
                        raise ValueError("comment has an invalid commit identity")
                    entry = dict(kind=kind, id=item["id"], body=body, author=author, url=link,
                                 sha=sha, updated_at=item.get("updated_at") or item.get("submitted_at") or "")
                    metadata["comments"].append(entry)
                    if len(encoded(metadata)) > MAX_BYTES - 512:
                        metadata["comments"].pop()
                        metadata["issues"].append("Not all comments fit the 64 KiB context limit; omitted comments were not assessed.")
                        full = True
                        break
                if full or len(rows) < 100:
                    break
            else:
                raise ValueError("comment pagination exceeded 10 pages")
        except (HTTPError, URLError, OSError, TimeoutError, ValueError, TypeError, AttributeError) as exc:
            reason = "HTTP " + str(exc.code) if isinstance(exc, HTTPError) else "timeout, pagination or response error"
            metadata["issues"].append(f"Could not finish collecting {kind}: {reason}. Missing comments were not assessed.")
            if deadline <= time.monotonic():
                metadata["issues"].append("The 60-second collection deadline ended before all comment feeds were read.")
                break
    # Error summaries must fit too; never silently trim collected declarations.
    if len(encoded(metadata)) > MAX_BYTES:
        raise ValueError("scope context and collection issues exceed 64 KiB")
    return metadata


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--target", required=True)
    parser.add_argument("--baseline", required=True)
    parser.add_argument("--output", required=True)
    args = parser.parse_args()
    token = os.environ.pop("PROVEN_CONTEXT_TOKEN", "")
    metadata = collect(Path(args.target), args.baseline, os.environ, token)
    fd = os.open(args.output, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
    with os.fdopen(fd, "wb") as output:
        output.write(encoded(metadata))
    if metadata["issues"]:
        print("Scope context is incomplete; the report will retain the available sources and name the collection failures.")


if __name__ == "__main__":
    try:
        main()
    except (ValueError, OSError, subprocess.SubprocessError) as exc:
        raise SystemExit("proven scope context: " + (str(exc) if isinstance(exc, ValueError) else type(exc).__name__))
