// Writes docs-site/content/docs from the repository's own Markdown and the
// generated pages. Frontmatter is added here, at build time; docs/ is only read.
//
//   node scripts/build-content.mjs
//
// GENERATED_FRESH_DIR names a directory the generator wrote just now (the image
// build's Go stage); without it the committed docs-site/generated is used, and
// the stale-page guard compares the two.
import fs from 'node:fs';
import path from 'node:path';
import {
  repoRoot, contentDir, publicDir, generatedDir, sourceUrl, routeOf,
  repoPages, generatedPages, lines,
} from './lib.mjs';

const freshDir = process.env.GENERATED_FRESH_DIR ? path.resolve(process.env.GENERATED_FRESH_DIR) : generatedDir;
const images = /\.(png|gif|jpe?g|svg|webp)$/i;

// Apply fn to the text outside code fences, keeping the fences as they are.
function outsideFences(text, fn) {
  const out = [];
  let chunk = [];
  const flush = () => { if (chunk.length) out.push(fn(chunk.join('\n'))); chunk = []; };
  for (const l of lines(text)) {
    if (l.inFence) { flush(); out.push(l.text); } else chunk.push(l.text);
  }
  flush();
  return out.join('\n');
}

// The few raw-HTML tags the repository's pages use, written as Markdown, and
// comments dropped. The pages in docs/ are not edited.
function plainHtml(text) {
  return text
    .replace(/<!--[\s\S]*?-->/g, '')
    .replace(/<img\s[^>]*?src="([^"]+)"[^>]*?alt="([^"]*)"[^>]*>/g, '![$2]($1)')
    .replace(/<a\s+href="([^"]+)"\s*>([\s\S]*?)<\/a>/g, '[$2]($1)')
    .replace(/<b>([\s\S]*?)<\/b>/g, '**$1**')
    .replace(/<\/?p(\s[^>]*)?>/g, '');
}

function rewriteLink(target, source, sourceToSlug, copied) {
  if (/^[a-z][a-z0-9+.-]*:/i.test(target) || target.startsWith('#') || target.startsWith('/')) return target;
  const [rawPath, anchor] = target.split('#');
  const resolved = path.posix.normalize(path.posix.join(path.posix.dirname(source), decodeURI(rawPath)));
  const hash = anchor ? `#${anchor}` : '';
  if (sourceToSlug.has(resolved)) return routeOf(sourceToSlug.get(resolved)) + hash;
  const onDisk = path.join(repoRoot, resolved);
  if (resolved.startsWith('..') || !fs.existsSync(onDisk)) return target; // left, and the link guard names it
  if (fs.statSync(onDisk).isDirectory()) return `${sourceUrl.replace('/blob/', '/tree/')}${resolved}${hash}`;
  if (images.test(resolved)) {
    const dest = path.join(publicDir, 'repo', resolved);
    fs.mkdirSync(path.dirname(dest), { recursive: true });
    fs.copyFileSync(onDisk, dest);
    copied.add(resolved);
    return `/repo/${resolved}`;
  }
  return `${sourceUrl}${resolved}${hash}`;
}

function convert(raw, source, sourceToSlug, copied) {
  let text = outsideFences(raw, plainHtml);
  // the first level-one heading is the title; the page shows it from the frontmatter
  let title = null;
  const body = [];
  for (const l of lines(text)) {
    const m = !l.inFence && title === null && /^#\s+(.*?)\s*$/.exec(l.text);
    if (m) { title = m[1].replace(/`/g, ''); continue; }
    body.push(l.text);
  }
  text = body.join('\n');
  text = outsideFences(text, (chunk) =>
    chunk.replace(/(!?\[[^\]]*\]\()\s*<?([^)\s>]+)>?((?:\s+"[^"]*")?\s*\))/g,
      (all, open, target, close) => open + rewriteLink(target, source, sourceToSlug, copied) + close));
  return { title: title ?? path.basename(source, '.md'), body: text.replace(/^\s*\n/, '') };
}

function write(slug, title, body) {
  const file = path.join(contentDir, `${slug}.md`);
  fs.mkdirSync(path.dirname(file), { recursive: true });
  fs.writeFileSync(file, `---\ntitle: ${JSON.stringify(title)}\n---\n\n${body.replace(/\s+$/, '')}\n`);
}

fs.rmSync(contentDir, { recursive: true, force: true });
fs.rmSync(path.join(publicDir, 'repo'), { recursive: true, force: true });
fs.mkdirSync(publicDir, { recursive: true });

const pages = repoPages();
const sourceToSlug = new Map(pages.map((p) => [p.source, p.slug]));
const copied = new Set();

for (const p of pages) {
  const { title, body } = convert(fs.readFileSync(path.join(repoRoot, p.source), 'utf8'), p.source, sourceToSlug, copied);
  write(p.slug, title, body);
}

// generated pages: read from the generator's output, never from docs/
const adapters = [];
for (const [file, slug] of generatedPages) {
  const raw = fs.readFileSync(path.join(freshDir, file), 'utf8');
  const fm = /^---\ntitle: (.*)\ndescription: (.*)\n---\n/.exec(raw);
  if (!fm) throw new Error(`build-content: ${file} has no frontmatter`);
  // A generated page's links are written as docs/ writes them (settings-doc
  // copies rows from docs/CONFIGURATION.md), so they resolve from docs/.
  const body = outsideFences(outsideFences(raw.slice(fm[0].length), plainHtml), (chunk) =>
    chunk.replace(/(!?\[[^\]]*\]\()\s*<?([^)\s>]+)>?((?:\s+"[^"]*")?\s*\))/g,
      (all, open, target, close) => open + rewriteLink(target, `docs/${file}`, sourceToSlug, copied) + close));
  write(slug, fm[1], `${fm[2]}\n\n${body.replace(/^\s+/, '')}`);
  if (slug.startsWith('reference/adapters/')) adapters.push({ slug, title: fm[1], description: fm[2] });
}
write('reference/adapters/index', 'Adapter interfaces',
  'The Go interfaces a language, a tool, a data source, a guardrail or a model plugs in through. Each page is written from `go doc`.\n\n' +
  adapters.map((a) => `- [${a.title}](${routeOf(a.slug)}): ${a.description}`).join('\n'));

const meta = (dir, value) => {
  fs.mkdirSync(path.join(contentDir, dir), { recursive: true });
  fs.writeFileSync(path.join(contentDir, dir, 'meta.json'), `${JSON.stringify(value, null, 2)}\n`);
};
const roots = ['index', 'guides', 'reference', 'adr', 'security'];
if (pages.some((p) => p.slug.startsWith('examples/'))) roots.push('examples');
meta('.', { pages: roots.filter((r) => r === 'index' || fs.existsSync(path.join(contentDir, r))) });
meta('guides', { title: 'Guides', pages: ['index', '...'] });
meta('reference', { title: 'Reference', pages: ['cli', 'settings', 'adapters'] });
meta('reference/adapters', { title: 'Adapter interfaces', pages: ['index', '...'] });
meta('adr', { title: 'Decisions', pages: ['index', '...'] });
if (fs.existsSync(path.join(contentDir, 'security'))) meta('security', { title: 'Security', pages: ['...'] });
if (fs.existsSync(path.join(contentDir, 'examples'))) meta('examples', { title: 'Examples', pages: ['index', '...'] });

console.log(`build-content: ${pages.length} pages from the repository, ${generatedPages.length + 1} generated, ${copied.size} images`);
