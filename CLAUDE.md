# CLAUDE.md

<!-- Root file = cross-cutting invariants where drift is expensive. Region-specific detail
     lives in .claude/rules/*.md with `paths:` frontmatter and loads on demand.
     Root CLAUDE.md is re-injected after /compact; rules are NOT — they reload the next
     time a matching file is read. That is the criterion for what stays here. -->

## Model — read first

Before changing anything in the **Go network** (`nodes/`, `nodes/wire/paced_wire.go`,
`nodes/Wiring/loader.go`, `nodes/Wiring/builders.go`) or the **content buffer**
(`Buffer/`, the render tree under `tools/topology-vscode/src/webview/three/`),
read [MODEL.md](MODEL.md). It pins the model. Do not propose multi-step
plans with options for network/wire work; name the single concrete next
step and get the model agreed first. "Agreed first" gates the START of the
work, not each step of it — once the model is settled, build the feature
through to done; do not halt after every step to re-ask.

Go owns the one clock and times its own bead delivery. It packs the whole scene (bead
positions, node/port geometry, edge curves, shading params, camera pose, selection,
overlays) into a **binary content buffer** and streams it. The render tree under
`three/` (rooted at `buffer-scene.tsx`, which composes it) decodes
and draws that buffer; it computes no positions, no geometry, and no traversal timing,
and never tells Go when a bead arrived. There is no JSON-trace render path and no
`pump.ts`; the TS layer is **render + forward only** and holds no domain state (guard:
`tools/check-no-webview-state.sh`).

The model's real entities live in [MODEL.md](MODEL.md): bead, wire (`PacedWire` — an
ACTIVE goroutine that owns its own beads, with a channel on each end), node goroutine,
input port, clock. The active node kinds are the structs under `nodes/<Kind>/`.

**Drift rule:** see MODEL.md's "Drift rule" section for the full statement (guards:
`tools/check-no-webview-state.sh`, `tools/check-no-await-on-bridge.sh`).

## Primitive landing rule (narrowed)

**Node kinds:** adding a kind requires four things in the same commit:
1. An entry in `NODE_DEFS` (`tools/topology-vscode/src/schema/node-defs.ts`, generated).
2. Nothing else in the schema dir — `node-defs.ts` is the single registry.
3. The Go node package under `nodes/<Kind>/`.
4. `go run ./tools/gen-node-defs`. **Skip this and the kind does not exist in the binary** —
   it fails at runtime with `unknown type "X"` while everything else looks correct.
   Guard: `check-generated.sh`.

Detail: `.claude/rules/node-kinds.md`. Buffer columns: `.claude/rules/buffer-schema.md`.
Wire props: `.claude/rules/wire-props.md`.

**Bridge surface:** **Go → TS** is binary content buffers (`buffer-snapshot`) and NOTHING
ELSE — one dedicated inherited-stdio pipe per emitting goroutine (VIEW/edge/node/interior),
no shared fd3/single-writer packer, `WIREFOLD_STREAM_FDS` mandatory.

**TS → Go** is framed binary records on stdin: **addressed edits** (a single `edit` message
whose sole op is `update`, setting an ATTRIBUTE on a typed entity — new capability is a new
entity kind or attribute, NOT a new op), **bare commands** (`save` only), and **`raw-input`**
(raw pointer/wheel + raycast hit → Go's gesture FSM; camera orbit and node/port moves are
produced in-process by the FSM, they do not cross this seam as edits).

The TS → Go send is **fire-and-forget** — no `await`, no Promise chain, no request/response,
no delivery signal (guard: `tools/check-no-await-on-bridge.sh`).

Vocabulary detail, parity guards, and the no-sidecar rule: `.claude/rules/bridge-surface.md`.

## Bash hygiene (keep AI round-trips snappy)

Bash output goes straight into the AI's context. Wide-fan commands
return hundreds of irrelevant matches from `node_modules/`, planning
docs, and the auto-memory dir, costing tokens and time.

- **`grep`**: always scope. For code, use `--include="*.ts" --include="*.tsx"`. For repo-wide searches, exclude noise: `--exclude-dir={node_modules,out,.git,handoff-archive,memory}`.
- **`find`**: never run `find .` unguarded — `tools/topology-vscode/node_modules/` has multi-MB files. Use `-not -path "*/node_modules/*" -not -path "*/out/*" -not -path "*/.git/*"` or just scope to a specific subtree.
- **`ls`**: prefer a specific subdir over wide listings; pipe to `head` if you only need a sample.
- Planning docs (`docs/planning/visual-editor/`, `memory/`) contain domain vocabulary — grep them only when the question is about *planning state*, not when looking for code.

## Testing shape — read before writing a test

A test asserts what **one goroutine itself** decided, emitted, or persisted. Do **not**
test that two or more goroutines communicate properly — not delivery, not ordering, not
absence-of-deadlock, not absence-of-race. That correctness is guaranteed BY CONSTRUCTION
here (per-mover ownership, dedicated per-pair channels, no locks/atomics — guard:
`tools/check-no-network-locks.sh`, empty allowlist), so such a test asserts what the
structure already gives you and exercises Go's runtime instead of this codebase.

The one exception is **persistence**: bytes on disk through a real reload
(`memory/feedback_headless_repro_verifies_persistence`).

Full doctrine — the dividing line, why absence assertions can't be polled, the industry
patterns and which actually transfer, a decision procedure, and named anti-patterns — is
in [docs/testing-shape.md](docs/testing-shape.md). Read it before adding a test that needs
more than one goroutine running.

## Workflow

- **Commit and push freely on task branches.** Per-commit sign-off is no longer required (relaxed post-v0; editing or reverting committed code is cheap). Sign-off IS still required for: merging a task branch into `main`, force-pushes, branch deletion, dependency removal, and any other destructive or shared-state action called out in the system prompt's "Executing actions with care" section.
- Build and run before reporting a change as ready; verify output matches previous run. If verification fails, fix forward or revert — don't leave broken state on the branch. To exercise a TS change in the LIVE editor, run `npm run build` — the Stop hook does this automatically, but manual subagent verifications do not (why `tsc --noEmit` isn't enough is in the verify recipe below; stated once, there).
  - **Verify recipe (NEVER run the sim in the foreground):** the single source of truth is `bash scripts/stop-checks.sh` run from the repo root. **Read its STDOUT, not `$?` — it ALWAYS exits 0**, by design: it speaks the Stop-hook JSON protocol, so a failure is a `{"decision":"block","reason":...}` object printed to stdout while the exit code stays 0. Clean means *empty stdout*. Therefore `stop-checks.sh >/dev/null; echo $?` is not a check — it discards the only failure signal and reads a constant. Verify with `bash scripts/stop-checks.sh` and look at the output, or gate on it with `[ -z "$(bash scripts/stop-checks.sh 2>/dev/null)" ]`. It runs go build+test, tsc `--noEmit`, the npm webview build, staticcheck, eslint, vitest, and the full guard suite (incl. message-kind-parity, polar-only-nav, no-camera-roundtrip) — gated so the expensive per-language steps run only when that language changed or the branch is ahead of origin/main. Do NOT run the old `go build && go test`, `tsc`, `npm run build` steps separately; that just duplicates what stop-checks already does. Caveats it does NOT cover: the early-exit was removed so it also runs on a clean tree, but `tsc --noEmit` alone still won't refresh `out/webview.js` (stop-checks' npm build does), and an extension-host change still needs VS Code "Developer: Reload Window" (reopening a file only reloads the webview).
- One logical change per commit. Push each commit to the current task branch.
- **Cost markers:** only record a `($N.NN)` cost marker on a commit (or bundle of commits) when the work was sized at **≥$5 expected** beforehand. Sub-$5 work lands without a marker. Bundle small commits into ≥$5 chunks for marker purposes. Pre-v0 sub-$5 markers stay as historical record but are no longer the convention.
- **One worktree per change — start every change with `tools/new-task.sh <short-kebab-name> "one-line description"`.** It creates the `task/<name>` branch from `origin/main`, its worktree at `worktrees/<name>`, and the branch description `tools/next.sh` reads. Do NOT `git checkout -b` in the main checkout: that mutates the ONE tree every session shares, and this repo has already lost a merge to it (a `merge --ff-only` landed on whichever branch the shared tree happened to be on, and a later rebase had to autostash around someone else's uncommitted edits). A worktree means `git checkout` is never needed, concurrent sessions never fight over the tree, and stop-checks verifies the tree the work is actually in. `main` stays checked out in the main directory. `node_modules` is not installed per worktree — stop-checks symlinks the main checkout's on first run.
  - When done: merge to `main` (sign-off required), then `git worktree remove worktrees/<name> && git branch -d task/<name>` (`--force` on the remove if the branch installed its own `node_modules`). Deleting both is what removes the item from `tools/next.sh`.
  - **Concurrent sessions share three things — treat all three as other people's state.** (1) The main checkout holds `main` and nothing else; guard: `check-main-checkout-on-main.sh`. (2) `node_modules` is symlinked from the main checkout, and stop-checks refuses to link when your branch's `package-lock.json` differs — run `npm install` inside your worktree instead, so you never rewrite another session's dependencies. (3) `git stash` is repo-global: a stash pushed in one worktree is visible and poppable in every other, so prefer a WIP commit on your own branch over stashing.
  - Branches stay short-lived and merge to `main` quickly. Avoid long-lived feature branches like the v0 `visual-editor` pattern.
- Channel names encode which two nodes are connected — preserve this convention.
- **Medium vs. substance.** Before adopting a **medium** dependency (rendering library, framework, parser, bundler, file watcher, test runner, package manager, language/runtime version, editor integration), explicitly ask "what's the dominant choice the rest of the world converged on for this category?" and justify deviating if not adopting it. The medium is where industry has solved your problem; being weird there is pure overhead. Do **not** apply this heuristic to the **substance** of the system — the execution model, what a node is, how time/ticks work, what a wire is, how nodes coordinate, the Go network that runs nodes. Industry defaults there encode "logic in procedures, topology as plumbing," which is the inversion this project exists to challenge. For substance, ask "what does this system actually need?" and ignore industry — the whole point is that the answer is different. (Prior failure: the await/Promise execution model was the industry-correct JS translation of goroutines+channels, and it hid pacing inside the event loop, coupling nodes that should have been independent. Right answer for the medium, wrong answer for the substance.)

## Memory and doctrine layout

- `memory/` — one file per memory (`project_*`, `feedback_*`, `user_*`); `memory/MEMORY.md`
  is the index. Guard: `check-memory-hygiene.sh`.
- `.claude/rules/*.md` — region-specific doctrine, `paths:` frontmatter, loads on demand.
  Put detail here when it only matters for one part of the codebase.
- Root CLAUDE.md — cross-cutting invariants only. It is re-injected after `/compact`;
  rules are not.
- [docs/drift-checklist.md](docs/drift-checklist.md) — the periodic agent-health audit.

## Session handoff

Live task state is NOT stored in a hand-maintained doc (that kept drifting — it was a
cache of state whose authoritative source is git/memory/MODEL.md). It is DERIVED:
`task/*` branches carry a one-line `git config branch.<name>.description`, and
**`tools/next.sh`** prints the live view (current branch, task branches, recent merges,
pointers to durable docs). Run it first in a fresh session.

Do not recreate a handoff.md or a continuation-prompt template; if state needs a home,
it belongs on a branch description, in memory, or in MODEL.md — not in a synced snapshot.
Guard: `check-no-state-cache.sh`.

**Write before compacting.** Compaction keeps CLAUDE.md/MODEL.md instructions, `memory/`
files, git state, files on disk, and planning docs. It loses intermediate reasoning, file
contents you previously read, multi-step conversation context, tool-call history, and
preferences stated only verbally. Save anything in the second list to a file before
`/compact`.

## Posture (post-v0)

Visual editor reached v0. New work is friction-driven, not phase-driven (the old per-phase plans were deleted — git history is the archive, not a `docs/planning/visual-editor/archive/` dir); justify changes from real-world editor use logged in [session-log.md](docs/planning/visual-editor/session-log.md). Working mode: user drives the editor and narrates; assistant logs and makes changes.
