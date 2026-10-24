// Shared by build-content.mjs and guards.mjs: where the repository is, which
// Markdown becomes which page, and a few text helpers.
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));

export const siteDir = path.resolve(here, '..');
// DOCS_REPO_ROOT points the scripts at another checkout; the negative controls
// use it to break a copy, so docs/ is never edited.
export const repoRoot = path.resolve(process.env.DOCS_REPO_ROOT || path.join(siteDir, '..'));
export const contentDir = path.join(siteDir, 'content', 'docs');
export const publicDir = path.join(siteDir, 'public');
export const generatedDir = path.join(siteDir, 'generated');
export const routeBase = '/docs';
export const sourceUrl = 'https://github.com/saphan-ai/proven/blob/main/';

export function walk(dir, pred) {
  const out = [];
  if (!fs.existsSync(dir)) return out;
  for (const entry of fs.readdirSync(dir, { withFileTypes: true }).sort((a, b) => a.name.localeCompare(b.name))) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) out.push(...walk(full, pred));
    else if (pred(full)) out.push(full);
  }
  return out;
}

const isMarkdown = (f) => f.endsWith('.md');

// The pages that come from the repository's own Markdown, as
// { source: repo-relative path, slug }. examples/ is rendered when present.
export function repoPages() {
  const rel = (f) => path.relative(repoRoot, f).split(path.sep).join('/');
  const pages = [];
  if (fs.existsSync(path.join(repoRoot, 'README.md'))) pages.push({ source: 'README.md', slug: 'index' });
  for (const f of walk(path.join(repoRoot, 'docs'), isMarkdown)) {
    const source = rel(f);
    const inner = source.slice('docs/'.length, -'.md'.length);
    const parts = inner.split('/');
    const last = parts.pop();
    const name = last === 'README' ? 'index' : last.toLowerCase();
    const folder = parts.length === 0 ? 'guides' : parts.join('/');
    pages.push({ source, slug: `${folder}/${name}` });
  }
  for (const f of walk(path.join(repoRoot, 'examples'), isMarkdown)) {
    const source = rel(f);
    const parts = source.slice(0, -'.md'.length).split('/');
    const last = parts.pop();
    pages.push({ source, slug: [...parts, last === 'README' ? 'index' : last.toLowerCase()].join('/') });
  }
  return pages;
}

// slug -> the route a reader opens; the home page is /docs itself.
export function routeOf(slug) {
  if (slug === 'index') return routeBase;
  return `${routeBase}/${slug.replace(/\/index$/, '')}`;
}

// The generated pages, by file name in the generated directory.
export const generatedPages = [
  ['cli.md', 'reference/cli'],
  ['settings.md', 'reference/settings'],
  ['adapter-tool.md', 'reference/adapters/tool'],
  ['adapter-test.md', 'reference/adapters/test'],
  ['adapter-query.md', 'reference/adapters/query'],
  ['adapter-guardrail.md', 'reference/adapters/guardrail'],
  ['adapter-generator.md', 'reference/adapters/generator'],
  ['adapter-reviewer.md', 'reference/adapters/reviewer'],
];

// Lines outside code fences, with their line numbers. Fence-aware callers keep
// the fence state themselves; this returns each line with inFence.
export function lines(text) {
  let fence = null;
  return text.split('\n').map((line, i) => {
    const m = /^\s*(`{3,}|~{3,})/.exec(line);
    let inFence = fence !== null;
    if (m) {
      if (fence === null) fence = m[1][0];
      else if (m[1][0] === fence) fence = null;
      inFence = true;
    }
    return { n: i + 1, text: line, inFence };
  });
}

// The anchor GitHub and Fumadocs both give a heading.
export function anchorOf(heading) {
  return heading
    .replace(/`/g, '')
    .replace(/\[([^\]]*)\]\([^)]*\)/g, '$1')
    .toLowerCase()
    .trim()
    .replace(/[^\p{L}\p{N}\s_-]/gu, '')
    .replace(/\s/g, '-');
}

export function headingAnchors(text) {
  const seen = new Map();
  const out = new Set();
  for (const l of lines(text)) {
    if (l.inFence) continue;
    const m = /^#{1,6}\s+(.*?)\s*#*\s*$/.exec(l.text);
    if (!m) continue;
    let a = anchorOf(m[1]);
    const k = seen.get(a) ?? 0;
    seen.set(a, k + 1);
    if (k > 0) a = `${a}-${k}`;
    out.add(a);
  }
  return out;
}

// Markdown links and images outside code, as { n, target }.
export function linksOf(text) {
  const out = [];
  for (const l of lines(text)) {
    if (l.inFence) continue;
    const noCode = l.text.replace(/`[^`]*`/g, (s) => ' '.repeat(s.length));
    for (const m of noCode.matchAll(/!?\[[^\]]*\]\(\s*<?([^)\s>]+)>?(?:\s+"[^"]*")?\s*\)/g)) {
      out.push({ n: l.n, target: m[1] });
    }
  }
  return out;
}
