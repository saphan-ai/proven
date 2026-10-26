// Scout Intent shim. Forwards each call to proven's Go engine (internal/scout) through meterless-conform.
// The engine is the deterministic Stage-1 scan; the spec's optional Stage-2 model is not built.

import { rpc } from "../bridge.ts";

export class Scout {
  registry: string;
  policyPack: unknown;

  constructor(opts: { intentRegistry: string; policyPack?: string | string[] }) {
    if (typeof opts?.intentRegistry !== "string") {
      throw new Error("Scout: intentRegistry must be a file path");
    }
    this.registry = opts.intentRegistry;
    this.policyPack = opts.policyPack ?? "default";
  }

  decide(input: { prompt: string; user?: { id?: string; role?: string }; surface?: string }) {
    const { registry, policyPack } = this;
    return Promise.resolve().then(
      () =>
        rpc({
          engine: "scout",
          op: "decide",
          registry,
          policyPack,
          prompt: input.prompt,
          user: input.user ?? {},
          surface: input.surface ?? "chat",
        }).decision,
    );
  }

  classify(input: { prompt: string; surface?: string }) {
    const { registry, policyPack } = this;
    return Promise.resolve().then(() => ({
      candidates: rpc({
        engine: "scout",
        op: "classify",
        registry,
        policyPack,
        prompt: input.prompt,
        surface: input.surface ?? "chat",
      }).candidates,
    }));
  }

  replan(input: { parent: string; reason: string }) {
    const { registry } = this;
    return Promise.resolve().then(() => rpc({ engine: "scout", op: "replan", registry, ...input }));
  }
}
