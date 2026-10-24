import { extract as extractBase } from "../base/extract";
import { extract as extractChange } from "../change/extract";
import { shippedCases, shippedSuiteFailures } from "../change/suite";
import { read as readReference } from "./reference";

interface Input {
  name: string;
  markdown: string;
}

const extra: readonly Input[] = [
  {
    name: "well-formed",
    markdown: "```\necho hello\n```\n",
  },
  {
    name: "info-string-backtick",
    markdown: "``` foo ` bar\necho hello\n```\n",
  },
  {
    name: "indent-four",
    markdown: "    ```\n    echo hello\n    ```\n",
  },
];

function encode(text: string): string {
  return text.replace(/\\/g, "\\\\").replace(/\n/g, "\\n").replace(/\r/g, "\\r");
}

function field(result: { kind: "command"; command: string } | { kind: "refuse"; class: string }): string {
  if (result.kind === "refuse") {
    return `refuse:${result.class}`;
  }
  return `command:${encode(result.command)}`;
}

function report(input: Input): string {
  const base = field(extractBase(input.markdown));
  const change = field(extractChange(input.markdown));
  const reference = field(readReference(input.markdown));
  return `${input.name} base=${base} change=${change} reference=${reference}`;
}

const failures = shippedSuiteFailures();
if (failures.length > 0) {
  for (const failure of failures) {
    process.stderr.write(`${failure}\n`);
  }
  process.exit(1);
}
process.stderr.write(`shipped-suite ok ${shippedCases.length}\n`);

const inputs: Input[] = [
  ...shippedCases.map((item) => ({ name: item.name, markdown: item.markdown })),
  ...extra,
];

for (const input of inputs) {
  process.stdout.write(`${report(input)}\n`);
}
