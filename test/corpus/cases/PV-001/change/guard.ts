// Change guard: tokenize tags, then fold each tag's attributes into a map.
// A later duplicate attribute overwrites an earlier one. The HTML standard
// keeps the first duplicate; this map does not.

export interface Answer {
  pass: boolean;
  missing: string[];
}

function isNameStart(ch: string): boolean {
  return (ch >= "A" && ch <= "Z") || (ch >= "a" && ch <= "z");
}

function isNameChar(ch: string): boolean {
  return isNameStart(ch) || (ch >= "0" && ch <= "9") || ch === ":" || ch === "." || ch === "_" || ch === "-";
}

function isWhitespace(ch: string): boolean {
  return ch === " " || ch === "\n" || ch === "\r" || ch === "\t" || ch === "\f";
}

function skipUntil(document: string, from: number, marker: string): number {
  const at = document.indexOf(marker, from);
  return at === -1 ? document.length : at + marker.length;
}

function skipRawText(document: string, from: number, tag: string): number {
  const closer = new RegExp("</" + tag + "\\s*>", "i");
  const found = closer.exec(document.slice(from));
  if (found === null || found.index === undefined) return document.length;
  return from + found.index + found[0].length;
}

export function listFrameIds(document: string): string[] {
  const ids: string[] = [];
  const n = document.length;
  let i = 0;

  while (i < n) {
    const lt = document.indexOf("<", i);
    if (lt === -1) break;
    i = lt;

    if (document.startsWith("<!--", i)) {
      i = skipUntil(document, i + 4, "-->");
      continue;
    }

    if (document.startsWith("<!", i) || document.startsWith("<?", i)) {
      i = skipUntil(document, i + 2, ">");
      continue;
    }

    if (document.startsWith("</", i)) {
      i = skipUntil(document, i + 2, ">");
      continue;
    }

    const nameStart = i + 1;
    const first = document[nameStart];
    if (first === undefined || !isNameStart(first)) {
      i = lt + 1;
      continue;
    }

    let j = nameStart + 1;
    while (j < n) {
      const ch = document[j];
      if (ch === undefined || !isNameChar(ch)) break;
      j++;
    }
    const tag = document.slice(nameStart, j).toLowerCase();
    i = j;

    const attrs = new Map<string, string>();
    let closed = false;
    while (i < n && !closed) {
      while (i < n) {
        const ch = document[i];
        if (ch === undefined || !isWhitespace(ch)) break;
        i++;
      }
      if (i >= n) break;
      const ch = document[i];
      if (ch === ">") {
        i++;
        closed = true;
        break;
      }
      if (ch === "/" && document[i + 1] === ">") {
        i += 2;
        closed = true;
        break;
      }

      const attrStart = i;
      while (i < n) {
        const attrCh = document[i];
        if (
          attrCh === undefined ||
          isWhitespace(attrCh) ||
          attrCh === "=" ||
          attrCh === ">" ||
          attrCh === "/"
        ) {
          break;
        }
        i++;
      }
      const attrName = document.slice(attrStart, i).toLowerCase();
      if (attrName === "") {
        i++;
        continue;
      }

      while (i < n) {
        const gap = document[i];
        if (gap === undefined || !isWhitespace(gap)) break;
        i++;
      }

      let value = "";
      if (document[i] === "=") {
        i++;
        while (i < n) {
          const gap = document[i];
          if (gap === undefined || !isWhitespace(gap)) break;
          i++;
        }
        const quote = document[i];
        if (quote === '"' || quote === "'") {
          i++;
          const valueStart = i;
          while (i < n && document[i] !== quote) i++;
          value = document.slice(valueStart, i);
          if (document[i] === quote) i++;
        } else {
          const valueStart = i;
          while (i < n) {
            const valueCh = document[i];
            if (valueCh === undefined || isWhitespace(valueCh) || valueCh === ">") break;
            i++;
          }
          value = document.slice(valueStart, i);
        }
      }

      // Later duplicate wins. This is the property the case exists to show.
      attrs.set(attrName, value);
    }

    const frame = attrs.get("data-screen");
    if (frame !== undefined) ids.push(frame);

    if (tag === "script" || tag === "style") {
      i = skipRawText(document, i, tag);
    }
  }

  return ids;
}

export function judge(document: string, required: readonly string[]): Answer {
  const found = new Set(listFrameIds(document));
  const missing = required.filter((id) => !found.has(id));
  return { pass: missing.length === 0, missing };
}
