"""No accounts or Docker required: exercise credential/container boundaries."""
import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch

import cli


class CLIPreset(unittest.TestCase):
    def state(self, root, oauth=False):
        state = Path(root) / "proven-cli-test"
        state.mkdir(mode=0o700)
        (state / "settings.json").write_text(json.dumps(dict(backend="claude", image="sha256:fixed",
            home="proven-claude-home", seconds=300, model="sonnet", oauth=oauth, test_image="sha256:tests")))
        return state

    def test_profile_container_gets_only_its_named_home(self):
        for backend in cli.BACKENDS:
            with self.subTest(backend=backend), tempfile.TemporaryDirectory() as root:
                state = self.state(root)
                cfg = cli.load(state)
                cfg.update(backend=backend, home="proven-" + backend + "-home")
                (state / "settings.json").write_text(json.dumps(cfg))
                seen = []
                with patch.object(subprocess, "call", side_effect=lambda argv: seen.append(argv) or 0), patch.object(subprocess, "run"):
                    self.assertEqual(cli.model(state, "chosen; $(touch injected)"), 0)
                argv = seen[0]
                self.assertIn("type=volume,src=proven-" + backend + "-home,dst=/opt/proven", argv)
                self.assertEqual(argv[-2:], ["proven-" + backend, "chosen; $(touch injected)"])
                self.assertIn("--read-only", argv)
                self.assertIn("295s", argv)
                self.assertNotIn("/var/run/docker.sock", " ".join(argv))
                self.assertNotIn("GITHUB_WORKSPACE", " ".join(argv))

    def test_oauth_file_is_private_never_in_docker_arguments_and_removed_on_failure(self):
        for probe_code in (0, 3):
            with self.subTest(probe_code=probe_code), tempfile.TemporaryDirectory() as root:
                state = self.state(root, oauth=True)
                secret = "test-only-oauth-value"
                calls = []

                def invoke(argv, **kwargs):
                    env = kwargs["env"]
                    self.assertNotIn("INPUT_CLAUDE_OAUTH_TOKEN", env)
                    self.assertEqual((state / "oauth-token").stat().st_mode & 0o777, 0o600)
                    self.assertEqual((state / "oauth-token").read_text(), secret)
                    self.assertEqual(env["CLAUDE_CODE_OAUTH_TOKEN"], secret)
                    if "check" in argv:
                        self.assertNotIn("INPUT_GITHUB_TOKEN", env)
                        self.assertEqual(env["PROVEN_GENERATOR_TIMEOUT"], "60s")
                    calls.append(argv)
                    return probe_code if "check" in argv else 3

                with patch.dict(os.environ, {"INPUT_CLAUDE_OAUTH_TOKEN": secret, "INPUT_GITHUB_TOKEN": "posting-token"}), patch.object(subprocess, "call", side_effect=invoke), patch.object(cli, "remove_containers") as cleanup:
                    self.assertEqual(cli.review(state), 3)
                    cleanup.assert_called_once_with(state)
                self.assertFalse((state / "oauth-token").exists())
                self.assertEqual(len(calls), 2 if probe_code == 0 else 1)
                seen = []
                with patch.object(subprocess, "call", side_effect=lambda argv: seen.append(argv) or 0), patch.object(subprocess, "run"):
                    cli.model(state, "sonnet")
                self.assertNotIn(secret, " ".join(seen[0]))
                self.assertNotIn("type=volume", " ".join(seen[0]))
                self.assertIn("readonly", " ".join(seen[0]))

    def test_local_network_policy_stops_before_preparation(self):
        with patch.dict(os.environ, {"PROVEN_NETWORK": "local"}), patch.object(subprocess, "run") as calls:
            with self.assertRaisesRegex(ValueError, "public"):
                cli.prepare()
            calls.assert_not_called()

    def test_missing_hosted_claude_login_explains_oauth_setup(self):
        with patch.dict(os.environ, {"INPUT_BACKEND": "claude", "INPUT_RUNNER_ENVIRONMENT": "github-hosted"}, clear=True), patch.object(subprocess, "run", return_value=subprocess.CompletedProcess([], 1)):
            with self.assertRaisesRegex(ValueError, "claude setup-token.*CLAUDE_CODE_OAUTH_TOKEN"):
                cli.prepare()

    def test_secondary_cli_never_mounts_the_primary_oauth_credential(self):
        with tempfile.TemporaryDirectory() as root:
            state = self.state(root, oauth=True)
            cfg = cli.load(state)
            cfg["secondary"] = dict(backend="copilot", image="sha256:copilot", home="proven-copilot-home", seconds=300, model="auto", oauth=False)
            (state / "settings.json").write_text(json.dumps(cfg))
            seen = []
            with patch.object(subprocess, "call", side_effect=lambda argv: seen.append(argv) or 0), patch.object(subprocess, "run"):
                cli.model(state, "second-model", "secondary")
            argv = " ".join(seen[0])
            self.assertIn("proven-copilot-home", argv)
            self.assertNotIn("oauth-token", argv)
            self.assertNotIn("proven-claude-home", argv)
            self.assertEqual(seen[0][-2:], ["proven-copilot", "second-model"])

    def test_cleanup_cannot_remove_an_arbitrary_directory(self):
        with tempfile.TemporaryDirectory() as root, patch.dict(os.environ, {"RUNNER_TEMP": root}):
            with self.assertRaises(ValueError):
                cli.state_path(str(Path(root) / "unrelated"))
            with self.assertRaises(ValueError):
                cli.state_path(str(Path(root) / "proven-cli-a" / "nested"))


if __name__ == "__main__":
    unittest.main()
