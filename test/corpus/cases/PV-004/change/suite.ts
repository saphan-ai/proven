import { extract } from "./extract";

const OPEN = "````";
const COMMAND = "echo hello";

function fenced(closer: string | null): string {
  const lines = [OPEN, COMMAND];
  if (closer !== null) {
    lines.push(closer);
  }
  return lines.join("\n") + "\n";
}

export interface ShippedCase {
  name: string;
  markdown: string;
  refusal: string;
}

export const shippedCases: readonly ShippedCase[] = [
  {
    name: "unclosed",
    markdown: fenced(null),
    refusal: "unclosed-fence",
  },
  {
    name: "closer-too-short",
    markdown: fenced("```"),
    refusal: "closing-fence-wrong-length-or-character",
  },
  {
    name: "closer-too-long",
    markdown: fenced("`````"),
    refusal: "closing-fence-wrong-length-or-character",
  },
  {
    name: "closer-longer-still",
    markdown: fenced("``````"),
    refusal: "closing-fence-wrong-length-or-character",
  },
  {
    name: "closer-tildes",
    markdown: fenced("~~~~"),
    refusal: "closing-fence-wrong-length-or-character",
  },
  {
    name: "closer-tildes-short",
    markdown: fenced("~~~"),
    refusal: "closing-fence-wrong-length-or-character",
  },
  {
    name: "text-after-closer",
    markdown: fenced("```` trailing"),
    refusal: "text-after-closing-fence",
  },
  {
    name: "text-after-closer-word",
    markdown: fenced("````note"),
    refusal: "text-after-closing-fence",
  },
];

export function shippedSuiteFailures(): string[] {
  const failures: string[] = [];
  for (const item of shippedCases) {
    const result = extract(item.markdown);
    const got =
      result.kind === "refuse" ? `refuse:${result.class}` : `command:${result.command}`;
    const want = `refuse:${item.refusal}`;
    if (got !== want) {
      failures.push(`${item.name} expected ${want} got ${got}`);
    }
  }
  return failures;
}
