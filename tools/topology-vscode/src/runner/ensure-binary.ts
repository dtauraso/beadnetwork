import * as fs from "fs";
import { buildBinary, maxGoMtime } from "../goBuild";

// ensureBinaryBuilt builds the Go binary at binPath if it's missing or stale.
// A rebuild is needed when binPath does not exist OR any *.go source under
// repoRoot is newer than binPath. Up-to-date → no build, returns ok. This
// replaces `go run .` (which relinks a throwaway binary every launch) with a
// single prebuilt binary reused across animation start/restart.
//
// Lazy safety net: even with the eager .go watcher (see extension.ts) keeping
// the binary fresh, this still rebuilds at launch when the watcher missed an
// event or wasn't armed. It delegates to the guarded buildBinary, so if the
// watcher is mid-build this call coalesces (busy → ok) wait-free and never
// blocks run().
// ensureBinaryBuilt retries buildBinary/existsSync this many times while waiting out an
// in-flight coalesced Go binary build (see below) before giving up and reporting an error.
// Each attempt is cheap (a build call that returns immediately when coalesced, or a stat
// check), so 50 is a generous ceiling meant to absorb a slow first-open Go build without
// hanging the extension host indefinitely on a build that never completes.
const BUILD_BINARY_MAX_ATTEMPTS = 50;

export function ensureBinaryBuilt(
  repoRoot: string,
  binPath: string,
): { ok: true } | { ok: false; error: string } {
  let binMtime = -1;
  try {
    binMtime = fs.statSync(binPath).mtimeMs;
  } catch { /* missing → rebuild */ }
  const needsRebuild = binMtime < 0 || maxGoMtime(repoRoot) > binMtime;
  if (!needsRebuild) return { ok: true };
  // buildBinary may COALESCE (returns ok with busy:true) when a watcher build is
  // in flight against the same output path. On first open the binary can be
  // absent AND a watcher build in flight — a coalesced ok would let run() spawn a
  // non-existent path (ENOENT, runner stuck). So only report ok once the binary
  // actually exists on disk: retry buildBinary until it runs to completion (the
  // guard is released) or the in-flight build has produced the binary.
  for (let attempt = 0; attempt < BUILD_BINARY_MAX_ATTEMPTS; attempt++) {
    const res = buildBinary(repoRoot, binPath);
    if (!res.ok) return res;
    if (!res.busy) {
      // Our own build ran synchronously to completion (ok). Trust it — but sanity
      // check the file so a silent absence still surfaces as an error, not ENOENT.
      if (fs.existsSync(binPath)) return { ok: true };
      return { ok: false, error: `go build reported success but ${binPath} is missing` };
    }
    // Coalesced against an in-flight build. If that build has already produced the
    // binary, we're done; otherwise retry (the guard will clear and our own build runs).
    if (fs.existsSync(binPath)) return { ok: true };
  }
  return {
    ok: false,
    error: `binary ${binPath} not built after ${BUILD_BINARY_MAX_ATTEMPTS} attempts`,
  };
}
