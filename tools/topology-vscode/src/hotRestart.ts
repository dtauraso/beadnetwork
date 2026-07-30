// Decision logic for the .go-change hot-restart feature (extension.ts's goWatcher). Split
// out from extension.ts so it can be unit-tested directly (per docs/testing-shape.md: this
// is single-actor decision logic, not two goroutines/processes communicating) — extension.ts
// itself is wiring that cannot be driven headlessly (it needs a live vscode.WebviewPanel).
import type { BuildResult } from "./goBuild";

// shouldRestartAfterBuild decides whether a completed goWatcher rebuild warrants
// restarting a LIVE sim. Two guards, both required:
//   - ok must be true — a failed compile must never kill a working sim (requirement 4:
//     "killing a working sim because someone saved a file mid-edit is worse than
//     staleness"). The caller still reports res.error the normal way; this function only
//     answers the restart question.
//   - busy must be falsy — buildBinary's coalescing return (`{ok: true, busy: true}`)
//     means SOME build against binPath finished, not necessarily the one covering this
//     caller's own trigger (see goBuild.ts's buildBinary doc comment: two overlapping
//     calls coalesce into one real `go build`, and the second caller gets told "ok" without
//     having caused a build itself). Treating a coalesced ok as "my rebuild landed" would
//     restart against a binary that might not even reflect the edit that fired the watcher
//     this time; skip the restart and let the NEXT rebuild (debounced) make the call.
export function shouldRestartAfterBuild(res: BuildResult): boolean {
  return res.ok && !res.busy;
}

// TrailingDebouncer coalesces bursty calls (several fs-watcher events for one save, or a
// git checkout touching hundreds of .go files) into a single trailing invocation of `fn`,
// `delayMs` after the LAST call to schedule(). Each schedule() call restarts the window —
// this is why a single git checkout of hundreds of files produces one rebuild/restart, not
// hundreds: every file-change event keeps pushing the timer out until the burst goes quiet.
// Generic (not goWatcher-specific) so it is trivially unit-testable with vi.useFakeTimers()
// and reusable by bundleWatcher's identical pattern if that is ever consolidated.
export class TrailingDebouncer {
  private pending: ReturnType<typeof setTimeout> | undefined;

  constructor(private readonly delayMs: number) {}

  schedule(fn: () => void): void {
    if (this.pending) clearTimeout(this.pending);
    this.pending = setTimeout(() => {
      this.pending = undefined;
      fn();
    }, this.delayMs);
  }

  // dispose cancels any pending trailing call — used on panel/watcher teardown so a debounce
  // timer never fires after its owner (goChannel, runner) has been disposed.
  dispose(): void {
    if (this.pending) clearTimeout(this.pending);
    this.pending = undefined;
  }
}
