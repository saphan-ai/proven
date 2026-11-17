import { lines } from './lib.mjs';

// Use exactly the page bodies prepared for the site; no second document list.
export function renderLLMS(pages, origin) {
  const base = new URL(origin);
  if (base.protocol !== 'https:' || base.pathname !== '/' || base.search || base.hash || base.username || base.password) {
    throw new Error('llms: expected an HTTPS documentation origin');
  }
  const seen = new Set();
  const index = ['# Proven documentation', '', '> Current repository-backed documentation for Proven, an evidence-based Git/PR reviewer that works without a model. Readiness and limitations are part of the product contract.', '',
    `- [Full documentation](${base.origin}/llms-full.txt): Every rendered documentation page in one text file.`,
    '- [Product facts](https://github.com/saphan-ai/proven/blob/main/llms.txt): Capabilities, model-free operation, evidence boundaries and interpretation guidance.', '', '## Pages', ''];
  const full = ['# Proven documentation — full text', '', '> Generated from the same pages as the documentation site. Recorded examples, measurements, current behavior and roadmap claims retain their original labels.', ''];
  for (const page of pages) {
    if (!page.title || !page.url.startsWith('/docs') || seen.has(page.url)) throw new Error('llms: missing title or duplicate/invalid page URL');
    seen.add(page.url);
    const url = new URL(page.url, base).href;
    const body = lines(page.body).map((line) => line.inFence ? line.text : line.text.replace(/(\]\()\/(?!\/)([^)\s]*)(\))/g, `$1${base.origin}/$2$3`)).join('\n').trim();
    index.push(`- [${page.title.replaceAll(']', '\\]')}](${url})`);
    full.push('---', '', `# ${page.title}`, '', `URL: ${url}`, '', body, '');
  }
  return { index: index.join('\n') + '\n', full: full.join('\n') };
}
