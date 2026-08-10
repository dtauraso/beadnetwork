# Decomposing MoveDispatch, then closing the ForTest hole

Two changes, in this order. The second depends on the first: the hole exists because
`MoveDispatch` is unconstructible from outside, and the shape of the fix changes once
`MoveDispatch` is no longer one object.

This doc states intent — the target, what breaks, the order, the verification, the risks.
It is not a status board. Delete it when the work lands; git history is its archive.

## Why

`nodes/Wiring` is 133 files (72 non-test, ~9,966 lines, 280 functions) after twenty-five
packages were lifted out of it. It stays large because it is not a collection of separable
concerns. It is one object's surface:

| type | methods | spread over |
|---|---|---|
| **`MoveDispatch`** | **88** (51 exported, 37 not) | **20 files** |
| `BuildArgs` | 23 | 8 files |
| `nodeGeometry` | 11 | 7 files |
| `buildCtx` | 10 | 6 files |

Every remaining lift fails the same way, and it is always this: `port_bindings` cannot move
because `pb.md` is a `*MoveDispatch`; the loader/spec unit cannot move because `NodeBuilder`
needs `PortSpec` while `RegisterBuilder` needs `Registry` back. That is not twenty coupling
problems. It is one, seen from twenty angles.

**The lifts have started making it worse.** Tracked across the branch, `MoveDispatch` went
91 methods → 87 (facade delegators deleted) → 88. It is climbing again because each late lift
adds an accessor or constructor to reach past the new package boundary — `Viewpoint()` so an
external test could read what `md.ui.vp` gave it directly, `NewMoveDispatchForTest` so an
external test could build one at all. Package-lifting now relocates coupling into the export
list instead of reducing it. Stop lifting; decompose the object.

## What MoveDispatch already is

The STATE is already decomposed. Twelve named fields, each with an owner:

```
mr       moverRegistry                  gs/GS    geomseeds.GeomSeeds
persist  persisters                     sw       streamWiring
ui       uiState                        lq       layoutQuantizer
Scenes   sceneswitch.SceneSwitch        RT       rowtables.RowTables
inboxes  nodeInboxes                    tr       *T.Trace
ctx      context.Context                tapToInstall  func(...)
```

What was never decomposed is the FACADE. Most of the 88 methods forward to one of those
owners. An earlier pass on this branch deleted the delegators for three owners (`rowtables`,
`geomseeds`, `sceneswitch` lifted to packages) and stopped when the remaining four —
`persisters`, `streamWiring`, `uiState`, `layoutQuantizer` — each turned out to take
`*nodeGeometry`, `map[string]*edgeMover`, or `movemsg.Msg` in their signatures. Deleting a
delegator does not help when the PARAMETERS are the entanglement.

So the target is not "delete more delegators". It is: make each owner a type whose methods
take only what that owner owns, then let the owner move.

## Target

`MoveDispatch` becomes a struct of owners with no behaviour of its own beyond construction
and shutdown. Each owner is reachable directly by its callers. The 88-method facade is gone,
not relocated.

Concretely, callers reach `md.UI.SetHover(...)` rather than `md.SetHoverUI(...)`, and an
owner whose methods no longer name an actor type becomes a package.

Method distribution today, by file, is the work list:

```
14 move_dispatch_api.go     6 scene_structure.go      2 scene_switch.go
13 overlay_gen.go           6 move_streams.go         2 scene_speed_persist.go
10 gesture_actions.go       5 gesture_handlers.go     2 scene_lattice_persist.go
 7 viewpoint_state.go       4 distance_groups.go      2 move_persist.go
                            3 view_stream.go          2 gesture_hit.go
                            3 pair_node_self.go       1 × 4 more
                            3 gesture_graph.go
```

## Order

Each step ends green and pushed. Do not start the next until the previous is committed.

**1. Establish the actor boundary first.** The four blocked owners are blocked by
`*nodeGeometry`, `*edgeMover`, `*nodeMover` appearing in their signatures. Before touching
any owner, determine for each such signature whether the actor is passed because the owner
genuinely drives it, or only to read a value off it. Where it is only a read, change the
signature to take the value. This is the step that unblocks everything after it, and it is
the one that can fail — see Risks.

**2. `uiState`.** 7 methods in `viewpoint_state.go` plus the selection/hover surface.
Largest single win and the one external callers touch most (`SetViewpoint` 12 sites,
`PanViewpoint` 5, `OrbitLockedViewpoint` 4, `Viewpoint` 4).

**3. `persisters`.** Six debounced disk persisters, already grouped. Blocked today only by
`viewpoint`/`overlayState`/`sceneSphere` passed by value from `uiState` — so this follows
step 2 directly.

**4. `streamWiring`.** Writes into `edgeMover.streamOut` and `nodeGeometry.stream.*`. Needs
step 1 to have resolved whether those writes belong to the stream owner or to the actor.

**5. `layoutQuantizer`.** Field is a single bool; the logic around it is inseparable from
`nodeGeometry` today. Likely last, possibly never — see Risks.

**6. `moverRegistry`.** Owns `nodeMover`/`edgeMover`/`nodeGeometry` — the actors MODEL.md
pins. It is the root of the entanglement and is NOT a target. It stays, and whatever remains
of `MoveDispatch` stays with it.

## Ripple list

- **`runtopology/`** — 14 call sites across `edge_stream.go`, `node_stream.go`,
  `view_stream.go`, `scene_state.go`, `startup_report.go`, `topology_run.go`. This is the
  process's startup sequence; every rename lands here.
- **`tools/gen-node-defs/overlay_gen.go`** — generates `overlay_gen.go`'s 13 methods. If
  those methods move to an owner, THE GENERATOR EMITS THE NEW SHAPE, not a hand-edit.
  Regenerate and confirm no diff.
- **`nodes/Wiring/` internals** — 20 files hold the methods; ~40 more call them.
- **`pair_node_mover_absence_test.go`** (root package) holds a `*MoveDispatch` directly.
- **Docs naming these symbols**: MODEL.md, `.claude/rules/persistence-ownership.md`,
  `.claude/rules/bridge-surface.md`, `docs/pair-node/**`. `check-doc-drift` and
  `check-docs-symbols` will catch stale paths — treat their failure as the checklist.
- **Guards keyed to unexported names** stop matching silently once a symbol is exported.
  `check-composer-fields.sh` inspects the composer struct specifically and WILL need
  updating; it must stay meaningful, not be loosened to pass.

## Verification

- `bash scripts/stop-checks.sh` after every commit. It ALWAYS exits 0 — clean means EMPTY
  stdout. Read the output, never `$?`.
- `go build ./...`, `go vet ./...`, `go test ./...` are authoritative. Editor diagnostics go
  stale after a move and are NOT evidence.
- `go run ./tools/gen-node-defs` then `git status` shows no generated diff;
  `BUF_LAYOUT_FINGERPRINT` byte-identical to `origin/main`.
- **No behaviour change** is the whole claim, and the tests here cannot prove it — this
  repo's doctrine (docs/process/testing-shape.md) forbids cross-goroutine tests, so nothing
  asserts that the movers still coordinate. The real check is driving the editor. Do it at
  step 2 and step 4, not only at the end.
- For any guard touched: make it fail once on purpose, restore, and record the failure text.

## Risks

- **Step 1 may not be resolvable.** If an owner genuinely drives an actor rather than reading
  from it, its methods belong with the actor and the owner should not move. That is a real
  answer, not a failure — record it and stop, rather than forcing it with interface
  indirection, a `types`/`common` package, or an alias shim. All three are banned here; a
  `geom_bridge.go` of type aliases was built and deleted on this branch for exactly that.
- **The facade may be load-bearing for the external API.** 51 exported methods are the
  surface `runtopology` and the root package use. Removing it widens what those packages must
  know about `Wiring`'s internals. If the export count goes UP, the change is not paying for
  itself — that is the same trap the late package-lifts fell into.
- **This touches the pinned model.** `MoveDispatch` owns the mover goroutines. Ownership,
  goroutine structure, channel wiring and timing must not change. If a step requires changing
  them, it is no longer decomposition and needs agreeing separately.
- **`SetMsgTap` is a live test seam** on `MoveDispatch` (`move_streams.go`). It has no
  production caller and must not acquire one during this work.

---

# Part 2 — closing the ForTest hole

Do this AFTER the decomposition, because the fix depends on what `MoveDispatch` ends up
being. Not before: closing it first would make step 1 harder for no gain.

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
4. Only then consider whether any of the three hatches can be removed outright. If the
   decomposition made an owner constructible without the full registry, its `ForTest` hatch
   may no longer be needed at all — deleting it beats guarding it.

## Verification

- Guard count before vs after, from the suite's own discovery glob: +1 exactly.
- The deliberate-failure text, recorded.
- `bash scripts/stop-checks.sh` clean, read as empty stdout.

## Risk

The guard must distinguish a definition from a reference, or it will flag the `func
NewXForTest` line itself and be permanently red. It must also not fire on doc comments naming
the symbol. Both are pattern problems; get them wrong and the usual failure applies — a guard
that matches nothing reads exactly like a guard that passes.
