"""Real Git metadata and fake HTTP: no accounts, writes to GitHub or models."""
import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch
from urllib.error import HTTPError

import delivery


class ScopeCollector(unittest.TestCase):
    def setup_context(self, root):
        repo = Path(root) / "repo"
        repo.mkdir()
        def git(*args):
            return subprocess.run(["git", "-C", str(repo), *args], check=True, text=True, stdout=subprocess.PIPE, stderr=subprocess.DEVNULL).stdout.strip()
        git("init", "-q")
        git("-c", "user.name=Example", "-c", "user.email=example@example.com", "commit", "--allow-empty", "-m", "Baseline example commit")
        base = git("rev-parse", "HEAD")
        git("-c", "user.name=Example", "-c", "user.email=example@example.com", "commit", "--allow-empty", "-m", "Deliver requested example behavior")
        sha = git("rev-parse", "HEAD")
        event = Path(root) / "event.json"
        event.write_text(json.dumps(dict(number=7, pull_request=dict(number=7, head=dict(sha=sha), title="Deliver requested example behavior", body="Reject all negative amounts"))))
        env = dict(GITHUB_REPOSITORY="example/project", GITHUB_EVENT_PATH=str(event))
        return repo, base, sha, env

    def row(self, id, body="Reject all negative amounts", author="example"):
        return dict(id=id, body=body, user=dict(login=author))

    def test_all_comment_kinds_are_separate_and_pagination_is_followed(self):
        with tempfile.TemporaryDirectory() as root:
            repo, base, sha, env = self.setup_context(root)
            calls = []
            def fetch(path, timeout):
                calls.append(path)
                self.assertLessEqual(timeout,20)
                if "/issues/" in path:
                    if "page=2" in path:
                        return [self.row(101), self.row(102,"<!-- proven-review run=1 -->", "github-actions[bot]"), self.row(103,"<!-- proven-review copied by another author -->")]
                    return [self.row(i) for i in range(1,101)]
                if "/reviews?" in path: return [self.row(201)]
                if "/pulls/" in path: return [self.row(301)]
                return [self.row(401)]
            m = delivery.collect(repo, base, env, "test-token", fetch)
            self.assertEqual(m["issues"],[])
            self.assertEqual(len(m["comments"]),105)
            self.assertEqual({c["kind"] for c in m["comments"]},{"pr_comment","review","review_comment","commit_comment"})
            self.assertEqual(m["commits"][0]["sha"],sha)
            self.assertNotIn("test-token",delivery.encoded(m).decode())
            self.assertTrue(any("page=2" in c for c in calls))

    def test_permission_failure_is_recorded_and_other_sources_survive(self):
        with tempfile.TemporaryDirectory() as root:
            repo, base, _, env = self.setup_context(root)
            def fetch(path, timeout):
                if "/issues/" in path: raise HTTPError("https://api.github.com",403,"denied",{},None)
                return [self.row(1)] if "/reviews?" in path else []
            m=delivery.collect(repo,base,env,fetch=fetch)
            self.assertTrue(any("HTTP 403" in issue for issue in m["issues"]))
            self.assertEqual(len(m["comments"]),1)

    def test_personal_token_summary_is_excluded_only_for_its_actual_account(self):
        with tempfile.TemporaryDirectory() as root:
            repo,base,_,env=self.setup_context(root)
            def fetch(path,timeout):
                if path=="/user": return {"login":"example-author"}
                if "/issues/" in path: return [self.row(1,"<!-- proven-review run=1 -->","example-author"),self.row(2,"<!-- proven-review copied -->","someone-else")]
                return []
            m=delivery.collect(repo,base,env,"test-only-token",fetch)
            self.assertEqual([c["id"] for c in m["comments"]],[2])

    def test_local_policy_never_calls_http_and_wrong_head_is_refused(self):
        with tempfile.TemporaryDirectory() as root:
            repo, base, _, env=self.setup_context(root)
            env["PROVEN_NETWORK"]="local"
            def fetch(*args): self.fail("network used in local mode")
            m=delivery.collect(repo,base,env,fetch=fetch)
            self.assertTrue(m["issues"])
            event=Path(env["GITHUB_EVENT_PATH"])
            data=json.loads(event.read_text()); data["pull_request"]["head"]["sha"]=base; event.write_text(json.dumps(data))
            with self.assertRaisesRegex(ValueError,"exact PR head"):
                delivery.collect(repo,base,env,fetch=fetch)

    def test_repeating_page_and_overflow_never_claim_complete_coverage(self):
        with tempfile.TemporaryDirectory() as root:
            repo,base,_,env=self.setup_context(root)
            for body in ("Reject all negative amounts", "x"*65536):
                rows=[self.row(1,body)]*100
                m=delivery.collect(repo,base,env,fetch=lambda *_:rows)
                self.assertTrue(m["issues"])
                self.assertLessEqual(len(delivery.encoded(m)),delivery.MAX_BYTES)

    def test_oversized_git_message_is_refused_before_http_collection(self):
        with tempfile.TemporaryDirectory() as root:
            repo,base,_,env=self.setup_context(root)
            message=Path(root)/"message.txt"
            message.write_text("A"*(delivery.MAX_BYTES+1))
            subprocess.run(["git","-C",str(repo),"-c","user.name=Example","-c","user.email=example@example.com","commit","--allow-empty","-F",str(message)],check=True,stdout=subprocess.DEVNULL)
            event=Path(env["GITHUB_EVENT_PATH"])
            data=json.loads(event.read_text())
            data["pull_request"]["head"]["sha"]=delivery.git(repo,"rev-parse","HEAD").strip()
            event.write_text(json.dumps(data))
            with self.assertRaisesRegex(ValueError,"64 KiB"):
                delivery.collect(repo,base,env,fetch=lambda *_:self.fail("HTTP started for oversized context"))


if __name__ == "__main__":
    unittest.main()
