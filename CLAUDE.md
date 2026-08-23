# CLAUDE.md

## Model — read first

Before changing anything in the **Go network** (`Categories/Node/`, `Categories/NodeKinds/`, `Categories/Node/BeadAnimation/bead_line.go`,
`Categories/Scene/scenebuild/load.go`, `Categories/Scene/scenebuild/spec_seed.go`) or the **Go → TS
surface** (the block files and their `*_values.go`, `Start/extension/webview/`), read
[MODEL.md](MODEL.md). It pins the model. Do not propose multi-step
plans with options for network/bead work; name the single concrete next
step and get the model agreed first. "Agreed first" gates the START of the
work, not each step of it — once the model is settled, build the feature
through to done; do not halt after every step to re-ask.

Go owns the one clock and times its own bead delivery. It writes the whole scene (bead
positions, node/port geometry, edge curves, shading params, camera pose, selection,
overlays) to **block files**, each written by the goroutine that owns it, the row in the PATH
so there is one writer per file and no lock. The render tree under `Start/extension/webview/` (rooted at
`Start/extension/webview/scene/scene-root.tsx`, which composes it) READS those files and draws them; it
computes no positions, no geometry, no traversal timing, and never tells Go when a bead
arrived. There is no JSON-trace render path and no
`pump.ts`; the TS layer is **render + forward only** and holds no domain state (guard:
`Start/extension/webview/check-no-webview-state.sh`).

The model's real entities live in [MODEL.md](MODEL.md): bead (data carrying its own segment
and step count), bead line (`BeadLine` — the line a bead travels, holding the beads on it,
stepped by its SOURCE NODE's animation goroutine, which owns a bead from placement to
delivery), node goroutine, node input, and clock
([docs/model/entities.md](docs/model/entities.md); a line's step count is
`Categories/Node/Edge/edgegeom/chain_length.go`). The active node kinds are the structs under `Categories/NodeKinds/<Kind>/`.

**Drift rule:** see MODEL.md's "Drift rule" section for the full statement (guards:
`Start/extension/webview/check-no-webview-state.sh`, `check-no-await-on-bridge.sh`).

## Primitive landing rule (narrowed)

**Node kinds:** adding a kind requires four things in the same commit:
1. An entry in `NODE_DEFS` (`Categories/NodeKinds/node-defs.ts`, generated).
2. No separate `registry.ts` — `node-defs.ts` is the single node-kind registry, and it lives
   in `Categories/NodeKinds/` with the kinds it describes. **There is no `src/schema/`**: a registry
   lives with its concern, so `messages.ts` and the input codec are `Categories/Scene/Drag/`,
   `scenes-gen.ts` is `Categories/Scene/`, `wire-defs.ts` is `Categories/Scene/loadspec/`, and the trace
   events are `Trace/`. Adding a node kind touches only `node-defs.ts`.
3. The Go node package under `Categories/NodeKinds/<Kind>/`, with its logic always in `node.go` (never
   `<Kind>.go`) plus `SPEC.md`. Directory casing is mixed and both are live: PascalCase
   (`Time`, `TimeEnd`, `TimeStart`, `PulseLeft`, `PulseRight`) and lowercase (`input`,
   `pacer`, `pulse`, `selectleft`, `selectright`) — don't infer one from the other.
4. `go generate ./...`. **Skip this and the kind is in no switch**, so loading a scene
   that names it fails saying exactly that. Guard: `check-generated.sh`.

Detail: `.claude/rules/node-kinds.md`. Wire props: `.claude/rules/wire-props.md`.

**Bridge surface:** **Go → TS** is the BLOCK FILES the owning goroutine writes and NOTHING ELSE —
one writer per file, no frames, no pipes, no host→webview message at all (trace events were the
last to stream). Never add a stream to "push" a change: the file IS the current value.

**TS → Go** is framed binary records on stdin: **addressed edits** (a single `edit` message
whose sole op is `update`, setting an ATTRIBUTE on a typed entity — new capability is a new
entity kind or attribute, NOT a new op), **bare commands** (`save` only), and **`raw-input`**
(raw pointer/wheel + raycast hit → Go's gesture FSM; camera orbit and node/port moves are
produced in-process by the FSM, they do not cross this seam as edits).

The TS → Go send is **fire-and-forget** — no `await`, no Promise chain, no request/response,
no delivery signal (guard: `check-no-await-on-bridge.sh`).

Vocabulary detail, parity guards, and the no-sidecar rule: `.claude/rules/bridge-surface.md`.

## Repo layout — a thing and everything about it share a directory

There is **no `tools/`**. It was removed because it had stopped meaning anything: it held
the whole editor (107 Go files in production packages), the generators, and every guard.

There is **no `cmd/`** either, and **no `gen/`**: both grouped code by technique. Each generator lives with the thing it generates and is NAMED for what it produces — `values` (a block file's value names and paths), `wire` (message/attr definitions), `shading`, `flags`, `kindports`, `portsdef`, `kindtables` — one directory out, since a directory is one Go package. `go generate ./...` runs all of them, and the guards find them by the `//go:generate go run .` directive, never by directory name. A generator must NOT import the package it writes a `.go` file into (delete the output and the generator stops compiling) — why `Categories/Node/kindtables` is a sibling of `Categories/Node/values`, not inside it.

- **`src/`** — the npm package's source root, and the editor: each concern directory holds
  the Go that writes the thing and the TS that draws it, plus its `*_values.go`, generated
  `*-values-gen.ts`, and the guards that protect it. `src` keeps that name because
  npm, tsconfig and esbuild all assume it; directory naming for an npm package is medium,
  not substance. The package root is the REPO root — `package.json`, `tsconfig.json` and
  `node_modules/` live there, so there is one npm project and no path mappings.
- **`Categories/NodeKinds/`** — the node kinds, plus only the registry naming them (`builder.go`'s
  `Builder`, the generated `kinds_gen.go` switch). A kind is a LEAF: it declares its OWN port
  `bindings`, `deps` and `kindBuilder` and imports no shared kind package — Go matches an interface
  by SHAPE, so `scenebuild.BuildDeps` satisfies a kind's `deps` and its builder satisfies
  `NodeKinds.Builder`, neither naming the other. A directory counts as a kind only if it declares
  `var Builder = BuilderFor(…)`; `go generate` writes its `case` into `NodeKinds.BuilderFor` and
  nothing registers at runtime. A kind's ports come from its SPEC.md `## Ports`
  table and NOWHERE else — direction `in`, `out`, or `broadcast` (an out that fans to every
  downstream edge); deriving them from Go field types instead grew phantom ports, and stays out
  (`.claude/rules/node-kinds.md`). Both `node-defs.ts` and each kind's own generated `kindPorts`
  come from that table, so the kind passes no port list, and `go generate` FAILS if a kind binds a
  name the table lacks (which would silently bind a dead-end channel). The check after touching this directory is a `node-defs.ts` diff, not a build.
- **`Categories/Clock/`** — the human-speed clock, one of MODEL.md's own entities alongside the
  bead and the node, so it is a sibling of `Categories/Node/` rather than a part of it: `MsPerTick`, the
  `Clock` interface every goroutine holds its own `Copy()` of, and the sleep/speed delivery.
  It is the ONLY place a `time.Sleep`/`After`/`NewTicker` may park a goroutine
  (`check-no-wall-clock-wait.sh`, whose exempt list names two files here and nothing else).
- **`Categories/Node/`** — the node itself: its actor and geometry, its movers and rules, the
  per-owner FILES it writes (`BeadAnimation/`, `Edge/`, `Interior/`), its poles, its block file. A directory here is NOT a kind — the scanner and `check-dep-rules.sh` decide that by
  the `Register(...)` call, not by placement.
- **The TS→Go half of the bridge has no category of its own** — each part sits with what it is about:
  `Scene/Drag/` the raw pointer/wheel record, wire format, hit kinds, NDC and rect; `Scene/Gesture/` the
  FSM making orbits and handholds, `Node/Gesture/` the half that grabs, drags and hovers a node.
- **`src/spatial/`** — `Vec3`, `Segment`, eight operations, 37 lines, importing only `math`. It
  is the MEDIUM, deliberately unremarkable; the substance sits on top in **`Categories/Polar/`** —
  `polar` (coordinate/composition) and `polarindex` (index × constant) — never inside it.
  The value file — one primitive per file, fixed-width LE, atomic rename — is **not a
  package**: each concern owns its `value_file.go` and `leaf-values.ts`, because the layout
  need only agree between the Go writing that concern's block file and the TS reading it.
- **`Categories/Scene/`** — starting the program: claim the stream fds, resolve the scene,
  load the graph, wire every per-owner stream, seed the static columns, then launch one
  goroutine per node and block. `main.go` calls it and nothing else does. It is NOT a
  coordinator — it constructs and starts, and the network runs itself from there.
- **`Categories/Ring/`** — the ring, the shape a node and a bead are both drawn as: canonical torus
  points computed in Go, meshed in TS. `Categories/Ring/NodeShape/` and `Categories/Ring/Bead/` are the two that
  share it. "Categories/Ring" over "torus" because the codebase already votes that way — `ringM0`,
  `ringPick` and `ringBand` against a handful of torus names in the low-level math.
- **`Categories/Ring/Bead/`** — ONE bead: ring surface, style, its block file. SEVERAL beads — spacing, chaining, framing — is what a node does with beads, and belongs to the next bullet.
- **`Categories/Node/BeadAnimation/`** — the whole bead process, which is what a node uses beads
  for: `BeadLine` (the line beads travel, state with no goroutine of its own), the `Sender`
  and `Receiver` on each end, the slot `lattice/`, and the animation goroutine that steps it.
  The split line is `inflightBead` — the files that share it are the bead animation.
- **`Categories/Scene/Camera/`** — the camera, all of it: the basis/projection/angles math and `Viewpoint`,
  the files it persists under `view/camera/`, its block file, and the TSX drawing through it.
- **Trace events are not a package.** Each owner declares its own `RowEvent` and labels and
  appends to the file of the item it tracks — the emitter writes it ITSELF, so an event never
  crosses a package, and a label is a NAME, not an index. `scripts/readtrace` is the only reader — one program for every owner's trace.bin.
- **`Categories/Chrome/`** — the UI that is NOT the diagram: the pills, panels, dropdowns, tab strip
  and fit chip, plus the `chrome-theme.ts` they share ("Categories/Chrome" is the industry word for the
  frame around the content, and this repo reached for it twice on its own). The test is a
  `draw-*.ts`: chrome is drawn onto `ChromeCanvas`'s canvas, while the diagram is drawn in
  the scene. Each piece holds ALL of itself: its layout/hit-testing Go,
  `*_values.go`, generated `*-values-gen.ts`, `draw-*.ts`. A chrome piece does not perform
  topology edits — node create/delete is `Categories/Scene/structuraledit`, not the dropdown offering it.
  `Categories/Overlay/` and `Categories/RingPoint/` are NOT chrome — they are block files for the diagram.
- **`Start/extension/`** — the VS Code extension: our code, which RUNS IN the extension host
  (the Node process VS Code spawns) and is not that host — naming it `Host` said we were the
  container rather than the guest. Everything that is neither Go nor the
  webview: activation and spawn (`extension.ts`, `runCommand.ts`, `goBuild.ts`), the VS Code
  side (`html.ts`, `handle-message.ts`, the three dev-loop watchers), and `runner/`, the Go
  side — stdio sizing, per-owner stream demux, last-frame cache. `esbuild.mjs`'s entry point
  is `Start/extension.ts`; the bundle is still `out/extension.js`, which `package.json`'s
  `main` names.
- **`scripts/`** — what serves the repo rather than one concern: `stop-checks.sh`, the
  git-workflow scripts, `lib/`, `checks/` for guards that guard nothing in particular
  (clustered by concern: prose, hooks, lang, meta, source), `readtrace/`, the one decoder for every
  owner's trace.bin, and `genpaths/` — roots, Fatalf, Announce and `params/`, imported not copied.
- **A guard lives beside what it guards**, named by its own `PLACEMENT:` header — there is
  no table mapping guard to folder that could disagree with the header. `scripts/guard-list.sh`
  finds them by searching the repo, and refuses to report fewer than 40.
- Two guards keep this from rotting: `check-guard-paths-exist.sh` fails when any script
  names a path that no longer resolves (a guard greping a moved path still exits 0, which
  reads as coverage), and `check-dir-size.sh` fails when a directory outgrows its ceiling.
  Deliberate absences are declared in `scripts/checks/meta/paths-may-be-absent.tsv`, one
  reason per line.

## Bash hygiene (keep AI round-trips snappy)

Bash output goes straight into the AI's context. Wide-fan commands
return hundreds of irrelevant matches from `node_modules/`, planning
docs, and the auto-memory dir, costing tokens and time.

- **`grep`**: always scope. For code, use `--include="*.ts" --include="*.tsx"`. For repo-wide searches, exclude noise: `--exclude-dir={node_modules,out,.git,handoff-archive,memory}`.
- **`find`**: never run `find .` unguarded — `node_modules/` has multi-MB files. Use `-not -path "*/node_modules/*" -not -path "*/out/*" -not -path "*/.git/*"` or just scope to a specific subtree.
- **`ls`**: prefer a specific subdir over wide listings; pipe to `head` if you only need a sample.
- `memory/` and any branch-local `docs/planning/` doc contain domain vocabulary — grep them only when the question is about *planning state*, not when looking for code.

## There are no tests. Do not add any.

Every test file in this repo was deleted. A test cannot constrain an AI, because the AI edits
the test as freely as the code and at the same cost — the human effort-asymmetry the whole
mechanism rests on does not exist here. Measured on the branch that removed them: one refactor
churned 13,271 lines across 174 test files purely to keep them compiling through requested
changes, and caught zero production regressions. A test also cannot tell "the AI broke it" from
"David asked for this", so on every requested change its failure is cost, never signal.

**Verification is loud runtime failure plus driving the editor.** A failure that announces
itself in a live session reaches the human; a silent green check does not. That is the whole
difference — not tamper-proofness. **An AI can edit the loud failures too**, and the guards, and
this sentence. Their value is only that in the DEFAULT case they announce, so neglect makes them
fire instead of fall silent. Nothing here is a guarantee; the only real check is the human
noticing behavior.

Prefer, in order: an assertion that fires in the running system with a site tag
(`Categories/Node/check-panic-message.sh`); a guard whose failure state is loud and
whose allowlist is empty (`Categories/Node/check-no-network-locks.sh`); a `.probe`
breadcrumb (`.claude/rules/go-debugging.md`). Never a test.

## Workflow

- **MERGING WAITS FOR THE USER'S APPROVAL. Always. No exceptions, no categories.** Work on a task branch, verify it (`scripts/stop-checks.sh` output EMPTY), report it, and STOP. The user says when it merges to `main`. Green checks are permission to ASK, never permission to land.
  - **Why, and why the opposite rule used to be written here:** the old text said verified work merges "on the spot, without asking", justified by "a change on a task branch is INVISIBLE to the person driving the editor". That was true only under the worktree layout, where the editor was pinned to the main checkout. That layout is gone — there is ONE checkout, so `git checkout task/<name>` is what the editor runs and a branch is fully testable in place. Merging to make a change visible is now pointless AND harmful: it lands unreviewed work in the tool being used. In one session that shipped vanished edge beads, wrong animation colours, and a node that dragged backwards — each found by the user, in the editor, after the merge.
  - Do NOT reintroduce an exemption for "plumbing", "docs", "just tests", or "it's default-off so it can't affect anything". Every one of those is a category invented to merge without asking. There is no such category.
  - Sign-off is still required for genuinely destructive or shared-state actions: force-pushing `main` or any branch you do not own, rewriting published history, removing a dependency, deleting someone else's branch, and anything else called out in the system prompt's "Executing actions with care" section.
  - After merging, say so and name what to reload — a webview change needs the file reopened, an extension-host change needs "Developer: Reload Window" (`memory/feedback/ux/feedback_two_process_editor_reload.md`).
- Build and run before reporting a change as ready; verify output matches previous run. If verification fails, fix forward or revert — don't leave broken state on the branch. To exercise a TS change in the LIVE editor, run `npm run build` — the Stop hook does this automatically, but manual subagent verifications do not (why `tsc --noEmit` isn't enough is in the verify recipe below; stated once, there).
  - **Verify recipe (NEVER run the sim in the foreground):** the single source of truth is `bash scripts/stop-checks.sh` run from the repo root. **Read its STDOUT, not `$?` — it ALWAYS exits 0**, by design: it speaks the Stop-hook JSON protocol, so a failure is a `{"decision":"block","reason":...}` object printed to stdout while the exit code stays 0. Clean means *empty stdout*. Therefore `stop-checks.sh >/dev/null; echo $?` is not a check — it discards the only failure signal and reads a constant. Verify with `bash scripts/stop-checks.sh` and look at the output, or gate on it with `[ -z "$(bash scripts/stop-checks.sh 2>/dev/null)" ]`. It runs go build+test, tsc `--noEmit`, the npm webview build, staticcheck, eslint, vitest, and the full guard suite (incl. message-kind-parity, polar-only-nav, no-camera-roundtrip) — gated so the expensive per-language steps run only when that language changed or the branch is ahead of origin/main. Do NOT run the old `go build && go test`, `tsc`, `npm run build` steps separately; that just duplicates what stop-checks already does. Caveats it does NOT cover: the early-exit was removed so it also runs on a clean tree, but `tsc --noEmit` alone still won't refresh `out/webview.js` (stop-checks' npm build does), and an extension-host change still needs VS Code "Developer: Reload Window" (reopening a file only reloads the webview).
- One logical change per commit. Push each commit to the current task branch.
- **Git hooks live in `.githooks/` (tracked) and need one command per clone:** `git config core.hooksPath .githooks`. `.githooks/pre-push` runs `scripts/verify.sh` and aborts the push if it fails, so anything that commits WITHOUT the Claude Stop hook firing — a `git commit` you type, a subagent on a task branch, a session that ends another way — is still verified before it leaves the machine. This is local git, NOT GitHub Actions: no remote service, no minutes, works offline. `check-hooks-allowlist.sh` fails when `core.hooksPath` is unset, because a tracked hook git was never pointed at is inert while reading as coverage. Deliberate override: `git push --no-verify`. `.githooks/pre-commit` strips added comments from the staged files, so the comment policy costs one commit instead of two; it is a FIXUP, not a gate — it edits staged content and re-stages it, never rejects, touches only staged `.go`/`.ts`/`.tsx` (a whole-tree rewrite would edit other sessions' work in this shared checkout), and SKIPS a file with staged AND unstaged edits, naming it on stderr, since re-adding it would sweep the unstaged half into the commit. `scripts/strip-comments.sh [--apply] [paths...]` is the same by hand; override with `git commit --no-verify`.
- **Cost markers:** only record a `($N.NN)` cost marker on a commit (or bundle of commits) when the work was sized at **≥$5 expected** beforehand. Sub-$5 work lands without a marker. Bundle small commits into ≥$5 chunks for marker purposes. Pre-v0 sub-$5 markers stay as historical record but are no longer the convention.
- **Plain branches in the one checkout — start every change with `scripts/new-task.sh <short-kebab-name> "one-line description"`.** It creates the `task/<name>` branch from `origin/main`, checks it out, and sets the branch description `scripts/next.sh` reads. There is no worktree step: this repo tried one-worktree-per-change and dropped it — the extra tree cost more than it bought in practice, so plain `git checkout -b`-style branch switching in the single checkout is back to normal.
  - When done: merge to `main` (no sign-off — see the landing rule above), then `git branch -d task/<name>`. Deleting it is what removes the item from `scripts/next.sh`.
  - **Concurrent sessions share this one checkout — treat its state as other people's too.** Before switching branches, make sure your own work is committed (`git status`), since checking out a different branch changes what everyone in this tree sees. **Do not use `git stash`.** The stash stack is repo-global: an entry pushed by one session is visible and poppable by every other session working in this checkout, so two sessions stashing concurrently can pop each other's work, and `stash@{0}` means something different depending on who pushed last. Commit the WIP on your own branch instead (`scripts/wip.sh`) — it is private to your branch, survives a crash, and costs one `git reset --soft HEAD~1` to undo. This includes `git rebase --autostash`, which is the same global stack: commit first, then rebase with a clean tree. (Prose-only: git has no pre-stash hook, so there is nothing to enforce it with.)
  - Branches stay short-lived and merge to `main` quickly. Avoid long-lived feature branches like the v0 `visual-editor` pattern.
- Channel names encode which two nodes are connected — preserve this convention.
- **Medium vs. substance.** Before adopting a **medium** dependency (rendering library, framework, parser, bundler, file watcher, test runner, package manager, language/runtime version, editor integration), explicitly ask "what's the dominant choice the rest of the world converged on for this category?" and justify deviating if not adopting it. The medium is where industry has solved your problem; being weird there is pure overhead. Do **not** apply this heuristic to the **substance** of the system — the execution model, what a node is, how time/ticks work, what a bead is, how nodes coordinate, the Go network that runs nodes. Industry defaults there encode "logic in procedures, topology as plumbing," which is the inversion this project exists to challenge. For substance, ask "what does this system actually need?" and ignore industry — the whole point is that the answer is different. (Prior failure: the await/Promise execution model was the industry-correct JS translation of goroutines+channels, and it hid pacing inside the event loop, coupling nodes that should have been independent. Right answer for the medium, wrong answer for the substance.)

## Memory and doctrine layout

- `memory/` — one file per memory (`project_*`, `feedback_*`, `user_*`); `memory/MEMORY.md`
  is the index. Guard: `check-memory-hygiene.sh`.
- `.claude/rules/*.md` — region-specific doctrine, `paths:` frontmatter, loads on demand.
  Put detail here when it only matters for one part of the codebase.
- Root CLAUDE.md — cross-cutting invariants only. It is re-injected after `/compact`;
  rules are not.
- [docs/process/drift-checklist.md](docs/process/drift-checklist.md) — the periodic agent-health audit.

## Session handoff

Live task state is NOT stored in a hand-maintained doc (that kept drifting — it was a
cache of state whose authoritative source is git/memory/MODEL.md). It is DERIVED:
`task/*` branches carry a one-line `git config branch.<name>.description`, and
**`scripts/next.sh`** prints the live view (current branch, task branches, recent merges,
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

Visual editor reached v0. New work is friction-driven, not phase-driven; justify changes from real-world editor use. Working mode: user drives the editor and narrates; assistant makes changes. Friction arrives live, in conversation — what survives the session goes to `memory/` as a lesson, and what became code is already in git. Do not reintroduce a session log: the one that existed was write-only, never read to justify a change, and duplicated `git log` for everything except a handful of corrected measurements that now live in `memory/`.

**Plan docs are allowed, per change, in `docs/planning/`.** The old blanket "per-phase plans were deleted, git history is the archive" is gone: a change that reverses a documented invariant, or ripples across code and several pages at once, is worth writing down BEFORE it is made — what breaks, in what order, and how it is verified. Write one when the change is that shape; skip it when the change is a page edit or a rename.

A plan doc states intent, not live state: the target, the ripple list, the order, the verification, the risks. It must NOT become a status board — no "current step", no checkboxes updated as work lands, no summary of what is done. That is what `check-no-state-cache.sh` bans and what branch descriptions plus `scripts/next.sh` already answer. Delete the plan when the change lands; git history is the archive for the plan itself.
