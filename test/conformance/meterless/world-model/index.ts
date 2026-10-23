// World Model shim. Forwards each call to proven's store through meterless-conform.
// Stable ids, merge bands, provenance, ingest idempotency and the audit log
// are the store's. A bitemporal graph is outside that minimum.

import { rpc } from "../bridge.ts";

export function normalizeName(name: string): string {
  return rpc({ engine: "world", op: "normalize", name }).name;
}

export function entityIdFromName(type: string, name: string): string {
  return rpc({ engine: "world", op: "entityId", type, name }).entityId;
}

export function decide(similarity: number): string {
  return rpc({ engine: "world", op: "decide", similarity }).decision;
}

export class WorldModel {
  session: string;

  constructor(opts: { storage?: string; namespace?: string; clock?: () => Date }) {
    if (!opts?.namespace) throw new Error("WorldModel: namespace is required");
    this.session = rpc({
      engine: "world",
      op: "open",
      namespace: opts.namespace,
      storage: opts.storage ?? "memory",
    }).session;
  }

  upsertEntity(input: Record<string, unknown>) {
    const session = this.session;
    return Promise.resolve().then(() => rpc({ engine: "world", op: "upsert", session, input }));
  }

  ingest(observations: unknown[]) {
    const session = this.session;
    return Promise.resolve().then(() => {
      const ans = rpc({ engine: "world", op: "ingest", session, observations });
      return { changed: ans.changed === true };
    });
  }

  stateShapeCanonical(): string {
    return rpc({ engine: "world", op: "canonical", session: this.session }).canonical;
  }

  audit(): { seq: number }[] {
    return rpc({ engine: "world", op: "audit", session: this.session }).audit ?? [];
  }

  relate(input: Record<string, unknown>) {
    const session = this.session;
    return Promise.resolve().then(() => rpc({ engine: "world", op: "relate", session, input }));
  }

  view(name: string) {
    const session = this.session;
    return {
      neighborsOf(id: string, opts?: { via?: string; at?: string }) {
        return rpc({
          engine: "world",
          op: "neighbors",
          session,
          view: name,
          entity: id,
          via: opts?.via,
          atTime: opts?.at,
        }).neighbors;
      },
    };
  }
}
