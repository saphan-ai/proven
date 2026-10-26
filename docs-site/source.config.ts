import { defineConfig, defineDocs } from 'fumadocs-mdx/config';

// content/docs is written by scripts/build-content.mjs from the repository's own Markdown.
export const docs = defineDocs({ dir: 'content/docs' });

export default defineConfig();
