---
name: feedback_two_process_editor_reload
description: VS Code webview vs extension host are separate processes; reopen-file reloads only the webview, Developer-Reload-Window reloads the extension host
metadata:
  type: feedback
---

The topology editor runs as TWO separate VS Code processes: the **webview** (`out/webview.js` — React/three render) and the **extension host** (`out/extension.js` — Go-process spawn, stdout trace parsing, the bridge). Reopening the topology file reloads ONLY the webview; the extension host keeps running its old code. To pick up extension-host changes you must run **Developer: Reload Window**. `npm run build` refreshes the on-disk bundles but does NOT reload the running host.

**Why:** A session was nearly lost chasing a "stale `flagged` bundle" and "Go not emitting geometry" when the real issue was the running extension host executing pre-rebuild code; reopen-file cleared the webview crash but left the stale host.

**The mechanism that makes this invisible:** the extension host is what allocates the dedicated stream fds and passes `WIREFOLD_STREAM_FDS` to Go (`runCommand.ts`). A stale host can hand Go plumbing that no longer matches, and Go then streams nothing for that kind — so an ENTIRE ENTITY CLASS silently disappears from the scene while every file on disk checks out clean. 2026-07-28: all edges vanished this way; the bundle was byte-identical to a fresh build and the Go headless edge test passed, because neither is where the fault was. Go now reports the fds-absent case on stderr (`main.go`), so it names itself instead of looking like a code defect.

**How to apply:** Reload Window BEFORE theorizing — and the trigger is not only "my change didn't take." It is also **"a whole class of thing stopped rendering"** (all edges, all nodes, all beads), which reads like a code regression and is the same cause. Recognising that second signature is what was missed both times this memory was written. For runtime truth prefer `.probe/*.jsonl` logs and `go run`/`go test` over gopls (which also goes stale). See [[feedback_webview_devtools_frame]], [[feedback_runtime_breadcrumbs_beat_static_analysis]].
