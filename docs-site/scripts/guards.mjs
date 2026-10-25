// The site's guards. Each one exits the build non-zero when its rule is broken:
//
//   page-intact  a page carries a merge-conflict marker
//   links        a link inside the site, or to an image, or to an anchor, goes nowhere
//   public-text  a page matches a known shape of an internal path, a private
//                host name or address, or a credential format (publicTextRules);
//                a value outside those shapes is not caught — a guard, not a proof
//   stale        a generated page differs from what the generator writes now
//
// They read docs-site/content/docs, which build-content.mjs wrote, so they see
// what a reader would see. A guard prints one line per rule: ok, or FAIL and
// every finding. All of them run, so one run names everything that is wrong.
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import {
  repoRoot, siteDir, contentDir, publicDir, generatedDir, routeBase, routeOf,
  generatedPages, walk, lines, linksOf, headingAnchors,
} from './lib.mjs';

// ---- the rules, as data -----------------------------------------------------

// Public-text rule: a page renders only what is already public in the repository.
export const publicTextRules = [
  { kind: 'internal path', re: /\/Users\/[^\s/]+|\/home\/[^\s/]+|\/Volumes\/|\/var\/folders\/|[A-Za-z]:\\Users\\/ },
  { kind: 'internal path', re: /\bsaphan-studio\/|\bworktrees\/|\.provision\b|\.saphan\/|\.saphan-guardrail|\b_EXEC_(ORDER|STATUS)\b|\bbridge\/(governance|registry|development|correspondence)\b/ },
  // a host name inside a private zone; an e-mail address in one is a placeholder account, not a machine
  { kind: 'machine name', re: /(?<![@\w.-])[a-z0-9][a-z0-9-]*(\.[a-z0-9-]+)*\.(local|internal|lan|corp|intranet|home\.arpa)\b/i },
  { kind: 'machine name', re: /\b[a-z0-9-]+\.[a-z0-9-]+\.ts\.net\b/i },
  { kind: 'machine name', re: /\b(10\.\d{1,3}\.\d{1,3}\.\d{1,3}|192\.168\.\d{1,3}\.\d{1,3}|172\.(1[6-9]|2\d|3[01])\.\d{1,3}\.\d{1,3})\b/ },
  { kind: 'credential', re: /\bAKIA[0-9A-Z]{16}\b/ },
  { kind: 'credential', re: /\bgh[pousr]_[A-Za-z0-9]{30,}\b/ },
  { kind: 'credential', re: /\bnvapi-[A-Za-z0-9_-]{20,}/ },
  { kind: 'credential', re: /\bsk-[A-Za-z0-9_-]{20,}/ },
  { kind: 'credential', re: /\bxox[baprs]-[A-Za-z0-9-]{10,}/ },
  { kind: 'credential', re: /-----BEGIN [A-Z ]*PRIVATE KEY-----/ },
  { kind: 'credential', re: /\beyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{5,}/ },
  // also after an underscore-joined name: GITHUB_TOKEN=…, PROVEN_GUARDRAIL_KEY=…
  { kind: 'credential', re: /(?:\b|_)(api[_-]?key|key|token|secret|password)\b\s*[:=]\s*["']?[A-Za-z0-9/+_-]{24,}/i },
];

// ---- the checks: pages are [{ slug, text }]; each returns a list of findings --

export function checkConflicts(pages) {
  const out = [];
  for (const p of pages) {
    for (const l of lines(p.text)) {
      // <<<<<<< and >>>>>>> are markers at any indent; ======= only at column one,
      // because an indented run of = can be a heading underline.
      if (/^\s*(<{7}|>{7})(\s|$)|^={7}(\s|$)/.test(l.text)) out.push(`${p.slug}:${l.n}: merge-conflict marker "${l.text.slice(0, 20)}"`);
    }
  }
  return out;
}

export function checkLinks(pages, publicFiles = new Set()) {
  const byRoute = new Map(pages.map((p) => [routeOf(p.slug), p]));
  const anchors = new Map(pages.map((p) => [p.slug, headingAnchors(p.text)]));
  const out = [];
  let checked = 0;
  for (const p of pages) {
    for (const { n, target } of linksOf(p.text)) {
      if (/^[a-z][a-z0-9+.-]*:/i.test(target)) continue;
      checked += 1;
      const [rawPath, anchor] = target.split('#');
      const here = `${p.slug}:${n}: ${target}`;
      if (rawPath === '') {
        if (anchor && !anchors.get(p.slug).has(anchor)) out.push(`${here}: no heading has that anchor in this page`);
      } else if (rawPath === routeBase || rawPath.startsWith(`${routeBase}/`)) {
        const dest = byRoute.get(rawPath.replace(/\/$/, ''));
        if (!dest) out.push(`${here}: no page has that address`);
        else if (anchor && !anchors.get(dest.slug).has(anchor)) out.push(`${here}: no heading has that anchor in ${dest.slug}`);
      } else if (rawPath.startsWith('/repo/')) {
        if (!publicFiles.has(rawPath)) out.push(`${here}: the image is not in the site`);
      } else {
        out.push(`${here}: the target is not a page, an image or a file of the repository`);
      }
    }
  }
  return { findings: out, checked };
}

export function checkPublicText(pages) {
  const out = [];
  for (const p of pages) {
    for (const l of lines(p.text)) {
      for (const { kind, re } of publicTextRules) {
        const m = re.exec(l.text);
        if (m) out.push(`${p.slug}:${l.n}: ${kind} "${m[0].slice(0, 40)}"`);
      }
    }
  }
  return out;
}

// committedDir and freshDir hold the generator's output; configText is
// docs/CONFIGURATION.md, whose generated section must be what settings-doc writes now.
export function checkStale(committedDir, freshDir, configText) {
  const out = [];
  for (const [file] of generatedPages) {
    const a = path.join(committedDir, file);
    const b = path.join(freshDir, file);
    if (!fs.existsSync(b)) { out.push(`${file}: the generator did not write it`); continue; }
    if (!fs.existsSync(a)) { out.push(`${file}: no committed copy; run: sh docs-site/scripts/generate.sh docs-site/generated`); continue; }
    const x = fs.readFileSync(a, 'utf8').split('\n');
    const y = fs.readFileSync(b, 'utf8').split('\n');
    for (let i = 0; i < Math.max(x.length, y.length); i++) {
      if (x[i] !== y[i]) {
        out.push(`${file}:${i + 1}: committed "${(x[i] ?? '').slice(0, 60)}" but the generator writes "${(y[i] ?? '').slice(0, 60)}"; regenerate: sh docs-site/scripts/generate.sh docs-site/generated`);
        break;
      }
    }
  }
  // the page docs/CONFIGURATION.md carries the same table
  const region = /<!-- generated: settings -->\n([\s\S]*?)<!-- end generated -->/.exec(configText);
  const settingsFresh = path.join(freshDir, 'settings.md');
  if (!region) out.push('docs/CONFIGURATION.md: the generated settings section is missing');
  else if (fs.existsSync(settingsFresh)) {
    const fresh = fs.readFileSync(settingsFresh, 'utf8').split('<!-- generated by docs-site/scripts/generate.sh: do not edit -->\n\n')[1] ?? '';
    if (region[1].trim() !== fresh.trim()) {
      out.push('docs/CONFIGURATION.md: its generated settings section differs from what cmd/settings-doc writes now; run: go run ./cmd/settings-doc');
    }
  }
  return out;
}

// ---- running them -----------------------------------------------------------

function freshGenerated() {
  if (process.env.GENERATED_FRESH_DIR) return path.resolve(process.env.GENERATED_FRESH_DIR);
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'proven-docs-fresh-'));
  const r = spawnSync('sh', [path.join(siteDir, 'scripts', 'generate.sh'), dir], { encoding: 'utf8', env: { ...process.env, DOCS_REPO_ROOT: repoRoot } });
  if (r.status !== 0) throw new Error(`the generator failed (exit ${r.status}): ${r.stderr}`);
  return dir;
}

function main() {
  const pages = walk(contentDir, (f) => f.endsWith('.md')).map((f) => ({
    slug: path.relative(contentDir, f).split(path.sep).join('/').replace(/\.md$/, ''),
    text: fs.readFileSync(f, 'utf8'),
  }));
  if (pages.length === 0) { console.error('guards: no pages in content/docs; run build-content first'); process.exit(2); }
  const publicFiles = new Set(walk(publicDir, () => true).map((f) => `/${path.relative(publicDir, f).split(path.sep).join('/')}`));
  const links = checkLinks(pages, publicFiles);
  let stale;
  try {
    stale = checkStale(generatedDir, freshGenerated(), fs.readFileSync(path.join(repoRoot, 'docs', 'CONFIGURATION.md'), 'utf8'));
  } catch (e) {
    stale = [String(e.message)];
  }
  const results = [
    ['page-intact', checkConflicts(pages), `${pages.length} pages`],
    ['links', links.findings, `${pages.length} pages, ${links.checked} links`],
    ['public-text', checkPublicText(pages), `${pages.length} pages, ${publicTextRules.length} rules`],
    ['stale', stale, `${generatedPages.length} generated pages and the settings section`],
  ];
  let failed = 0;
  for (const [name, findings, what] of results) {
    if (findings.length === 0) { console.log(`guard ${name}: ok (${what})`); continue; }
    failed += 1;
    console.log(`guard ${name}: FAIL (${findings.length})`);
    for (const f of findings) console.log(`  ${f}`);
  }
  process.exit(failed === 0 ? 0 : 1);
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) main();
