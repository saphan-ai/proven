#!/bin/sh
set -eu
python3 - <<'CHECK'
from pathlib import Path
import re, subprocess
from urllib.parse import unquote, urlsplit
root=Path.cwd()
files=subprocess.check_output(['git','ls-files','-co','--exclude-standard','-z','--','*.md']).decode().split('\0')
errors=[]
for name in sorted(set(files)):
    p=root/name
    if not name or not p.is_file(): continue
    base=root/'docs' if name.startswith('docs-site/generated/') else p.parent
    fence=False
    for number,line in enumerate(p.read_text().splitlines(),1):
        if re.match(r'^\s*(```|~~~)',line): fence=not fence; continue
        if fence: continue
        line=re.sub(r'`[^`]*`','',line)
        for target in re.findall(r'!?\[[^\]]*\]\(<?([^\s)>]+)>?(?:\s+"[^"\n]*")?\)',line):
            parsed=urlsplit(target)
            if parsed.scheme or parsed.netloc or not parsed.path: continue
            path=unquote(parsed.path)
            dest=root/path.lstrip('/') if path.startswith('/') else base/path
            if not dest.exists(): errors.append(f'{name}:{number}: {target}')
if errors:
    print('Broken links:\n'+'\n'.join(errors));raise SystemExit(1)
print('check-links: every relative documentation link resolves')
CHECK
