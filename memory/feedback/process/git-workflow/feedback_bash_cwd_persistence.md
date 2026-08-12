---
name: Bash cwd persists across calls — use absolute paths for destructive ops
description: Bash tool's working directory persists between calls; relative-path rm/mv after an earlier cd can hit the wrong target
type: feedback
---

The Bash tool's cwd persists across invocations. A `cd tools/foo && npm install` leaves cwd at `tools/foo` for the next call. A subsequent `rm -rf node_modules package.json` intended for the repo root then deletes the wrong package.json.

**Why:** Near-miss on 2026-05-03 — deleted `../tools/topology-vscode/package.json` while trying to remove a stray root one. Recovered via `git checkout HEAD --`, but the file could have been untracked.

**How to apply:** For any destructive op (`rm`, `mv`, overwriting redirects), use absolute paths — never rely on cwd. Or run `pwd` first if in doubt. Especially after a recent `cd` in another Bash call.
