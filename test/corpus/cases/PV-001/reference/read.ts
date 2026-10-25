// Reference reading: frame ids a spec-conformant HTML parser keeps.
// parse5 follows the HTML standard, which drops a duplicate attribute and
// keeps the first. This module does not use either guard.

import { parse, type DefaultTreeAdapterMap } from "parse5";

type Node = DefaultTreeAdapterMap["node"];

function walk(node: Node, ids: string[]): void {
  if ("attrs" in node) {
    for (const attr of node.attrs) {
      if (attr.name === "data-screen") {
        ids.push(attr.value);
      }
    }
  }
  if ("childNodes" in node) {
    for (const child of node.childNodes) {
      walk(child, ids);
    }
  }
}

export function listFrameIds(document: string): string[] {
  const ids: string[] = [];
  walk(parse(document), ids);
  return ids;
}
