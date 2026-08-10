---
branch: task/god-objects
---

# What the god-objects branch has not fixed

The branch decomposed the god files and directories. 27 packages came out of `nodes/Wiring`,
no file anywhere is over 391 lines, and no directory outside `nodes/Wiring` is over ~31 files.
This is the remaining list. Delete this doc when it is empty.

## 1. `loadspec` never lifted — unfinished, not blocked

`topo_spec.go`, `validate.go`, `loader_tree.go`, `node_registry.go`, `builders.go` are still
in `nodes/Wiring`. Every edge cleared except one: `loader_tree.go` reads `positionFileJSON`
(a 10-field JSON DTO, `quant_offset_persist.go:86`) and `positionFilePath` (one
`filepath.Join`, `node_mover.go:37`).

Move both into `nodes/Wiring/positionfile`, update writer and reader, re-attempt the lift.
`check-persist-write-ownership.sh` and `check-scene-path-resolution.sh` police these paths and
may match by unexported name — they go silently green once the symbols are exported, so fix
and prove teeth.

It was previously recorded as blocked on the reasoning that `persistence-ownership.md`
co-locates a file's reader and writer to stop schema drift. That rule governs write
ownership, not where a type lives. A schema in its own package that both sides import
prevents drift better than proximity.

## 2. `MoveDispatch` is still 63 methods across 20 files (was 88)

Its state is already twelve named owners; the facade over them was never fully decomposed.
The 13 generated pure delegators are done: `overlayToggles` now binds
`(*overlayState).ToggleX` directly (`tools/gen-node-defs/overlay_gen.go`'s `writeOverlayGen`,
`nodes/Wiring/overlay_gen.go`), `stdin_dispatch.go`'s `"toggle"` handler calls
`fn(&md.ui.ov, tr)`, and the scenePoles/nodePoles debug breadcrumb (the one non-pure part of
those 13 — it needed `md.EmitBreadcrumb`, a `MoveDispatch` method) moved to that same call
site via two small generated tables, `overlayFlagBreadcrumbScope` and `overlayFlagValue`.
88 → 75 methods (the drop is exactly 13).

Round 2 deleted the 15 remaining one-line pure forwards to a single owner field: `Bind`,
`EdgeOut`, `centerOfNode`, `enqueueFor`, `finalizeActors` (→ `md.mr`); `RootMove`,
`commitNodeMoveLocal`, `heldCenters` (→ `md.lq`); `setHoverUI` (→ `md.ui`); `NodeSeeds`,
`EdgeSeeds`, `loadTimeCenters` (→ `md.GS`, already exported — external callers in
`runtopology` now call `md.GS.NodeSeedsFn()`/`md.GS.EdgeSeedsFn()` directly, no new export).
Callers now address the owner field directly (`md.mr.X`, `md.lq.X(md, ...)`, `md.ui.X`).
`EdgeOut`'s own body (`mr.edgeOutFor`) turned out to have zero live callers anywhere — it
was deleted too as dead code once the delegator was gone (staticcheck U1000 caught it).

Three of the 15 turned out NOT to be pure/single-owner and were excluded, kept as-is:
`SliderSpeed` (`ui`) computes `EffectiveClockSpeed(md.ui.speed, md.ui.clockDivisor)` — a
derived value from two fields via a package function, not a plain forward. `SetViewpoint`
and `Viewpoint` (`ui`) both have OUT-OF-PACKAGE callers that cannot reach the unexported
`md.ui.vp` field directly — `runtopology.loadSceneState` passes `md.SetViewpoint` as a func
value to `scenecamera.SeedInitialViewpoint`, and `nodes/Wiring/scenecamera`'s own tests call
`md.Viewpoint()`. Deleting either would force exporting `ui`/`vp`, which the branch's own
constraint rules out.

75 → 63 methods (drop of 12, `grep -h "^func ([a-z]* \*MoveDispatch)"
nodes/Wiring/*.go | grep -v _test | wc -l`). Exported-symbol count for package `Wiring`
dropped 119 → 114 (no new exports).

Round 3 rehomed the 8 more single-owner methods whose bodies called NOTHING but their one
owner field (plus params/locals/package funcs) and had no cross-package callers requiring
their delegator to keep growing exports: `NodeFromHit`/`EdgeFromHit` (→ `RT`, now
`rowtables.RowTables.NodeFromHit`/`EdgeFromHit`, deleted from `MoveDispatch`); `NodeKind`,
`nodeBodyRadius`, `linkRefusal`, `nearestNodeTo` (→ `mr`, deleted from `MoveDispatch`);
`dropPointFromNDC` (→ `ui`, deleted); `LoadLatticePoints` (→ `ui`, deleted).
`HasNodeMover`/`NodeSelfDriven`/`NodeQuantOffset` moved their bodies to `mr` too but kept
thin nil-free `MoveDispatch` delegators, since `pair_self_drive_persist_test.go`/
`pair_node_mover_absence_test.go` (package `main`, outside `nodes/Wiring`) call them and
`mr` is unexported — deleting the delegator would break those tests, not just grow exports.
`BroadcastLatticePoints` kept its `MoveDispatch` delegator for the same nil-`*MoveDispatch`
guard reason (defensive, never exercised) but moved its send loop to a new
`nodeInboxes.broadcastLatticePoints`. 63 → 55 methods (drop of 8); exported `MoveDispatch`
method count 33 → 31 (`NodeKind`/`nodeBodyRadius`... only `NodeKind` was exported, so -1
export there, and `nodeBodyRadius` was already unexported, -1 more from the 3 that kept
delegators net-zero — see the commits for the exact per-owner diff). Package `Wiring`
exported-symbol count unchanged, 158 → 158 (no new exports; `rowtables.RowTables` gained
two exported methods, a different package).

`SelectScene`/`EnableSceneSwitch` were checked and declined: `sceneswitch`'s own package
doc comment already records the reason (genuine orchestration referencing `scene.SceneTabs`
and persistence helpers, not a thin delegator). `ResolveSceneDistanceGroups`/`LoadOverlays`/
`LoadSpeed`/`SetViewpoint`/`EmitViewpoint`/`SetViewStream`/`EnableSceneSwitch` remain the 7
export-blocked methods (unchanged from the task that measured them) — several of these ALSO
call another `MoveDispatch` method (e.g. `LoadOverlays` calls `emitViewFrame`), so they are
doubly blocked, not just export-blocked.

Remaining, unstarted: most of the rest of the ~63-methods list touches its owner field
AND calls at least one other `MoveDispatch` method (`emitViewFrame`, `sendMove`,
`SetViewpoint`, etc.) — the *mutate state, then emit a view frame*/route-through-dispatch
shape — or has a cross-package caller through an unexported field, so it stays until the
write-then-emit question is answered once (see below). 12 span two or more owners
outright; the dominant shape among them is *mutate state, then emit a view frame*, 8 times.

### The write-then-emit answer: the owner mutates, the view-owner goroutine emits

Measured on the 55 remaining methods: **31 are blocked by calling another `MoveDispatch`
method**, not by ownership — 9 are multi-owner, and only 15 are still cleanly rehomable. So
this question gates the majority of what is left, not a tail of 12. (The original "68 of 88
rehome mechanically" figure counted owner FIELDS and ignored method calls. It measured the
wrong thing; every estimate built on it was too optimistic.)

What they call, and why one method dominates:

```
14 emitViewFrame   4 sendMove   2 each: seedOrbitPivot, refuseStructuralEdit,
                                        dragPlaneHit, distanceGroupMax, SetViewpoint
```

`emitViewFrame` is not a helper — it is the VIEW frame SERIALIZER. It reads nearly all of
`ui` (all 11 overlay flags, viewpoint, sceneSphere, speed, sceneKinds, sceneEditable,
editRefused, lastDraggedNode), plus `sw` (viewOut/viewBuildFrame/viewTick) and `RT`
(NodeRowFor). Of course it touches three owners: a frame is a snapshot of the whole view.

**Decision: split mutation from emission at the CALL SITE.** The owner method mutates and
nothing else; the caller — which is always the view-owner goroutine — emits the frame after
it. Owners never gain a way to emit, and never gain a back-reference.

This is not a new idea; **it is already proven in-tree on this branch.** The 13 generated
overlay toggles were converted to exactly this shape: `stdin_dispatch.go` now calls
`fn(&md.ui.ov, tr)` to flip the flag, then emits the frame itself. Thirteen methods, no
back-reference, no behaviour change.

It is also what the codebase's own ownership already says. The VIEW stream is single-owner —
`RunStdinReader` — and the source says so in several places ("the view-owner goroutine
(RunStdinReader) is the SOLE caller"). Every `emitViewFrame` call site is reached from that
one goroutine (gesture handlers, dispatch handlers, scene persisters). So moving the emit
from callee to caller does not change which goroutine writes the VIEW stream. **Verify that
per call site before moving it** — if any call site turns out to run on a mover goroutine,
that one is a genuine ownership change and stops, rather than being pushed through.

Order: apply the split to the 14 `emitViewFrame` callers first (largest and best-precedented),
then `sendMove`'s 4, then the rest. Do not answer this by giving owners a back-reference to
the hub — that is the cycle returning under a new name.

**The `emitViewFrame` pass is done.** Applied: `viewpoint_state.go`'s 5 (`EmitViewpoint`/
`OrbitViewpoint`/`OrbitLockedViewpoint`/`ZoomViewpoint`/`PanViewpoint` no longer call
`emitViewFrame` themselves — their callers in `gesture_handlers.go`/`gesture_actions.go` do,
after the mutation, matching the toggle precedent); `gesture_actions.go`'s `setHover` (now
returns `(events, changed)`, `updateHover` emits) and `applySelect` (its `emitSelectViewFrame`
delegator was deleted, replaced by a pure `selectViewEvent` builder `applySelect` itself calls
`emitViewFrame` on, at each of its 3 branches); `scene_structure.go`'s `refuseStructuralEdit`
(now mutates `md.ui.editRefused` only; each of its 12 `CreateNode`/`DeleteNode` call sites
emits its own frame); `distance_groups.go`'s `ApplyDistanceGroupTarget` (returns `moved`
without emitting; its one caller, `applyUpdateDistanceGroup` in `stdin_apply.go`, emits when
`moved`). Every split call site was confirmed reached from `RunStdinReader`'s own dispatch
(gesture handlers, `applyUpdate`'s tables) before moving it.

Held back, NOT split, for named reasons:
- `stdin_dispatch.go`'s 2 (`clockAttrHandlers["speed"]`, `overlayAttrHandlers["toggle"]`) —
  already at the target shape: each is an inline closure that mutates then emits directly,
  with no intermediate `MoveDispatch` method to split (`overlayAttrHandlers["toggle"]` is the
  literal precedent this whole pass follows).
- `view_stream.go`'s `EmitBreadcrumb` — no owner mutation exists in its body (it only sets
  fields on its `ev` parameter and forwards to `emitViewFrame`), so there is nothing to hoist
  to an owner type; the write-then-emit split does not apply to a method that never mutates.
- `scene_sphere_persist.go`'s `LoadSceneSphere`, `scene_speed_persist.go`'s `LoadSpeed`,
  `scene_overlays_persist.go`'s `LoadOverlays` — each has an out-of-package caller
  (`runtopology/scene_state.go`) that cannot reach the unexported `emitViewFrame` method
  directly; same export-block class as the 7 already-documented methods below (`LoadSpeed`/
  `LoadOverlays` are literally 2 of that 7).
- `gesture_graph.go` — its 1 listed mention was a doc comment describing a call already
  deleted in an earlier pass (`the local-polar drag-log reset (emitViewFrame(KindAbcDragReset)
  ...) that used to run here was deleted`); no live `emitViewFrame` call exists in the file.

Method/export counts unchanged by this pass (55 `MoveDispatch` methods total, 158 exported
`Wiring` symbols) — the split moves emit calls between existing methods and their callers, it
does not delete or add a `MoveDispatch` method (one exception nets to zero:
`emitSelectViewFrame` deleted, `selectViewEvent` added, both unexported).

`moverRegistry` is not a target. It owns `nodeMover`/`edgeMover`/`nodeGeometry`, the actors
MODEL.md pins, and whatever remains of `MoveDispatch` stays with it.

## 3. `buildCtx` is a 26-field shared mutable blackboard

11 build phases read and write it; no signature says what any phase reads or produces. Same
defect as `pb.md` and `currentBuildMD` — reaching for state instead of being handed it — and
the reason the build files cannot leave `Wiring`.

Fix: each phase takes what it reads and returns what it produces, so ordering is data flow
the compiler enforces. Do NOT do this with a registry: node kinds self-register because they
are an open set of independent things; the phases are a closed ordered sequence, and
registering them would need priority numbers or a dependency DAG — more machinery for eleven
known steps.

## 4. ~3½ build phases recompute what the movers already derive

One coordinate system: MODEL.md — *"a centre is a sum of polar vectors from ONE centre… every
node hangs off the scene centre directly, one hop."* `nodegeom.EdgeStepCount` is called from
`build_wires.go:45` at build and from `chain_beads.go:154` and
`beadcrud/touching_beads.go:95` at runtime — same function, same frame, same inputs.
`computeNodeGeometry`, `computeQuantizedLayout`, `computeReachRadii` are the same shape.

The structural phases are genuinely one-time (you cannot bind a channel that does not exist).
The derived-geometry ones need not exist: if a node derives its geometry from its position and
the scene sphere, as it already does at runtime, construction is the first update and those
phases delete rather than move. Worth more than item 3, and item 3 does not block it.

---

## The one rule

**Name the specific values the leaf reads from the hub.** If you can list them, it is not a
blocker — pass them and continue. Only if the leaf needs the hub *to drive it* is it real, and
then quote the compiler error verbatim.

`beadcrud` and `portwiring` were each declared genuine compiler-confirmed cycles by two
separate passes. Both lifted once someone asked that question. Every blocker on this branch
has been the same category, and it is indistinguishable from an impossibility at the compiler.

Be suspicious of a blocker justified by architectural prose citing a rule that does not cover
the case, offered in place of an error message. Prefer the compiler.

The same test catches a bad FIX, not just a bad blocker: when something is introduced as
unavoidable, name the alternative that was rejected and why. `currentBuildMD` was justified by
"`PortBindings` can no longer carry this" — true, and it says nothing about `BuildArgs`, which
was sitting next to it in the same package with ten fields already on it. `loadspec` was
justified by a persistence rule that governs a different thing. Both justifications were
locally true statements that did not reach the conclusion drawn from them.

## Constraints

- No interfaces, no `types`/`common` package, no alias shims, no dot-imports, no new
  package-level actor globals, no new `ForTest` hatches — enforced by
  `tools/network/structure/check-fortest-has-no-production-caller.sh`, empty allowlist. All
  of these make the compiler stop complaining while leaving the coupling in the design.
- No package under `nodes/Wiring/` may import `nodes/Wiring`. Verify across EVERY subpackage,
  not just the one being moved — scoping this check to the current change is how
  `stdinreader`/`scenecamera`'s two violations survived while the invariant was repeatedly
  reported as held. (Both are now fixed; the invariant holds across every subpackage.)
- No behaviour change: ownership, goroutine structure, channel wiring and timing are fixed.
- **The tests cannot prove no-behaviour-change.** `docs/process/testing-shape.md` forbids
  cross-goroutine tests, so nothing asserts the movers still coordinate. Drive the editor.
- Guards go silently green two ways: a non-recursive glob stops seeing moved files, and a
  guard keyed to an unexported name stops matching once exported. For any guard touched, make
  it fail once on purpose and record the text.
- If the exported surface grows, the change is not paying for itself — that is what happened
  to the late package-lifts, which is why this list stops at decomposition rather than
  continuing to lift.
