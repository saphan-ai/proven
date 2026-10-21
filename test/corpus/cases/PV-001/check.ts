// One line per §2 input: name, base result, change result, reference result.
// A result is the guard's pass/fail answer, with missing ids when it fails.
// The reference result applies that same required-id test to the parser's
// frame ids. It does not call either guard.

import { judge as judgeBase } from "./base/guard.js";
import { judge as judgeChange } from "./change/guard.js";
import { listFrameIds as referenceFrameIds } from "./reference/read.js";

declare const console: { log(message: string): void };

const REQUIRED = ["home"] as const;

const REAL_FRAME = '<section data-screen="home"></section>';

function withRealFrame(attack: string): string {
  return attack + "\n" + REAL_FRAME;
}

const comment = '<!-- <div data-screen="home"></div> -->';
const script = "<script>var markup = '<div data-screen=\"home\"></div>';</script>";
const style = "<style>.card::before { content: '<div data-screen=\"home\"></div>'; }</style>";
const attributeValue = "<img alt='data-screen=\"home\"' src=\"spacer.gif\">";
const cdata = '<![CDATA[<div data-screen="home"></div>]]>';
const splitLine = '<p>data-screen=\n"home"</p>';
const duplicate = '<div data-screen="other" data-screen="home"></div>';

const inputs: readonly { name: string; html: string }[] = [
  { name: "comment-without-frame", html: comment },
  { name: "comment-with-frame", html: withRealFrame(comment) },
  { name: "script-without-frame", html: script },
  { name: "script-with-frame", html: withRealFrame(script) },
  { name: "style-without-frame", html: style },
  { name: "style-with-frame", html: withRealFrame(style) },
  { name: "attribute-value-without-frame", html: attributeValue },
  { name: "attribute-value-with-frame", html: withRealFrame(attributeValue) },
  { name: "cdata-without-frame", html: cdata },
  { name: "cdata-with-frame", html: withRealFrame(cdata) },
  { name: "split-line-without-frame", html: splitLine },
  { name: "split-line-with-frame", html: withRealFrame(splitLine) },
  { name: "duplicate-attribute", html: duplicate },
];

function formatAnswer(pass: boolean, missing: readonly string[]): string {
  if (pass) return "pass";
  return "fail missing=" + missing.join(",");
}

function referenceResult(html: string): string {
  const found = new Set(referenceFrameIds(html));
  const missing = REQUIRED.filter((id) => !found.has(id));
  return formatAnswer(missing.length === 0, missing);
}

for (const input of inputs) {
  const base = judgeBase(input.html, REQUIRED);
  const change = judgeChange(input.html, REQUIRED);
  console.log(
    [
      input.name,
      "base=" + formatAnswer(base.pass, base.missing),
      "change=" + formatAnswer(change.pass, change.missing),
      "reference=" + referenceResult(input.html),
    ].join(" "),
  );
}
