# Decomposing MoveDispatch, then closing the ForTest hole

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

**3. Answer the write-then-emit question ONCE.** Same defect as steps 1–2 from the other
side: an owner mutates, then something must emit a view frame. Decide where the frame
boundary lives. This is the one design decision in the plan; everything after it is
mechanical. Do NOT answer it by giving owners a back-reference to `MoveDispatch` — that is
the cycle returning under a new name.

**4. Delete the 26 pure delegators.** Now safe and purely mechanical. Change the
`overlay_gen.go` emitter for the 13 generated ones and regenerate; hand-edit the rest.
Callers move to `md.ui.ov.X` / `md.ui.vp.X`. Takes 88 → ~62, touching no logic.

**5. Rehome the ~50 single-owner methods,** heaviest owner first: `ui` (43), then `mr` (12),
then the small tails (`lq`, `RT`, `GS`, `Scenes`, `sw`, `inboxes`). Mechanical.

**6. Reassess.** With the facade gone and the graph acyclic, re-ask whether the
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

# Part 2 — closing the ForTest hole

Do this AFTER the decomposition, because the fix depends on what `MoveDispatch` ends up
being. Closing it first would make step 1 harder for no gain.

## The hole

`newMoveDispatch` builds the entire mover graph — every `nodeMover`, `edgeMover`, and their
dedicated channels. The only supported way to obtain one was `LoadTopology`, which goes
through spec parsing, validation, and a kind registry that only `package main` can populate
(via `kinds_generated.go`'s blank imports). The unexported constructor WAS the enforcement:
you could not build one without the front door.

`NewMoveDispatchForTest` (`move_dispatch_construct.go`) is a public function taking raw maps
that returns a live one. Any package can now construct a `MoveDispatch` over arbitrary
geometry, bypassing validation and the registry.

**Nothing enforces the naming.** A grep of every guard script finds no check that a `ForTest`
symbol has no production caller. Three exist today:

- `Wiring.NewMoveDispatchForTest` (`move_dispatch_construct.go`)
- `Wiring.NewDrivenOutForTest` (`driven_out.go`)
- `wire.NewOutChanForTest` (`nodes/wire/out_port.go`)

`driven_out.go`'s own header documents that its unexported constructor IS the guarantee —
that a plain `Out` passed where a `DrivenOut` is required is a compile error. A public
`ForTest` sibling with nothing watching it is that guarantee on the honour system.

This repo machine-checks weaker things: a buffer column with no production consumer fails
`check-no-dead-buffer-column.sh`; `useSyncExternalStore` outside an allowlist fails
`check-no-webview-state.sh`; a hand-rolled scene path fails `check-scene-path-resolution.sh`.

## Target

A guard — `tools/network/structure/check-fortest-has-no-production-caller.sh` — that fails
when any `*ForTest` symbol is referenced from a non-`_test.go` file. Empty allowlist, like
`check-no-network-locks.sh`.

## Order

1. Write the guard. Confirm it passes on the current tree (all three have test-only callers
   today — verify, do not assume).
2. **Make it fail on purpose**: add a production reference to one `ForTest` symbol, confirm
   the guard names it, remove it. A guard that has never failed is indistinguishable from one
   that cannot.
3. Register it in the suite. `scripts/stop-checks.sh` discovers guards via
   `tools/*/check-*.sh tools/*/*/check-*.sh` — confirm the suite's guard COUNT goes up by
   exactly one. A guard the glob does not find is silently absent and the suite still prints
   nothing.
4. Only then consider whether any hatch can be removed outright. If the decomposition made an
   owner constructible without the full registry, its `ForTest` hatch may not be needed at
   all — deleting it beats guarding it.

## Verification

- Guard count before vs after, from the suite's own discovery glob: +1 exactly.
- The deliberate-failure text, recorded.
- `bash scripts/stop-checks.sh` clean, read as empty stdout.

## Risk

The guard must distinguish a definition from a reference, or it will flag the `func
NewXForTest` line itself and be permanently red. It must also not fire on doc comments naming
the symbol. Both are pattern problems; get them wrong and the usual failure applies — a guard
that matches nothing reads exactly like a guard that passes.
