// Markovian shim. Forwards each call to proven's engine (internal/markov) through meterless-conform.
// The runner's generator is a function in this process, so the run is driven here one step at a time:
// "start" names the first prompt, the generator answers it, "feed" takes the answer with the machine
// the last answer carried. Which prompt comes next, the markers, the carried state, the accounting and
// the reflection are all the Go machine's; this file only carries text both ways (ADR-0048 point 3).

import { rpc } from "../bridge.ts";

export function modelTotals(n: number) {
  return rpc({ engine: "markov", op: "modelTotals", n }).totals;
}

export function runEfficiency(chunks: unknown[], config: unknown) {
  return rpc({ engine: "markov", op: "runEfficiency", chunks, chunkConfig: config }).efficiency;
}

type Generator = (prompt: string, attachments: unknown[], systemPrompt?: string) => Promise<{ text: string }>;

export class Markovian {
  config: Record<string, number>;
  clock: () => number;
  generator?: Generator;

  constructor(opts: { chunkConfig?: Record<string, number>; clock?: () => number; generator?: Generator } = {}) {
    // The configuration is checked by the engine; a refused value throws here, as the spec asks.
    this.config = rpc({ engine: "markov", op: "config", chunkConfig: opts.chunkConfig ?? {} }).config;
    this.clock = opts.clock ?? Date.now;
    this.generator = opts.generator;
  }

  async run(options: {
    goal?: string;
    prompt?: string;
    mode?: string;
    generator?: Generator;
    memoryContext?: string;
    attachments?: { name: string; content: string }[];
    abortSignal?: AbortSignal;
    reflect?: boolean;
  }) {
    const gen = options.generator ?? this.generator;
    if (!gen) throw new Error("Markovian: provide a generator");
    let ans = rpc({
      engine: "markov",
      op: "start",
      chunkConfig: this.config,
      goal: options.goal ?? options.prompt ?? "",
      mode: options.mode ?? "",
      reflect: options.reflect !== false,
      memoryContext: options.memoryContext ?? "",
      attachments: options.attachments ?? [],
      at: this.clock(),
    });
    while (!ans.done) {
      const aborted = options.abortSignal?.aborted === true;
      let text = "";
      let failure: string | undefined;
      if (!aborted) {
        try {
          const attachments = ans.step === 1 ? options.attachments ?? [] : [];
          text = (await gen(ans.prompt, attachments, ans.system || undefined)).text;
        } catch (err) {
          failure = (err as Error)?.message ?? String(err);
        }
      }
      ans = rpc({ engine: "markov", op: "feed", machine: ans.machine, text, failure, aborted, at: this.clock() });
    }
    return ans.run;
  }
}
