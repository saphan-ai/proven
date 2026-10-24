import { Parser } from "commonmark";

export type ExtractResult =
  | { kind: "command"; command: string }
  | { kind: "refuse"; class: string };

export function read(markdown: string): ExtractResult {
  const parser = new Parser();
  const doc = parser.parse(markdown);
  const walker = doc.walker();
  let event = walker.next();
  while (event) {
    const node = event.node;
    if (event.entering && node.type === "code_block" && node._isFenced) {
      let command = node.literal ?? "";
      if (command.endsWith("\n")) {
        command = command.slice(0, -1);
      }
      return { kind: "command", command };
    }
    event = walker.next();
  }
  return { kind: "refuse", class: "no-fenced-block" };
}
