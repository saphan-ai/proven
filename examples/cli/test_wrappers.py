"""Exercise stdin, argv, profile isolation and cleanup without vendor accounts."""
import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parent


class Wrappers(unittest.TestCase):
    def test_text_protocol_and_no_tools(self):
        for vendor in ("claude", "copilot"):
            with self.subTest(vendor=vendor), tempfile.TemporaryDirectory() as tmp:
                home = Path(tmp)
                fake = home / vendor
                record = home / "record.json"
                fake.write_text("#!/usr/bin/env python3\nimport json,os,sys\n"
                                "json.dump({'argv':sys.argv[1:],'prompt':sys.stdin.read(),"
                                "'cwd':os.getcwd(),'claude':os.getenv('CLAUDE_CONFIG_DIR'),"
                                "'copilot':os.getenv('COPILOT_HOME')},open(os.environ['RECORD'],'w'))\n"
                                "print('{\"risks\":[]}')\n")
                fake.chmod(0o755)
                env = dict(os.environ, HOME=tmp, RECORD=str(record), PATH=tmp + os.pathsep + os.environ["PATH"])
                prompt = "untrusted prompt; $(touch must-not-exist)"
                proc = subprocess.run(["sh", str(ROOT / (vendor + "-text.sh")), "chosen-model"],
                                      input=prompt, text=True, capture_output=True, env=env, check=True)
                self.assertEqual(json.loads(proc.stdout), {"risks": []})
                got = json.loads(record.read_text())
                self.assertEqual(got["prompt"], prompt)
                args = got["argv"]
                self.assertEqual(args[args.index("--model") + 1], "chosen-model")
                self.assertEqual(args[args.index("--output-format") + 1], "text")
                self.assertFalse(Path(got["cwd"]).exists(), "scratch was not removed")
                self.assertEqual(got[vendor], str(home / ".config" / "proven" / vendor))
                if vendor == "claude":
                    self.assertEqual(args[args.index("--tools") + 1], "")
                    self.assertIn("--safe-mode", args)
                    self.assertIn("--strict-mcp-config", args)
                else:
                    self.assertEqual(args[args.index("--available-tools") + 1], "--model")
                    self.assertIn("--disable-builtin-mcps", args)
                    self.assertIn("--no-custom-instructions", args)
                self.assertFalse((home / "must-not-exist").exists())


if __name__ == "__main__":
    unittest.main()
