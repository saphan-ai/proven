export type ExtractResult =
  | { kind: "command"; command: string }
  | { kind: "refuse"; class: string };

const FENCE = /^(`{3,}|~{3,})(.*)$/;

function linesOf(markdown: string): string[] {
  const lines = markdown.split(/\r?\n/);
  if (lines.length > 0 && lines[lines.length - 1] === "") {
    lines.pop();
  }
  return lines;
}

export function extract(markdown: string): ExtractResult {
  const lines = linesOf(markdown).map((line) => line.trim());
  let openAt = -1;
  let openChar = "";
  let openLen = 0;
  for (let i = 0; i < lines.length; i++) {
    const match = FENCE.exec(lines[i]);
    if (match) {
      openAt = i;
      openChar = match[1][0];
      openLen = match[1].length;
      break;
    }
  }
  if (openAt < 0) {
    return { kind: "command", command: "" };
  }
  const content: string[] = [];
  for (let i = openAt + 1; i < lines.length; i++) {
    const match = FENCE.exec(lines[i]);
    if (!match) {
      content.push(lines[i]);
      continue;
    }
    const fence = match[1];
    const rest = match[2];
    if (rest.length > 0) {
      return { kind: "refuse", class: "text-after-closing-fence" };
    }
    if (fence[0] !== openChar || fence.length !== openLen) {
      return { kind: "refuse", class: "closing-fence-wrong-length-or-character" };
    }
    return { kind: "command", command: content.join("\n") };
  }
  return { kind: "refuse", class: "unclosed-fence" };
}
