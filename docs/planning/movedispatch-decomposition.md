---
branch: task/god-objects
---

# Decomposing MoveDispatch, then closing the holes the boundaries opened

Two changes, in this order. The second depends on the first: the hole exists because
`MoveDispatch` is unconstructible from outside, and the shape of the fix changes once
`MoveDispatch` is no longer one object.

This doc states intent — the target, what breaks, the order, the verification, the risks.
It is not a status board. Delete it when the work lands; git history is its archive.

## What MoveDispatch is, and is not

**It is not in the model.** `MoveDispatch` appears zero times in MODEL.md. The pinned
entities are bead, wire (`PacedWire`), node goroutine, input port, clock. `MoveDispatch`
holds the mover goroutines but is not itself a model entity — it is an implementation
artifact sitting beside the model. This work therefore does NOT require agreeing a change to
the pinned model. It must still not change ownership, goroutine structure, channel wiring, or
timing, because it holds the things that do.

**What it actually is:** the editor's command surface. Read the 51 exported methods as a set
— 13 `Toggle*`, 6 `*Viewpoint`, 4 `Load*`, 3 `Enable*`, 4 `Set*Streams`, plus `CreateNode`,
`DeleteNode`, `SelectScene`, `RootMove`, `HandleRawInput`, `SliderSpeed`. That is the
receiving end of the TS → Go bridge. An editor with ~50 commands has ~50 entry points; that
part is a genuine API, not accumulated mess.

`nodes/Wiring` is ~10k non-test lines because two different things are fused in it: **network
assembly** (build, loader, movers) and **the editor's command surface**. The seam between
those two is the real one.

## The measurement this plan rests on

Of the 88 `MoveDispatch` methods, counting only real owner FIELDS (`mr`, `ui`, `sw`,
`persist`, `lq`, `GS`, `RT`, `Scenes`, `inboxes`, `ctx`):

| owner fields touched | methods |
|---|---|
| **1** | **68** |
| 0 | 8 |
| 2 | 6 |
| 3 | 6 |

**68 of 88 touch exactly one owner** and rehome mechanically. Only **12 span two or more**.
Separately, **26 methods are one-line pure delegators** (`func (md *MoveDispatch)
ToggleHandholds(tr) { md.ui.ov.ToggleHandholds(tr) }`) — those do not move, they VANISH once
callers address the owner.

Where the 68 single-owner methods go:

```
43 ui       12 mr       3 lq       3 RT       3 GS       2 Scenes    1 sw    1 inboxes
```

**13 of the 26 delegators are GENERATED** (`overlay_gen.go`, emitted by
`tools/gen-node-defs` from `OVERLAY_FLAG_NAMES` in `messages.ts`). Deleting them is one
emitter change, not thirteen hand-edits. Do not hand-edit that file.

### Two earlier claims this measurement corrects

- A first pass at counting reported ~40 "multi-owner" methods. That count was wrong: its top
  pairing was `emitViewFrame,ui`, and `emitViewFrame` is a `MoveDispatch` METHOD, not an
  owner field. It was counting a method calling a sibling method as cross-owner coupling.
- The `MoveDispatch` method count moved 91 → 87 → 88 across this branch, and that trend was
  cited as evidence that decomposition merely relocates coupling. It is not evidence about
  THIS change: it measured the late package-lifts, where each new package boundary needed an
  accessor or constructor to reach past it. Different activity, different conclusion.

## The three import cycles — the same defect, three times

Three package extractions were attempted on this branch and each was reverted on a
compiler-confirmed cycle. **There should be no import cycles.** They are not a fact of the
domain; every one has the same cause, and it is the same cause as the 12 spanning methods:
**a would-be-leaf package holds a back-reference to the hub.**

Go forbids cycles, so these never exist in built code — they exist as latent cycles in the
dependency graph, and they surface the moment anything tries to leave `Wiring`. That is why
the directory cannot get smaller.

| would-be package | outbound edge (pkg → Wiring) | inbound edge (Wiring → pkg) |
|---|---|---|
| `portwiring` (`port_bindings.go`, `port_wiring.go`) | `PortBindings.md` is a `*MoveDispatch`, read for `pb.md.sw.interiorOuts`, `pb.md.RT.NodeRowFor` | `build_args.go` needs `PortBindings`/`PortSpec` |
| `beadcrud` (`touching_beads.go`, `bead_crud.go`) | `dragTouchingBeads` takes `*MoveDispatch` and `*nodeGeometry`, reads `nm.topo`, `nm.id`, `md.mr.edgeMovers` | `commit_node_move.go` calls `dragTouchingBeads`/`resolveBeadCrudMove` |
| `loadspec` (`topo_spec.go`, `validate.go`, `loader_tree.go`, `node_registry.go`, `builders.go`) | `NodeBuilder.Ports []PortSpec` and `Build(pb PortBindings)` need types from `port_bindings.go` | `build_args.go`'s `RegisterBuilder` writes `Registry[kind]` |

Read the outbound column: in all three the leaf reaches UP to the hub — for a channel, a row
number, or a struct type. None of them needs the hub itself; each needs a value the hub
happens to hold.

### The rule this plan enforces

**Dependencies point one way. A leaf takes what it needs as a parameter; it does not hold the
hub.** Concretely, for each cycle, replace the back-reference with the value:

- `PortBindings.md *MoveDispatch` → the interior-stream writer and the row-lookup it actually
  uses, passed at construction. `pb.md` disappears.
- `dragTouchingBeads(md, nm, ...)` → the bead/edge data it reads, passed as values. It does
  not need a mover, it needs that mover's topo and centres.
- `NodeBuilder` → once `PortSpec`/`PortBindings` no longer reach the hub, `loadspec` and the
  registry stop cycling on their own.

This is the SAME work as step 3's write-then-emit question, seen from the other side: both
are "a thing reaches the hub for one value it could have been given." Solve it once and the
12 spanning methods and the 3 cycles resolve together.

### Banned non-fixes

An interface, a `types`/`common` package, or an alias shim will each make the compiler stop
complaining while leaving the cycle in the design. A `geom_bridge.go` of type aliases was
built and deleted on this branch for exactly this. If a cycle cannot be broken by passing
values, the honest answer is that the code belongs together — record that and leave it.

### Verification for this part

- The three extractions above are the acceptance test: after the signature work, each must
  lift with `go build ./...` clean. If one still cycles, the compiler names the edge — record
  it verbatim rather than paraphrasing.
- `tools/network/check-dep-rules.sh` enforces the import graph; confirm it still has teeth
  after any package appears (its regex missed subpackage paths once already this session and
  went silently green — that fix is in, but re-verify).
- No new package may import `nodes/Wiring`. That is the invariant; state it as the check.

## Target

`MoveDispatch` becomes a struct of owners with no behaviour of its own beyond construction
and shutdown. Callers address the owner (`md.UI.PanViewpoint(...)`) rather than the hub.
The facade is deleted, not relocated. **No package under `nodes/Wiring/` imports
`nodes/Wiring`** — the dependency graph is acyclic by construction, not by the compiler
refusing it.

Expected: ~76 of 88 methods removed or cleanly rehomed — 26 deleted outright, ~50 moved to
the single owner they already touch. The 12 spanning methods are the real remainder and are
handled explicitly, below.

## The 12 spanning methods

These are the whole difficulty, so they are named rather than discovered later:

```
setHover              [RT, ui, + emitViewFrame]      SetEdgeStreams   [GS, mr, sw]
setSelectionUI        [ctx, mr, ui, + sendMove]      SetNodeStreams   [GS, mr, sw]
EnableViewpointPersist[persist, ui]                  Start            [ctx, mr]
EnableEditPersist     [Scenes, mr, persist]          sendMove         [ctx, mr]
CreateNode            [Scenes, ui, + 4 helpers]      sendTiltEdit     [ctx, inboxes]
DeleteNode            [RT, Scenes, ui]               emitViewFrame    [RT, sw, ui]
```

The dominant shape among them is **mutate state, then emit a view frame** — `PanViewpoint`
is `md.ui.vp.PanViewpoint(...)` followed by `md.emitViewFrame(...)`, and that pattern recurs
8 times. It is one repeated shape, not twelve separate problems. Decide the answer to it ONCE
(see step 3) and most of the 12 follow.

`Start`, `sendMove`, `sendTiltEdit` are `ctx` + a mover directory — process lifetime plus
delivery. Those stay with whatever remains of `MoveDispatch`; they are not commands.

## Order

Each step ends green and pushed. Do not start the next until the previous is committed.

**The cycles come first.** They are the enabling change: every later step is easier in an
acyclic graph, and doing the mechanical work first would churn call sites that the cycle work
then churns again. It is also the step most likely to teach us the plan is wrong — so it
should run while that is still cheap to act on.

**1. Break the three back-references** (the table above). Replace each hub reference with the
values it actually reads: `PortBindings.md` → the interior writer and row lookup it uses;
`dragTouchingBeads(md, nm, …)` → the topo and centres it reads; then `NodeBuilder` follows
once `PortSpec`/`PortBindings` are hub-free. No behaviour change — same values, passed
instead of reached for.

**2. Prove it: lift the three packages** (`portwiring`, `beadcrud`, `loadspec`). This is the
acceptance test for step 1, not new work — if they lift with `go build ./...` clean, the
cycles are genuinely gone. If one still cycles, the compiler names the edge; record it
verbatim and fix that edge before continuing.

**3. Close the holes the earlier work opened.** Do this BEFORE any further decomposition:
each is a back-reference or an escape hatch that a boundary pushed into the code, and leaving
them means later steps build on top of them. Full list and rationale in **Part 2** below —
it is no longer only the `ForTest` hatches. Two of the five were created by step 1 itself.

**4. Answer the write-then-emit question ONCE.** Same defect as steps 1–2 from the other
side: an owner mutates, then something must emit a view frame. Decide where the frame
boundary lives. This is the one design decision in the plan; everything after it is
mechanical. Do NOT answer it by giving owners a back-reference to `MoveDispatch` — that is
the cycle returning under a new name.

**5. Delete the 26 pure delegators.** Now safe and purely mechanical. Change the
`overlay_gen.go` emitter for the 13 generated ones and regenerate; hand-edit the rest.
Callers move to `md.ui.ov.X` / `md.ui.vp.X`. Takes 88 → ~62, touching no logic.

**6. Rehome the ~50 single-owner methods,** heaviest owner first: `ui` (43), then `mr` (12),
then the small tails (`lq`, `RT`, `GS`, `Scenes`, `sw`, `inboxes`). Mechanical.

**7. Reassess.** With the facade gone and the graph acyclic, re-ask whether the
assembly/command-surface split is now a real seam. Do not assume it is. If the export count
has gone UP, stop — see Risks.

`moverRegistry` is NOT a target at any step. It owns `nodeMover`/`edgeMover`/`nodeGeometry`,
the actors MODEL.md pins, and whatever remains of `MoveDispatch` stays with it.

## Ripple list

- **`runtopology/`** — 14 call sites across `edge_stream.go`, `node_stream.go`,
  `view_stream.go`, `scene_state.go`, `startup_report.go`, `topology_run.go`. This is the
  process startup sequence; every rename lands here.
- **`tools/gen-node-defs/overlay_gen.go`** — the emitter for 13 of the delegators. Change the
  emitter, regenerate, confirm no diff.
- **`nodes/Wiring/` internals** — 20 files hold the methods; ~40 more call them.
- **`pair_node_mover_absence_test.go`** (root package) holds a `*MoveDispatch` directly.
- **Docs naming these symbols**: MODEL.md, `.claude/rules/persistence-ownership.md`,
  `.claude/rules/bridge-surface.md`, `docs/pair-node/**`. `check-doc-drift` and
  `check-docs-symbols` catch stale paths — treat their failure as the checklist.
- **`check-composer-fields.sh`** inspects the composer struct specifically and WILL need
  updating. It must stay meaningful, not be loosened to pass.
- **Guards keyed to unexported names** stop matching silently once a symbol is exported.

## Verification

- `bash scripts/stop-checks.sh` after every commit. It ALWAYS exits 0 — clean means EMPTY
  stdout. Read the output, never `$?`.
- `go build ./...`, `go vet ./...`, `go test ./...` are authoritative. Editor diagnostics go
  stale after a move and are NOT evidence.
- `go run ./tools/gen-node-defs`, then `git status` shows no generated diff and
  `BUF_LAYOUT_FINGERPRINT` is byte-identical to `origin/main`.
- **Two progress measures.** Cycles: the three packages lift clean (step 2), and no package
  under `nodes/Wiring/` imports `nodes/Wiring`. Method count: 88 → ~62 after step 4, → ~12
  after step 5. If the method count rises, something is being added to reach past a new
  boundary — stop and look.
- **No behaviour change is the whole claim, and the tests here cannot prove it.** This repo's
  doctrine (docs/process/testing-shape.md) forbids cross-goroutine tests, so nothing asserts
  the movers still coordinate. The real check is driving the editor. Do it after step 1 and
  again after step 3, not only at the end.
- For any guard touched: make it fail once on purpose, restore, record the failure text.

## Risks

- **Step 1 is where this plan can be proven wrong, and that is why it runs first.** If a
  back-reference cannot be replaced by values — because the leaf genuinely drives the hub
  rather than reading from it — then that code belongs together and the package should not
  exist. Record which edge, and stop; do not force it. Failing at step 1 costs two steps of
  work instead of five.
- **Step 3 is a design decision, not code motion.** If it cannot be answered without giving
  owners a back-reference to the hub, the honest outcome is to stop after step 2 and say so.
  That is a good resting point, not a failure.
- **The export surface may grow.** 51 exported methods are what `runtopology` and the root
  package use. If deleting the facade means those packages must know more about `Wiring`'s
  internals, the change is not paying for itself — the same trap the late package-lifts fell
  into. Measure exports before and after.
- **Banned escapes**: no interface indirection, no `types`/`common` package, no alias shim.
  A `geom_bridge.go` of type aliases was built and deleted on this branch for exactly that.
- **`SetMsgTap` is a live test seam** on `MoveDispatch` (`move_streams.go`). It has no
  production caller and must not acquire one during this work.

---

# Part 2 — closing the holes the boundary work opened

This is step 3, immediately after the cycles. It ran late in the first draft, on the reasoning
that the fix depends on what `MoveDispatch` becomes. That was wrong: every later step builds
on these, so they get closed first.

There are five, and **two of them were created by step 1 itself** — which is the argument for
doing this now rather than at the end. Each has the same shape: a package boundary needed a
value from the other side, and instead of passing it, the code reached for it.

## Hole 1 — `currentBuildMD`, a package-level mutable hub reference

`nodes/Wiring/build_args.go` now holds `var currentBuildMD *MoveDispatch`, set once by
`buildNodes` and read by `build_args_lattice.go`, `build_args_tilt_vector.go` and
`build_args_selfdrive.go` for `md.ui` / `md.inboxes` / `md.mr`.

This was introduced to break Cycle A: `PortBindings` could not carry that state once it moved
to `portwiring`, so the reference moved to a package-level var instead. **The import cycle is
gone and the coupling is not — it is now global and invisible.** That is a worse position
than the struct field it replaced: a field back-reference is at least visible in the type.

It is also a shared mutable global in a codebase whose whole architecture is ownership plus
message passing with zero shared memory (`check-no-network-locks.sh`, empty allowlist). It is
safe only by the argument that the build phase is single-threaded — an argument, not an
enforcement. Nothing checks it. Two overlapping loads would race.

**Fix:** those three `BuildArgs` methods need specific values (`ui` lattice/tilt state,
`inboxes`, a mover lookup). Pass them into `BuildArgs` at construction like every other field
it carries. Then delete the var.

## Hole 2 — the `portwiring` re-export aliases

`nodes/Wiring/port_bindings.go` now contains:

```go
type PortDir = portwiring.PortDir
type PortSpec = portwiring.PortSpec
type PortBindings = portwiring.PortBindings
```

Alias shims are banned in this plan, and were banned in the brief that produced them — a
`geom_bridge.go` of exactly this shape was built and deleted earlier on this branch.

The constraint behind them is real: fourteen node-kind packages (`nodes/Time`, `nodes/pulse`,
`nodes/PairNode`, …) call `Wiring.PortSpec` / `Wiring.PortIn` / `Wiring.RegisterBuilder` as
their public construction API, and `.claude/rules/node-kinds.md` documents that surface. So
the aliases keep 14 packages compiling.

**Fix:** update those 14 call sites to name `portwiring` directly, then delete the aliases.
That is a mechanical rename across a documented API, and `.claude/rules/node-kinds.md` and
`nodes/SPEC-FORMAT.md` must be updated in the same commit. If the decision instead is that
`Wiring` should keep re-exporting the node-kind API deliberately, then say so in the rule file
and drop the ban — but do not leave it undecided, which is where it is now.

## Holes 3–5 — the `ForTest` hatches

`newMoveDispatch` builds the entire mover graph — every `nodeMover`, `edgeMover`, and their
dedicated channels. The only supported way to obtain one was `LoadTopology`, which goes
through spec parsing, validation, and a kind registry that only `package main` can populate
(via `kinds_generated.go`'s blank imports). The unexported constructor WAS the enforcement:
you could not build one without the front door.

`NewMoveDispatchForTest` (`move_dispatch_construct.go`) is a public function taking raw maps
that returns a live one. Any package can now construct a `MoveDispatch` over arbitrary
geometry, bypassing validation and the registry.

**Nothing enforces the naming.** A grep of every guard script finds no check that a `ForTest`
symbol has no production caller. Three exist:

- `Wiring.NewMoveDispatchForTest` (`move_dispatch_construct.go`)
- `Wiring.NewDrivenOutForTest` (`driven_out.go`)
- `wire.NewOutChanForTest` (`nodes/wire/out_port.go`)

`driven_out.go`'s own header documents that its unexported constructor IS the guarantee —
that a plain `Out` passed where a `DrivenOut` is required is a compile error. A public
`ForTest` sibling with nothing watching it is that guarantee on the honour system.

This repo machine-checks weaker things: a buffer column with no production consumer fails
`check-no-dead-buffer-column.sh`; `useSyncExternalStore` outside an allowlist fails
`check-no-webview-state.sh`; a hand-rolled scene path fails `check-scene-path-resolution.sh`.

## These are closed by DELETING them, not by guarding them

A guard is a detector. Adding one leaves the hole in place and posts a sentry next to it —
the code still permits the thing, and the repo grows another shell script. **Every hole below
is closed by removing the construct, so that the thing it permitted is no longer expressible.
No new guard is required for any of them.**

That is this repo's own preference: make the bug class unrepresentable rather than police it
(`memory/feedback/architecture/feedback_make_bug_class_unrepresentable.md`), and prefer code
structure over rules that describe the structure
(`memory/feedback/process/feedback_code_self_defends.md`).

All five are the same move: a boundary needed a value from the other side, and the code
reached across instead of being handed it. Holes 1 and 2 were created while *fixing* exactly
that defect — which is how easily it recurs, and why the fix has to be structural.

## Order — each step DELETES something

1. **Hole 1: delete `currentBuildMD`.** Pass the three values those `BuildArgs` methods need
   (`ui` lattice/tilt state, `inboxes`, a mover lookup) in at construction, like every other
   field `BuildArgs` already carries. Then the var is unreferenced and is removed. After this
   there is no package-level hub reference to reach for — the shape is gone, not watched.

2. **Hole 2: delete the three aliases.** Repoint the 14 node-kind packages to name
   `portwiring` directly and update `.claude/rules/node-kinds.md` and `nodes/SPEC-FORMAT.md`
   in the same commit. The alternative is equally structural: decide `Wiring` re-exports the
   node-kind API deliberately, write that in the rule file, and drop the ban. Either is fine;
   leaving it undecided is not.

3. **Holes 3–5: delete the hatches, don't guard them.** Each `ForTest` constructor exists for
   ONE reason — a test in an external package cannot reach an unexported constructor. So
   remove the reason:
   - `NewMoveDispatchForTest` exists because `stdinreader`'s test moved out of `Wiring`.
     Either move that test back into `Wiring` (it tests `Wiring` internals; per
     `docs/process/testing-shape.md` it asserts what one goroutine decided, and nothing
     requires it to live in the sibling package), or give `stdinreader` a seam that does not
     require constructing the whole mover graph.
   - `NewDrivenOutForTest` and `NewOutChanForTest` predate this work. Check each: if its
     caller can live in the same package as the type, the hatch deletes with it.
   - Only if a hatch genuinely cannot be removed does a guard become the fallback — and then
     it is a fallback, recorded as such, not the plan.

**If, and only if, a hatch survives step 3**, add
`tools/network/structure/check-fortest-has-no-production-caller.sh` (fails when a `*ForTest`
symbol is referenced from a non-`_test.go` file; empty allowlist). Then: confirm it passes on
the current tree, **make it fail on purpose** and record the text (a guard that has never
failed is indistinguishable from one that cannot), and confirm `stop-checks`' guard COUNT
rises by exactly one — its discovery glob is `tools/*/check-*.sh tools/*/*/check-*.sh`, and a
guard it does not find is silently absent while the suite still prints nothing.

## What actually prevents recurrence

Not a script. The reason holes 1 and 2 appeared is that a boundary was created before the
values crossing it were identified — the cycle got broken by relocating the reference rather
than removing it. The preventive is in step 1 of Part 1: *identify what the leaf reads, then
pass it*. If that is done first, there is nothing to guard.

The one thing worth considering as a guard is hole 1's SHAPE — a package-level `var` holding
a `*MoveDispatch` or any actor pointer. That is genuinely hard to make unrepresentable in Go,
nothing caught it, and it is a shared-mutable-global in a codebase whose architecture forbids
exactly that. Decide it after step 1, when it is clear whether the shape can recur.

## Verification

- `grep -rn "currentBuildMD" nodes/` returns nothing.
- `grep -rn "= portwiring\." nodes/Wiring/port_bindings.go` returns nothing, or the rule file
  documents the re-export as deliberate.
- `grep -rn "ForTest" nodes/ --include="*.go"` returns only definitions that survived step 3
  with a recorded reason.
- Net guard count unchanged, unless step 3 left a survivor — in which case +1 exactly.
- `bash scripts/stop-checks.sh` clean, read as empty stdout.

## Risk

The `ForTest` deletions change where tests live, and this repo's testing doctrine
(`docs/process/testing-shape.md`) constrains that: a test asserts what ONE goroutine decided,
and cross-goroutine tests are forbidden. Moving a test back into `Wiring` must not turn it
into a test of two goroutines communicating. If it would, the test stays external and its
hatch is the one that earns the fallback guard.

