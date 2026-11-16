// The parent consumes native events, not test stdout or the child's exit alone.
import { run } from 'node:test';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { spawnSync } from 'node:child_process';
// Node's test worker inherits --eval from an eval parent. Bootstrap a trusted
// file first, so workers execute the selected test file rather than this helper.
if (typeof process._eval === 'string') {
  const temp = fs.mkdtempSync(path.join(os.tmpdir(), 'proven-node-adapter-'));
  try {
    const entry = path.join(temp, 'runner.mjs');
    fs.writeFileSync(entry, process._eval, { mode: 0o600 });
    const child = spawnSync(process.execPath, [entry, ...process.argv.slice(1)], { stdio: ['ignore', 'inherit', 'inherit'] });
    process.exitCode = child.status ?? 1;
  } finally { fs.rmSync(temp, { recursive: true, force: true }); }
} else {
const [mode, file, name] = process.argv.slice(2);
const project = fs.mkdtempSync(path.join(os.tmpdir(), 'proven-node-project-'));
fs.cpSync(process.cwd(), project, { recursive: true });
if (!fs.existsSync(path.join(project, 'node_modules')) && fs.existsSync('/deps/node_modules')) {
  fs.symlinkSync('/deps/node_modules', path.join(project, 'node_modules'));
}
process.chdir(project);
const result = { proven_test_report: 1, adapter: 'node', file, name,
  passed: 0, failed: 0, skipped: 0, errors: 0, completed: false, output: '' };
const escape = s => s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
const options = { concurrency: false, execArgv: [] };
if (file !== '.') options.files = [file];
if (mode === 'test') options.testNamePatterns = [new RegExp('^' + escape(name) + '$')];
try {
  const stream = run(options);
  const parents = [];
  for await (const { type, data } of stream) {
    if (type === 'test:start') {
      parents.length = data.nesting ?? 0;
      parents.push(data.name);
    }
    if (type === 'test:pass' || type === 'test:fail') {
      const fullName = data.fullName ?? [...parents.slice(0, data.nesting ?? 0), data.name].join(' ');
      if (data.details?.type === 'suite') {
        if (type === 'test:fail' && data.details?.error?.failureType !== 'subtestsFailed') result.errors++;
        continue;
      }
      if (mode === 'test' && fullName !== name && !fullName.startsWith(name + ' ')) continue;
      if (data.skip || data.todo) result.skipped++;
      else if (type === 'test:pass') result.passed++;
      else if (data.details?.error?.failureType === 'testCodeFailure') result.failed++;
      else result.errors++;
    }
    if (type === 'test:stdout' || type === 'test:stderr' || type === 'test:diagnostic') {
      result.output = (result.output + (data.message ?? '')).slice(0, 512);
    }
  }
  result.completed = true;
} catch { result.errors++; }
console.log(JSON.stringify(result));
fs.rmSync(project, { recursive: true, force: true });
process.exitCode = result.failed || result.errors ? 1 : 0;
}
