// H-MEM shim. Forwards each call to proven's store through meterless-conform.
// Sleep, dreaming, conflicts and ledger history each reach their own operation
// in the command. Each call is one process, so what a later call needs from an
// earlier one (a sleep plan, a conflict) is handed back to it.

import { rpc } from "../bridge.ts";

function fill(rec: any): any {
  if (rec == null) return rec;
  rec.tags = rec.tags ?? [];
  rec.entities = rec.entities ?? [];
  rec.relatedTo = rec.relatedTo ?? [];
  return rec;
}

export class HMEM {
  clock: () => number;
  session: string;
  store: { get(id: string): any; all(): any[] };
  ledger: { history(id: string): any };
  sleep: { preview(): any; execute(preview: unknown): any };
  dreaming: { dream(): any; approve(id: string): any };
  conflicts: { scan(): any; autoResolve(found: unknown): any };

  constructor(opts: { clock?: () => number } = {}) {
    this.clock = opts.clock ?? (() => Date.now());
    this.session = rpc({ engine: "hmem", op: "open" }).session;
    const session = this.session;
    const clock = () => this.clock();
    const call = (op: string, extra: Record<string, unknown> = {}) =>
      rpc({ engine: "hmem", op, session, at: clock(), ...extra });
    this.store = {
      get(id: string) {
        return fill(call("get", { memoryId: id }).record);
      },
      all() {
        const rows = call("all").records ?? [];
        return rows.map(fill);
      },
    };
    this.ledger = {
      history(id: string) {
        return call("ledger", { memoryId: id }).history;
      },
    };
    this.sleep = {
      preview() {
        return call("sleep").preview;
      },
      execute(preview?: unknown) {
        return call("sleepExecute", { preview: preview ?? null }).report;
      },
    };
    this.dreaming = {
      dream() {
        return call("dream").proposals;
      },
      approve(id: string) {
        return call("approve", { memoryId: id }).record ?? undefined;
      },
    };
    this.conflicts = {
      scan() {
        return call("conflicts").conflicts;
      },
      autoResolve(found: unknown) {
        return call("autoResolve", { conflict: found }).outcome;
      },
    };
  }

  add(input: Record<string, unknown> = {}) {
    const rec = fill(rpc({
      engine: "hmem",
      op: "add",
      session: this.session,
      at: this.clock(),
      input,
    }).record);
    const session = this.session;
    const clock = () => this.clock();
    return new Proxy(rec, {
      set(target, prop, value) {
        target[prop as string] = value;
        if (prop === "supersededBy") {
          rpc({
            engine: "hmem",
            op: "supersede",
            session,
            at: clock(),
            memoryId: target.id,
            by: value,
          });
        }
        return true;
      },
    });
  }

  query(query: string, opts?: { topN?: number; threshold?: number }) {
    // The runner's 0 means keep every score. The Go store uses 0 for its default.
    let threshold = opts?.threshold;
    if (threshold === 0) threshold = -1;
    return rpc({
      engine: "hmem",
      op: "query",
      session: this.session,
      at: this.clock(),
      query,
      topN: opts?.topN ?? 0,
      threshold: threshold ?? 0,
    }).result;
  }

  feedback(memoryId: string, kind: string) {
    rpc({
      engine: "hmem",
      op: "feedback",
      session: this.session,
      at: this.clock(),
      memoryId,
      kind,
    });
  }
}
