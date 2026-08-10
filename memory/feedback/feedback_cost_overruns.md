---
name: Where the larger session costs come from
description: Sessions that overshot their expected size were driven by speculative tooling fixes layered onto an unverified diagnosis — pause and re-check the diagnosis before adding tooling
type: feedback
---

Concrete cost overruns and the pattern behind them. Update as more
data points appear.

- **2026-05-04, `task/probe-rerun-smoothness` ($14.21).** Real work
  was a small webview→host bridge to dump the probe log to a file.
  The cost ballooned because a single bad diagnosis ("symbol
  `undefined` → bundle isn't running") drove four speculative
  fixes (eager init refactor, `acquireVsCodeApi` re-throw guard,
  cache-bust query string, auto-reload watcher). Two of those
  caused VS Code-internal `toUrl` errors and had to be reverted.
  The actual cause was the wrong devtools frame — the bundle had
  been working all along. See
  [feedback_webview_devtools_frame.md](feedback_webview_devtools_frame.md).

**Why:** A wrong diagnosis compounds — each "fix" adds new state
(reverted commits, broken caches, restarted hosts) that pollutes the
next diagnostic. The cheap signal (canvas renders → bundle runs) was
ignored in favor of the expensive one (console returns the wrong
answer in the wrong frame).

**How to apply:** When a diagnostic conflicts with a working UI
("symbol missing" but "feature works"), trust the UI and suspect the
diagnostic *before* writing any fix code. Stop and ask "what does it
mean if both observations are true?" — usually the diagnostic is
attached to the wrong thing. Don't add tooling (cache-busters,
watchers, hot-reload bridges) on speculation; demand a clear failing
test or repro first. Tooling that breaks the dev loop costs more
than the original bug.
