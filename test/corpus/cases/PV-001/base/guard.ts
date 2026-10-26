const FRAME_ID = /data-screen\s*=\s*(?:"([^"]*)"|'([^']*)')/g;

export interface Answer {
  pass: boolean;
  missing: string[];
}

export function listFrameIds(document: string): string[] {
  const ids: string[] = [];
  for (const match of document.matchAll(FRAME_ID)) {
    ids.push(match[1] ?? match[2] ?? "");
  }
  return ids;
}

export function judge(document: string, required: readonly string[]): Answer {
  const found = new Set(listFrameIds(document));
  const missing = required.filter((id) => !found.has(id));
  return { pass: missing.length === 0, missing };
}
