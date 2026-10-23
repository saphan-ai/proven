// One JSON request, one JSON answer, one process.
// Node's pipe handles do not expose a file descriptor, so a long-running
// process cannot be read synchronously. The stores live on disk and the
// next call opens the same session.

import { spawnSync } from "node:child_process";

export function rpc(req: Record<string, unknown>): any {
  const bin = process.env.METERLESS_CONFORM_BIN;
  if (!bin) throw new Error("METERLESS_CONFORM_BIN is not set");
  const res = spawnSync(bin, [], {
    input: JSON.stringify(req) + "\n",
    encoding: "utf8",
    maxBuffer: 8 * 1024 * 1024,
  });
  if (res.error) throw res.error;
  if (res.status !== 0) {
    const err = (res.stderr || "").trim();
    throw new Error(err || "meterless-conform exit " + res.status);
  }
  const line = (res.stdout || "").trim();
  if (!line) throw new Error("meterless-conform returned no answer");
  const ans = JSON.parse(line);
  if (!ans.ok) throw new Error(ans.error || "meterless-conform error");
  return ans;
}
