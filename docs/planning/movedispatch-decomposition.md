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

### `layoutQuantizer`'s five `md *MoveDispatch` parameters — pre-existing, now closed

Five `layoutQuantizer` methods (`RootMove`, `commitNodeMoveLocal`, `heldCenters`, `heldEdges`,
`broadcastToEdgesAndPartners`) took `*MoveDispatch` back as an explicit parameter — an owner
that had moved off the hub in round 2 but still reached back into it. This was PRE-EXISTING
from `ff726c63` ("extract layoutQuantizer owner from MoveDispatch"), not introduced by this
branch, and round 3's "23 stayed" list (reason 1, "calls something that itself needs
`*MoveDispatch` by signature") explicitly deferred revisiting it. Measured what each actually
read off `md`: none called a `MoveDispatch` method, only owner fields (`RootMove` → ctx, mr;
`commitNodeMoveLocal` → mr, ui; `heldCenters`/`heldEdges`/`broadcastToEdgesAndPartners` → mr).
Converted all five to take those owners directly (plus `dragTouchingBeads`, a package-level
helper `commitNodeMoveLocal` called that only read `mr`, needed for the conversion to compile
without smuggling `md` back in). Every call site was in-package (`nodes/Wiring`, including
tests) — `RootMove` is exported but its receiver field (`md.lq`) is unexported, so no outside
package could reach it regardless of the method's own export status.

This unblocked 3 of the 5 methods round 3 held back for reason 1: `applyNodeDragTarget`,
`ApplyDistanceGroupTarget`, `beginSphereRotation` converted to package-level functions
(`ui`/`mr`/`lq`/`ctx`, no `MoveDispatch` method calls left in their bodies).
`ApplyDistanceGroupTarget`'s helper `waitForCenterSettle` converted alongside it (read only
`mr`). `gestHome` and `gestWheel` stay blocked — both still call `md.SetViewpoint`/
`md.EmitViewpoint`/`md.PanViewpoint`/`md.emitViewFrame` (reason 2, not reason 1).

`MoveDispatch` method count: 40 → 37 (drop of 3). Exported `Wiring`-package symbol count
unchanged, 162 → 162 (`ApplyDistanceGroupTarget` was already counted via the general
`^func [A-Z]` pattern only when unbound; as a method it never matched that grep, so
unexporting it as `applyDistanceGroupTarget` doesn't move the count). No new interfaces,
`types`/`common` package, or `ForTest` hatch. The per-subpackage no-imports-`Wiring`
invariant holds (empty loop output). No guard is keyed to these five method names —
`check-composer-fields.sh` (the one guard textually matching `layoutQuantizer`) checks
struct field counts, not method signatures, and was re-run clean, unmodified.

## 2. `MoveDispatch` is still 48 methods across ~19 files (was 88)

Round 4 re-measured the 55 remaining methods for the mechanical criterion (touches at most
one owner field, calls no other `MoveDispatch` method) rather than trusting the prior
estimate: 22 looked eligible by field count alone, but only 7 survived once cross-package
callers/func-values were checked one at a time (the same audit trap the task brief warned
about — a `grep -v` filter without a leading `./` had silently matched nothing earlier in
this pass and had to be re-run). Moved: `SliderSpeed`, `dragPlaneHit`,
`refuseStructuralEdit`, `OrbitViewpoint`, `OrbitLockedViewpoint`, `ZoomViewpoint` (all →
`uiState`, `scene_speed_persist.go`/`gesture_actions.go`/`scene_structure.go`/
`viewpoint_state.go`); `selectViewEvent` → `rowtables.RowTables.SelectViewEvent`
(`row_tables.go` — this one needed a new import of `Trace`/`nodes/wire` into the `rowtables`
package, no cycle, no new export since `RowTables`/`NodeRowFor` were already exported).
`refuseStructuralEdit` moving required updating `check-refusal-emits-frame.sh`'s definition-
skip regex from a receiver-specific pattern (`\(md \*MoveDispatch\)`) to a generic one
(`\([a-zA-Z_]+ \*\w+\)`) — confirmed with a deliberate break (deleting one call site's
follow-up emit) both before and after the regex change. 55 → 48 methods (drop of 7).
Exported `Wiring`-package symbol count unchanged (160 → 160, verified against `git stash`).

Held back and reconfirmed, not re-litigated: the 7 export-blocked methods
(`ResolveSceneDistanceGroups`, `LoadOverlays`, `LoadSpeed`, `SetViewpoint`, `EmitViewpoint`,
`SetViewStream`, `EnableSceneSwitch`) plus `HasNodeMover`/`NodeSelfDriven`/`NodeQuantOffset`
(external `mr` callers) — `Viewpoint()` and `PanViewpoint` join this set this round: both are
called as func values or directly from `nodes/Wiring/scenecamera`'s external test package.
`SelectScene` stays declined per its package doc comment (genuine orchestration). Everything
else among the 55 touches 2+ owners or calls another `MoveDispatch` method (mostly
`emitViewFrame`/`sendMove`, the gesture/dispatch entry points) and is not a rehome target —
see the "write-then-emit" section below for why those stay.

## 2a. Prior state (round 3): `MoveDispatch` was 63 methods across 20 files (was 88)

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

### The floor is 14, not 48 — "they are entry points" was not a reason

At 48 methods the remainder was described as "command entry points that legitimately stay".
That was an unexamined stopping point, not a constraint. Measured:

| | count |
|---|---|
| owner fields touched: 1 | 24 |
| owner fields touched: 2 | 12 |
| owner fields touched: 3 | 8 |
| owner fields touched: 0 | 4 |

A free function taking two or three owners is unremarkable, so "it is an entry point" is a
preference for method syntax over passing parameters — the same shape as every other blocker
this branch has dissolved. The real constraint is WHO CALLS IT:

| | count | |
|---|---|---|
| production callers outside `nodes/Wiring` | **14** | the genuine external API; the hub is the natural receiver because `runtopology` holds only `md` |
| **test-only** callers outside (prod = 0) | **3** | `NodeSelfDriven`, `HasNodeMover`, `NodeQuantOffset` |
| in-package only | **31** | could be free functions taking the owners they read |

So `MoveDispatch`'s honest floor is **14 methods** — the external API — and everything else is
open. The 31 in-package ones are the next tranche: make each a package-level func taking the
owners it reads (`func gestHome(lq *layoutQuantizer, mr *moverRegistry, ui *uiState, …)`),
which is this doc's own rule applied one level up.

The 3 test-only ones are the same class as the `ForTest` hatches already closed: a public
method existing solely because a test in `package main` cannot reach an unexported field.
Close them the same way — remove the reason (move the test, or give it a real entry point),
not by keeping the delegator.

**Re-investigated and confirmed genuinely blocked, not by default.** `HasNodeMover`/
`NodeSelfDriven`/`NodeQuantOffset` are exercised only by `pair_self_drive_persist_test.go`
and `pair_node_mover_absence_test.go`, both `package main`, both needing a REAL `PairNode`
kind registered to build the topology they drive — `PairNode` lives in `nodes/PairNode`,
which imports `nodes/Wiring` to call `RegisterBuilder` in its own `init()`. Two escape
routes were tried and both fail at the compiler, not by inconvenience:
- Moving the assertions into an internal `nodes/Wiring` test file (`package Wiring`, which
  can reach `md.mr` directly) so the accessor is unnecessary: fails to compile —
  `go vet` on a probe file gives `imports .../nodes/PairNode from probe_test.go / imports
  .../nodes/Wiring from node.go: import cycle not allowed in test`. Go's internal-test
  exception (test files may import a package that itself imports the package under test)
  does not cover importing a package that imports the CURRENT package while also being
  compiled as part of it.
- Moving them into an external test package under `nodes/Wiring` (`package Wiring_test`)
  sidesteps the cycle but gains nothing: an external package has exactly the same access
  as `package main` — no unexported field, so the accessor is still required.

So all three stay, unchanged, for the same reason recorded when they were first kept: a
real cross-package test (package main, the only place `PairNode`+`Wiring` are both live)
needs to observe `moverRegistry`'s unexported per-node facts, and there is no route to that
which does not either fail to compile or still need an exported accessor. No code change.

### The 31 in-package-only conversion pass — 8 converted, 23 stayed, and one correction

Re-verified the 31 one at a time (owner fields read, and every downstream call). 8 became
package-level functions, in this order/grouping (each its own commit):
`distanceGroupMax`/`DistanceGroupLens` (2 owners: `uiState`, `moverRegistry` —
`distance_groups.go`), `sendMove`/`sendTiltEdit` (`moverRegistry`+`context.Context`,
`nodeInboxes`+`context.Context` — `move_dispatch_api.go`), `setSelectionUI` (`uiState`,
`moverRegistry`, `context.Context`, calling the now-free `sendMove` directly instead of
threading a func value — `move_dispatch_api.go`), `commitDragStart` (`uiState`,
`moverRegistry`, `context.Context` — `gesture_graph.go`, wired into `commitEdges` via a
closure since the table's `action` field still needs the `func(md *MoveDispatch, ...)`
shape the other two entries use), `SelectScene` (`sceneswitch.SceneSwitch` alone —
`scene_switch.go`), `setHover` (`uiState`, `moverRegistry`, `rowtables.RowTables`,
`context.Context` — `gesture_actions.go`). `DistanceGroupLens`/`SelectScene` keep their
existing exported names (they were already exported methods); everything else stays
unexported.

**Every one of the remaining 23 is genuinely blocked, not merely inconvenient** — three
distinct reasons, none of them "it is an entry point":

1. **Calls something that itself needs `*MoveDispatch` by signature, and that something is
   out of THIS task's scope to change.** `layoutQuantizer.RootMove(md *MoveDispatch, ...)`
   and `layoutQuantizer.heldCenters(md *MoveDispatch)` were converted to free methods in an
   *earlier* round that kept `md` as an explicit parameter (not a receiver) precisely
   because their own bodies call `md.sendMove`/reach several owners — that choice predates
   this task and isn't being revisited here. Anything that calls either still needs to hand
   over `*MoveDispatch`, which is the same back-reference this task exists to remove, just
   one hop later: `applyNodeDragTarget`, `ApplyDistanceGroupTarget`, `beginSphereRotation`,
   `gestHome`, `gestWheel`.
2. **Calls `emitViewFrame` or `SetViewpoint`, and those stay.** `emitViewFrame` was measured
   at 4 owners (`sw`, `ui`, `RT`, plus `mr` transitively via `DistanceGroupLens`) — the
   "four or more is a signal" line in this doc's own rule, and the doc's own prose
   ("a frame is a snapshot of the whole view") says why: it is not a leaf, it is the VIEW
   serializer. `SetViewpoint` is one of the 14 held back for a genuine external caller.
   Every one of `updateHover`, `applyOrbit`, `applyOrbitLocked`, `applySelect`,
   `seedOrbitPivot`, `commitHandholdStart`, `commitRotateStart`, `CreateNode` (13 call
   sites), `DeleteNode` (4 call sites) calls one of these two directly, and `HandleRawInput`/
   `gestPointerDown`/`gestPointerMove`/`gestPointerUp` reach one transitively through the
   `rawInputHandlers`/`hitClassifiers`/`commitEdges`/`applyAction` dispatch tables, whose
   entries are typed `func(md *MoveDispatch, ...)` because the handlers they hold need `md`
   for reason 1 or 2. `emitViewFrame`, `Viewpoint`, `PanViewpoint` themselves were also part
   of the 31; see below for why `Viewpoint`/`PanViewpoint` are a MEASUREMENT correction
   rather than a real "stayed" case.
3. **Two field-identity reasons, not owner count.** `SetMsgTap` writes
   `md.tapToInstall`, a field that lives directly on `MoveDispatch` (not inside any named
   owner substruct) — converting it would need a field-pointer parameter for one
   test-only seam, for marginal benefit; left as is. `BroadcastLatticePoints` guards
   `if md == nil { return }` — that check is ON THE RECEIVER ITSELF; a converted
   function's `*nodeInboxes` parameter (`&md.inboxes`) is never nil even when `md` is, so
   preserving this guard EXACTLY (this doc's own rule) requires keeping `md` as the
   receiver, not just threading a pointer through.

**One correction to the 31's own measurement, found by re-verifying rather than trusting
it.** `Viewpoint` and `PanViewpoint` are counted as "in-package-only" only because the
counting grep (`grep -vE "^(\./)?nodes/Wiring/"`) treats every file under
`nodes/Wiring/<subpackage>/` as "inside" `nodes/Wiring` — but `nodes/Wiring/scenecamera`
is a DIFFERENT Go package (`scenecamera_test`, an external test package), and its
`scene_camera_test.go` calls `md.Viewpoint()` / `md.PanViewpoint(...)` as exported methods
on `*MoveDispatch`. Converting either to an unexported free function would not compile
that test. This is the exact "grep -v without a leading ./" trap this doc names elsewhere
in a different spot — caught here by trying to convert and checking the caller, not by
trusting the directory-prefix count.

**Result:** `MoveDispatch` method count 48 → 40 (drop of 8, `grep -h "^func ([a-z]*
\*MoveDispatch)" nodes/Wiring/*.go | grep -v _test | wc -l`). The naive
`grep -hE "^(func|type|var|const) [A-Z]"` exported-symbol count reads 160 → 162, but that
+2 is `DistanceGroupLens`/`SelectScene` becoming textually visible to a `^func [A-Z]`
pattern that never matched a method declaration (`func (md *MoveDispatch) X`) in the first
place — both names were ALREADY exported and reachable cross-package as methods before this
pass. As free functions they now take `*uiState`/`*moverRegistry`/`*sceneswitch.SceneSwitch`
— all unexported-package-internal or already-exported types — so nothing outside
`nodes/Wiring` gained a new way to reach them; if anything the real cross-package-callable
surface shrank, since a package holding only `*MoveDispatch` can no longer call
`md.DistanceGroupLens()`/`md.SelectScene()` (neither had an outside caller, so nothing
broke). No new interfaces, `types`/`common` package, or `ForTest` hatch were added. The
per-subpackage no-imports-`Wiring` invariant holds (empty loop output, re-verified after
every commit). Guards touched:
`tools/network/structure/check-refusal-emits-frame.sh` (still finds all 13
`refuseStructuralEdit`/`emitViewFrame` pairs, untouched by this pass since `CreateNode`/
`DeleteNode` were not converted) and gofmt (one reformat, `gesture_graph.go`, own commit).

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

`scene_structure.go`'s 13 `refuseStructuralEdit`/`emitViewFrame` call-site pairs (not 12 —
the true count, `refuseStructuralEdit(` calls excluding its own definition) are now covered
two ways: `nodes/Wiring/refuse_structural_edit_emit_test.go` drives `CreateNode`/`DeleteNode`
down their cheapest refusal branch (`!md.ui.sceneEditable`) and asserts both the
`ui.editRefused` counter and a captured VIEW frame; `tools/network/structure/check-refusal-emits-frame.sh`
statically checks every call site, not just the two the test reaches. Both were confirmed to
fail when an emit was deleted, and both are clean on the current tree.

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

**Measured per-phase field touch counts** (10 phase methods on `buildCtx`):

```
computeReachRadii          reads 2   writes 1
allocateVectorChannels     reads 3   writes 2
bindDispatch                reads 3   writes 0
buildTypeMaps               reads 3   writes 2
computeNodeGeometry          reads 4   writes 2
computeQuantizedLayout       reads 5   writes 3
buildEdgeMaps                reads 6   writes 4
--- wide, NOT converted ---
allocateWires                reads 8   writes 6
buildMoveDispatch             reads 11  writes 1
buildNodes                    reads 18  writes 10
```

**Converted the seven narrow phases to package-level functions.** Each now takes what it
reads and returns what it produces; `buildFromSpec` threads the values explicitly:

```go
func computeNodeGeometry(spec loadspec.TopoSpec, sphere geom.SceneSphere) (map[string]nodegeom.NodeGeom, map[string]vec3)
func computeQuantizedLayout(spec loadspec.TopoSpec, sphere geom.SceneSphere, centers map[string]vec3, nodeGeoms map[string]nodegeom.NodeGeom) map[string]quantoffset.QuantizedOffset
func computeReachRadii(spec loadspec.TopoSpec, nodeGeoms map[string]nodegeom.NodeGeom)
func allocateVectorChannels(spec loadspec.TopoSpec) (vectorOutByNode, vectorInByNode map[string]chan tiltvector.TiltVectorMsg)
func buildTypeMaps(spec loadspec.TopoSpec) (nodeType map[string]string, kindBroadcastPorts map[string]map[string]bool)
func buildEdgeMaps(spec loadspec.TopoSpec, nodeType map[string]string, kindBroadcastPorts map[string]map[string]bool) (inbound map[string]map[string]string, outbound map[string]map[string][]string, outboundHandle map[string]map[string][]string)
func bindDispatch(md *MoveDispatch, outSink map[string]*wire.Out, destWire map[string]*wire.PacedWire)
```

`computeQuantizedLayout`/`computeReachRadii` still MUTATE the `centers`/`nodeGeoms` maps
handed to them in place — Go maps are reference types, so that mutation is visible to the
caller through the same map identity `buildFromSpec` already threads, with no behaviour
change and no extra return needed for those two.

**Left as methods, on purpose:** `allocateWires` (8/6), `buildMoveDispatch` (11/1), and
`buildNodes` (18/10) — a free function with 8, 11, or 18 parameters is worse than the
blackboard it would replace, and `buildNodes` in particular is a hub in its own right
(it reads nearly everything earlier phases produced to actually construct each node). These
three still need `*buildCtx`, so `buildCtx` itself is not removed by this pass — only its
seven narrowest phases stopped depending on it.

Verification: `go build ./...`, `go vet ./...`, `go test -count=1 ./...` all clean;
`go run ./tools/gen-node-defs` produces no generated diff; the per-subpackage
no-imports-`Wiring` invariant holds (empty loop output); exported `Wiring`-package symbol
count unchanged (163 → 163). No guard is keyed to any of the seven phase names.

**Coverage gap found, not assumed away.** Deliberately dropping the call to
`computeQuantizedLayout`, `computeReachRadii`, or swapping `allocateVectorChannels`'s two
return values did NOT fail `go test ./...` — the existing suite exercises `buildFromSpec`
only through fixtures small/simple enough that the resulting zero-value/swapped state
happens not to diverge from what the assertions check. This is the same limitation this
doc's own Constraints section already names ("the tests cannot prove no-behaviour-change" —
`docs/process/testing-shape.md`): these three phases' correctness on this pass rests on manual
review of the diff plus `go build`/`go vet` catching a signature mismatch, not on a red test.
Each injected bug was restored immediately after observing the (lack of) failure.

**Closed.** `nodes/Wiring/build_load_derive_test.go` and
`nodes/Wiring/vector_channel_threading_test.go` now assert what `LoadTopology` actually
produces for all three: `TestLoadTopologyComputesReachRadii` reads a load-time node's own
`ReachR` off `md.mr.nodeGeoms`; `TestLoadTopologyComputesQuantizedOffsets` reads its
`quantOffset` and derives it back to (approximately) the node's own scene-polar center;
`TestAllocateVectorChannelsKeysSourceOutTargetIn` pins `allocateVectorChannels`' own
source→Out/target→In contract; `TestPairNodeVectorChannelsThreadSourceOutTargetIn` (external
`Wiring_test` package, so it can import PairNode) follows that threading all the way through
`buildFromSpec`/`BuildArgs`/each PairNode's own build func by reading the built node's
`vec.VectorOut`/`vec.VectorIn` via reflection. Each test was independently confirmed to fail
under its own injection (dropped `computeReachRadii` call, dropped `computeQuantizedLayout`
call, swapped `allocateVectorChannels` return values at the `buildFromSpec` call site) and to
pass again once restored.

## 4. WITHDRAWN — the build phases do not duplicate the movers, they initialise them

This item claimed ~3½ build phases "recompute what the movers already derive" and should be
deleted, on the evidence that `nodegeom.EdgeStepCount` is called from `build_wires.go:45` at
build and from `chain_beads.go:154`/`beadcrud/touching_beads.go:95` at runtime, and that
`reachRFromPolar` is called from both `computeReachRadii` and `commitNodeMoveLocal`.

**Same function called twice is not the same as redundant.** Checked: the build phases produce
the INITIAL geometry and seed it into the movers — `newMoveDispatch(geoms, …)` takes exactly
what `computeNodeGeometry`/`computeReachRadii` computed. The movers then MAINTAIN it: a
mover recomputes reach when a node moves. No mover derives its own initial geometry, and none
reads the spec — the spec is on disk and only the loader opens it. So deleting
`computeReachRadii` leaves every node at `ReachR = 0` until its first drag, and deleting
`computeNodeGeometry` leaves the movers with nothing to be seeded with.

Initialisation and maintenance calling the same pure function is correct, not duplication —
it is the same frame producing the same answer at two different times, which is what a single
coordinate system is FOR.

The claim was built from a grep showing one function in two places, turned into an
architectural conclusion without testing whether either call could actually be removed. That
is the same shape as the `loadspec` and `currentBuildMD` rationalisations recorded above: a
locally true observation that does not reach the conclusion drawn from it. Left here as the
record; nothing to do.

---

## 5. `uiState` lift probed and declined — `gestureState` is entangled with entry points that stay on `MoveDispatch`

Checked whether `uiState` (`ui_state.go`, 14 fields incl. `vp viewpointState`,
`ov overlayState`, `gest gestureState`) could lift into `nodes/Wiring/uistate`, matching the
pattern already established by `GS geomseeds.GeomSeeds`/`Scenes sceneswitch.SceneSwitch`/
`RT rowtables.RowTables` — export follows the type moving, not the other way round.

**Declined without writing any code, by measurement alone (revert condition 3 from the task
brief, confirmed rather than assumed).** `gestureState`'s 18 unexported fields (`phase`,
`dragNode`, `dragGrabOffset`, `dragStartCenter`, `rotPivot`, `rotCx`, `rotCy`, `rotPxPerRad`,
`fov`, `rect`, `button`, `downX`, `downY`, `prevX`, `prevY`, `emptyDown`, `handholdDown`,
`secondary`, plus the method `reset`) are read and written directly by field, not through any
`uiState`/`gestureState` method, from `gesture_dispatch.go`, `gesture_handlers.go`,
`gesture_graph.go`, `gesture_hitclassify.go` — every one of whose handler functions is typed
`func(md *MoveDispatch, g *gestureState, ...)`. These are exactly the gesture ENTRY POINTS
this doc's own round 4 already confirmed stay on `MoveDispatch`
(`gestPointerDown/Move/Up`, `HandleRawInput`, `gestHome`, `gestWheel` — "reason 1: calls
something that itself needs `*MoveDispatch` by signature"). Moving `uiState` would either drag
all four of those files (and everything downstream of them, transitively `emitViewFrame`'s 14
callers) into the new package — reintroducing the `*MoveDispatch` back-reference under a new
name — or require exporting `gestureState` and all 18 of its fields so the entry points left
behind in `Wiring` could keep reaching into it by field. Either way this is not "export
`uiState`'s own surface" (the task's stated acceptable cost, matched by `GS`/`Scenes`/`RT`):
it is exporting the FSM's entire private bookkeeping to satisfy callers that were never going
to move, which the task explicitly named as a revert condition ("`gestureState` turns out to
be entangled with the gesture entry points ... in a way that would drag those into the new
package or require a back-reference").

`viewpointState`/`overlayState` were not separately measured once `gestureState` alone was
sufficient to decline — `gestureState` embeds inside `uiState`, so `uiState` cannot lift
without it regardless of the other two owners' own entanglement.

No file was edited; `git status --short` was empty throughout. `MoveDispatch`'s method/export
counts are unchanged from round 4 (37 methods; the 7 export-blocked methods
`ResolveSceneDistanceGroups`/`LoadOverlays`/`LoadSpeed`/`SetViewpoint`/`EmitViewpoint`/
`SetViewStream`/`EnableSceneSwitch` remain blocked, for the same reason as before — `md.ui`
staying unexported).

**Pattern recorded for the next attempt at this class of lift:** exporting an owner field is a
consequence of ITS TYPE moving to its own package (`GS`/`Scenes`/`RT`'s precedent) — but that
is only cheap when the type's fields are reached exclusively through its own methods. When a
type's fields are reached directly, by field, from entry-point functions elsewhere in the same
package (the gesture FSM's shape here), the type cannot lift alone; the entry points have to
move with it or the whole thing stays. Measure field-level cross-file access before proposing
a lift like this one, the same way `RootMove`/`heldCenters`/`applyNodeDragTarget` etc. were
measured for method-level calls in round 3.

## 6. Gesture-cluster lift (gesture-actor.md step 3) — probed and declined, then the decline itself corrected

**Correction (superseding the decline below without deleting the record).** The decline was
built on the four cluster functions' SIGNATURES (`beginSphereRotation`/`applyNodeDragTarget`/
`commitDragStart`/`setHover` take `*moverRegistry`/`*layoutQuantizer`), not on their BODIES.
Measured: none of the four USES `mr`/`lq` for anything except passing them on to exactly
three operations — `sendMove(mr, ctx, id, msg)`, `heldCenters(mr)` (via `lq`),
`RootMove(ctx, mr, id, target)` (via `lq`). That is pure pass-through plumbing, the same
shape `nodeGeometry` already solves with a bound func value
(`ng.msg.sendMove = md.mr.enqueueFor(ng)`, `move_dispatch_construct.go:161`). Replaced the
owner-pointer parameters with bound func values —

```go
sendMove    func(id string, msg movemsg.Msg)
heldCenters func() map[string]vec3
rootMove    func(id string, target vec3) bool
```

— giving `beginSphereRotation(ui *uiState, heldCenters func() map[string]vec3, ev ...)`,
`applyNodeDragTarget(ui *uiState, rootMove func(id string, target vec3) bool, ev ...) bool`,
`commitDragStart(ui *uiState, sendMoveFn func(id string, msg movemsg.Msg), g, ev, tr)`,
`setHover(ui *uiState, sendMoveFn func(id string, msg movemsg.Msg), RT *rowtables.RowTables, node, port string, isInput bool, tr *T.Trace) (events, changed)`.

No new `MoveDispatch` field: the composer is already at its 12-field cap
(`check-composer-fields.sh`), so the closures are built INLINE at each of the six call sites
(`gesture_hitclassify.go` ×2, `gesture_graph.go` ×2, `gesture_handlers.go`, `gesture_actions.go`)
— e.g. `lq, mr := &md.lq, &md.mr; beginSphereRotation(&md.ui, func() map[string]vec3 { return lq.heldCenters(mr) }, ev)`
— rather than stored as fields. `ctx` is captured by value at closure-construction time
(`mr, ctx := &md.mr, md.ctx`), not held live through `md`, matching how `sendMove`/`RootMove`
already receive `ctx` as an explicit parameter today; since `md.ctx` is set once in `Start`
and never reassigned, this is behaviourally identical to threading `md.ctx` through, just
without a `*MoveDispatch`/`*moverRegistry`/`*layoutQuantizer` parameter.

`sendMove`'s send itself is UNCHANGED — still `mr.sendMove(ctx, id, msg)`, the blocking
send-with-`ctx.Done()`-escape (not `enqueueFor`'s non-blocking flush; those are two different
`sendMove`s in this codebase and this task's closures wrap the blocking one, exactly as the
functions did before). `go test -race -count=1 ./...` passes clean. Deliberately dropped the
`rootMove` closure's return value (`return false` instead of calling `lq.RootMove(...)`) and
confirmed `TestGestureDragOffCenterPreservesGrabPoint`/`TestGestureDragCenterGrabUnchanged`
fail, then restored — the binding is covered, not just built.

`grep -nE "\*moverRegistry|\*layoutQuantizer"` across the nine cluster files
(`gesture.go`, `gesture_actions.go`, `gesture_handlers.go`, `gesture_graph.go`,
`gesture_hitclassify.go`, `gesture_dispatch.go`, `ui_state.go`, `viewpoint_state.go`,
`view_stream.go`) is now empty. Exported `Wiring`-package symbol count unchanged, 167 → 167.
The no-imports-`Wiring` loop is unaffected (still empty; no package moved).

**This does NOT reopen the package lift.** The task was binding only, on purpose — the lift
itself (moving the nine files + `uiState`/`viewpointState`/`gestureState`/`overlayState` +
the view-stream half of `streamWiring` into a new package) is a separate, larger change to
measure fresh, since removing the owner-pointer signatures is a precondition for it but not
the whole of it (the field-level direct-access entanglement recorded in item 5 for
`gestureState` is untouched by this pass).

## 6b. The hub-free half lifted: GestureState/GestureRect/ViewpointState → gesturefsm; uiState still does not come free

Re-measured the whole cluster (`gesture.go`, `gesture_actions.go`, `gesture_handlers.go`,
`gesture_graph.go`, `gesture_hitclassify.go`, `gesture_dispatch.go`, `ui_state.go`,
`viewpoint_state.go`) at method/function granularity, not just files. Two disjoint halves:

- **17 `func (md *MoveDispatch)` methods** — the entry layer that builds closures from
  `&md.mr`/`&md.lq`/`md.ctx`. Unchanged, confirmed the same count as gesture-actor.md's own
  measurement.
- **The FSM's non-method state**: `gestureState`/`gestureRect` (gesture.go) and
  `viewpointState` (viewpoint_state.go) plus their own methods (`Aspect`, `PixelToNDC`,
  `Reset`, `SetViewpoint`/`EmitViewpoint`/`OrbitViewpoint`/`OrbitLockedViewpoint`/
  `ZoomViewpoint`/`PanViewpoint`) — none of these reference `md`, `*moverRegistry`, or
  `*layoutQuantizer` at all; they only read/write their own fields plus `geom`/`Trace`.

**Moved these three types + their own methods into `nodes/Wiring/gesturefsm`** as exported
types (`GestureState`, `GestureRect`, `ViewpointState`, `GesturePhase` +
`GestIdle`/`GestPending`/`GestRotating`/`GestDragging`/`GestHandhold`), all 18+ fields
exported since they are read directly by field from entry-layer files that stay in `Wiring`
(`gesture_handlers.go`, `gesture_dispatch.go`, `gesture_hitclassify.go`, `ui_state.go`,
`view_stream.go`, tests) — this is the type's OWN surface, the same cost the task accepted
for `GS`/`RT`/`Scenes`. `nodes/Wiring/gesture.go`, `nodes/Wiring/viewpoint_state.go` now hold
plain type aliases (`type gestureState = gesturefsm.GestureState`, etc. — the same shape as
`vec_alias.go`'s `vec3 = wire.Vec3`) so every remaining call site keeps its short unqualified
name; only the alias declarations themselves are package-qualified.

**`uiState` did NOT come free — confirmed by grep, not assumed.** `beginSphereRotation`,
`dragPlaneHit`, `applyNodeDragTarget`, `setHover`, and `commitDragStart` (the leaf actions the
task brief's own "hub-free half" measurement counted as non-method) all take `ui *uiState`
directly and read/write `ui.gest`/`ui.vp`/`ui.sel`/`ui.lastDraggedNode` by field. `uiState`
itself cannot move (item 5's finding stands, re-confirmed): `view_stream.go` — explicitly
out of scope for this task — reads `md.ui.ov.<13 overlay fields>`, `md.ui.editRefused`,
`md.ui.sceneEditable`, `md.ui.sceneKinds`, `md.ui.speed`, `md.ui.lastDraggedNode`, and
`md.ui.sceneSphere` all by unexported field, and `move_persist.go` writes `md.ui.vp.Persist`
by field. If `uiState` moved, ALL of those would need to be exported too — not the state
type's own surface but a second type's (`overlayState`'s 13 booleans) forced open by a THIRD
file's access pattern, which is the "export surface grows beyond the type's own API" revert
condition. So `beginSphereRotation`/`dragPlaneHit`/`applyNodeDragTarget`/`setHover`/
`commitDragStart` stay in `Wiring`, unchanged in shape (still take `*uiState`), and the 7
export-blocked `MoveDispatch` methods (`ResolveSceneDistanceGroups`/`LoadOverlays`/
`LoadSpeed`/`SetViewpoint`/`EmitViewpoint`/`SetViewStream`/`EnableSceneSwitch`) remain
blocked for the same `md.ui` unexported reason as every prior round — no second commit was
possible or attempted.

`MoveDispatch` method count: still 17 in the cluster (unchanged — none of the 17 moved or
were deleted). Exported `Wiring`-package symbol count is not the useful measure here (the
alias declarations `type gestureState = ...` etc. are themselves package-level type
declarations, already counted before as delegator methods were, net effect on the
`^func [A-Z]`/`^type [A-Z]` grep is a wash since the OLD unexported `gestureState`/
`gesturePhase`/`gestureRect`/`viewpointState` type declarations already existed under those
same unexported names). Test-name-set equality: `gesture_selection_test.go` 9/9,
`gesture_camera_outcomes_test.go` 5/5, `gesture_drag_offset_test.go` 2/2 (before/after,
`grep -c "^func Test"`) — identical names, only field-selector renames inside assertions.
`go test -race -count=1 ./...` passes clean (verbatim result in the task report). No
interfaces, `types`/`common` package, alias-shim-as-cycle-workaround, dot-import, or
`ForTest` hatch were added — the type aliases are the same pattern already in
`nodes/Wiring/vec_alias.go`, not a workaround for an import cycle (there was no cycle:
`gesturefsm` has zero dependency on `Wiring`, confirmed by `go build`). The no-imports-
`Wiring` loop is empty. `go run ./tools/gen-node-defs` shows no diff.

**Deliberately broke `GestureState.Reset`'s phase assignment** (`GestIdle` → `GestPending`)
and confirmed 4 tests fail (`TestGestureEmptyDragOrbits`, `TestGestureHandholdOrbits`,
`TestGesturePressReleaseNoMoveSelects`, `TestGestureClickNoCameraChange`), then restored —
the moved `Reset` method is covered. **Uncovered, reported rather than silently accepted:**
no test asserts `GestureState.DragNode` is cleared back to `""` by `Reset` specifically (only
`Phase` is checked post-reset); a deliberate comment-out of that one line did not fail any
test. This gap pre-dates the lift (the assertion was never written) and the lift did not
introduce it — named here per the task's coverage-reporting requirement, not fixed (out of
this task's scope to add tests).

## 6c. The nine `= gesturefsm.X` alias shims from 6b removed — every call site qualified

6b's own alias declarations (`type gesturePhase = gesturefsm.GesturePhase`, `gestIdle`/
`gestPending`/`gestRotating`/`gestDragging`/`gestHandhold` const aliases, `type gestureState
= gesturefsm.GestureState`, `type gestureRect = gesturefsm.GestureRect`,
`type viewpointState = gesturefsm.ViewpointState`) were deleted. They were justified at the
time by citing `vec_alias.go`'s `vec3 = wire.Vec3` as precedent — same shape, not a
license: `geom_bridge.go`'s `polar = geom.Polar` aliases and `port_bindings.go`'s six
`= portwiring.X` aliases were each removed for the same reason (a shim makes the package
boundary cosmetic), and this is the third instance of that same pattern.

Every reference in `nodes/Wiring` (production and tests) now names
`gesturefsm.GestureState`/`gesturefsm.GesturePhase`/`gesturefsm.GestIdle`/etc. directly —
`gesture.go`, `gesture_dispatch.go`, `gesture_graph.go`, `gesture_graph_test.go`,
`gesture_handlers.go`, `gesture_hitclassify.go`, `ui_state.go`, `viewpoint_state.go`,
`viewpoint_ops_test.go`, `gesture_camera_outcomes_test.go`, `gesture_drag_offset_test.go`,
`gesture_selection_test.go`. `grep -rn "= gesturefsm\." nodes/Wiring/*.go` now returns only
real comparisons (`g.Phase != gesturefsm.GestPending` etc.), no `type`/`const` alias
declarations. Exported `Wiring`-package symbol count unchanged, 167 → 167. `go build ./...`,
`go vet ./...`, `go test -race -count=1 ./...` all clean; `go run ./tools/gen-node-defs`
produces no generated diff; the no-imports-`Wiring` loop is empty.

Also closed the coverage gap 6b reported but did not fix: added an assertion to
`TestGesturePressReleaseNoMoveSelects` that `GestureState.Reset` clears `DragNode` back to
`""`. The 4 tests 6b named as "already exercising `Reset`" turned out not to observe this
specifically — each one's `mr` is a zero-value `moverRegistry` with no seeded `centerMirror`,
so their pointerdown's "node" hit classifier never actually arms `DragNode` (`centerOfNode`
returns `ok==false`), leaving it at `""` before Reset runs either way. Gave
`TestGesturePressReleaseNoMoveSelects` a real `nodeGeometry` for "N7" (mirroring
`gesture_drag_offset_test.go`'s `dragOffsetMD` pattern, whose self-seeded `centerOut`
channel is what makes `centerOfNode` succeed) so the pointerdown genuinely arms
`DragNode="N7"` before the no-move pointerup's `Reset`. Commented out `g.DragNode = ""` in
`Reset` and confirmed the new assertion fails:
`gesture_selection_test.go:175: after click DragNode="N7" want "" (Reset must clear it)` —
then restored it; `go test -race -count=1 ./nodes/Wiring/...` is clean again.

## 6d. `uiState`/`overlayState`/the VIEW emitter LIFTED — `docs/planning/gesture-actor.md`'s "Step 4"

Superseding 6b/6c's "uiState still does not come free": the blocker there was `view_stream.go`
reading `md.ui.ov`/`md.ui.vp`/`md.ui.sceneSphere` by field from a THIRD file, out of scope for
that task. A follow-up task named `view_stream.go` in scope and moved it alongside `uiState`/
`overlayState` into `nodes/Wiring/viewstate`, which dissolves the objection outright (the
access is now intra-package). Full detail — what moved, the bound-func treatment for
`moverRegistry`/`rowtables.RowTables`, the second commit deleting `SetViewpoint`/
`EmitViewpoint`/`SetViewStream`, VIEW-stream-ownership verification, and guard fixes — is in
`docs/planning/gesture-actor.md`'s "Step 4" section (that document owns this lift's narrative;
this entry exists so a reader of THIS doc's "5."/"6."/"6b."/"6c." history sees the outcome
without switching files). `MoveDispatch` method count 37 → 35; the 7 export-blocked methods
are now 4 (`ResolveSceneDistanceGroups`/`LoadOverlays`/`LoadSpeed`/`EnableSceneSwitch`),
each blocked for a genuine reason unrelated to `uiState` (each reaches at least one more
owner beyond `md.UI`, in a file outside this lift's scope).

**Correction (6e below): that "reaches ≥1 more owner beyond `md.UI`" measurement was the
`emitViewFrame` delegator itself** — a one-line forward left behind by this same lift,
counted as a second owner. Once it was deleted, three of the four (`ResolveSceneDistanceGroups`/
`LoadOverlays`/`LoadSpeed`) reach exactly one owner (`md.UI`) and were kept anyway because
each does real multi-step work beyond forwarding; only `EnableSceneSwitch` was a pure
two-field forward and was deleted.

## 6e. `emitViewFrame` delegator deleted; `EnableSceneSwitch` deleted, the other three kept

The `emitViewFrame` delegator 6d left behind (`func (md *MoveDispatch) emitViewFrame(events
[]wire.RowEvent) { md.UI.EmitViewFrame(events) }`, `view_stream.go`) was deleted; its ~30
in-package call sites now call `md.UI.EmitViewFrame(...)` directly. That resolved the "reaches
≥1 more owner beyond `md.UI`" reason 6d gave for all four remaining export-blocked methods —
each of them touched `md.UI` plus the now-gone `md.emitViewFrame`, so once that delegator was
gone each touched only ONE owner. Re-measuring what each actually does, not just what it
touches:

- **`ResolveSceneDistanceGroups`** (`distance_groups.go`) — three separate `scene.*` lookups
  (`SceneHasDistanceGroups`/`SceneIsEditable`/`SceneKindMask`) each assigned to a different
  `md.UI` field. Real logic (three independent computed values), not a forward. **Kept.**
- **`LoadOverlays`** (`scene_overlays_persist.go`) — loads `overlays.json`, installs it via
  `md.UI.OV.SetGuideVisibility`, then emits a 13-event VIEW frame by hand. Real logic. **Kept.**
- **`LoadSpeed`** (`scene_speed_persist.go`) — loads `speed.json`, sets two `md.UI` fields,
  computes the effective speed, and broadcasts it to every speed sink before emitting a VIEW
  frame. Real logic. **Kept.**
- **`EnableSceneSwitch`** (`scene_switch.go`) — was exactly two field assignments
  (`md.Scenes.AnchorPath = anchorPath; md.Scenes.Quit = quit`), nothing else. A pure forward
  onto the already-exported `md.Scenes` field. **Deleted** — its one caller
  (`runtopology/topology_run.go`) now sets `md.Scenes.AnchorPath`/`md.Scenes.Quit` directly.

`MoveDispatch` method count and export-blocked count: see the commit report for before/after
numbers (this doc is not a status board; the count is a measurement made once, not tracked
here going forward).

## 6f. Three of the four remaining pure delegators deleted; `SliderSpeed` kept as a derived value

Checked the four export-exempt-only delegators onto `md.UI`: `EmitBreadcrumb` (`view_stream.go`)
and `Viewpoint` (`viewpoint_state.go`) were plain one-line forwards — deleted, callers now
address `md.UI.EmitBreadcrumb(...)` / `md.UI.VP.Viewpoint` directly
(`runtopology/startup_report.go`, `nodes/Wiring/stdin_dispatch.go` in-package;
`nodes/Wiring/scenecamera/scene_camera_test.go` out-of-package, an external test package that
already reaches `md.UI` since it is exported). `view_stream.go` had nothing left in it once
`EmitBreadcrumb` was gone and was deleted as a file. `PanViewpoint` (`viewpoint_state.go`) was
also a pure one-line forward (`md.UI.VP.PanViewpoint(delta, tr)`) despite its long doctrine
comment about pan-vs-zoom-vs-scene-sphere — the comment moved with the deletion (now sits
directly above `md.UI.VP.PanViewpoint`'s call sites' context in the file, not attached to a
body); deleted, callers rewritten in `gesture_handlers.go` (in-package) and
`scene_camera_test.go` (out-of-package). `SliderSpeed` (`scene_speed_persist.go`) was
re-examined and KEPT: its body computes `EffectiveClockSpeed(md.UI.Speed, md.UI.ClockDivisor)`,
a derived value from two `md.UI` fields via a package function — the same "real work beyond
forwarding" shape 6e already used to keep `ResolveSceneDistanceGroups`/`LoadOverlays`/
`LoadSpeed`, not a plain field pass-through.

`MoveDispatch` method count: 21 → 18 (drop of 3, `grep -h "^func ([a-z]* \*MoveDispatch)
[A-Z]" nodes/Wiring/*.go | grep -v _test | wc -l`). Exported `Wiring`-package symbol count
unchanged, 162 → 162 (same grep pattern as prior rounds in this doc — note it filters on
`_test` appearing in the MATCHED LINE TEXT, not the source filename, since `-h` strips
filenames; it has never actually excluded test files, and this round's before/after used the
identical command so the comparison is apples-to-apples regardless). No new interfaces,
`types`/`common` package, alias shim, or `ForTest` hatch. The no-imports-`Wiring` loop is
empty. `go run ./tools/gen-node-defs` produces no diff. `go test -race -count=1 ./...` is
clean.

VIEW-stream ownership verified by reading `runtopology/topology_run.go`: `emitStartupBreadcrumbs`
(which now calls `md.UI.EmitBreadcrumb`) still runs at line 63, and `startStdinReader` — which
spawns the gesture/view-owner goroutine — still runs at line 81, in the same function, same
order, unchanged by this pass; only the receiver access path changed
(`md.EmitBreadcrumb`→`md.UI.EmitBreadcrumb`), not which goroutine calls it or when.

Coverage: deliberately broke `gesturefsm.ViewpointState.PanViewpoint` (dropped the `v.Pan(delta)`
call) — 7 tests failed across `nodes/Wiring` and `nodes/Wiring/scenecamera` (both in-package
and external-test-package callers), confirming the rewritten call sites are exercised; restored,
`go test -race -count=1 ./...` clean again. **Uncovered, reported rather than silently accepted:**
no test catches `viewstate.UIState.EmitBreadcrumb` becoming a no-op — deliberately emptied its
body and every test still passed. This gap pre-dates this pass (the method's own logic was
untouched; only its now-deleted `MoveDispatch` wrapper was removed) and is out of this task's
scope to fix.

## 6a. Original decline (superseded above, kept as the record of what was measured wrong)

Attempted to lift the nine gesture+view files, `uiState`/`viewpointState`/`gestureState`/
`overlayState`, and the view-stream half of `streamWiring` into `nodes/Wiring/gestureview`,
per the task brief's own measurement (`md.mr`/`md.sw`/`md.lq` reduce to seven operations:
`centerOfNode`, `nodeBodyRadius`, `heldCenters`, `viewOut`, `viewBuildFrame`, `viewTick`,
`ensureClaims`).

**That measurement undercounted.** A wider grep for `md\.mr\b`/`md\.lq\b` (not just
`md\.mr\.`/`md\.lq\.`, which only catches direct method calls) finds `&md.mr`/`&md.lq`
passed BY POINTER into six functions, three of which live inside the files that would move:

```
beginSphereRotation(ui *uiState, mr *moverRegistry, lq *layoutQuantizer, ...)   gesture_actions.go (moves)
applyNodeDragTarget(ctx, ui *uiState, mr *moverRegistry, lq *layoutQuantizer, ...) gesture_actions.go (moves)
commitDragStart(ui *uiState, mr *moverRegistry, ctx, g, ev, tr)                  gesture_graph.go (moves)
setSelectionUI(ui *uiState, mr *moverRegistry, ctx, node, edge)                  move_dispatch_api.go (STAYS)
sendMove(mr *moverRegistry, ctx, id string, msg movemsg.Msg)                     move_dispatch_api.go (STAYS)
DistanceGroupLens(ui *uiState, mr *moverRegistry)                                distance_groups.go (STAYS)
```

plus `layoutQuantizer.RootMove`/`layoutQuantizer.heldCenters`, both already on record (item 1,
"pre-existing, now closed") as needing `*moverRegistry`/full owner access, not a leaf read.

**`moverRegistry` and `layoutQuantizer` are unexported types defined in package `Wiring`**
(`mover_registry.go`, `quantized_move.go`). Go's export rule — not a design preference — means
no other package can name them in a parameter list at all, exported alias or not: renaming
them `MoverRegistry`/`LayoutQuantizer` doesn't fix it, because the new package would then have
to `import ".../nodes/Wiring"` to reach the exported name, which is banned outright
(`nodes/Wiring/<sub>` may not import `nodes/Wiring`) and would cycle against the OTHER
required edge anyway: `MoveDispatch` must hold a field of the new package's type
(`GV gestureview.GestureView`, the task's own required pattern, matching `GS`/`RT`/`Scenes`),
so `Wiring` already imports the new package. Both edges existing together is
`import cycle not allowed`, the literal error the compiler gives — this is `beadcrud`/
`portwiring`'s "genuine, compiler-confirmed cycle" class, not the usual dissolvable blocker.

**Both directions of the edge, as the task asked:**
- new package → `Wiring`: `beginSphereRotation`/`applyNodeDragTarget`/`commitDragStart`
  (in-cluster) and their callees `setSelectionUI`/`sendMove`/`DistanceGroupLens`/
  `layoutQuantizer.RootMove`/`layoutQuantizer.heldCenters` (stay in `Wiring`) all require
  `*moverRegistry`/`*layoutQuantizer` — unexported `Wiring` types, unreachable from outside.
- `Wiring` → new package: `MoveDispatch` needs a `GV gestureview.GestureView`-shaped field
  to keep being the composer, per the task's own prescribed pattern.

This is a bigger surface than "give it three func values" — `mr`/`lq` are read AND written
(hover messaging, `RootMove`, sphere-rotation math) through many operations, not the three
named lookups, and `moverRegistry`/`layoutQuantizer` are explicitly declared not a lift target
elsewhere in this doc ("`moverRegistry` is not a target... whatever remains of `MoveDispatch`
stays with it"). Turning every one of those operations into an individually-injected func
value would be a different, much larger task (effectively exporting `moverRegistry`'s/
`layoutQuantizer`'s behaviour piecemeal), not this one.

**No code was written.** `git status --short` was empty throughout this probe; `go build
./...`/`go test ./...` were run against the untouched tree only to confirm the starting state
was clean. `MoveDispatch`'s method/export counts are unchanged from item 5's last measurement
(37 methods; the same 7 export-blocked methods remain blocked).

Pattern recorded for the next attempt: grepping `md\.owner\.` (dot-call only) undercounts —
`&md.owner` passed as a function argument is the same coupling and needs its own grep
(`md\.owner\b`, then check what the receiving parameter's TYPE is, not just that the value
came from one field).

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

## 7. The build cluster (14 `build*.go` + `loader.go`/`loader_layout.go`) — probed and DECLINED

Applied "the one rule": named the specific values the build cluster reads from `MoveDispatch`'s
actor internals, not just from the constructor. `newMoveDispatch`/`finalizeActors`/`bind` alone
would have been a legitimate ask (an exported constructor + two exported entry-point methods —
matching the `NewMoveDispatch` note in the task brief and the `GS`/`RT`/`Scenes`/`UI` precedent
of "export follows the type's own surface"). That is not what the grep shows.

**Exact export list the lift would require**, from `build_move_dispatch.go`/`build_nodes.go`/
`build.go` (verbatim call sites):

- `newMoveDispatch(...)` → an exported constructor. **Defensible alone.**
- `md.mr` (`moverRegistry`, unexported by design — `move_dispatch.go`'s own field comment says
  `bind`/`centerOfNode`/`enqueueFor`/`finalizeActors` are meant to be called `md.mr.X` by
  **in-package** callers only) — reached directly six separate ways:
  - `md.mr.nodeGeoms` — the actor map itself, iterated and indexed by node id
    (`build_move_dispatch.go` lines seeding `flags.coplanarEdges`, `flags.upAxis`,
    `quantOffset`, `selfKind`, `tilt.topTiltVectorThetaIdx`, `topo.neighborKinds`,
    `outs.outTargets`) — **7 unexported `nodeGeometry` fields written by direct field
    assignment**, not through any method.
  - `md.mr.bind(outSink, inputcodec.SlotRegistry(destWire))` (`build_nodes.go:111`)
  - `md.mr.finalizeActors(&b.speedSinks)` (`build.go:126`)
  - `&b.md.mr` handed to `PortBindings.mr` (`build_nodes.go:28`)
- `md.lq.quantizedLayout` (`layoutQuantizer`, unexported) — direct field write
  (`build_move_dispatch.go:61`)
- `b.md.inboxes` (`nodeInboxes`, unexported) — pointer handed out (`build_nodes.go:27`)
- `b.md.sw.interiorOuts` / `.driveOuts` / `.buildInteriorFrame` (`streamWiring`, unexported) —
  three more unexported fields, pointers handed out (`build_nodes.go:49-51`)
- `b.md.selfDriveClaimed` — referenced from `build_args_selfdrive.go`

**Count: 1 legitimate constructor + 6 distinct unexported-internals reach-ins (one of them —
`nodeGeoms`' 7 per-node fields — is itself a compound list), across `moverRegistry`,
`layoutQuantizer`, `nodeInboxes`, and `streamWiring`.** This is not "the constructor plus a
couple of real entry points"; it is four different actor-owned types' private state, reached by
field, from outside their own methods.

**This is exactly the class item 6a's "one rule" was written to separate from a real blocker,
but on the other side of the line.** `beadcrud`/`portwiring` lifted because their coupling
reduced to a small, named set of func values/exported methods once measured. Here the measured
set is the actor registry's OWN unexported state, written directly instead of through any
method — the shape item 6a called "not a lift target" for `moverRegistry`/`layoutQuantizer.
`nodeGeometry`'s own file (`node_geometry.go`) states, for nearly every mutable field, that it
is written ONLY by that node's own goroutine post-construction — the reason this build-time
seeding is safe today is that it runs single-threaded, before any node goroutine exists, still
inside the actor's own package. Exporting `nodeGeoms`, its 7 fields, `bind`, `finalizeActors`,
`inboxes`, and `streamWiring`'s three fields would not create a new capability the build phase
lacks — it would remove the enforcement (unexported = only this package's methods can touch it)
that currently backs every one of `node_geometry.go`'s single-writer claims, to satisfy a file
move with no behavioural motivation.

**Declined.** `MoveDispatch` and every actor it owns (`moverRegistry`, `layoutQuantizer`,
`nodeInboxes`, `streamWiring`) stay in `Wiring`, and so does their construction — `loader.go`,
`loader_layout.go`, and all 14 `build*.go` files. `newMoveDispatch` remaining unexported (used
only by `buildMoveDispatch`, its sole caller, already in-package) costs nothing; exporting only
it while leaving the other five reach-ins unexported would still fail to compile the lift, since
`build_move_dispatch.go`/`build_nodes.go`/`build.go` would still need to move with it or the
constructor's own body would gain a back-reference to `Wiring`. No code was written; no
constructor, method, or field was exported. `git status --short` was empty throughout this
probe; `go build ./...` was run against the untouched tree only to confirm the starting state.

No interface, `types`/`common` package, alias shim, dot-import, package-level actor global, or
`ForTest` hatch was used to route around this — the decline is the outcome, not a placeholder
for one of those.

## 8. Scene-persistence pure helpers lifted into `nodes/Wiring/scenepersist`

Measured `nodes/Wiring` by coupling rather than by category: 42 pure standalone functions
(~550 lines) touch no `md.`, no `buildCtx`, no actor type. Fourteen of them clustered in one
subject — scene persistence I/O across `scene_lattice_persist.go`/`scene_speed_persist.go`/
`scene_overlays_persist.go`/`scene_sphere_persist.go` — and `scenepersist` already existed
(holding `scene_selection_persist.go` from an earlier lift), so this had a named home rather
than a new grab-bag package.

**Moved, verified pure:** `WriteSceneLattice`/`FormatLatticeJSON`/`LoadSceneLattice`/
`LoadLatticePoints`/`SendLatticePointsNonBlocking` (+ `DefaultLatticePoints`, exported
alongside since three other `Wiring` files read it directly);
`EffectiveClockSpeed`/`WriteSceneSpeed`/`FormatSpeedJSON`/`LoadSceneSpeed`/`BroadcastSpeed`
(+ `DefaultPlaybackSpeed`, same reason); `WriteSceneOverlays`/`LoadSceneOverlays`;
`LoadSceneSphere`/`WriteSceneSphere`. All 14 matched the task brief's list exactly — none
turned out to secretly close over a `Wiring` type.

**Stayed, as directed:** the `MoveDispatch` methods `LoadOverlays`/`LoadSpeed`/
`LoadSceneSphere`/`SliderSpeed`, and the four persister types (`latticePersister`/
`speedPersister`/`overlaysPersister`/`sceneSpherePersister`) — each now calls
`scenepersist.X(...)` instead of the in-package function. `HumanEditSpeed` also stayed (an
interaction-mode override, not a persistence default). The four Wiring-side files kept their
exact basenames (`scene_lattice_persist.go` etc.) — `scene_selection_persist.go`'s own header
already recorded why: `check-persist-write-ownership.sh` matches by basename only, so a
package split that keeps the name needs no guard edit. Both persistence guards were re-run
clean after the move with no edits, then proven with teeth: a `jsonpersist.WriteJSONAtomic`
call dropped into `nodes/Wiring/teeth_probe.go` AND into
`nodes/Wiring/scenepersist/teeth_probe.go` (recursive scan, basename-only match — both
non-owner basenames) each produced `unauthorized-write`; a hand-rolled
`filepath.Join(root, "view", "probe.json")` in `nodes/Wiring/scenepersist/teeth_probe.go`
produced `hand-rolled-join`. All three probe files were deleted immediately after observing
the failure; `git status --short` was empty before each commit.

Six call sites elsewhere in `Wiring` (`build_move_dispatch.go`, `build_args_lattice.go`,
`move_dispatch.go`, `move_dispatch_construct.go`, `loader.go`, `stdin_dispatch.go`,
`stdin_apply.go`) were updated to call the new `scenepersist.` names. Test files
(`scene_camera_persist_test.go`, `scene_clock_divisor_test.go`, `scene_edit_persist_test.go`,
`scene_lattice_edit_test.go`, `scene_lattice_persist_test.go`, `scene_speed_persist_test.go`,
`scene_sphere_persist_test.go`, `stdin_input_integration_test.go`, `tilt_edit_speed_test.go`)
were qualified the same way rather than moved, because every one of them depends on the
`Wiring`-only `writeTree`/`loadTreeMD`/`MoveDispatch` test harness and so cannot move to
`scenepersist` without `scenepersist` importing `Wiring` (forbidden). The one exception —
`TestSceneSphereRoundTrip` in `scene_sphere_persist_test.go`, which drove only
`writeSceneSphere`/`loadSceneSphere` and `t.TempDir()`, no `MoveDispatch` — moved to a new
`nodes/Wiring/scenepersist/scene_sphere_persist_test.go` verbatim (same assertions, same
name). No test was weakened; no `[allow-test-weakening: ...]` marker was added.

`nodes/Wiring` non-test `.go` file count: 61 → 61 (no file deleted or added at that level;
the four files shrank, `scenepersist` gained 4 new files + 1 moved test). JSON byte-identity
was verified by keeping every struct tag, key name, and polarity/ordering literal
byte-for-byte identical in the moved copy (`sceneLatticeFile`, `sceneSpeedFile`,
`sceneOverlaysFile`'s 13 tagged fields, `sceneSphereJSON`) — confirmed by the full existing
persistence-round-trip suite passing unchanged
(`TestPersistLatticePointsRoundTrips`/`TestPersistSpeedRoundTrips`/
`TestOverlaysPersistPreservesCamera`/`TestSceneSphereRoundTrip`/
`TestSceneSphereContentFitSurvivesReloadAfterMove` among others), which reads the actual
bytes off disk (`memory/feedback_headless_repro_verifies_persistence.md`'s shape), not an
in-memory stub.

The no-imports-`Wiring` loop is empty (`scenepersist` imports only `jsonpersist`,
`scenepaths`, `tiltvector`, `viewstate`, `geom`, `nodes/wire`, `nodes/wire/clock` — none of
them `Wiring`). `go build ./...`, `go vet ./...`, `go run ./tools/gen-node-defs` (no
generated diff) all clean. `go test -race -count=1 ./...` passes clean, including the new
`nodes/Wiring/scenepersist` package.

No interface, `types`/`common` package, alias shim, dot-import, package-level actor global,
or `ForTest` hatch was added.

## 9. Six pure derive phases lifted into `nodes/Wiring/topoderive`

Item 7's decline was scoped to the build cluster's actor-internals reach-ins
(`md.mr`/`md.lq`/`md.inboxes`/`md.sw`); it explicitly did not cover the pure DERIVE phases
item 3 had already converted to package-level functions. Re-verified each of the six named
functions' signature AND body (not just the signature, per this doc's own "grep -v without
a leading ./" trap and the `md\.owner\b` vs `md\.owner\.` lesson from item 6a) before moving
anything:

**Moved, confirmed pure** (touch no `md.`, `buildCtx`, or actor type, in body or signature):
`computeNodeGeometry` → `topoderive.ComputeNodeGeometry` (`geometry.go`),
`computeQuantizedLayout` → `topoderive.ComputeQuantizedLayout` (`quantized_layout.go`),
`computeReachRadii` → `topoderive.ComputeReachRadii` (`reach.go`),
`buildEdgeMaps` → `topoderive.BuildEdgeMaps` (`edge_maps.go`),
`allocateVectorChannels` → `topoderive.AllocateVectorChannels` (`vector_channels.go`).
`vec3` (Wiring's `= wire.Vec3` alias) is spelled `wire.Vec3` directly in the new package, per
this doc's own standing rule against alias shims (items 6c's three prior removals) — no new
alias file was added.

**`reachRFromPolar` moved too**, alongside `computeReachRadii` (its only load-path caller) —
`topoderive.ReachRFromPolar`, exported so `nodes/Wiring/commit_node_move.go`'s
`commitNodeMoveLocal` (a live-drag path staying in `Wiring`, item 4's "same function called
twice is not the same as redundant") can still call it. One-way import confirmed: `Wiring`
now imports `topoderive`, `topoderive` imports nothing under `nodes/Wiring/`.

**Stayed, with reason:** `buildTypeMaps` — its signature is pure (`loadspec.TopoSpec` in,
two plain maps out), but its BODY reads the package-level `Registry` var
(`node_registry.go`), whose value type `NodeBuilder` is declared in package `Wiring` itself
(`Build func(..., deps buildDeps) (wire.Node, error)`, closing over the unexported hub type
`buildDeps`). This is exactly the class the task brief warned the measuring script might
miss: the SIGNATURE looked identical to the other five, but the BODY closes over a Wiring
type via a package-level global, not a parameter — moving it would require `topoderive` to
import `Wiring` (forbidden) or `Wiring` to export `NodeBuilder`'s internals further than
they already are, for no behavioural gain. `buildEdgeMaps` was already parameterized on
`buildTypeMaps`' two outputs (`nodeType`, `kindBroadcastPorts` — both plain `map[string]...`
types), so it moved cleanly while its supplier function did not; `build.go` still calls
`buildTypeMaps(b.spec)` in `Wiring`, then hands the two plain-map results to
`topoderive.BuildEdgeMaps(...)`.

**Phase order and mutation visibility verified, not just asserted.** `buildFromSpec`
(`build.go`) calls the five in the exact same order as before, now qualified:
`topoderive.ComputeNodeGeometry` → `topoderive.ComputeQuantizedLayout` →
`topoderive.ComputeReachRadii` → `b.allocateWires()` → `topoderive.AllocateVectorChannels`
→ `b.buildMoveDispatch()` → `buildTypeMaps` → `topoderive.BuildEdgeMaps` → `b.buildNodes()`
— no line moved position, only the call target's package qualifier changed.
`ComputeQuantizedLayout`/`ComputeReachRadii` still take `centers`/`nodeGeoms` as `map[...]`
parameters and mutate them IN PLACE exactly as their `Wiring`-method predecessors did — Go
maps are reference types regardless of which package holds the function, so the caller's
copy in `buildCtx` sees the same writes through the same map identity as before; moving the
function to a different package changes nothing about that, since the map itself was never
copied, only passed by its existing reference. Confirmed with three deliberate injections at
the `build.go` call sites (dropped `topoderive.ComputeReachRadii` call, dropped
`topoderive.ComputeQuantizedLayout` call, swapped `topoderive.AllocateVectorChannels`'s two
return values) — each reproduced the exact failure item 3's coverage-gap tests were written
to catch:
```
--- FAIL: TestLoadTopologyComputesReachRadii (0.01s)
    build_load_derive_test.go:65: node 1 ReachR = 0, want 50 (distance to its edge partner)
--- FAIL: TestLoadTopologyComputesQuantizedOffsets (0.01s)
    build_load_derive_test.go:108: node 2 quantOffset derives to {X:0 Y:0 Z:0}, want close to {X:50 ...} (scene-polar center); off by 50 > tol 10
--- FAIL: TestPairNodeVectorChannelsThreadSourceOutTargetIn (0.00s)
    vector_channel_threading_test.go:107: source node's (id 1) own VectorOut is nil — source end never wired
```
All three restored immediately after observing the failure; `go build ./...`/`go test
-race -count=1 ./...` clean again.

**Test move.** `build_load_derive_test.go`'s third test,
`TestAllocateVectorChannelsKeysSourceOutTargetIn`, drove only
`allocateVectorChannels(spec)` plus `encoding/json` — no `MoveDispatch`/`writeSpecTree`/
`LoadTopology` harness — so it moved verbatim (same name, same 6 assertions) to
`nodes/Wiring/topoderive/vector_channels_test.go`. The other two tests in that file
(`TestLoadTopologyComputesReachRadii`, `TestLoadTopologyComputesQuantizedOffsets`) read
`md.mr.nodeGeoms` off a real `LoadTopology`, so they stayed in `Wiring` unchanged — this is
the file SPLIT the task called for, not a weakening (test-name-set before: 3; after: 2 in
`Wiring` + 1 in `topoderive` = 3, assertion counts unchanged per test).
`vector_channel_threading_test.go` (external `Wiring_test` package, needs `PairNode`) was
untouched except for one stale doc-comment line the `check-doc-symbols` guard caught
(`allocateVectorChannels` → `topoderive.AllocateVectorChannels`).

`nodes/Wiring` non-test `.go` file count: 61 → 59 (`build_geometry.go`, `loader_layout.go`
deleted outright — each held exactly one moved function and nothing else survived in
either). `topoderive` gained 5 production files + 1 test file. Exported `Wiring`-package
symbol count is unaffected in the way this doc's own grep tracks it (none of the five moved
functions were ever capitalized in `Wiring`, so their departure removes zero from that
count; `topoderive`'s own exported surface is new but is a DIFFERENT package, matching the
`scenepersist`/`gesturefsm`/`viewstate` precedent of "export follows the type/function's own
surface").

Guards re-audited for silent-green risk (glob or unexported-name keying):
`tools/repo-hygiene/check-no-untracked-source.sh` (glob-based file inventory) DID catch the
six new files while unstaged — `git add -N` alone would have kept them permanently invisible
to it, the doc's own instructions call this out by name (`git add -N` is explicitly
disallowed by the task; real `git add` used instead). `tools/network/doc/check-doc-symbols.sh`
(name unconfirmed exactly, invoked via `check-doc-symbols` in stop-checks) caught the one
stale backtick reference to `allocateVectorChannels` left behind in
`vector_channel_threading_test.go`'s doc comment after the rename — fixed to
`topoderive.AllocateVectorChannels`. No guard is keyed to `build_geometry.go`/
`loader_layout.go`/`build_edge_maps.go`/`broadcast_move.go` by filename or to any of the six
function names (`grep -rl` against `tools/*/*.sh tools/*/*/*.sh` for every function/filename
above returned empty before this change, confirming there was nothing to re-key).

The no-imports-`Wiring` loop is empty (`for p in $(go list ./nodes/Wiring/... | grep -v
'nodes/Wiring$'); do go list -deps "$p" | grep -qx github.com/dtauraso/wirefold/nodes/Wiring
&& echo "IMPORTS WIRING: $p"; done` — no output). `go build ./...`, `go vet ./...` clean;
`go run ./tools/gen-node-defs` produces no generated diff; `go test -race -count=1 ./...`
passes clean including the new `nodes/Wiring/topoderive` package. No interface,
`types`/`common` package, alias shim, dot-import, package-level actor global, or `ForTest`
hatch was added.

## 10. Six leftover tests moved to the package holding their subject; five stayed, one split

Checked which test files never referenced `md.`/`MoveDispatch`/`moverRegistry`/
`nodeGeometry`/`edgeMover`/`nodeMover`/`buildCtx`/`uiState`/the `loadTreeMD`/
`writeSpecTree`/`LoadTopology` harness — these were left behind by the 32 earlier lifts on
this branch, each of which moved the code but not always its test.

**Moved, verbatim, one commit each:**
- `speed_delivery_test.go` (5 tests) → `nodes/wire/clock` — exercises
  `SendSpeedNonBlocking`/`ApplySpeedNonBlocking`/`NewRealClock`/`TickPeriod`, all of which
  live there; only the `clock.` qualifier was stripped (in-package now).
- `lit_bead_index_test.go` (4 tests, incl. `TestChainBeadsEveryIndexIsReachable`) →
  `nodes/Wiring/beadindex` — every assertion calls only `beadindex.LitBeadIndex` and
  `lattice.DwellTicksPerBead`; despite its doc comment discussing `chainBeads` (still in
  `Wiring`, `chain_beads.go`), no test in the file calls it — the file was NOT actually
  mixed, just documented misleadingly. Confirmed by grep: `chainBeads` has zero non-comment,
  non-doc references anywhere in this test file.
- `pan_polar_test.go` (1 test) → `nodes/Wiring/geom` — exercises
  `PanDisplacementPolar`/`PlaneSlide`/`DeltaToPolar`/`BasisFromViewpoint`, all exported
  `geom` functions.
- `pulse_speed_parity_test.go` (1 test) → `nodes/Wiring/nodegeom` — a drift guard between
  `nodegeom.CurveParamPulseSpeedWuPerMs` and `lattice.PulseSpeedWuPerMs`; neither symbol is
  `Wiring`'s, and `nodegeom` already hosts the same-shaped `curve_params_test.go` guard for
  its sibling constant (`TestTiltVectorAngleStepIsPiOver12`), so this is the established
  pattern, not a new one.

**Split:** `scene_path_safety_test.go`'s `TestSafeTreePathComponent` moved to a new
`nodes/Wiring/jsonpersist/safe_tree_path_test.go` (pure `jsonpersist.SafeTreePathComponent`,
no `Wiring` symbol). `TestWriteQuantOffsetRejectsTraversalID` stayed in `Wiring` — it calls
the unexported `writeQuantOffset` (`quant_offset_persist.go`), which was never a lift
candidate.

**Kept, verified rather than assumed (per the task's own "probably staying" list):**
- `vec_close_test.go` — `vecClose` is still called by 4 other `Wiring`-package test files
  (`gesture_camera_outcomes_test.go`, `gesture_drag_offset_test.go`,
  `gesture_home_test.go`, `scene_camera_persist_test.go`); its subject is its Wiring-package
  callers, not a subpackage.
- `drive_slot_claim_test.go` — constructs `BuildArgs` directly (unexported `Wiring` fields).
- `dispatch_keys_test.go` — reads `Wiring`'s own unexported dispatch tables
  (`rawInputHandlers`, `hitClassifiers`, etc.).
- `state_seed_test.go` — constructs `BuildArgs` directly, calls `BuildArgs.StateSeed`.
- `tilt_vector_phi_removed_persist_test.go` — needs the `writeTreeFile` harness helper
  (`wire_test_helpers_test.go`, unexported, `Wiring`-package only).
- `interior_sphere_test.go` — its own doc comment claims it needs "Wiring's own nodeRadius",
  but that function no longer exists anywhere in `Wiring` (grep confirms); the test now
  calls only `interior.*`/`nodegeom.NodeRadius`, both exported, and neither package imports
  the other (confirmed: `go list -deps` each way, empty both directions). It stays anyway —
  moving it would mean picking one of two unrelated packages as an arbitrary host for a test
  that needs both, which is a new kind of coupling, not a subject match; neither `interior`
  nor `nodegeom` was in this task's four candidate destinations
  (`nodes/wire/clock`, `beadindex`, `jsonpersist`, `geom`). Left as a stale-comment finding,
  not fixed (out of this task's scope).

**Verification.** Each of the 5 destination packages (`clock`, `beadindex`, `geom`,
`nodegeom`, `jsonpersist`) had its subject deliberately broken once and the moved test
confirmed to fail from its new location, then restored — see the task's own commit history
for the exact breaks (a `+0.5` bearing offset in `PanDisplacementPolar`, a `+1` floor offset
in `LitBeadIndex`, a `ch <- speed` unconditional send in `SendSpeedNonBlocking`, `0.04` →
`0.05` in `CurveParamPulseSpeedWuPerMs`, and dropping the `""`/`"."`/`".."` guards in
`SafeTreePathComponent`). `nodes/Wiring` top-level `.go` file count: 117 → 113 (4 files
fully departed: `speed_delivery_test.go`, `lit_bead_index_test.go`, `pan_polar_test.go`,
`pulse_speed_parity_test.go`; `scene_path_safety_test.go` shrank but stayed since it still
holds one test). `go build ./...`, `go vet ./...` clean; `go test -race -count=1 ./...`
passes with no failures across every package. `check-test-integrity.sh` clean (no net
assertion removal, no skip/only/exit/recover added — every moved test kept its exact body).
The no-imports-`Wiring` loop is empty. No interface, `types`/`common` package, alias shim,
dot-import, package-level actor global, or `ForTest` hatch was added.

## 11. `chain_beads.go`/`bead_chain.go` re-measured twice — the 265-line loop BODY decomposed into three phases in `beadindex`, actor management declined by mechanism

**First pass (superseded by the correction below, kept for the record):** measured only
`chainBeads()`'s TOP LEVEL and lifted two inline scalar formulas
(`BeadPlacementOffset`/`PulsePlacementOffset`) into `beadindex`, then declined `chainBeads()`
itself whole because its top-level body calls four impure operations. That verdict was true
of the top level and said nothing about the 265 lines (of `chainBeads`' 283-line body)
inside the single `for _, to := range m.outs.outTargets` loop between them — a function that
had not been given a name, not evidence the loop's arithmetic was impure.

**Correction — decomposed the loop body by phase**, matching the shape the impure statements
and the pure arithmetic actually take: resolve target geometry + step count, publish the
count (impure), gather live pulses (impure), build the breadcrumb text (pure) next to the
breadcrumb send (impure), place beads along the chain and decide which are lit (pure). Added
`nodes/Wiring/beadindex/chain_edge_layout.go`, three pure functions (still `beadindex` — no
new package):

- `ChainEdgeGeometry(selfCenter, targetCenter wire.Vec3, selfTorusR float64, selfKind, targetKind string) (dist float64, dir wire.Vec3, count int, ok bool)`
  — `nodegeom.EdgeCenterDistAndDir` + `nodegeom.EdgeStepCount`, unchanged, given a name.
- `ChainBeadRows(dir, chainSep wire.Vec3, base, step float64, count int, resolved []wire.Vec3, resolvedValid []bool, pulses []Pulse) (ox, oy, oz []float32, lit []uint8, litVal []int32)`
  — the placeholder-bead loop and the lit-pulse loop, unchanged arithmetic. `resolved`/
  `resolvedValid` are the bead-actor chain's own ALREADY-DRAINED snapshot positions, read
  once in `chain_beads.go` and handed down as a plain slice — the function itself never
  touches the actor, only a copy of its last-known values, which is what keeps it pure.
- `ChainAimBreadcrumbText(to string, count int, dist float64, dir wire.Vec3) string` — the
  `fmt.Sprintf` breadcrumb VALUE string, pure formatting from local values, split from the
  `m.tr.Breadcrumb(...)` call next to it (that call is the side effect; the string is not).

`chain_beads.go`'s loop now contains only: the `m.topo.partnerCenters`/`m.topo.neighborKinds`/
`m.topo.mutualTargets`/`m.topo.nodeRowFor` lookups (bucket (a), gathering call arguments),
the `PublishSteps`/`SendStepsNonBlocking` channel sends, the `LiveBeadFractions` live-wire
read that builds `pulses`, the `m.tr.Breadcrumb` send, the `m.reconcileBeadChain` call and
the read of its returned snapshot into `resolved`/`resolvedValid`, and appending each phase's
returned slices.

**(a)/(b) classification of the original 265-line loop body (lines 122–386 of the
pre-decomposition file), by statement group:**

| statement group | lines (approx) | bucket |
|---|---|---|
| `partnerCenters`/`neighborKinds` lookups + skip | 122–135 | (a) reads `m.*` |
| `EdgeCenterDistAndDir` + `EdgeStepCount` call | 137–155 | (b) pure (now `ChainEdgeGeometry`) |
| publish `count` onto `outWireOuts`/`outStepsIn` | 156–174 | (a) channel sends |
| gather `pulses` from `outWires[i].LiveBeadFractions` | 176–214 | (a) reads live wire state (the `pulse` struct build itself is pure once the wire is read; kept with the read since the read dominates the group) |
| breadcrumb gate + text formatting | 240–270 | text formatting (b, now `ChainAimBreadcrumbText`); the `if m.tr != nil`/`m.tr.Breadcrumb` framing (a) |
| `offsetAt`/`aimUnit`/`chainSep` (incl. `ParallelChainOffset`) setup | 278–306 | (b) pure, except the `m.topo.mutualTargets[to]`/`m.geom.SceneCenter`/`m.id` reads that gather its arguments (a) |
| `beadTickFn` check + `reconcileBeadChain` call | 307–316 | (a) impure (starts/stops actor goroutines) |
| placeholder-bead loop | 319–344 | (b) pure (now `ChainBeadRows`, given `resolved`/`resolvedValid`) |
| lit-pulse loop | 345–385 | (b) pure (now `ChainBeadRows`) |

Roughly 190 of the 265 lines (placeholder loop, lit-pulse loop, the geometry/step-count call,
the breadcrumb text, most of the offset/aim/chainSep setup) were bucket (b) and are now
lifted; the rest — every channel send, every live-wire/actor read, the `reconcileBeadChain`
call, and the small `m.*` lookups that gather each phase's arguments — stays in
`chain_beads.go` because it is a read of node-owned state or a side effect, not because it
sits in a loop.

**Declined, by mechanism, not by category (unchanged from the first pass):**

- `reconcileBeadChain` (`bead_chain.go`) — writes the unexported field `m.beads.beadChains`
  (a map) by direct assignment/mutation, starts goroutines (`beadchain.NewBead(...).Start()`),
  closes channels (`close(c.stops[i])`), and calls `c.group.BroadcastGeometry(...)` (a
  channel-based broadcast). Exporting `beadChains` to let an outside package write it would
  delete the single-writer enforcement `node_geometry.go` already documents for
  `nodeGeometry`'s mutable fields.
- `startBeadDrag`/`endBeadDrag` — each is one line, `c.group.StartDrag()` /
  `c.group.EndDrag()`, a channel-close operation on the shared `BeadWakeGroup` actor state.
  No computation to lift; the body is the side effect.
- `edgeBeadChain` (the type `bead_chain.go` declares) — holds live actor handles
  (`*beadchain.Bead`, `<-chan beadchain.BeadSnapshot`, `chan struct{}` stop channels) as its
  own fields; it is actor bookkeeping, not a data type a pure function could take/return
  without also exporting the goroutine-owning fields above.

**LOC:** `chain_beads.go` 391 → 317 lines (was 388 after the first-pass extraction; the
phase decomposition removed another 71). `beadindex/chain_edge_layout.go`: new file, 122
lines. `bead_chain.go` unchanged, 169 lines (nothing moved out of it — see the declines
above). `nodes/Wiring` non-test top-level `.go` file count: unchanged, 59 (no file moved from
`Wiring`; the new file landed in the already-existing `beadindex` package).

**Verification.** `go build ./...`, `go vet ./...` clean; `go test -race -count=1 ./...`
passes with no failures and no race across every package. The no-imports-`Wiring` loop
(`for p in $(go list ./nodes/Wiring/... | grep -v 'nodes/Wiring$'); do go list -deps "$p" |
grep -qx github.com/dtauraso/wirefold/nodes/Wiring && echo "IMPORTS WIRING: $p"; done`) is
empty. `tools/network/beads/check-no-sqrt-in-chain-beads.sh` re-run clean. Deliberately broke
`ChainEdgeGeometry`'s `count` (`+1`) and confirmed 5 of the 7 `TestChainBeads*` tests in
`chain_beads_geometry_test.go` fail (`TestChainBeadsStayOutsideBothNodes`,
`TestChainBeadsAlwaysAtLeastOneBead`, `TestChainBeadsCountIsSpanProportional`,
`TestChainBeadsExactDoubleTangency`, `TestChainBeadsLastBeadOnTargetTorusOffAxis`), then
restored and confirmed all 7 pass again. No interface, `types`/`common` package, alias shim,
dot-import, package-level actor global, or `ForTest` hatch was added.

**Verification:** `go build ./...`, `go vet ./...` clean; `go test -race -count=1 ./...`
passes with no failures (verbatim `ok` for every package, no race reported). The
no-imports-`Wiring` loop (`for p in $(go list ./nodes/Wiring/... | grep -v 'nodes/Wiring$');
do go list -deps "$p" | grep -qx github.com/dtauraso/wirefold/nodes/Wiring && echo "IMPORTS
WIRING: $p"; done`) is empty.
`tools/network/beads/check-no-sqrt-in-chain-beads.sh` re-run clean (chain_beads.go still
calls no cartesian-sqrt helper directly — the two lifted formulas are plain scalar
arithmetic, no `Normalize`/vector-length inside them). No file was renamed, so no guard
keyed to `chain_beads.go`/`bead_chain.go` needed updating (both are named by
`check-bead-actor-has-call-site.sh` and `check-no-sqrt-in-chain-beads.sh`, unaffected by an
in-place edit). No interface, `types`/`common` package, alias shim, dot-import,
package-level actor global, or `ForTest` hatch was added.

## 12. `scene_structure.go`/`distance_groups.go` re-measured — six pure functions landed (three in existing packages, three in a new `distancegroups` package), everything else declined by mechanism

Measured both files function-by-function.

**First pass declined four `distance_groups.go` functions on their SIGNATURE**
(`*moverRegistry`/`*layoutQuantizer` as a parameter type) rather than their BODY — the exact
mistake the brief calls out and the coordinator caught and reversed. The bodies of
`distanceGroupMax`/`waitForCenterSettle` call exactly one method through the hub
(`mr.centerOfNode(id)`); `applyDistanceGroupTarget` calls that plus `lq.RootMove(ctx, mr,
...)`; `DistanceGroupLens` calls neither `mr.` nor `lq.` at all — it is a wrapper over
`distanceGroupMax` with zero direct hub touches. None of the four bodies read/write an
unexported FIELD; each read is one method call, so each becomes a plain func-value
parameter (`centerOf func(string) (vec3, bool)`, `rootMove func(ctx, target, newPos) bool`),
bound at the Wiring call site — the same pattern `move_dispatch_construct.go` already uses
(`ng.msg.sendMove = md.mr.enqueueFor(ng)`).

**Lifted (second pass, corrected):**

- `kindForID` → `loadspec.KindForID` (`nodes/Wiring/loadspec/builders.go`). Body only calls
  `Buffer.KnownKinds()`/`Buffer.NodeKindID(...)` — no Wiring state.
- `newNodeID` → `loadspec.NewNodeID` (same file). Body is `strconv.Itoa(LargestNodeID(root)
  + 1)` — `LargestNodeID` already lives in `loadspec` (`loader_tree.go`).
- `firstPortOfDir`'s inner scan → `portwiring.FirstPortOfDir(ports []PortSpec, dir PortDir)
  (string, bool)` (`nodes/Wiring/portwiring/port_bindings.go`). The Registry lookup half
  (`Registry[kind]`, keyed by the Wiring-unexported `NodeBuilder` type) stays in Wiring's own
  two-line wrapper; only the pure scan over `[]portwiring.PortSpec` moved.
- `distanceGroupMax`, `DistanceGroupLens`, `applyDistanceGroupTarget`, `waitForCenterSettle`,
  plus the `distancePair`/`distanceGroupOrder`/`distanceGroups` table → new package
  `nodes/Wiring/distancegroups` (`Max`/`Lens`/`ApplyTarget`/`Pair`/`GroupOrder`/`Groups`,
  `waitForCenterSettle` stayed unexported inside the new package). Each now takes
  `centerOf func(string) (wire.Vec3, bool)` (`ApplyTarget` also takes `rootMove func(ctx
  context.Context, target string, newPos wire.Vec3) bool`) instead of `*moverRegistry`/
  `*layoutQuantizer`. Wiring's `distance_groups.go` keeps `DistanceGroupLens`/
  `applyDistanceGroupTarget` as thin wrappers (same exported names/signatures the two
  existing call sites — `move_dispatch_construct.go`'s `DistanceGroupLensFn` bind and
  `stdin_apply.go`'s dispatch — and both test files already used) that build the two bound
  funcs and forward. `wire.Vec3` (imported as `wire` in the new package) is what `vec3` in
  Wiring is already a type ALIAS for (`vec_alias.go`: `type vec3 = wire.Vec3`), the same
  alias `geom` uses for its own exported signatures — no new type crossed a boundary.
  **New package, not an existing one**, because the boundary is genuinely different from
  every existing `nodes/Wiring` subpackage: `geom` is stateless vector math with no domain
  table, `topoderive` computes structural facts once at LOAD time from a `loadspec.TopoSpec`,
  and this is RUNTIME toolbar-panel math (one arrow click) over live per-node centers reached
  through injected accessor/mover functions — no existing package is that boundary.

**Declined, by mechanism, not by category — these hold:**

- `MoveDispatch.CreateNode`/`MoveDispatch.DeleteNode` — both write files
  (`WriteNewNodeFiles`, `RemoveNodeDir`, `edgefile.WriteEdgeFile`/`RemoveEdgesTo`,
  `countspersist.WriteCounts`), call `md.Scenes.Quit()` (ends the process), and read/write
  `md.UI` (`RefuseStructuralEdit`, `EmitViewFrame`, `SceneEditable`, `SceneKinds`,
  `SceneSphere`) — an entry point with side effects on every path, not a computed result.
- `moverRegistry.linkRefusal` (already in `mover_registry.go`) — reads `mr.nodeGeoms`
  (`moverRegistry`'s own unexported MAP field) directly, so the actual statement pinning it
  is a field read, not the receiver type alone; declined for the same mechanism as the build
  cluster (§7).
- `ResolveSceneDistanceGroups` (`distance_groups.go`) — a `*MoveDispatch` method that writes
  three `md.UI` FIELDS (`HasDistanceGroups`, `SceneEditable`, `SceneKinds`) by direct
  assignment. Its own body calls into `scene.SceneHasDistanceGroups`/`SceneIsEditable`/
  `SceneKindMask`, already pure and already in `scene` — nothing further to lift.

**LOC:** `scene_structure.go` 260 → 243 lines. `distance_groups.go` 203 → 76 lines (the
group table and the four math functions moved out; `ResolveSceneDistanceGroups` and the two
thin wrappers stayed). New file `nodes/Wiring/distancegroups/distance_groups.go`, 149 lines.
`nodes/Wiring` non-test top-level `.go` file count: still 59 (no file removed from the
top level; one new file added in a new subpackage, which this task counts separately).

**Verification:** `go build ./...`, `go vet ./...` clean; `go test -race -count=1 ./...`
passes with no failures (verbatim `ok`/`[no test files]` for every package, no race
reported). The no-imports-`Wiring` loop is empty. Deliberately broke
`distancegroups.Groups` (all three groups pointed at nonexistent node ids) and confirmed
`TestRingResolvesItsDistanceGroups` FAILS with `ring streamed all three group lengths as 0 —
the ring owns these groups, so the panel would not render`, then restored and confirmed it
passes again. Deliberately removed the `check-no-wall-clock-wait.py` allowlist entry for
`waitForCenterSettle`'s `time.Sleep` (now at the new path) and confirmed the guard fails,
then restored it and confirmed clean. No interface, `types`/`common` package, alias shim,
dot-import, package-level actor global, or `ForTest` hatch was added. One test file
(`distance_groups_scene_test.go`) had two unexported-symbol references
(`distanceGroups`/`distanceGroupOrder`) updated to the moved package's exported names
(`distancegroups.Groups`/`distancegroups.GroupOrder`) — same assertions, same bodies, no
test renamed/dropped/weakened.

## 13. `nodeGeometry` cluster (`node_geometry*.go`, `node_mover.go`, `edge_mover*.go`,
`pair_node_self.go`) re-measured statement by statement — one pure derivation lifted, the
rest confirmed genuinely single-writer

Re-derived the prior decline rather than trusting it (the task brief's own instruction): the
"7 unexported `nodeGeometry` fields written by direct assignment" cited in §7 belong to
`build_move_dispatch.go` (`nm.flags.coplanarEdges`, `nm.flags.upAxis`, `nm.quantOffset`,
`nm.selfKind`, `nm.tilt.topTiltVectorThetaIdx`, `nm.topo.neighborKinds`,
`nm.outs.outTargets`) — the BUILD cluster, not this one. §7's decline never covered
`node_geometry.go`/`node_geometry_parts.go`/`node_geometry_stream.go`/`node_mover.go`/
`edge_mover.go`/`edge_mover_stream.go`/`edge_mover_run.go`/`pair_node_self.go`; those eight
files had never been measured statement-by-statement before this task.

**Classified every function's body, not its signature or its file's header comment:**

- `nodeGeometry.handle` (`node_geometry.go`) — every branch either assigns an unexported
  field directly (`m.ui.selected`/`.hovered`/`.hoverPort`/`.hoverIsInput`/`.latchedSel`,
  `m.tilt.topTiltVectorThetaIdx`, `m.topo.partnerCenters[msg.SenderID]`) or calls another
  single-writer method (`m.applyCenter`, `m.startBeadDrag`, `m.endBeadDrag`,
  `m.persistTiltVectorAngle`, `m.emitGeometry`, `m.writeStreamFrame`) — 100% bucket (a), a
  pure dispatcher over its own receiver's state.
- `nodeGeometry.emitGeometry` (`node_geometry_stream.go`) — one call, `m.writeStreamFrame`.
  Bucket (a).
- `nodeGeometry.writeStreamFrame` (`node_geometry_stream.go`, 176 lines before this task) —
  the panic guard (reads events, panics — assertion, stays), roughly **100 lines of pure
  arithmetic on field READS** (`m.geom`, `m.flags`, `m.topo.partnerCenters`, `m.tilt`, `m.ui`
  — no assignment to any of them), then the `NodeFrameInput` struct literal build and the
  `m.stream.buildFrame`/`m.stream.streamOut.Write` calls (bucket (a) — the dedicated-fd
  write, single-writer by construction). **Lifted the ~100-line pure middle** —
  `nodegeom.DeriveFrameGeometry` (`nodes/Wiring/nodegeom/frame_geometry.go`), taking a
  `FrameGeometryInputs` (the node's `NodeGeom`, its two scene flags, its `partnerCenters`
  map, its four tilt indices + `receivedVectorSet` + `latticePoints`) and returning a
  `FrameGeometryOutputs` (center, sphereR, pole θ/φ, ring-axis θ/φ, lattice points, and the
  four tilt/received angle+length columns) — pole/ring-axis/tilt-angle derivation that used
  to run inline now calls `nodegeom.NodeWorldPos`/`EffectiveRadius`/`geom.InwardPole`/
  `TorusDefaultAxisAngles`/`UprightRingAxis`/`PoleContainingEdge`/`NodeRadius`, all of which
  already lived in `nodegeom`/`geom` — this call was already cross-package before the move,
  only the orchestrating arithmetic between those calls was inline. `writeStreamFrame` itself
  now: panic guard → one `nodegeom.DeriveFrameGeometry` call → label fallback (2 lines,
  read-only, left inline as not worth a second call) → `m.chainBeads()` (already a separate
  pure function, unchanged) → struct build → write. 176 → ~100 lines.
- `node_geometry_parts.go` — zero functions; every declaration is a type. Nothing to
  classify.
- `node_mover.go` — `nodeMetaFilePath`/`nodeDirPath` are pure `filepath.Join` computations,
  but `.claude/rules/persistence-ownership.md`'s own guard
  (`check-scene-path-resolution.sh`) requires `nodes/` path `Join` calls live ONLY in
  `node_mover.go`/`edge_mover.go`/`loader_tree.go` — moving them would fail that guard by
  construction, so they stay despite being pure. `WriteNewNodeFiles`/`RemoveNodeDir` write
  files (bucket (a), same guard). `newNodeMover` is a two-field literal constructor — no
  computation to lift. `run` is the actor loop: channel receives, `SleepCycle`, wire drive —
  100% bucket (a).
- `edge_mover.go`'s `handle`/`recomputeGeometry` — `handle` assigns `m.selected` or calls
  `nodegeom.SetNodeWorld(&m.srcGeom, ...)` (mutates the receiver's own field through a
  pointer) then `m.recomputeGeometry()`; `recomputeGeometry` calls the already-pure
  `nodegeom.EdgeSegment` (unchanged, already cross-package) then three sends
  (`m.out.PublishSegment`, `m.dest.ReviseInFlightGeometry`, `m.writeStreamFrame`). Both
  functions are field-write/send at every statement — no local arithmetic worth a name.
- `edge_mover_stream.go`'s `writeStreamFrame` — panic guard, one `nodegeom.EdgeSegment` call
  (already pure/cross-package), then `m.dest.DrainPendingEvents()`/`DrainBreadcrumbEvents()`
  — QUEUE DRAINS, which mutate the wire's own pending-event queue on read (the same
  single-goroutine-ownership shape a channel receive has), not a pure read — so the
  `RowEvent` construction immediately downstream of each drain is inseparable from a
  stateful call and was left in place rather than split into a same-file "pure" helper that
  would still take the drained slice as its only real input (measured, not assumed: the
  loop body itself is 6 lines of struct-literal field copies per event, not enough
  computation to justify a second call for its own sake).
- `edge_mover_run.go`'s `run` — actor loop, 100% bucket (a) (channel selects, `SleepCycle`,
  `DriveOneCycle`).
- `pair_node_self.go` — every method (`Breadcrumb`, `EmitGeometryOnce`, `Step`,
  `SetTiltIndex`, `SetRoundsToParallel`, `SetReceivedVector`, `SetLatticePoints`,
  `ClearOutBeads`, plus `MoveDispatch.NodeSelfDriven`/`HasNodeMover`/`NodeQuantOffset`) either
  assigns an unexported `nodeGeometry` field directly (`g.tilt.*`, `g.readout.*`) or is a
  channel-driven actor step (`Step`) or a thin delegator to `md.mr` (kept for the external
  `package main` test callers, per §2a's already-recorded reason — unchanged here). 100%
  bucket (a).

**Moved:** `nodegeom.DeriveFrameGeometry` (+ `FrameGeometryInputs`/`FrameGeometryOutputs`),
`nodes/Wiring/nodegeom/frame_geometry.go`. Nothing else in the cluster.

**Declined, with the exact statement pinning each:** every other function above — see the
per-function breakdown; the pinning statement in every case is either a direct assignment to
an unexported `nodeGeometry`/`edgeMover` field (`m.ui.selected = ...`, `g.tilt.topTiltVectorThetaIdx = theta`,
`m.selected = 1`, etc.), a channel send/receive, a `SleepCycle` call, a file write
(`jsonpersist.WriteJSONAtomic`, `os.RemoveAll`), or a `nodes/` path `Join` the persistence
guard pins to this file by name.

**LOC:** `node_geometry_stream.go` 219 → 149 lines. New file
`nodes/Wiring/nodegeom/frame_geometry.go`, 101 lines. `node_geometry.go` (309),
`node_geometry_parts.go` (275), `node_mover.go` (197), `edge_mover.go` (244),
`edge_mover_stream.go` (98), `edge_mover_run.go` (96), `pair_node_self.go` (257) all
unchanged. `nodes/Wiring` non-test top-level `.go` file count unchanged (59) — this was a
line move within an existing top-level file into an existing subpackage, not a file removal.

**Verification:** `go build ./...`, `go vet ./...` clean. `go test -race -count=1 ./...`
passes with no failures and no race reported (every package `ok` or `[no test files]`). The
no-imports-`Wiring` loop is empty. `nodes/Wiring/node_geometry_lattice_points_test.go`'s two
tests (`TestWriteStreamFrameDefaultLatticeMatchesOldConstant`,
`TestWriteStreamFrameFollowsSetLatticePoints`) already drive
`nodeGeometry.writeStreamFrame` through its own `buildFrame` injection seam and therefore
exercise `DeriveFrameGeometry` end-to-end with no test change needed. Deliberately broke
`DeriveFrameGeometry`'s `TopTiltVectorTheta` line (`out.TopTiltVectorTheta = 0` instead of
`float64(in.TopTiltVectorThetaIdx) * latticeThetaStep`) and confirmed both tests FAIL by
name:
```
--- FAIL: TestWriteStreamFrameDefaultLatticeMatchesOldConstant (0.00s)
    node_geometry_lattice_points_test.go:65: default-lattice angles = (top=0 bottom=1.3089969 coplanar=1.3089969 received=1.3089969), want all 1.3089969 (idx=5 * nodegeom.CurveParamTiltVectorAngleStep)
--- FAIL: TestWriteStreamFrameFollowsSetLatticePoints (0.00s)
    node_geometry_lattice_points_test.go:94: 12-point-lattice angles = (top=0 bottom=2.6179938 coplanar=2.6179938 received=2.6179938), want all 2.6179938 (idx=5 * 2π/12, NOT the fixed nodegeom.CurveParamTiltVectorAngleStep 1.3089969)
```
then restored and confirmed both pass again. No guard is keyed to `writeStreamFrame`,
`node_geometry_stream.go`, or any symbol in the cluster by name (`grep -rl` across
`tools/*/*.sh tools/*/*/*.sh` for every moved/touched function and filename returned empty
before this change, so no guard needed re-keying and none was touched). No interface,
`types`/`common` package, alias shim, dot-import, package-level actor global, or `ForTest`
hatch was added.

## 14. `mover_registry.go`/`move_dispatch_construct.go` re-measured statement by statement —
§12's `linkRefusal` decline corrected, three pure functions lifted

§12 declined `moverRegistry.linkRefusal` on the grounds that it "reads `mr.nodeGeoms`
directly, so the actual statement pinning it is a field read". That measured the whole
function by its one field read, not by counting statements — re-measured: the body is 13
statements, of which **1** touches `mr` (`srcGeom, found := mr.nodeGeoms[src]`) and **9**
are pure string/port-lookup logic (`firstPortOfDir` calls, `fmt.Sprintf` message building,
the found/hasIn/hasOut branches). Same shape `nodeBodyRadius`/`nodeKind` etc. already
show in this file: a real field touch does not make the rest of the body impure.

**Classified every function in `mover_registry.go` and `move_dispatch_construct.go`:**

`mover_registry.go`:
- `bind` — every statement writes into `mr.edgeOut`/`mr.edgeMovers`/`mr.nodeGeoms`-derived
  struct fields (`em.out`, `em.dest`, `srcNM.outs.*`). 100% bucket (a).
- `start` — launches one goroutine per mover. 100% bucket (a) (goroutine start).
- `finalizeActors` — reads `mr.selfDriveClaimed`/`mr.nodeGeoms`, writes `mr.nodeMovers`,
  makes channels. 100% bucket (a).
- `drainCenterMirror` — channel receive into `mr.centerMirror`. 100% bucket (a).
- `centerOfNode` — calls `mr.drainCenterMirror`, reads `mr.centerMirror`. 100% bucket (a).
- `sendMove` — channel send (with ctx-cancel escape). 100% bucket (a).
- `enqueueFor` — returns a closure that appends to `nm.msg.pending`, calls
  `nm.flushPending`, panics past the bound. 100% bucket (a) (field writes + panic
  assertion).
- `nodeKind` — 1 statement, reads `mr.nodeGeoms[nodeID]`. Trivial, not worth splitting.
- `nodeBodyRadius` — 1 statement, `nodegeom.NodeRadius(mr.nodeKind(id))`; the pure call is
  already in `nodegeom`, nothing left to lift.
- `hasNodeMover`/`nodeSelfDriven`/`nodeQuantOffset` — 1-3 statements each, all direct
  `mr.nodeGeoms`/`mr.nodeMovers` reads returned as-is. Trivial.
- `linkRefusal` — **1/13 statements touch `mr`** (the `nodeGeoms` lookup); the other 9
  (port-direction lookups + message formatting) never read or write `mr`. **Lifted** the 9
  into a new package-level `linkRefusalFor(src, srcKind string, srcFound bool, kind
  string)`, called by `linkRefusal` after it resolves `srcKind`/`srcFound` off
  `mr.nodeGeoms`. `linkRefusalFor` stays in `mover_registry.go` (not a subpackage) because
  it calls `firstPortOfDir` (`scene_structure.go`), which reads the package-level
  `Registry` global — a dependency no subpackage may take on `nodes/Wiring` itself.
- `nearestNodeTo` — **1/6 statements touch `mr`** (the `range mr.nodeGeoms` loop header);
  the distance comparison itself (`c.Sub(p)`, `d2`, best-tracking) never reads `mr`.
  **Lifted** to `nodegeom.NearestTo(centers map[string]vec3, p vec3) (string, bool)`
  (`nodes/Wiring/nodegeom/nearest.go`, new file) — `mover_registry.go` now only builds the
  `id -> center` map off `mr.nodeGeoms` and hands it to `nodegeom.NearestTo`. Landed in
  `nodegeom` (existing subpackage, already the home for `NodeWorldPos`/`EdgeSegment`/etc.,
  and depends on nothing but its own `vec3 = wire.Vec3` alias) rather than a new package.

`move_dispatch_construct.go` (`newMoveDispatch`, one function, single-threaded setup):
- The per-node seed loop (label-default, world-center-from-`HasPos`, row-from-id parsing,
  struct literal) — every statement was a pure function of `id`/`i`/`g`/the already-computed
  `row`, appended into `md.GS.NodeSeeds` (the append is the only `md`-touching statement).
  **Lifted** to `geomseeds.BuildNodeSeed(id string, i int, g nodegeom.NodeGeom, row int)
  NodeGeomSeed`.
- The per-edge seed loop (endpoint lookup + `nodegeom.EdgeSegment` + struct literal, with a
  loud error on a missing endpoint) — same shape, pure given `label`/`ep`/the `geoms` map
  parameter (not `md.mr.nodeGeoms` — the load-time map `newMoveDispatch` was already handed).
  **Lifted** to `geomseeds.BuildEdgeSeed(label string, ep inputcodec.EdgeEndpoints, geoms
  map[string]nodegeom.NodeGeom) (EdgeGeomSeed, error)`.
- The mutual-pair detection (`hasEdge` set built from `edgeEndpoints`, then a second pass
  checking the reverse direction) — pure given `edgeEndpoints` alone; the only `md`-touching
  statement was the final `nm.topo.mutualTargets[ep.Target] = true` write, which stayed at
  the call site. **Lifted** to `geomseeds.MutualPairs(edgeEndpoints
  map[string]inputcodec.EdgeEndpoints) map[string]map[string]bool`.
- Everything else in `newMoveDispatch` — the `md := &MoveDispatch{...}` construction, the
  `md.mr.nodeGeoms[id] = ng` node-geometry population loop (closures capturing `md.mr`/
  `md.tapToInstall`/`md.lq`/`md.UI` — these closures ARE the cross-goroutine wiring the
  model requires, not incidental), the partner-center/edge-id seeding loops (write into
  `nm.topo.partnerCenters`/`nm.topo.edgeIDs`, read `md.mr.nodeGeoms`/`md.mr.edgeMovers`),
  and the `RT.Build`/`DistanceGroupLensFn` binding at the end — all bucket (a): every
  statement writes to or reads through `md`/`mr`. `newMoveDispatch` is single-threaded
  SETUP, not an actor loop, but it is still a hub by construction: it exists to populate
  `MoveDispatch`, so most of its body writing to `md` is not a defect the way it would be in
  a request-handling method.

All three `geomseeds` additions live in the SAME existing subpackage `newMoveDispatch`
already imports (`geomseeds.NodeGeomSeed`/`EdgeGeomSeed` — the caller already builds these
exact struct literals, so their own constructors are their type's natural home, same
reasoning as `nodegeom` above). `geomseeds` gained two new imports it previously didn't
need (`nodegeom`, `inputcodec`, `loadspec`, `fmt`) — checked for cycles: neither
`nodegeom` nor `inputcodec` nor `loadspec` imports `geomseeds`, `nodes/Wiring`, or each
other in a way that closes a loop (`go build ./...` is authoritative and stayed clean).

**Declined:** nothing else in either file was a plausible split — see the per-function
breakdown above; every remaining function's own body is 100% mr/md field
touches/channel ops/goroutine starts, not merely "impure at the top level".

**LOC:** `mover_registry.go` 384 → 392 (net +8: the `linkRefusal`/`linkRefusalFor` split
added a function boundary and doc comments; no logic added). `move_dispatch_construct.go`
272 → 240 (−32, the three inlined loops shrank to single calls). New file
`nodes/Wiring/nodegeom/nearest.go`, 19 lines. `geomseeds/geom_seeds.go` 68 → 144 (+76, the
three new pure builders plus their doc comments). `nodes/Wiring` non-test top-level `.go`
file count unchanged (both touched files already existed at the top level; only
`nodegeom/nearest.go`, inside an existing subpackage, is new).

**Verification:** `go build ./...`, `go vet ./...` clean. `go test -race -count=1 ./...`
passes with no failures and no race reported (every package `ok` or `[no test files]`). The
no-imports-`Wiring` loop (`for p in $(go list ./nodes/Wiring/... | grep -v
'nodes/Wiring$'); do go list -deps "$p" | grep -qx
github.com/dtauraso/wirefold/nodes/Wiring && echo "IMPORTS WIRING: $p"; done`) is empty.
`grep -rln` across `tools/` for `nearestNodeTo`, `linkRefusal`, `mutualTargets`,
`BuildNodeSeed`, `BuildEdgeSeed`, `mover_registry.go`, `move_dispatch_construct.go` returned
nothing — no guard is keyed to any of these names or files, none needed re-keying.

Deliberately broke `geomseeds.BuildNodeSeed`'s `Row: row` (changed to `Row: i`, silently
falling back to spec-order position instead of `id-1`) and confirmed
`TestMoveDispatchRowTablesUseNodeIDMinusOne` FAILS by name:
```
--- FAIL: TestMoveDispatchRowTablesUseNodeIDMinusOne (0.01s)
    node_move_row_table_test.go:51: LookupNodeRow(2)=("30",true) want ("3",true)
```
then restored and confirmed the test passes again (`go test -race -count=1
./nodes/Wiring/...` clean, no failures).

**Coverage gap found, not papered over.** Neither `linkRefusal`/`linkRefusalFor` nor
`nearestNodeTo`/`nodegeom.NearestTo` has any direct test, and the one production call site
for both (`scene_structure.go`'s drop-to-create-linked-node path) is reached only through
`MoveDispatch.CreateNode`, whose one existing test
(`refuse_structural_edit_emit_test.go`'s `TestCreateNodeRefusalEmitsViewFrame`) exercises
only the cheapest refusal branch (`!md.ui.sceneEditable`) and never reaches the drop/link
logic at all. This gap pre-dates this task (the split did not remove or weaken any
assertion — there was none to weaken) and is reported per this task's own coverage
requirement, not fixed (adding a `CreateNode`-with-link integration test is a separate,
larger change than a decomposition pass).

No interface, `types`/`common` package, alias shim, dot-import, package-level actor global,
or `ForTest` hatch was added. Tests moved with their subject: none — no test file in either
cluster needed a change (the coverage gap above is exactly why).

## 15. The gesture FSM cluster (6 files, 680 lines) — one genuine lift, the rest re-confirmed pinned, `gesture_state.go`'s stale claim corrected

Re-measured `gesture_actions.go`/`gesture_handlers.go`/`gesture_graph.go`/`gesture.go`/
`gesture_dispatch.go`/`gesture_hitclassify.go` statement by statement per this task's own
"measure bodies, not signatures" instruction, since `gesture_state.go`'s package comment
made a "cannot move without dragging uiState" claim that predates Step 4 of
`docs/planning/gesture-actor.md` (uiState → exported `viewstate.UIState`).

**`beginSphereRotation` lifted** to `gesturefsm.GestureState.BeginSphereRotation` — its body
was already `ui *viewstate.UIState` only to read `ui.VP.Viewpoint`/`ui.Gest` and write
`ui.Gest`'s own fields; changing the first param to `vp geom.Viewpoint` (by value) removed
the `viewstate` dependency entirely, leaving a pure method reading `vp`/its own `Rect` field
and a `heldCenters` closure, writing only `RotPivot`/`RotCx`/`RotCy`/`RotPxPerRad` (its own
type's fields). Two call sites (`gesture_hitclassify.go`'s `"handhold"`/`"empty"` branches)
updated to `g.BeginSphereRotation(md.UI.VP.Viewpoint, heldCentersFn, ev)`.

**The other three uiState-taking leaf actions (`applyNodeDragTarget`, `commitDragStart`,
`setHover`) do NOT move, for a reason narrower than the old comment's:** `viewstate` imports
`gesturefsm` (`UIState.Gest gesturefsm.GestureState`), so `gesturefsm` importing `viewstate`
back is a real cycle, not an unexported-field problem — `uiState` being exported changed
nothing for these three, because each one's body genuinely reads/writes OTHER `UIState`
fields (`ui.Sel.*`, `ui.LastDraggedNode`, `ui.SetHoverUI(...)`, `ui.DragPlaneHit(...)`), not
just `Gest`. A function that needs the rest of `UIState` belongs in `viewstate`, not
`gesturefsm`, regardless of how its parameter is spelled — and moving it into `viewstate`
was out of this task's scope (matching the same "third file forces a surface open" shape
Step 3 of `gesture-actor.md` already declined once). The FSM entry points
(`gestPointerDown/Move/Up`, `HandleRawInput`, `gestHome`, `gestWheel`) and
`updateHover`/`applySelect` stay in package Wiring because, beyond the `UIState` field
reads above, they also reach unexported `MoveDispatch` fields (`md.mr`, `md.lq`, `md.RT`,
`md.ctx`) that cannot be named outside package Wiring at all — the same blocker
`gesture-actor.md`'s "Step 3 — probed and declined" section already found for the
`*moverRegistry`/`*layoutQuantizer` cluster, re-confirmed here rather than assumed.

**Four unused parameters deleted**, each confirmed dead by grep of the function body before
removal: `setHover`'s `tr *T.Trace`, `applySelect`'s `tr *T.Trace`, `gestPointerDown`'s
`tr *T.Trace`, `gestPointerUp`'s `slotReg inputcodec.SlotRegistry`. Removing `applySelect`'s
`tr` and `setHover`'s `tr` left `gestPointerUp`'s and `updateHover`'s own `tr` params
unused too (they existed only to forward); both dropped as well, since neither is
constrained by a dispatch-table type — only the five `rawInputHandlers` map closures are,
and those keep their full `(md, ev, slotReg, tr)` shape unchanged, simply not passing the
now-dropped args through to `gestPointerDown(ev)`/`gestPointerUp(ev)`/`updateHover(ev)`.

**Classification (statement buckets: (a) = writes a uiState/MoveDispatch field, sends on a
channel, starts a goroutine, or writes a file; (b) = computation on locals/params/field
reads):**

| function | file | (a) | (b) | outcome |
|---|---|---|---|---|
| `beginSphereRotation`/`BeginSphereRotation` | gesture_actions.go → gesturefsm | 0 | ~25 | **lifted** |
| `applyNodeDragTarget` | gesture_actions.go | 1 (`rootMove(...)` call) | ~4 | stays — already minimal, the one statement IS the action |
| `commitDragStart` | gesture_graph.go | 2 (`ui.LastDraggedNode = ...`, `sendMoveFn(...)`) | ~2 | stays — mixed, not separable further without a 3-line split not worth a new file |
| `setHover` | gesture_actions.go | 1 (`ui.SetHoverUI(...)`) | ~8 | stays — dedupe compare + RowEvent construction is bucket (b) but the one authoritative write can't leave `viewstate`'s reach |
| `applySelect` | gesture_actions.go | 3 (`setSelectionUI` ×3, `EmitViewFrame` ×3) | ~4 (hit-kind branching) | stays — dominated by (a) |
| `updateHover` | gesture_actions.go | 1 (`EmitViewFrame`) | ~6 | stays — reaches `md.RT`/`md.mr`/`md.ctx` (unexported `MoveDispatch` fields) |
| `seedOrbitPivot` | gesture_actions.go | 1 (`SetViewpoint`) | ~3 | stays — reaches `md.UI.VP` |
| `applyOrbit`/`applyOrbitLocked` | gesture_actions.go | 2 each (`OrbitViewpoint`/`OrbitLockedViewpoint`, `EmitViewFrame`) | ~6 each | stays — the (b) prefix (basis/prev/curr/dir math) is a plausible FUTURE lift as a pure `gesturefsm` helper taking `vp geom.Viewpoint` by value, same treatment as `BeginSphereRotation`; not done this pass to keep the change reviewable as one mechanical idea |
| `gestHome`/`gestPointerDown`/`gestPointerMove`/`gestPointerUp`/`gestWheel` | gesture_handlers.go | 2–6 each | ~4–10 each | stay — each is dominated by dispatch-table walks and `md.mr`/`md.lq`/`md.UI` reach |
| `commitHandholdStart`/`commitRotateStart` | gesture_graph.go | 1 each (`seedOrbitPivot` call, itself (a)) | ~2 each | stay — the pure prefix is 2 lines, not worth a file |
| hit classifiers (`hitClassifiers` map) | gesture_hitclassify.go | 1–2 each | ~2 each | stay — each reaches `md.RT`/`md.mr` |

**`gesture_state.go`'s package comment corrected** — see the file itself; the stale
"cannot move without dragging uiState" sentence is replaced with the actual boundary
(`viewstate`↔`gesturefsm` import cycle + unexported `MoveDispatch` fields), stated so it
does not drift the same way again.

**LOC:** `gesture_actions.go` 204 → 171 (−33, `beginSphereRotation` moved out).
`gesture_handlers.go` 193 → 193 (param deletions, no net line change — same line count,
narrower signatures). `gesture_graph.go`/`gesture.go`/`gesture_dispatch.go`/
`gesture_hitclassify.go` unchanged in line count (dispatch-table call-site edits only).
`nodes/Wiring` non-test top-level file count: unchanged at 59 (no new top-level file; the
new file is `gesturefsm/gesture_camera_seed.go`, inside the existing subpackage).
`gesturefsm` package: 133 → 178 lines (+45, the new file).

**Verification:** `go build ./...`, `go vet ./...` clean. `go test -race -count=1 ./...`:
every package `ok` or `[no test files]`, no failure, no race reported.
`nodes/Wiring/gesturefsm` itself has `[no test files]` — it is exercised only indirectly,
through `nodes/Wiring`'s own gesture tests. The no-imports-`Wiring` loop (same command as
§14 above) is empty.

**Coverage confirmed, not assumed — deliberately broke `BeginSphereRotation`** (changed
`g.RotPivot = pivot` to `g.RotPivot = pivot.Add(wire.Vec3{X: 1})`) and reproduced
`TestGestureEmptyDragOrbits`'s exact failure:
```
--- FAIL: TestGestureEmptyDragOrbits (0.00s)
    gesture_camera_outcomes_test.go:28: rotPivot={1.0000000000000056 5.510910596163082e-15 90} want focus-ahead (0,0,90)
```
then restored and confirmed `go test github.com/dtauraso/wirefold/nodes/Wiring -run
TestGestureEmptyDragOrbits` passes again. This is the only lifted function this pass
produced, and it has a test that can fail by name — no coverage hole on the lifted surface.
The functions that were NOT lifted (declined above) keep whatever coverage they already
had; this task did not audit or change it.

**No guard renamed or re-keyed.** `grep -rln "beginSphereRotation\|gestPointerDown\|gestPointerUp" tools/ --include="*.sh"` and the same names against `tools/` scoped away from
`node_modules`/`topology-vscode` are both empty — no shell guard is keyed to any of these
Go symbol names. (A first, unscoped pass of this grep across all of `tools/` DID hit —
React's own `setHover`/`useState` setters in unrelated `.tsx` panel files, e.g.
`TiltVectorAnglePanel.tsx`, `NodePalette.tsx` — confirmed as unrelated by inspection, not a
guard, before discounting them; recorded here so the "grep returned nothing" claim is
scoped honestly rather than repeating the unscoped false-positive.) None needed re-keying
and none was made to fail-once as a check (there was nothing to prove teeth on).

No interface, `types`/`common` package, alias shim, dot-import, package-level actor global,
or `ForTest` hatch was added. No test was renamed, dropped, weakened, or skipped — none
needed to change since no test named any of the moved/edited symbols directly (they are
reached only through `HandleRawInput`).

## 16. The four scene-level `*Persister` types were one actor written four times — unified

`overlaysPersister`/`sceneSpherePersister`/`speedPersister`/`latticePersister`
(`scene_overlays_persist.go`/`scene_sphere_persist.go`/`scene_speed_persist.go`/
`scene_lattice_persist.go`) each declared an almost-identical unexported type: one
`path string` field, and one method (`schedule`/`flushNow`) that checked `p == nil ||
p.path == ""`, called a package `WriteScene*` func in `scenepersist`, and logged any error
under a fixed site tag. `camerapersist.ViewpointPersister` (already its own lifted package)
has the exact same shape, confirming the pattern rather than inventing it, but stays its own
type since it also owns the FSM-viewpoint→`PolarCamera` conversion, not just a write call.

None of the four had a real debounce despite the historical "schedule" naming — each write
is SYNCHRONOUS, inline on the caller's own (view-owner) goroutine; the prior
debounce/coalescing window was removed earlier and never came back (each file's own header
already said so). So there was no timing/ordering/flush-on-quit semantics to preserve beyond
"nil/unarmed is a no-op, one write call, log on error" — verified by grep before touching
anything, per this task's own "measure bodies" rule, not assumed from the names.

**Unified into `scenepersist.Persister[T]`** (`nodes/Wiring/scenepersist/persister.go`),
parameterized by payload type T with the write function bound as a func value at
construction — the codebase's own bound-func-value pattern
(`move_dispatch_construct.go`'s `ng.msg.sendMove = md.mr.enqueueFor(ng)`), not an
`interface{}`/`any` payload. `Persister[T]{Path, Write func(string, T) error, Tag}.Schedule(v
T)` replaces all four unexported types and their two method spellings (`schedule`/
`flushNow` → one `Schedule`). `move_persist.go`'s `persisters` struct fields and
`EnableEditPersist`'s four constructions now name `*scenepersist.Persister[viewstate.OverlayState]`/
`[geom.SceneSphere]`/`[float64]`/`[int32]` directly — no type alias was introduced to keep
the old short names (an alias would be the exact "package boundary is cosmetic" shim already
removed three times on this branch, §6c). The four now-dead unexported types and their
methods were deleted outright from the four `nodes/Wiring/scene_*_persist.go` files, which
keep only their `MoveDispatch`-facing `Load*`/`SliderSpeed`/`BroadcastLatticePoints` methods
(pinned — each writes `md.UI.*` directly). Every call site (`stdin_dispatch.go`,
`stdin_apply.go`, and 5 test files) updated `.schedule(...)`/`.flushNow(...)` →
`.Schedule(...)`.

`nodes/Wiring/quant_offset_persist.go`'s `nodeGeometry.persistQuantOffset`/
`persistTiltVectorAngle` were measured and left alone — genuinely a different actor: no
debounce type at all, direct methods on the PER-NODE mover (`nodeGeometry`, one goroutine
per node, many concurrent writers to DIFFERENT files) rather than the view-owner goroutine's
single-file writers, and each call does a whole-struct marshal with fields carried forward
(`topTiltVectorThetaIdx`) rather than a bare single-value write — forcing it into
`Persister[T]` would be the "wrong shared abstraction" the task warned against, not a
missed unification.

Verification: `go build ./...`, `go vet ./...`, `go test -race -count=1 ./...` clean (one
transient failure, `TestGestureEmptyDragOrbits`/`TestGestureHomeThenOrbitBuildsOnHomePose`,
traced to the concurrently-edited `gesture_*.go` files mid-WIP in the same shared checkout —
gone on the next run, confirmed unrelated by grep showing neither test references
`persist`/`Persister`). `python3 tools/network/concurrency/check-no-wall-clock-wait.py`
clean; grepping the allowlist for any of the six touched filenames found no entries, so no
re-keying was needed (none of these files ever used a wall-clock wait). The no-imports-
`Wiring` loop is empty. Deliberately broke `Persister.Schedule` (early `return` before the
write) and confirmed `TestPersistOverlaysRoundTrips`, `TestPersistLatticePointsRoundTrips`,
`TestPersistSpeedRoundTrips`, `TestSceneSpherePersisterFlushNow` all fail by name — one test
per persister instance, covering the shared write path for every payload type — then
restored; clean again. Drove a real write through the unified type and read the actual bytes
off disk (`os.ReadFile`, temp probe test, deleted after): `overlays.json` contained
`{"sceneToriVisible":false}` after `ToggleSceneTori`+`Schedule`.

No interface, `types`/`common` package, alias shim, dot-import, package-level actor global,
or `ForTest` hatch was added.

## 17. `edgeMover` lifted into `nodes/Wiring/edgemover` (§13's `edge_mover*.go` trio)

§13 measured `edge_mover.go`/`edge_mover_stream.go`/`edge_mover_run.go` as self-contained
(zero references to `mr.`/`md.`/`lq.`) but left them in place as a line-move-only pass. This
task made the actual package move, as `EdgeMover`.

**16-member classification** (construction-time = assigned once during the single-threaded
wiring phase, before any goroutine starts, and either passed to `New` or set via a one-shot
setter method; post-construction = read or sent on repeatedly while the actor is running):

| member | class | how it crosses the package boundary |
|---|---|---|
| `srcID`/`dstID` | post-construction (read every `flushPending` retry via `resolveDest`) | exported read-only accessors `SrcID()`/`DstID()` |
| `srcH`/`dstH` | construction-time (read once, `bind`) | exported read-only accessors `SrcHandle()`/`DstHandle()` |
| `out` | construction-time (write-once, `bind`) | setter `SetOut` |
| `dest` | construction-time write (`bind`) + one post-construction read (`setEdgeStreams`' `SetStreamsActive` check) | setter `SetDest`, read-only accessor `Dest()` |
| `extIn` | post-construction (sent on every edge click-select, `sendEdgeSelect`) | **channel stays unexported**; wrapped as method `Select(ctx, on)` |
| `srcIn`/`dstIn` | post-construction (sent on every `flushPending` retry, from the node's own goroutine) | **channels stay unexported**; wrapped as methods `TrySendFromSrc`/`TrySendFromDst`, handed back as the bound func value `resolveDest` now returns |
| `stepsIn` | post-construction (sent every `chainBeads` pass, from the source node's own goroutine) | **channel stays unexported**; wrapped as method `SendSteps`, stored as the bound func value in `nodeOuts.outStepsIn` |
| `speedCh` | construction-time (write-once, per-edge speed sink) | setter `SetSpeedCh` |
| `streamOut` | construction-time (write-once, `setEdgeStreams`) | setter, bundled into `SetStream` |
| `edgeRow` | construction-time (write-once, `setEdgeStreams`) | setter, bundled into `SetStream` |
| `nodeRowFor` | construction-time (write-once, `setEdgeStreams`) | setter, bundled into `SetStream` |
| `buildFrame` | construction-time (write-once, `setEdgeStreams`) | setter, bundled into `SetStream` |
| `run` | the goroutine entry point (`moverRegistry.start`) | exported `Run(ctx)` |

**No channel field was exported.** The four genuinely POST-CONSTRUCTION channels
(`extIn`/`srcIn`/`dstIn`/`stepsIn`, all repeatedly sent-on from package `Wiring` while the
actor is running, not just at wiring time) stayed unexported and are reached only through
methods that close over them (`Select`/`TrySendFromSrc`/`TrySendFromDst`/`SendSteps`) — the
same "hand back a bound func value" shape `ng.msg.sendMove = md.mr.enqueueFor(ng)` already
uses in the other direction. This required two real signature changes in package `Wiring`,
not just call-site renames: `nodeMessaging.resolveDest` changed from
`func(id string) (chan movemsg.Msg, bool)` to `func(id string) (func(movemsg.Msg) bool,
bool)` (the node-to-node `neighborIn` branch now wraps its raw channel in a same-shape
closure, `trySendMsg`, so `flushPending` (`node_geometry_retry.go`) has one uniform call
regardless of which kind of destination it resolved), and `nodeOuts.outStepsIn` changed from
`[]chan int` to `[]func(int)`, storing `em.SendSteps` (a method value) instead of the raw
channel — `chain_beads.go`'s call site became `m.outs.outStepsIn[i](count)`. The now-fully-
inlined `nodes/Wiring/stepdeliver` package (its `SendStepsNonBlocking` helper existed for
exactly the one call site this replaced) was deleted; `EdgeMover.SendSteps` carries the same
latest-wins logic.

**The `claimedStream` obstacle.** `streamOut`'s old type, `Wiring.claimedStream`
(`stream_claim.go`), is unexported with an unexported constructor ON PURPOSE (its own header
comment: a narrow capability, no `Unwrap()`), so `edgemover` cannot reference it and cannot
import `Wiring` to ask for one. Duplicated the small mechanism as `edgemover.StreamHandle` +
`edgemover.ClaimRegistry` + `edgemover.Claim` (`nodes/Wiring/edgemover/stream_claim.go`) —
same claim-or-reject-with-stderr-report shape, own registry. `streamWiring` now holds two
claim registries side by side (`claims` for node/view, `edgeClaims edgemover.ClaimRegistry`
for edges) instead of one shared map; this changes nothing observable, since the three kinds
were already namespaced by a `kind:key` prefix that never collided across kinds.

**Exported surface of `EdgeMover`:** `New`, `Run`, `SrcID`/`DstID`/`SrcHandle`/`DstHandle`,
`SetOut`, `SetDest`/`Dest`, `SetSpeedCh`, `SetStream`, `Select`, `TrySendFromSrc`/
`TrySendFromDst`, `SendSteps`, plus the package-level `StreamHandle`/`ClaimRegistry`/`Claim`/
`NewClaimRegistry`/`InboxDepth`. Every export is either a one-shot wiring-time setter/getter
for a value that has no other route across the package boundary, or a method wrapping a
channel operation so the channel itself never crosses — no export exists solely for
convenience.

**Goroutine count, channel set, and send/receive order are unchanged.** Still exactly one
goroutine per edge (`EdgeMover.Run`, launched from `moverRegistry.start`, unchanged call
site shape: `go func() { em.Run(ctx) }()`), still the same four channels per edge
(`extIn`/`srcIn`/`dstIn`/`stepsIn`, same `InboxDepth`/buffer-1 capacities), still the same
select order in `Run`'s drain loop, still the same non-blocking/latest-wins semantics for
every send this move touched (`TrySendFromSrc`/`TrySendFromDst` mirror the old inline
`select`-with-`default`; `SendSteps` mirrors `stepdeliver.SendStepsNonBlocking`'s two-phase
drain-then-retry; `Select` mirrors the old `sendEdgeSelect`'s blocking-with-`ctx.Done()`-escape
send exactly, statement for statement).

**Guards re-keyed and proven with teeth.** `check-scene-path-resolution.sh` and
`check-persist-write-ownership.sh` both match by FILENAME (`edge_mover.go`), and the new
package's file is named `edge_mover.go` too (same filename, new directory, still under
`nodes/Wiring/` recursively) — both guards' `find "$WIRING_DIR" -name "*.go"` walk picks it
up unchanged, no allowlist edit needed. Confirmed both still bite: injected a stray
`filepath.Join("nodes", "x", "y.json")` into `mover_registry.go` — `check-scene-path-resolution.sh`
reported `hand-rolled-node-path: .../mover_registry.go: 27:...` and exited 1; injected a
stray `jsonpersist.WriteJSONAtomic("x.json", nil)` into the same file —
`check-persist-write-ownership.sh` reported `unauthorized-write: .../mover_registry.go:
27:...` and exited 1. Both probes removed immediately after, `go build ./...` and `git diff`
confirmed clean. `check-no-network-locks.sh`, `check-stream-fd-mismatch-reported.sh`,
`check-stream-kind-ts-parity.sh` all re-ran clean, unmodified (none matched any moved
symbol/filename by content).

**Verification:** `go build ./...`, `go vet ./...` clean. `go test -race -count=1 ./...`
passes with no failures and no race reported. The no-imports-`Wiring` loop
(`for p in $(go list ./nodes/Wiring/... | grep -v 'nodes/Wiring$'); do go list -deps "$p" |
grep -qx github.com/dtauraso/wirefold/nodes/Wiring && echo "IMPORTS WIRING: $p"; done`) is
empty — `edgemover` does not import `Wiring`. Deliberately broke `EdgeMover.SrcID`/`DstID`
(swapped to both return `dstID`) and confirmed 7 tests fail by name
(`TestTouchingBeadSourceIsOneBeadLengthFromCentre`,
`TestThirdAtRestIsOneBeadLengthNotSelfTorusR`, `TestAngleGateAdmitsAddAwayAndBlocksAddToward`,
`TestCommitNodeMoveLocalDrawsQuantizedNotRawTarget`,
`TestCommitNodeMoveLocalNeverMovesTowardMouseTarget`,
`TestCommitNodeMoveLocalRemoveTakesBeadsPlace`,
`TestCommitNodeMoveLocalAddMovesOneBeadBeyondNewBead`,
`TestCommitNodeMoveLocalPersistsQuantizedNotRawPolar`), then restored — clean again.

**Uncovered, reported rather than silently accepted.** `TrySendFromSrc`/`TrySendFromDst`/
`Select`/`SendSteps` have no test that can fail on their own delivery outcome — deliberately
made `TrySendFromSrc` always return `false` (never actually send) and the full suite stayed
green. This is not a gap introduced by the move: it is the SAME cross-goroutine-delivery
correctness `docs/process/testing-shape.md`'s own doctrine excludes from testing ("do not
test that two or more goroutines communicate properly — not delivery, not ordering"), now
just living behind a method instead of an inline channel send. `SrcID`/`DstID` (this same
move's other new accessors) ARE covered, as shown above — the split lines up exactly where
the doctrine says it should: pure state reads are testable, cross-goroutine delivery isn't.

`nodes/Wiring` non-test top-level `.go` file count: 3 fewer (`edge_mover.go`/
`edge_mover_stream.go`/`edge_mover_run.go` moved out) plus the deleted `stepdeliver` package
(1 file). No interface, `types`/`common` package, alias shim, dot-import, package-level actor
global, or `ForTest` hatch was added.

## 18. `nodeGeometry` does NOT move out of package `Wiring` — measured, declined

§17 moved `edgeMover` into `nodes/Wiring/edgemover`. The obvious next step was the same move
for the per-node actor (`nodeGeometry`/`nodeMover`) into a `nodeactor` package. It was
measured and declined. Recording it so the measurement is not redone.

**What made `edgeMover` movable.** ZERO production files outside its own three touched an
unexported field. Every external need resolved to one of 16 named members, each cleanly
construction-time-once or a genuinely-repeated post-construction operation — so a constructor
plus ~15 methods covered the whole surface, and all four inbox channels stayed unexported
behind methods that close over them.

**Why `nodeGeometry` is a different shape.** A first pass grepped the literal type name
`*nodeGeometry` in signatures and found 5 external files. That UNDERCOUNTS: it misses every
file that receives an already-typed `nm`/`ng` (ranged out of `md.mr.nodeGeoms`) and then reads
or WRITES its unexported sub-struct fields without re-declaring the type. Grepping by field
selector instead (`.geom`, `.topo.`, `.msg.`, `.outs.`, `.stream.`, `.tilt.`, `.readout.`,
`.clocks.`, `.beads.`, `.flags.`, `.quantOffset`, `.persistRoot`, `.selfKind`) finds 11 files
outside the actor's own files:

`mover_registry.go`, `move_dispatch_construct.go`, `commit_node_move.go`, `touching_beads.go`,
`stream_wiring.go`, `build_move_dispatch.go`, `build_args_selfdrive.go`, `move_streams.go`,
`move_persist.go`, `broadcast_move.go`, `distance_groups.go`.

These are not reads at the edges. `move_dispatch_construct.go` assigns five `ng.msg.*` fields
and creates `nm.msg.neighborIn` channels in both directions; `nm.topo.mutualTargets`,
`edgeIDs`, and `partnerCenters` are read AND written from outside the actor during the
single-threaded wiring pass, and `edgeIDs`/`partnerCenters` again from the runtime commit/drag
path (`commit_node_move.go`, `touching_beads.go`). The touched set spans nearly every one of
the type's named sub-owners.

**Why the §17 remedy does not transfer.** An exported accessor/setter per touched field would
be roughly 25–35 new exported methods, and would convert most of the loader's node-wiring code
from field writes to method calls in the same commit. That is not "move the actor" — it is a
materially larger change that would re-open the reach-back-in shape item 5 declined for
`uiState`, at larger scale (item 5 found ONE file reaching `uiState` by field; this is 11).

**The stop condition that applies.** Not the channel rule specifically — the surface simply
cannot be reduced to a constructor plus a small post-construction API without an oversized
exported surface or a genuine model-shaped change. Per §17's own rule, a correct decline beats
a move that opens the ownership hole.

Nothing was moved. This entry records the measurement, not a plan to revisit it.
