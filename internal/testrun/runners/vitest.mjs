import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { spawnSync } from 'node:child_process';
const [mode, file, name] = process.argv.slice(1);
const result = { proven_test_report: 1, adapter: 'vitest', file, name,
  passed: 0, failed: 0, skipped: 0, errors: 0, completed: false, output: '' };
const temp = fs.mkdtempSync(path.join(os.tmpdir(), 'proven-vitest-'));
const root = path.join(temp, 'project');
const escape = s => s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
try {
  // Vite needs writable scratch. The source mount and prepared dependencies
  // stay read-only; no package install or project script runs here.
  fs.cpSync(process.cwd(), root, { recursive: true });
  if (!fs.existsSync(path.join(root, 'node_modules')) && fs.existsSync('/deps/node_modules')) {
    fs.symlinkSync('/deps/node_modules', path.join(root, 'node_modules'));
  }
  const binary = path.join(root, 'node_modules/vitest/vitest.mjs');
  const report = path.join(temp, 'report.json');
  const args = [binary, 'run', '--reporter=json', '--outputFile=' + report,
    '--maxWorkers=1', '--no-file-parallelism', '--allowOnly=false'];
  if (file !== '.') args.push(file);
  if (mode === 'test') args.push('--testNamePattern=^' + escape(name) + '$');
  const run = spawnSync(process.execPath, args, { cwd: root, encoding: 'utf8', maxBuffer: 1024 * 1024 });
  result.output = ((run.stdout ?? '') + (run.stderr ?? '')).slice(0, 512);
  const json = JSON.parse(fs.readFileSync(report, 'utf8'));
  for (const suite of json.testResults ?? []) {
    if (suite.status === 'failed' && !suite.assertionResults?.some(test => test.status === 'failed')) result.errors++;
    for (const test of suite.assertionResults ?? []) {
      if (mode === 'test' && test.fullName !== name) continue;
      if (test.status === 'passed') result.passed++;
      else if (test.status === 'failed') result.failed++;
      else result.skipped++;
    }
  }
  result.errors += json.numRuntimeErrorTestSuites ?? 0;
  if (Array.isArray(json.unhandledErrors)) result.errors += json.unhandledErrors.length;
  if (run.status !== 0 && result.failed === 0) result.errors++;
  result.completed = !run.error && (run.status === 0 || run.status === 1);
} catch { result.errors++; }
finally { fs.rmSync(temp, { recursive: true, force: true }); }
console.log(JSON.stringify(result));
process.exitCode = result.failed || result.errors ? 1 : 0;
