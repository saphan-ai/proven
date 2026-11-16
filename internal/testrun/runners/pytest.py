# Executed by Proven's container adapter, never on the review host.
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import xml.etree.ElementTree as ET

mode, file, name = sys.argv[1:]
report = dict(proven_test_report=1, adapter="pytest", file=file, name=name,
              passed=0, failed=0, skipped=0, errors=0, completed=False, output="")
with tempfile.TemporaryDirectory(prefix="proven-pytest-") as temp:
    xml = Path(temp) / "results.xml"
    selected = file + "::" + name if mode == "test" else file
    env = dict(os.environ, PYTHONDONTWRITEBYTECODE="1", PYTHONPATH=os.getcwd())
    argv = [sys.executable, "-m", "pytest", "-q", "-p", "no:cacheprovider",
            "--basetemp=" + temp + "/scratch", "--junitxml=" + str(xml), selected]
    result = subprocess.run(argv, env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    report["output"] = result.stdout.decode("utf-8", "replace")[:512]
    try:
        root = ET.parse(xml).getroot()
        cases = list(root.iter("testcase"))
        for case in cases:
            if mode == "test":
                expected = name.split("::")[-1]
                actual = case.get("name", "")
                if actual != expected and not (actual.startswith(expected + "[") and actual.endswith("]")):
                    report["errors"] += 1
                    continue
            if case.find("error") is not None:
                report["errors"] += 1
            elif case.find("skipped") is not None:
                report["skipped"] += 1
            elif case.find("failure") is not None:
                report["failed"] += 1
            else:
                report["passed"] += 1
        # Collection/configuration/internal errors never become a failing test.
        report["completed"] = result.returncode in (0, 1) and bool(cases)
        if result.returncode not in (0, 1):
            report["errors"] += 1
    except (OSError, ET.ParseError):
        report["errors"] += 1
print(json.dumps(report, ensure_ascii=True))
sys.exit(0 if report["completed"] and not report["errors"] and not report["failed"] else 1)
