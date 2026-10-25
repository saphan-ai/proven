export type ExtractResult =
  | { kind: "command"; command: string }
  | { kind: "refuse"; class: string };

const FENCE = /^ {0,3}(`{3,}|~{3,})/;

function linesOf(markdown: string): string[] {
  const lines = markdown.split(/\r?\n/);
  if (lines.length > 0 && lines[lines.length - 1] === "") {
    lines.pop();
  }
  return lines;
}

export function extract(markdown: string): ExtractResult {
  const lines = linesOf(markdown);
  let openAt = -1;
  for (let i = 0; i < lines.length; i++) {
    if (FENCE.test(lines[i])) {
      openAt = i;
      break;
    }
  }
  if (openAt < 0) {
    return { kind: "command", command: "" };
  }
  const content: string[] = [];
  for (let i = openAt + 1; i < lines.length; i++) {
    if (FENCE.test(lines[i])) {
      return { kind: "command", command: content.join("\n") };
    }
    content.push(lines[i]);
  }
  return { kind: "command", command: content.join("\n") };
}
