import test from 'node:test';
import assert from 'node:assert/strict';
import { renderLLMS } from './llms.mjs';

test('each page reaches both files; source links become absolute outside code', () => {
  const pages = [
    { title: 'Start', url: '/docs', body: 'Try [the guide](/docs/guides/model_free).\n\n```sh\necho "[literal](/docs/x)"\n```' },
    { title: 'No model', url: '/docs/guides/model_free', body: 'Deterministic checks, not an exhaustive proof.' },
  ];
  const output = renderLLMS(pages, 'https://docs.example.com');
  for (const p of pages) {
    assert.ok(output.index.includes(`https://docs.example.com${p.url}`));
    assert.ok(output.full.includes(`URL: https://docs.example.com${p.url}`));
  }
  assert.ok(output.full.includes('[the guide](https://docs.example.com/docs/guides/model_free)'));
  assert.ok(output.full.includes('echo "[literal](/docs/x)"'));
  assert.ok(output.full.includes(pages[1].body));
});

test('duplicate pages and unsafe origins fail generation', () => {
  const p = { title: 'Page', url: '/docs', body: 'Text' };
  assert.throws(() => renderLLMS([p, p], 'https://docs.example.com'), /duplicate/);
  assert.throws(() => renderLLMS([p], 'https://secret@docs.example.com'), /origin/);
  assert.throws(() => renderLLMS([p], 'http://docs.example.com'), /origin/);
});
