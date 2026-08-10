---
name: VS Code webview devtools default to the outer wrapper frame
description: Console diagnostics like `typeof window.X` in a VS Code webview's devtools attach to the outer wrapper frame by default, not the iframe running the bundle — they return undefined for symbols that exist
type: feedback
---

In a VS Code webview's devtools, the Console tab's frame dropdown
defaults to the outer wrapper (`top`), which has none of our bundle's
globals. Running `typeof window.__pulseProbeDump` etc. returns
`'undefined'` even when the bundle is loaded and running fine.

**Why:** Burned ~half a session on the smoothness probe re-run
(2026-05-04, `task/probe-rerun-smoothness`) chasing this. Repeatedly
"fixed" non-bugs (eager-init guard, `acquireVsCodeApi` re-throw, even
a cache-bust attempt that broke things and had to be reverted) before
realising the symbols were defined all along — just in the inner
iframe's context. Canvas rendering and node drag working were the
real signal that the bundle was fine.

**How to apply:** Don't use `typeof window.X` in webview devtools as
the verification method. Prefer a webview→host file-bridge: have the
webview write a JSON to disk via `vscode.postMessage`, and read the
file directly. If a console call is unavoidable, tell the user to
switch the Console frame dropdown from `top` to the inner frame
first. If user reports a webview symbol as `undefined` but the UI
still renders, suspect frame attachment before chasing module-init
bugs.
