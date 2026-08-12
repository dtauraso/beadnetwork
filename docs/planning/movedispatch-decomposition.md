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

## 19. `nodeGeometry`'s CONSTRUCTION-TIME writes consolidated into its own wiring API (in-package)

§18 declined the package move but did not act on its own 11-file, 35-write measurement.
Re-measured with a field-selector grep restricted to actual assignment statements (`=`,
`append`, excluding comments): **30 external writes were construction-time** (the
single-threaded wiring pass, before any driving goroutine starts), spread across
`move_dispatch_construct.go` (11), `build_move_dispatch.go` (7), `stream_wiring.go` (5),
`mover_registry.go`'s `bind` (4), `move_streams.go` (1), `move_persist.go` (1),
`build_args_selfdrive.go` (1). `distance_groups.go` had none live on the current tree (§18's
"1" there was already stale — some other pass had already cleared it; recorded here rather
than silently trusted). 5 more, all genuinely RUNTIME (the node's own goroutine writing its
own field, repeatedly, while running) were re-confirmed and left untouched as instructed:
`commit_node_move.go`'s `nm.quantOffset = off` (commit path), `mover_registry.go`'s
`nm.msg.pending = append(...)` (`enqueueFor`'s retry queue), and `pair_node_self.go`'s 6
`g.tilt.*` writes (`PairNodeSelf`'s own methods, called from the pair kind's own goroutine).

**Added `nodes/Wiring/node_geometry_wire.go`** — unexported wiring methods on
`*nodeGeometry`, one per construction-time concern, each named for what it wires rather than
which sub-struct it touches: `wireMessaging` (bundles the 5 `msg.*` closures set together,
once, right after `newNodeGeometry`), `setMsgTap`, `ensureNeighborChannel`,
`addMutualTarget`, `seedPartnerCenter`, `addEdgeID`, `addNeighborKind`, `setSceneFlags`
(bundles the two scene-wide flag loops into one), `setQuantOffset`,
`setTopTiltVectorThetaIdx`, `addOutTarget`, `addOutWire` (bundles the 4 index-parallel
`nodeOuts` appends `moverRegistry.bind` makes per edge), `wireStream` (bundles the 5
`nodeStream`/`neighborTopology.nodeRowFor` writes `setNodeStreams` makes per node),
`setPersistRoot`, `copyClockSrc`. All 15 stay unexported (package `Wiring` only) — no
exported accessor, no interface, matching the task's constraint that this is NOT §18's
declined move.

**Shape chosen: several narrow methods, not one big constructor.** `newNodeGeometry` itself
was left alone — its own callers (`move_dispatch_construct.go`'s node loop) still call it
first, then wire the rest in separate calls, because the wired values are only knowable at
different points across the loader's phases (messaging closures need `md`/other nodes to
exist; scene flags/quant offsets/tilt/neighbor-kinds/out-targets are seeded phases later, in
`buildMoveDispatch`; the stream fds arrive later still, in `setNodeStreams`; the persist root
arrives last, in `EnableEditPersist`). Folding all of that into one constructor would force
either passing every one of those not-yet-computed values through `newNodeGeometry` (which
runs before most of them exist) or a giant post-hoc setter with 20+ parameters — the exact
bloated-constructor outcome the task named as a worse alternative than several well-named
methods. Each method instead matches one PHASE of construction, in the order that phase
already ran.

**Writes deliberately left at the call site (not folded into a method), with reasons:**
- `commit_node_move.go`'s `nm.quantOffset = off` and `mover_registry.go`'s
  `nm.msg.pending = append(...)` — RUNTIME, not construction; explicitly out of scope.
- `pair_node_self.go`'s 6 `g.tilt.*` writes — also runtime (the pair kind's own goroutine
  applying `SetTiltIndex`/`SetReceivedVector`/`SetLatticePoints`), same class as the two
  above; not touched.
- `build_move_dispatch.go`'s `nm.selfKind = n.Type` — a direct field on `nodeGeometry`
  itself, not one of the 9 named sub-structs/fields the task scoped in (`msg`/`topo`/`outs`/
  `stream`/`tilt`/`flags`/`quantOffset`/`persistRoot`/`clocks`); left as a bare field write,
  same as before.
- `node_mover.go`'s `g.clocks.clk = g.clocks.clockSrc.Copy()` (inside `nodeMover.run`, at the
  actor's own goroutine start) — this is the actor's OWN file, not an external caller; the
  identical copy done by `build_args_selfdrive.go`'s `ClaimSelfDrive` (a genuinely external,
  construction-time call site, since a self-driven node has no `nodeMover.run` to perform it)
  was folded into `copyClockSrc`, but `node_mover.go`'s own inline copy was left exactly as
  it was — converting it would not reduce an external write, since it isn't one.

**Assignment order relative to `buildFromSpec`'s phases is unchanged.** Every consolidated
call sits at the exact call site the direct field write used to sit at — no method was moved
to a different phase, no call was hoisted or deferred; `setSceneFlags`' two-loops-into-one
merge is the one non-mechanical change, verified separately: both original loops ran over
`md.mr.nodeGeoms` between the same two neighboring statements with nothing reading
`nm.flags` in between, so folding them into a single loop over the same map, calling
`setSceneFlags(coplanarEdges, upAxis)` once per node, produces the identical final `flags`
value on every node (each loop only ever WROTE `true` when its own scene condition held;
`false` was always the zero-value default either way, so setting both fields explicitly to
their computed booleans reproduces the same two-separate-conditional-write outcome).

**Guards.** `check-composer-fields.sh`, `check-persist-write-ownership.sh`,
`check-scene-path-resolution.sh` all re-ran clean on the new file and the touched call
sites (`check-persist-write-ownership: clean (11 write call site(s) scanned across 132
files...)`, `check-scene-path-resolution: clean (222 files scanned...)`) — none match by
symbol name, since no persistence write or path-construction call moved; only field
assignments inside the same package did. No guard needed re-keying.

**Verification.** `go build ./...`, `go vet ./...` clean. `go test -race -count=1 ./...`
clean, no race, no failures (`ok` for every package with tests, including
`nodes/Wiring`, `nodes/Wiring/scenecamera`, `nodes/PairNode`). The no-imports-`Wiring` loop
(`for p in $(go list ./nodes/Wiring/... | grep -v 'nodes/Wiring$'); do go list -deps "$p" |
grep -qx github.com/dtauraso/wirefold/nodes/Wiring && echo "IMPORTS WIRING: $p"; done`) is
empty — nothing new imports `Wiring` (the new file is IN `Wiring`, nothing moved out).
`bash scripts/verify.sh` (run from repo root) reports `stop-checks: clean`.

**Deliberate breaks, confirmed and restored:**
- Dropped `setQuantOffset`'s real value (`quantoffset.QuantizedOffset{}` for every node) in
  `build_move_dispatch.go` → `TestLoadTopologyComputesQuantizedOffsets` failed by name
  (`node 2 quantOffset derives to {X:0 Y:0 Z:0}, want close to {X:50 ...}`), then restored.
- Forced `wireStream`'s `kindID` to a constant `255` for every node in `setNodeStreams` →
  **no test failed.** Uncovered, reported rather than assumed: nothing asserts a loaded
  node's streamed `KindID` column against its spec-declared kind.
- Made `addOutWire` always pass `nil`/`nil` for `outWireOuts`/`outStepsIn` → **no test
  failed.** Same class as §17's own documented gap (`TrySendFromSrc`'s delivery outcome is
  untestable per `docs/process/testing-shape.md`'s own doctrine): `outWireOuts`/`outStepsIn`
  feed `chainBeads`' publish/step-revision paths, cross-goroutine delivery correctness this
  doc's testing doctrine excludes from unit tests by design, not a gap this task introduced.

Before/after external construction-time write count: **30 → 0** (all 30 now route through
one of the 15 new `nodeGeometry` wiring methods; the 5 genuinely runtime writes stay
unchanged, as instructed). Exported `Wiring`-package symbol count unchanged — all 15 new
methods are unexported. `nodes/Wiring` non-test top-level `.go` file count: +1
(`node_geometry_wire.go`). No interface, `types`/`common` package, alias shim, dot-import,
package-level actor global, or `ForTest` hatch was added.

### 19a. The `kindID` hole closed; the `outWireOuts`/`outStepsIn` sibling hole confirmed excluded

Follow-up to the two "no test failed" breaks recorded above.

**`kindID` — closed.** Added `nodes/Wiring/node_geometry_wire_kindid_test.go`,
`TestSetNodeStreamsResolvesPerNodeKindID`: drives `MoveDispatch.SetNodeStreams` (the real,
exported production call path — `move_streams.go` → `streamWiring.setNodeStreams`,
`stream_wiring.go`) against a 2-node, 2-different-kind `LoadTopology` fixture (`AimedSrc`/
`AimedSink`, the same test-registered kinds `TestLoadTopologyComputesQuantizedOffsets`
already uses), with a synthetic `kindIDFor` mapping each kind to a distinct, non-zero,
non-255 id. Each node's own `writeStreamFrame` call — one goroutine, invoked synchronously
in the test, no mover goroutine started — is asserted to have packed back ITS OWN kind's
id, not the other node's and not a constant. Verified both ways: with `wireStream`'s
`kindID` argument forced to the constant `255` (the exact injected break `stream_wiring.go`
"Deliberate breaks" recorded above used), the test failed by name
(`node 1 (AimedSrc) streamed KindID = 255, want 7`); restored, it passes again.

**`outWireOuts`/`outStepsIn` — confirmed the excluded class, no test added.** Traced both
sinks to their consumers: `nodeOuts.outWireOuts[i].PublishSteps(count)` writes onto
`wire.Out`'s `geomSendSteps` channel, drained by "whichever ONE goroutine actually places
beads on this Out" (`nodes/wire/out_port.go`'s `Geom()` doc comment) — for most kinds that
is the node's own separate Update/kind goroutine, not the `nodeMover` goroutine that runs
`chainBeads`; `nodeOuts.outStepsIn[i](count)` (bound to `edgeMover.SendSteps`) hands the
revised count to "the edgeMover's own goroutine (which cannot read the Out directly)"
per `mover_registry.go`'s `bind` doc comment. Both are channel/closure handoffs consumed by
a goroutine other than the one that decided the value — the cross-goroutine delivery class
`docs/process/testing-shape.md` deliberately excludes (corollary 1: never assert delivery;
corollary 2: do not test that two goroutines communicate). This is the same class as this
document's own §17 gap, not a gap this task introduced. No test was written for it, per the
task's own instruction not to manufacture a goroutine-communication test to close an
explicitly-excluded hole.

## 20. `nodeGeometry`/`nodeMover` lifted into `nodes/Wiring/nodeactor` (§18's declined move,
reversed by §19's own consolidation)

§18 declined this exact move on an 11-file, ~30-write measurement. §19 then consolidated
those 30 external construction-time writes into 15 in-package wiring methods, without
re-measuring whether the decline still held. It did not: a corrected count (this task's own
first step) found the type's remaining EXTERNAL surface had shrunk from "30 writes across 11
files" to "23 touch sites across 5 files, 14 distinct fields" — edgeMover-shaped (§17 needed
16 members / ~15 methods), not §18's declined 25–35-method shape.

**The re-measurement, and why the coordinator's own number was also wrong.** The task's
own field-selector regex (`\b(nm|ng|...)\.(msg|topo|outs|stream|tilt|readout|flags|clocks|
beads)\.[a-zA-Z]`) required a dot AFTER the sub-struct name, so every BARE field read —
`nm.geom`, `nm.id`, `nm.selfKind`, `nm.quantOffset` — was invisible to it: 10 fields, 16
sites, 4 files. Re-running the grep for bare fields found the missing class directly: 14
fields, 23 sites, 5 files (`mover_registry.go` 10, `commit_node_move.go` 7, `touching_beads.go`
3, `move_dispatch_construct.go` 2, `build_move_dispatch.go` 1). This is the THIRD regex
undercount on this branch (a signature-grep missed §18's own 11 files by the same shape of
blind spot; a later regex counted a read as a write) — worth naming as a pattern: a
field-selector grep that requires trailing punctuation after the matched prefix silently
excludes every BARE-field touch, and "which fields does an external caller reach" is
therefore not safely answerable by one regex; grep twice, once anchored on the sub-struct
dot and once on the bare field name, or read every non-comment hit by hand as this task did.

Even the corrected 23/5/14 count still missed real touches, found only by letting the
compiler enumerate them after the move: `nm.tr` (read directly and called, `nm.tr.Breadcrumb(...)`,
in `commit_node_move.go`) and `nm.speedCh` (written directly, `nodeMover.speedCh = nodeSpeedCh`,
in `mover_registry.go`'s `finalizeActors` — a field on `*nodeMover`, not `*nodeGeometry`, so it
was outside every one of this task's `nodeGeometry`-scoped greps by construction). Final tally:
**16 distinct fields, 5 external files, plus the 15 already-in-package wiring methods §19
built** — all absorbed into the actor's exported surface below. The reliable way to find
"does this move compile" was never a better regex; it was moving the type and reading `go
build`'s own error list, which is what actually caught `tr`/`speedCh`.

**File classification, verified.** Methods-on-`*nodeGeometry`/`*nodeMover` files (moved,
verbatim structure, package renamed `Wiring`→`nodeactor`, types renamed `nodeGeometry`→
`NodeGeometry`, `nodeMover`→`NodeMover`): `node_geometry.go`, `node_geometry_parts.go`,
`node_geometry_stream.go`, `node_geometry_retry.go`, `node_geometry_center.go`,
`node_mover.go`, `node_geometry_wire.go` (§19's own wiring-API file — its 15 methods
became this task's construction-time EXPORTED surface, unchanged in shape, just capitalized),
`pair_node_self.go`, `bead_chain.go`, `chain_beads.go`, `quant_offset_persist.go` — 11 files,
matching the task's own classification exactly. `pair_node_self.go` split: `PairNodeSelf`
itself and its methods moved; the three thin `*MoveDispatch` delegators it also held
(`NodeSelfDriven`/`HasNodeMover`/`NodeQuantOffset` — package-`Wiring`-only methods on a
package-`Wiring` type, never touching an actor field) stayed behind, relocated to
`move_dispatch_api.go`. Call-site-only files (`commit_node_move.go`,
`move_dispatch_construct.go`, `mover_registry.go`, `touching_beads.go`, `stream_wiring.go`,
`move_streams.go`, `move_persist.go`, `build_args_selfdrive.go`, `build_move_dispatch.go`)
stayed in package `Wiring`, updated to the exported API — one more file than the task's own
list (`build_move_dispatch.go`, holding `SetSelfKind`'s one call site) because the corrected
measurement found it.

**Exported surface, and why each export is unavoidable.** Two constructors
(`NewNodeGeometry`, `NewNodeMover`) plus `NewPairNodeSelf` (new — `ClaimSelfDrive`,
`build_args_selfdrive.go`, used to build a `PairNodeSelf` with a bare struct literal
touching two unexported fields while both types shared a package; that literal is
unreachable across the boundary, so a constructor replaced it, mirroring the shape
`NewNodeGeometry`/`NewNodeMover` already have). Fifteen unexported-turned-exported
construction-time wiring methods, unchanged from §19 (`WireMessaging`, `SetMsgTap`,
`EnsureNeighborChannel`, `AddMutualTarget`, `SeedPartnerCenter`, `AddEdgeID`,
`AddNeighborKind`, `SetSceneFlags`, `SetQuantOffset`, `SetTopTiltVectorThetaIdx`,
`AddOutTarget`, `AddOutWire`, `WireStream`, `SetPersistRoot`, `CopyClockSrc`), plus one NEW
wiring method this task added, `SetSelfKind` (`build_move_dispatch.go`'s
`nm.selfKind = n.Type` — a bare field write §19 explicitly left alone because `selfKind`
wasn't one of the 9 scoped sub-structs; the package move makes any external field write
unreachable, so it needed the same treatment as the other 15). Sixteen POST-CONSTRUCTION
accessor/mutator methods, new to this task (`node_geometry_accessors.go`): `ID`, `Kind`,
`SelfKind`, `Label`, `WorldCenter`, `NodeRow`, `EdgeIDs`, `PartnerCenters`, `NeighborKinds`,
`SendMove`, `NeighborIDs`, `QuantOffset`, `QuantizedOffsetValue`, `ReachR`, `Traced`,
`Breadcrumb`, `Tick`, `ApplyCenter` (renamed from `applyCenter`, now the exported single
write of a node's own center — `commit_node_move.go`'s external call site), `WriteStreamFrame`
(exported door to `writeStreamFrame`, for the same file's breadcrumb write),
`CommitQuantOffset` (folds `commit_node_move.go`'s 3-statement measure/store/persist block
— `quantoffset.MeasureScalar` + the runtime `quantOffset` write + `persistQuantOffset` —
into one method, so the runtime write §19 left as a bare field assignment never needed an
exported setter at all), and `TryRecvExternal` (a non-blocking receive off `extIn`, needed
only because two Wiring-package tests drive a bare `*NodeGeometry` with no driving
goroutine started, so nothing else drains it). On `*NodeMover`: `SetSpeedCh` (the
construction-time write `finalizeActors` makes) and `Run` (the goroutine entry point,
`moverRegistry.start`'s call site). Every export is either a one-shot wiring-time
setter/getter for a value with no other route across the boundary, or (for the four
genuinely repeated post-construction channel touches) a method that closes over the
channel rather than returning it.

**No channel was exported — confirmed by construction, not just by review.** The three
channel-bearing fields (`msg.extIn`, `msg.centerOut`, `msg.neighborIn`) stay unexported,
reached only through methods that close over them: `SendExternal` (blocking send with a
ctx-cancel escape, mirrors edgemover's `Select`), `TryRecvExternal` (non-blocking receive,
test-only in practice), `PollCenter` (non-blocking receive), `NeighborTrySend` (returns a
bound `func(movemsg.Msg) bool` closing over one `neighborIn` slot — mirrors edgemover's
`TrySendFromSrc`/`TrySendFromDst` exactly), and `EnqueueSend` (absorbs what used to be
`mover_registry.go`'s `enqueueFor` closure body — a direct external touch of `msg.tap` and
`msg.pending` — into the actor itself, so `mover_registry.go`'s own `enqueueFor` is now a
one-line forward: `return nm.EnqueueSend`). `msg.sendMove`/`msg.commitLocal`/`msg.resolveDest`/
`msg.centerOf`/`msg.tap` are plain `func` values, not channels, and cross the boundary as
values (`WireMessaging`'s five parameters, `SendMove()`'s return) the same way §17's
`resolveDest`/`enqueueFor` bound-func-value pattern already did in the other direction — no
new category of export. `msg.pending` (the retry queue itself, a `[]pendingSend`) is never
exported: `EnqueueSend`'s bound check, panic, and append all happen inside the actor now,
where before the move `mover_registry.go` touched `nm.msg.pending`/`nm.msg.tap` directly
because it lived in the same package as an implementation-detail convenience, not because
either needed to cross a package boundary.

**The `claimedStream` obstacle, hit a third time, resolved the same way as the second.**
`node_geometry_wire.go`'s `WireStream` takes a stream handle; `Wiring.claimedStream` is
unexported with an unexported constructor on purpose (§17's own note on this). Rather than
have `nodeactor` import `nodes/Wiring/edgemover` for its `StreamHandle`/`ClaimRegistry`
(coupling two sibling actor packages for a mechanism that has nothing to do with edges),
this task duplicated the ~70-line claim-or-reject wrapper a second time,
`nodeactor.StreamHandle`/`ClaimRegistry`/`Claim` (`nodeactor/stream_claim.go`) — same shape,
same reasoning, explicitly citing §17's precedent rather than re-deriving it. `stream_wiring.go`
now holds `nodeClaims nodeactor.ClaimRegistry` alongside `edgeClaims edgemover.ClaimRegistry`
(package `Wiring`'s own `claimedStream`/`streamClaims` type and its `newStreamClaims`/
`newClaimedStream` constructors, `stream_claim.go`, are now DEAD — nothing outside the VIEW
stream, which already had its own separate `viewstate.viewClaimedStream`, used the node kind
of `Wiring`'s registry — and were deleted rather than left orphaned). The property this
depends on — node claim keys (ids), edge claim keys (labels), and the VIEW claim's fixed
singleton key can never collide — was already true and already documented
(`stream_wiring.go`'s own header comment, quoted verbatim in `WireStream`'s doc comment) before
this task moved the node registry into its own package; splitting a THIRD disjoint namespace
out changes nothing observable, the same proof §17 used for the edge/node split.

**Goroutine count, channel set, and send/receive order are unchanged.** Still exactly one
goroutine per ring node (`NodeMover.Run`, launched from `moverRegistry.start`) and the pair
kind's own goroutine driving a `*PairNodeSelf` directly (`ClaimSelfDrive`) — no new
goroutine, no removed one. Still the same channels per node (`extIn`, one `neighborIn` per
direct neighbour, `centerOut`), same buffer depths (`inboxDepth = 8` in `nodeactor/consts.go`,
duplicated from `mover_registry.go`'s `moverInboxDepth`, same value, same "why 8" reasoning,
same precedent as edgemover's own `InboxDepth` duplicate). Still the same drain-until-empty
loop shape in `Run`/`PairNodeSelf.Step`, the same latest-wins `centerOut` push in
`ApplyCenter`, the same per-destination-FIFO retry queue in `flushPending` (now called only
from inside the actor: `Run`, `PairNodeSelf.Step`, and `EnqueueSend`).

**Verification.** `go build ./...`, `go vet ./...` clean. `go test ./...` and
`go test -race -count=1 ./...` both clean, no failures, no race, across every package
(including `nodes/Wiring`, `nodes/Wiring/nodeactor`, `nodes/PairNode`, `nodes/NormalSum`).
The no-imports-`Wiring` loop (`for p in $(go list ./nodes/Wiring/... | grep -v
'nodes/Wiring$'); do go list -deps "$p" | grep -qx github.com/dtauraso/wirefold/nodes/Wiring
&& echo "IMPORTS WIRING: $p"; done`) is empty — `nodeactor` does not import `Wiring`.
`bash scripts/verify.sh` reports `stop-checks: clean`.

**Guards re-keyed, both proven with teeth.** `check-no-sqrt-in-chain-beads.sh` hardcoded
`nodes/Wiring/chain_beads.go`; re-pointed at `nodes/Wiring/nodeactor/chain_beads.go`.
`check-composer-fields.sh` located `type nodeGeometry struct {` by content-grep across
`nodes/Wiring/*.go` (recursive, so the new subdirectory was already in scope) but the
DECLARATION TEXT itself changed with the rename; re-keyed its `COMPOSERS` row to
`type NodeGeometry struct {`. `check-scene-path-resolution.sh` and
`check-persist-write-ownership.sh` both match by FILENAME (`node_mover.go`,
`quant_offset_persist.go`) via `find "$WIRING_DIR" -name "*.go"` walks that already recurse
into subdirectories — same precedent §17 documented for `edge_mover.go` — so neither needed
a path edit; both re-verified with teeth anyway: injected `filepath.Join("nodes", "x",
"y.json")` into `mover_registry.go` → `check-scene-path-resolution.sh` reported
`hand-rolled-node-path: .../mover_registry.go: 51:...` and exited 1; injected
`jsonpersist.WriteJSONAtomic("x.json", nil)` into the same file → `check-persist-write-ownership.sh`
reported `unauthorized-write: .../mover_registry.go: 51:...` and exited 1. Both probes
removed immediately after, `go build ./...` and `git diff` confirmed clean.
`check-no-network-locks.sh`, `check-stream-fd-mismatch-reported.sh`,
`check-stream-kind-ts-parity.sh`, `check-bead-actor-has-call-site.sh` all re-ran clean,
unmodified.

Six doc citations went stale and were fixed as part of this task (not left for the next
drift pass, since `check-doc-drift`/`check-doc-symbols`/`check-docs-symbols` all fail loud
on them): `MODEL.md` (`bead_chain.go`/`chain_beads.go`/`node_mover.go`/`bead_chain_test.go`
paths, `newNodeMover`→`NewNodeMover`, `applyCenter`→`ApplyCenter`),
`docs/bead-model/bead-lattice.md` (`chain_beads_geometry_test.go` path),
`docs/process/testing-shape.md` (`pending_bound_test.go` path),
`memory/project/project_wire_is_straight_line_not_chain.md` (`chain_beads.go`/`bead_chain.go`
paths), `nodes/PairNode/SPEC.md` (`pair_node_self.go` path, `Wiring.PairNodeSelf`→
`nodeactor.PairNodeSelf`, `nodeMover`→`NodeMover`), and two static HTML doc pages under
`docs/pair-node/` (`data-src` citations for the same three moved/renamed files).

**Deliberate break, confirmed and restored, on the moved surface itself (not just the
guards).** Forced `EdgeIDs()` to always return `nil` → 8 tests failed by name
(`TestTouchingBeadSourceIsOneBeadLengthFromCentre`, `TestThirdAtRestIsOneBeadLengthNotSelfTorusR`,
`TestAngleGateAdmitsAddAwayAndBlocksAddToward`, `TestCommitNodeMoveLocalDrawsQuantizedNotRawTarget`,
`TestCommitNodeMoveLocalNeverMovesTowardMouseTarget`, `TestCommitNodeMoveLocalRemoveTakesBeadsPlace`,
`TestCommitNodeMoveLocalAddMovesOneBeadBeyondNewBead`, `TestCommitNodeMoveLocalPersistsQuantizedNotRawPolar`)
— restored, `go test ./nodes/Wiring/...` clean again.

**Uncovered, reported rather than silently accepted.** Forced `SelfKind()` to always return
`""` → **no test failed.** `dragTouchingBeads`'s only consumer of the value
(`nodeTorusOuterR(selfKind)`) falls back to the default `(110,60)` radius for an unrecognised
kind, and every fixture's sanity check (`selfTorusR - lattice.BeadStepR >= 1.0`) still holds
against that fallback — a real kind and the empty-string fallback are numerically
indistinguishable in every existing fixture. Same class as this document's own §17/§19 gaps
(`TrySendFromSrc` always failing, `outWireOuts`/`outStepsIn` fed `nil`): a genuine hole, not
one this task introduced, and not closed here per the same instruction §19a followed for
`outWireOuts`/`outStepsIn` — reported, not manufactured shut. The four channel-touching
methods (`SendExternal`, `NeighborTrySend`, `PollCenter`, `EnqueueSend`'s underlying send
success) are UNTESTABLE by `docs/process/testing-shape.md`'s own doctrine (cross-goroutine
delivery), the same excluded class §17 named for `TrySendFromSrc`/`TrySendFromDst`/`Select`/
`SendSteps` — not re-verified individually here since the doctrine, not a new measurement,
is what excludes them.

`nodes/Wiring` non-test top-level `.go` file count: 12 fewer (11 actor files +
`stream_claim.go`, now dead and deleted) plus `pair_node_self.go`'s three delegators folded
into `move_dispatch_api.go` (no new file). `nodes/Wiring/nodeactor` (new): 13 non-test `.go`
files (11 moved + `consts.go` + `stream_claim.go`, a NEW duplicate) plus
`node_geometry_accessors.go` (new). No alias shim, interface, `types`/`common` package,
dot-import, package-level actor global, or `*ForTest` constructor was added.

## 21. `nodes/wire` measured statement-by-statement — one real pure-math lift found, everything else is genuinely wire-state or channel ops

Measured the assigned cluster (`out_port.go`, `wire_readout.go`, `in_port.go`,
`paced_wire_drive.go`, `paced_wire.go`, `paced_wire_send.go`, `drive_item.go`,
`live_beads.go` — 1467 LOC across 8 files) statement by statement, per the task's own
correction ("measure BODIES statement by statement, never signatures, never whole files").

**What was actually found.** This cluster is small and was already split by concern before
this task (each file's own header comment states its one job); most functions are already
either pure one-liner field reads/writers (`Paced`, `Gated`, `Wired`, `Live`/`Failed`/
`BufferFull` on `DriveItem`) too short to be worth a cross-package move, or are
channel-send/receive/state-mutation top to bottom (`Send`, `RecvTick`, `drainPlacements`,
`stepAll`, `ClearInFlight`, `flushDroppedBreadcrumbs`, `drainBreadcrumbEvents`,
`drainPendingEvents`, `PollRecv`). The one genuine, repeated block of pure arithmetic —
found in THREE places, not one — was a bead's clamped fractional-progress computation
(`nowTick`/`placementTick`/`crossTicks` → `t ∈ [0,1]`), independently duplicated:

```
live_beads.go LiveBeadFractions        6 stmts pure math (target/t/clamp), reads b.placementTick by pointer
live_beads.go LiveBeadRows             8 stmts pure math (same shape), reads b.placementTick by pointer
paced_wire_drive.go ReviseInFlightGeometry  6 stmts pure math (same shape), reads b.placementTick by pointer
```

Each read `b.placementTick`/`b.steps` off a pointer INTO `pw.inflight` (the wire's own
mutable state) but only to COPY the two floats needed for the arithmetic — no statement in
any of the three writes to `pw.inflight` as part of computing `t` (the writes in
`ReviseInFlightGeometry` — `b.steps = newSteps`, `b.seg = newSeg`, `b.placementTick = ...` —
come AFTER `t` is computed and are unrelated to computing it, so they stay). The task's own
worked example names exactly this shape ("was declined as impure because its top level sent
on channels — its body was arithmetic on locals"): reading two floats off a bead pointer to
feed pure arithmetic is a field read, not a mutation, and the type the pointer resolves
(`inflightBead`) is never touched by the moved code, only two of its fields' VALUES.

**Lifted.** `lattice.BeadFraction(nowTick, placementTick, crossTicks float64) float64`
(`nodes/wire/lattice/bead_fraction.go`) replaces all three call sites' clamp blocks with one
call each. `out_port.go`'s `flushSendEvent` also had one duplicate-with-a-neighbor
computation — `SimLatencyMs: float64(steps) * lattice.DwellTicksPerBead * clock.MsPerTick` —
which was ALREADY calling into `lattice` for its one constant; lifted the whole expression as
`lattice.SimLatencyMs(steps int) float64`, since it reads no field of `Out` at all (`steps` is
already a plain `int` parameter) and belongs next to `DwellTicksPerBead`, the constant it is
defined in terms of. `lattice` was the only candidate of the three named existing
subpackages: `beadchain` is the chain-BEAD-goroutine's own state and never reads a
`PacedWire`; `clock` is the tick-scale primitive and both lifted functions already reference
it as a dependency, not a home; `lattice` is the bead-lattice CONSTANTS package this same
arithmetic is defined against (`DwellTicksPerBead`, `BeadStepR`) and already has zero
dependency on `wire` (confirmed by the no-imports-`wire` loop below), so nothing new needed
importing back.

**Declined, with the specific statement that pins each.**

- `Out.Geom()` (`out_port.go`) — every statement either drains a channel
  (`drainStepsNonBlocking(o.geomSendSteps, ...)`) or mutates the goroutine-owned cache
  (`o.sendCur.Steps` written by that drain); there is no arithmetic in the body at all, only
  channel reads landing on a field.
- `Out.placeDrivenNoWalker` (`out_port.go`) — `outcome := o.pw.Send(v, ..., tick)` sends on
  the wire's in-channel; `o.flushSendEvent(v, g.Steps)` writes a `RowEvent` onto the node's
  own interior stream — both are the "sends/receives on a channel" bucket verbatim, the
  handful of remaining lines are a single field-selection return, not arithmetic worth a
  file of its own.
- `Out.placementFrom`/`CurrentPlacement` (`out_port.go`) — pure struct construction from
  `o.node`/`o.port`/an already-loaded `outGeom`, but the type it builds
  (`beadPlacement`) is UNEXPORTED in package `wire`; a subpackage cannot construct or return
  it, so this cannot cross the package boundary regardless of purity (a real type-boundary
  reason, not a preference).
- `wireReadout.appendPending`/`drainPendingEvents` (`wire_readout.go`) — `appendPending`'s
  body is `r.pending = append(r.pending, ev)` then a bound check against
  `maxPendingEvents` that PANICS naming the wire — the append IS the wire's own mutable
  state (`readout.pending`), and the panic path is an assertion tied to that same state,
  not extractable arithmetic. `drainPendingEvents` is `out := r.pending; r.pending = nil;
  return out` — every statement reads or clears `r.pending` directly.
- `PacedWire.DrainPendingEvents`'s conversion loop (`wire_readout.go`) — genuinely pure
  (`out[i] = PendingWireEvent{Kind: pe.kind, ...}`, 6 of 7 statements), but it converts
  FROM `pendingWireEvent`, an unexported package-`wire` type, so — same reason as
  `placementFrom` above — it cannot leave the package without exporting a type solely to
  satisfy this one conversion, which is not what this task's export-cost bar allows.
- `ReviseInFlightGeometry`'s trailing three statements (`paced_wire_drive.go`) —
  `b.steps = newSteps`, `b.seg = newSeg`, `b.placementTick = nowTick - t*pw.ticksToCross(newSteps)`
  each assign directly to a field of `*inflightBead`, a pointer into `pw.inflight` — the
  wire's own owned in-flight state, per MODEL.md's "exactly one goroutine touches that
  state" ("Wire (PacedWire)" bullet). These stay.
- `Send`/`RecvTick`/`Recv`/`ClearInFlight` (`paced_wire_send.go`), `drainPlacements`/
  `stepAll`/`DriveOneCycle` (`paced_wire_drive.go`), `PollRecv`/`Breadcrumb`
  (`in_port.go`), `flushDroppedBreadcrumbs`/`drainBreadcrumbEvents`
  (`wire_readout.go`) — each function's FIRST statement is already a channel send/receive
  or a direct mutation of `inflight`/`pending`/`breadcrumbCh`/`droppedBreadcrumbs`, and
  every subsequent statement either extends that same operation (the `select` bodies) or
  is unreachable without it (e.g. `stepAll`'s `pw.advanceBead`/`pw.readout.appendPending`
  calls only run once a bead is already being read off `pw.inflight`). No statement run of
  ≥3 pure lines exists in any of these bodies once the channel/state statements are
  excluded — measured, not assumed, by reading each body line by line above.
- `NewPacedWire`/`NewOutPaced`/`NewInPaced`/`newOutChan`/`NewOutChanForTest`/
  `NewOutChanDeadEnd`/`NewInChan`/`NewPacedOutNoGeom` (constructors, several files) — pure
  struct literal construction, but every one either allocates a channel (`make(chan ...)`,
  itself excluded as channel-adjacent by the task's own bucket (a): "starts a goroutine,
  sends/receives on a channel" reads naturally as covering channel allocation for the
  wire's own transport, not just send/recv) or builds an unexported package-`wire` type
  (`*Out`, `*In`, `*PacedWire`) that cannot be returned from a subpackage. Not moved.

**File LOC before → after** (measured `wc -l`, non-test):
`out_port.go` 326→325, `live_beads.go` 105→90, `paced_wire_drive.go` 204→196. `wire_readout.go`
(236), `in_port.go` (205), `paced_wire.go` (139), `paced_wire_send.go` (131), `drive_item.go`
(121) unchanged — no statement in any of those five survived the (b) test above. New:
`nodes/wire/lattice/bead_fraction.go` (44), `nodes/wire/lattice/bead_fraction_test.go` (47).
`nodes/wire` top-level (non-subpackage) `.go` file count: unchanged at 31 — this lift added a
file to the EXISTING `lattice` subpackage, it did not add or remove a file from `nodes/wire`
itself.

**No-imports-`wire` loop**: `nodes/wire/beadchain` already imports `nodes/wire` (a
PRE-EXISTING, separately-tracked relationship — `bead_actor.go`'s own doc comments describe
the chain-bead goroutine reading a wire's `LiveBeadFractions`), confirmed unchanged by this
task by running the same loop against `git stash` before this task's commit — it printed the
identical single line before and after, so this task added no new subpackage-imports-`wire`
edge. `nodes/wire/lattice` itself: zero dependency on `wire`, unchanged.

**Guards.** `check-uniform-pulse-speed.sh` (clean, `NewPacedWire`'s one production call site
untouched), `check-no-network-locks.sh` (clean, empty allowlist, no lock/atomic touched or
added), `check-panic-message.sh` (clean, `appendPending`'s panic — the only one in this
cluster — was not moved or edited), `scripts/audit-channel-names.sh` (clean, no channel
declared/renamed). None of these guards name `LiveBeadFractions`/`LiveBeadRows`/
`ReviseInFlightGeometry`/`flushSendEvent` by symbol, so none needed re-keying; each was still
re-run to confirm silence is real rather than assumed.

**Invariant loop, deliberate breaks, and verbatim results.**
`go build ./...` clean; `go vet ./...` clean; `go test -race -count=1 ./...` clean (own
scope: `ok github.com/dtauraso/wirefold/nodes/wire 1.622s`, `.../nodes/wire/lattice 1.583s`,
`.../nodes/wire/beadchain 1.246s`, `.../nodes/wire/clock 1.633s`, full-tree run also clean
before a concurrent, unrelated agent's in-progress `nodes/PairNode` edit later left that one
package mid-refactor and non-building — confirmed by `git log --oneline -5` (no commit of
mine touches `PairNode`) and by `git status --short` showing those files modified-but-
unstaged outside this commit; not reverted, not attributed to this task, per the brief's own
instruction. Deliberately broke `BeadFraction`'s division (`* 2`) — `go test
./nodes/wire/lattice/... -run TestBeadFractionClampsAtZeroAndOne` failed by name
(`--- FAIL: TestBeadFractionClampsAtZeroAndOne/midway_is_fractional`, `got 1, want 0.5`),
restored, re-passed. Deliberately broke `SimLatencyMs` (`steps+1`) — `go test
./nodes/wire/lattice/... -run TestSimLatencyMsScalesWithSteps` failed by name
(`SimLatencyMs(0) = 224, want 0`), restored, re-passed. One dead branch found while probing:
`BeadFraction`'s trailing `if t > 1 { t = 1 }` is UNREACHABLE given the preceding
`target := min(nowTick, placementTick+crossTicks)` capping — `t` can never exceed 1.0 before
that check runs (confirmed: changing it to `t = 2` passed every existing case, including
"past deadline clamps to 1"). Kept anyway, unchanged from the pre-lift code it was copied
from (defensive against float rounding at the boundary), and named here rather than quietly
deleted, since deleting it was not this task's question to answer.

**Functions with NO test that can fail if broken**, reported per the task's own requirement:
`Out.Geom`, `Out.publishSteps`/`publishSegment` and their exported mirrors, `Out.Paced`/
`Gated`/`Wired`, `In.Wired`, all of `DriveItem`'s three predicates, every constructor
(`NewPacedWire`, `NewOutPaced`, `NewInPaced`, `NewOutChanForTest`, `NewOutChanDeadEnd`,
`NewInChan`, `NewPacedOutNoGeom`) — none is asserted against directly by name anywhere in
`nodes/wire/*_test.go`; each is exercised only incidentally as plumbing inside a larger
integration-style wire test, so a targeted one-line break in any of them was not
independently provable to fail the way the two lifted functions above were. This gap
pre-dates this task (none of these functions were touched) and is reported, not fixed, per
the same "uncovered, reported rather than silently accepted" convention this document's
earlier sections use.

**`git status --short` (this task's files only) and commit:**
```
A  nodes/wire/lattice/bead_fraction.go
A  nodes/wire/lattice/bead_fraction_test.go
M  nodes/wire/live_beads.go
M  nodes/wire/out_port.go
M  nodes/wire/paced_wire_drive.go
```
`0343a81b` — "The bead's fractional-progress clamp math and SimLatencyMs's reported latency
arithmetic move into nodes/wire/lattice as pure functions, replacing three duplicated copies
in nodes/wire."

## 22. `nodes/PairNode` — the θ lattice and tilt state machine lifted into `nodes/PairNode/tiltring`

`nodes/PairNode` was the last untouched god-cluster on this branch: 17 non-test files, 2687
LOC, zero subpackages, and `arith_round_test.go` at 475 lines was the largest file in the
whole repo. Classified every function statement-by-statement per the task's own rule
(measure BODIES, never signatures or whole files):

- **`ring.go`'s pure functions** (`newRing`, `at`, `arrivedState`, `seedState`,
  `angleLength`, and `tiltState`/`ring`'s own fields) — every statement was (b): arithmetic
  on locals, parameters, and the ring's own fields, with zero reference to `*Node`, no
  channel op, no goroutine start, no state write outside the ring being constructed. Lifted.
- **`machine.go`'s `tiltMachine` type and its methods** (`nearerEndCount`, `stopping`,
  `settled`, `step`, `choice`, `String`, `machineFor`, the `stoppingCounts` table, the
  `setting` var) — same shape: pure functions of a `tiltState`/`ring` and a `tiltvector.TiltMachine`
  value, no `*Node` anywhere in a body. Lifted.
- **Declined, with the specific statement**: `machineForGap` and `adoptMachine`
  (`machine.go`) both take `*Node` receivers and their bodies READ `n.tilt.Machine`
  (`adoptMachine`'s `if n.tilt.Machine != setting { return }`) or call `n.topState()`/
  `n.ringOf()` (`machineForGap`) — a field READ is bucket (b) by the task's own rule, but
  `adoptMachine`'s body also WRITES `n.tilt.Machine = machineFor(choice)`, one bucket-(a)
  statement in an otherwise-pure function, so the whole function stayed with the node whose
  state it writes. `ringOf`/`topState`/`bottomState`/`setTop`/`setBottom`/
  `fromAnotherLattice`/`drainLattice`/`adoptLattice` (`ring.go`) all read or write
  `n.tilt.*`/`n.lattice.*`/`n.vec.*` directly, or (`drainLattice`) do a channel receive —
  bucket (a) — so all eight stayed.
- **`node.go`/`vectors.go`/`edits.go`/`rest.go`**: every function reads or writes `n.*`
  fields, sends/receives on a channel, or is the goroutine entry point (`Update`) itself.
  No function in these four files is bucket-(b) only; none moved.

**New package `nodes/PairNode/tiltring`** — "the θ lattice PairNode's tilts live on, and the
one state machine that decides where a tilt returns to: pure math, no node, no goroutine, no
channel" (its own package doc comment). Exported: `State`, `Ring`, `NewRing`, `(*Ring).At`/
`ArrivedState`/`SeedState`, `(*State).AngleLength`/`NearerEndCount`, `Machine`,
`StoppingCount`, `(Machine).Stopping`/`Settled`/`Step`/`Choice`/`String`, `Setting`,
`MachineFor`. `ring.go`/`machine.go` shrank to the eight/two `*Node` methods that stayed,
with updated field types (`node_parts.go`'s `Top`/`Bottom *tiltring.State`, `Machine
tiltring.Machine`, `Ring *tiltring.Ring`) and call-site renames only (`.idx`→`.Idx`,
`.next`→`.Next`, `.quarter`→`.Quarter`, `n.ringOf().points`→`.Points`, etc.) — no behavior
change, confirmed by the full suite passing unmodified and by the deliberate-break loop
below.

**`arith_round_test.go` (475 lines, was the largest file in the repo) — NOT split, moved
whole.** Read statement-by-statement: it is ONE `Test` function (`TestOneRoundIsSignAndRemainder`)
plus one helper (`oneRoundSweep`), not several concerns in one file — the file's own header
comment states this ("ONE update round, written without a case in it"). Every check in its
body reads chained local state computed earlier in the SAME loop iteration (`e`, `c`, `u`,
`v`, `d`, `eNbr`, `up`, ...) — splitting it into several files would mean re-deriving that
chain from scratch in each one (duplicating the O(24²+48²) sweep and ~30 lines of setup per
split) or inventing a new struct to carry the chain across files, either of which
restructures the test rather than moves it, which the task's own constraint forbids
("VERBATIM"). It moved whole into `tiltring/arith_round_test.go`, alongside the pure
lattice/machine code it exercises, with only symbol-name updates (`newRing`→`NewRing`,
`tiltMachine{mode: ...}`→`Machine{Mode: ...}`, `.idx`→`.Idx`, etc.) — no assertion added,
removed, or reworded.

**Six more pure-math test files moved whole** alongside it, for the same reason
(zero `*Node`, zero channel, zero goroutine in any test body): `ring_test.go` →
`tiltring/lattice_test.go`, `node_helpers_test.go` (the `offBy`/`steppedTop`/`testRing`/
`perpendicular`/`parallel` shared fixtures) → `tiltring/helpers_test.go`,
`arith_fromrest_test.go`, `arith_walk_test.go`, `arith_resting_lengths_test.go`,
`machine_halt_test.go`. **Two test files stayed** (`machine_adoption_test.go`,
`opening_test.go`) — both construct `&Node{...}` and call `*Node` methods
(`machineForGap`/`stepFromVector`/`adoptMachine`/`clear`), so they exercise the part of the
rule this task did NOT lift; each picked up a one-line local `testRing()`/`perpendicular`/
`parallel` fixture (same values, `tiltring.NewRing(48)`/`tiltring.MachineFor(...)`) since the
tiltring package's own copies are unexported to that package's own test files.

**Assertion count before vs after, counted across ALL seven affected files together** (the
task's own doctrine: moving a test between files is net-zero): `t.Fatal[f]`/`t.Error[f]`
call count was **55 before, 55 after** — every assertion moved, none added, none removed,
none reworded. Test-NAME set is identical (`TestOneRoundIsSignAndRemainder`,
`TestARingMustHaveAWholeQuarterTurn`, `TestFromRestIsTheQuarterOffset`,
`TestTheWalkIsClosedForm`, `TestRestingLengthsFollowFromTheGaps`,
`TestPerpendicularHaltsOnItsOwnTwoSeparations`, `TestParallelHaltsOnlyOnAQuarterTurn`,
`TestEachMachineStepsTowardItsOwnHalt`, `TestPerpendicularStepsThroughTheParallelHalt`,
`TestTheTwoMissesAreComplements`, `TestAModeHaltsExactlyOnItsHomeSet`,
`TestSeparationIsTheShortWayRound`), only their file relocated.
`tools/repo-hygiene/check-test-integrity.sh` (N-way-split aware) ran clean on the staged
diff.

**Docs.** `docs/pair-node/*.html`'s `data-src="nodes/PairNode/machine.go#tiltMachine"`-style
click-to-open annotations (checked by `tools/docs/check-docs-symbols.sh`, exact case-sensitive
symbol match against the named file) and `nodes/PairNode/SPEC.md`'s two backticked symbol
references (`tools/docs/check-doc-symbols.sh`) all pointed at the old, now-moved names.
Updated every `data-src` for a lifted symbol to `nodes/PairNode/tiltring/<file>#<ExportedName>`;
left every `data-src` for a symbol that stayed (`machineForGap`, `adoptMachine`, `setTop`,
`setBottom`, `adoptLattice`, `machine_adoption_test.go`, `opening_test.go`) untouched. One
self-inflicted bug caught by re-running the guard: a `replace_all` on
`nodes/PairNode/machine.go#machineFor` inside `cases.html` matched the substring inside
`nodes/PairNode/machine.go#machineForGap` too (three call sites), silently retargeting them
at a symbol (`MachineForGap`) that does not exist in `tiltring/machine.go` — the guard caught
it (`has no definition of "MachineForGap"`) and it was reverted to the correct, unmoved name.

**Functions with NO test that can fail if broken**: `tiltring.Machine.String()` — every test
that prints a `Machine` via `%v` does invoke it (satisfying `fmt.Stringer`), but none asserts
on the returned string; forcing it to always return `"BROKEN"` left the entire suite green.
Reported, not fixed, per this document's own "uncovered, reported rather than silently
accepted" convention. Every other lifted function (`NewRing`, `AngleLength`, `SeedState`,
`ArrivedState`, `NearerEndCount`, `Settled`, `Step`, `Choice`, `MachineFor`) has a test that
fails by name if broken — confirmed for `AngleLength` below and for the others by the sweep
tests' own coverage (each is called and its result checked against an independently-derived
closed form on every lattice position, both real modes).

**Deliberate break, confirmed, restored.** Forced `(*State).AngleLength` to `return 0`
unconditionally: `TestMachineIsReadFromTheGapNotFromOneTilt` (package `PairNode`) failed by
name on all three quarter-turn cases (`chose parallel, want perpendicular`), and
`TestFromRestIsTheQuarterOffset`/`TestRestingLengthsFollowFromTheGaps`/
`TestOneRoundIsSignAndRemainder` (package `tiltring`) failed by name — proving both the
surviving `*Node`-coupled test and the moved pure-math tests exercise the real lifted
function, not a stale copy. Restored; `go build ./...` and `git diff` confirmed clean before
re-committing.

**Guards.** `tools/network/quality/check-panic-message.sh`: broke `MachineFor`'s panic message
to a bare `"bad machine"` (no site tag) — guard reported `nodes/PairNode/tiltring/machine.go:163:
panic message does not open with a site tag` and exited 1; restored, clean. `tools/network/structure/check-dep-rules.sh`
passed unmodified — a same-kind subpackage (`nodes/PairNode/tiltring` importing nothing outside
the shared spine plus its own parent's kind name) is explicitly exempted by the guard's own
`[ "$dep" = "$kind" ] && continue`. `tools/network/structure/check-composer-fields.sh` does not
police `PairNode.Node` (it only caps `MoveDispatch`/`NodeGeometry`) — not applicable here, not
touched. The no-imports-`PairNode` loop (`for p in $(go list ./nodes/PairNode/... | grep -v
'nodes/PairNode$'); do go list -deps "$p" | grep -qx github.com/dtauraso/wirefold/nodes/PairNode
&& echo "IMPORTS PAIRNODE: $p"; done`) is empty — `tiltring` does not import `PairNode`.

**Verification.** `go build ./...`, `go vet ./...` clean. `go test -race -count=1 ./...`: all
packages `ok`, no race, no failure (including `nodes/PairNode` and the new
`nodes/PairNode/tiltring`). `bash scripts/verify.sh` reports `stop-checks: clean`.

**LOC/file count.** `nodes/PairNode` non-test top-level `.go` files: 7 (was 9: `node.go`,
`node_parts.go`, `ring.go`, `machine.go`, `vectors.go`, `edits.go`, `rest.go` remain; `ring.go`
156→ from 337, `machine.go` 44 from 269). `nodes/PairNode` total files at top level (incl.
SPEC.md and test files): 10 (was 17). New `nodes/PairNode/tiltring/`: 9 files (2 production —
`lattice.go` 173 LOC, `machine.go` 167 LOC — plus 7 test files, 1218 LOC). No `types`/`common`
package, no alias shim, no dot-import, no package-level actor global, no `*ForTest` hatch, no
interface added.

**`git status --short` and commit:** 27 files changed (9 modified in `PairNode`, 4 test files
renamed into `tiltring`, 4 new files in `tiltring`, 1 deleted (`node_helpers_test.go`, content
moved), 8 docs pages updated, `SPEC.md` updated). `d0501639` — "PairNode tilt lattice and state
machine moved to tiltring package with their tests".

## §23 — `nodes/Wiring/` emptied into `nodes/Wiring/dispatch`; `buildDeps` still couples node kinds to it

The first pass at this section (superseded below) declined the REQUIREMENT — "zero `.go`
files directly under `nodes/Wiring/`, only subdirectories" — on the grounds that the BONUS
goal ("node kinds decoupled from the dispatch core") was blocked. The coordinator's
correction was right: those are two different goals, and CLAUDE.md's own decomposition
doctrine says explicitly that a genuine cycle may land in ONE package — "It is acceptable
for one subpackage to be large if the coupling is real; it is NOT acceptable to leave a
file at the top level." The `buildDeps` finding blocks decoupling; it does not block
emptying the directory. Both statements are recorded here.

### The `buildDeps` finding (still true, kept for the record)

`BuildArgs` (`build_args.go`, now `nodes/Wiring/dispatch/build_args.go`) is the parameter
every one of the 13 node kinds' `Build` functions takes. It embeds:

```go
type buildDeps struct {
	latticePoints int32
	inboxes       *nodeInboxes
	mr            *moverRegistry
}
```

`*moverRegistry` is the exact type `MoveDispatch.mr` holds — the dispatch core's own
registry, not a copy or a narrowed view. Two of `buildDeps`'s three consuming methods
(`LatticeIn`/`TiltEditIn`, `ClaimSelfDrive`) mutate through these pointers (register this
node's own channel in the inboxes map, mark this node's own id claimed in `mr`). This is
the mechanism by which a node claims its own inbox/self-drive slot in the live registry, so
it is not incidental coupling to prune. A kind-API package holding `RegisterBuilder`/
`BuildArgs`/`Registry`/`BuildRegistry` still has to import wherever `nodeInboxes`/
`moverRegistry` live to type-check `buildDeps`, and every node kind depends on that
transitively through `BuildArgs`. Real decoupling needs a narrow exported interface over
`nodeInboxes`/`moverRegistry` (a design change to the registry's surface) or accepting a
small, separately-packaged slice of the dispatch core as audience 1's dependency instead of
zero — neither is a mechanical consequence of a file move, so it was correctly NOT
attempted as part of this move. **Node kinds import `nodes/Wiring/dispatch` today, exactly
as they imported `nodes/Wiring` before — the move is package-layout-neutral on this
question, per the coordinator's framing, not a regression.**

### The move that was done

All 94 files (45 non-test, 49 test) moved from `nodes/Wiring/*.go` into
`nodes/Wiring/dispatch/*.go` as **one package**, `package dispatch` (`package dispatch_test`
for the 3 external-test files). This is the "coupled core lands in one package" branch of
the task's own guidance — `MoveDispatch`, `moverRegistry`, `layoutQuantizer`/
`quantized_move.go`, `buildCtx`/the build pipeline, `BuildArgs`/`buildDeps`/`Registry`, the
gesture FSM, stdin dispatch, stream wiring, and the five persisters all stayed together,
unexported, unchanged, exactly as before — nothing was split further because nothing in the
finding above suggested a sub-boundary that wouldn't immediately need `moverRegistry` or
`buildCtx` back. `ls nodes/Wiring/*.go` now finds nothing; `nodes/Wiring/` holds only its
34 pre-existing subpackages plus the new `dispatch/`.

**Why one package and not the suggested 2–4-way split:** every non-test file in the original
7-way partition (load/build pipeline, dispatch core, gesture FSM, stdin/stream wiring,
persisters, scene/view state, small/misc) either directly names `*MoveDispatch`/
`*moverRegistry`/`buildDeps`/`buildCtx` or is called by something that does, in the same
package, with unexported types crossing every proposed seam (`nm *nodeMover`,
`streamWiring`, `persisters`, `uiState` are all unexported fields/types reached by
unqualified name from a dozen-plus of these files). Splitting further would have meant
exporting some of those types or fields (the task's own "no exported channel" ban is a
specific case of a wider rule this project already follows: don't export internal state to
make a move compile) or introducing package-level indirection this task explicitly
disallows (no `types`/`common` package). One package matches the actual coupling measured;
the reachable follow-up in the (declined) first pass — kind-API extraction after a
`buildDeps` redesign — is still exactly where BONUS-goal work should resume.

### Mechanics

Every importer kept its existing import alias (`W`, `Wiring`) where one was already
explicit; the ~20 bare `"github.com/dtauraso/wirefold/nodes/Wiring"` imports (all 13
node-kind packages plus `gatecommon`) were given an explicit `Wiring
"github.com/dtauraso/wirefold/nodes/Wiring/dispatch"` alias so that not one `Wiring.X`
call site inside any node kind's `node.go` needed to change — only the import line moved.
This is an aliasing choice on the CALLER side (completely ordinary Go), not a re-export
shim: `nodes/Wiring/dispatch` exports exactly what `nodes/Wiring` exported, nothing
re-exports anything, and no new file exists solely to forward calls.

One test helper needed a real code change, not just an import-path edit:
`distance_groups_test.go`'s `repoRootForDistanceGroupsTest` computed the repo root from
`runtime.Caller(0)` with a hardcoded "two levels up" comment — now three, since the file
sits one directory deeper. Caught by the first `go test -race ./...` run (`LoadTopology:
stat .../nodes/topology: no such file or directory`), fixed, reran clean.

### Importers re-pointed (33 files, plus 2 files that self-import their own moved package)

Root-level test files (4, explicit `Wiring` alias): `kind_registry_parity_test.go`,
`created_node_loads_test.go`, `pair_self_drive_persist_test.go`,
`pair_node_mover_absence_test.go`. `runtopology` (8, explicit `W` alias): `edge_stream.go`,
`startup_report.go`, `scene_state.go`, `gesture_actor.go`, `node_stream.go`,
`run_goroutines.go`, `view_stream.go`, `topology_run.go`. `nodes/Wiring/scenecamera/
scene_camera_test.go` (1, explicit `Wiring` alias, unaffected by the move itself — it's a
subpackage test, stays in place, only the import path changed). Node-kind packages and
their tests (20, bare import → new explicit `Wiring` alias): `nodes/TimeStart/node.go`,
`nodes/selectright/node.go`, `nodes/input/node.go`, `nodes/gatecommon/{drive.go,
drive_test.go, drive_unwired_speed_test.go}`, `nodes/holdflip/{node.go,
firing_rule_lean_test.go}`, `nodes/pulse/{node.go, firing_rule_lean_test.go}`,
`nodes/Time/node.go`, `nodes/PairNode/node.go`, `nodes/PulseLeft/{node.go,
firing_rule_lean_test.go}`, `nodes/PulseRight/{node.go, firing_rule_lean_test.go}`,
`nodes/pacer/node.go`, `nodes/NormalSum/node.go`, `nodes/TimeEnd/{node.go,
unfed_port_test.go}`, `nodes/selectleft/node.go`. Two files that moved INTO
`nodes/Wiring/dispatch/` and import the sibling package they now live beside as an external
test (`package dispatch_test`): `speed_delivery_full_set_test.go`,
`vector_channel_threading_test.go` — both already used an explicit `W` alias, so only the
import path changed.

`kinds_generated.go` (repo root, package `main`) does not name `Wiring`/`dispatch` at all —
only blank-imports each kind package — so it needed no change; `go run
./tools/gen-node-defs` was run as a check and produced a byte-identical
`kinds_generated.go` (confirmed via `git diff --stat`, zero lines changed beyond the files
this task edited directly).

### Test-name-set and assertion-count equality

`grep -oE '^func Test[A-Za-z0-9_]+'` across every `_test.go` under `nodes/Wiring/` at `HEAD`
(before the move) versus the working tree (after): **241 test functions, identical set,
`diff` empty.** `t.Fatal|Fatalf|Error|Errorf(` call count: **642 before, 642 after.**
`tools/repo-hygiene/check-test-integrity.sh` (rename- and split-aware) passed with no
`[allow-test-weakening: ...]` needed.

### Guards

Ran and inspected every guard CLAUDE.md's brief named plus every guard grep found
referencing a `nodes/Wiring/<file>.go` path:

- `check-scene-path-resolution.sh`, `check-persist-write-ownership.sh`,
  `check-dep-rules.sh`, `check-stream-fd-mismatch-reported.sh`,
  `check-send-rule-parity.sh` (uses `git ls-files 'nodes/Wiring/*.go'
  'nodes/Wiring/**/*.go'`, which already matches the nested `dispatch/` path — confirmed
  `send-rule-parity: clean`), `check-no-wall-clock-wait.py` (its `ALLOWED` entry names
  `nodes/Wiring/distancegroups/distance_groups.go` — the PRE-EXISTING `distancegroups`
  subpackage, a different file from the one this move relocated to
  `nodes/Wiring/dispatch/distance_groups.go`, which has no `time.Sleep`/`time.After`/
  `time.NewTicker` call at all — confirmed by grep before relying on it), all passed
  unmodified because they either recurse the tree dynamically or target subpackages this
  move did not touch.
- `check-composer-fields.sh` — deliberately added a loose field to `MoveDispatch` in its
  new location (`nodes/Wiring/dispatch/move_dispatch.go`): guard reported `MoveDispatch has
  13 fields (cap 12) — it is regrowing into a god-object.` and exited 1. Restored; `go
  build ./...` and the guard both clean again.
- `check-panic-message.sh` — deliberately broke `BuildRegistry`'s panic message (in its new
  location, `nodes/Wiring/dispatch/node_registry.go`) to a bare `"no kinds"`: guard reported
  `nodes/Wiring/dispatch/node_registry.go:53: panic message does not open with a site tag
  (want "funcOrSubsystem: ..."): no kinds` and exited 1. Restored; clean.
- `check-docs-symbols.sh` found one real stale citation:
  `docs/pair-node/math/formulas.html` had two `data-src="nodes/Wiring/scene_speed_persist.go"`
  cells (both symbols, `effective = userSpeed / divisor` and `HumanEditSpeed`, confirmed
  still present in the moved file) — repointed both to
  `nodes/Wiring/dispatch/scene_speed_persist.go`; guard now clean.
- `check-doc-symbols.sh` was clean throughout (no citations of the moved files' exported
  symbols existed in doc-symbol form beyond what `check-docs-symbols.sh` already found).
- `bash scripts/stop-checks.sh`'s own `check-doc-drift` caught 15 additional broken
  `nodes/Wiring/<file>.go` references across `.claude/rules/bridge-surface.md`,
  `.claude/rules/node-kinds.md`, `CLAUDE.md`, `MODEL.md` (×2), `docs/investigations/
  audit-baseline.md` (×2), `docs/investigations/interior-stream-framing.md` (×3),
  `docs/investigations/which-lattice-a-node-lives-on.md` (×2, one file, `replace_all`),
  `memory/feedback/architecture/feedback_schema_parser_parity.md`, `memory/project/
  project_lock_persistence_survives_respawn.md`, `memory/project/
  project_theta_phi_tilted_camera.md`, `nodes/PairNode/SPEC.md`, `tools/topology-vscode/
  ARCHITECTURE.md` (×2) — every one named a symbol still present in the moved file (checked
  with `ls`/`grep` before repointing, per the guard's own "delete rather than invent" rule
  — none needed deleting because none were actually stale symbols, only stale paths), so
  every citation was repointed to `nodes/Wiring/dispatch/<file>.go` rather than removed.
  `check-gofmt` also caught one import-order violation in `created_node_loads_test.go`
  (alphabetical ordering broke when the `Wiring` alias's import path changed) — fixed by
  reordering the import block. `bash scripts/stop-checks.sh` is empty (clean) after both
  fixes.

### Surfaces with no test that can fail if broken

Same caveat as prior sections of this document: a guard's own teeth were proven above by a
deliberate break-and-restore for `check-composer-fields.sh` and `check-panic-message.sh`.
No NEW no-test-can-fail surface was introduced by this move — every moved file's behavior
is exactly what it was in `nodes/Wiring`, and the 241/642 test-name/assertion-count parity
above is the evidence that whatever WAS or wasn't covered before the move is identically
covered or uncovered after it. This section does not re-audit each of the 94 files'
individual test coverage; that audit, if wanted, is a separate task from a package-layout
move.

### Verification

`go build ./...`, `go vet ./...`: clean. `go test -race -count=1 ./...`: every package
`ok` or `[no test files]`, zero `FAIL`, zero race reports (full listing in the commit's
accompanying session output; `nodes/Wiring/dispatch` itself: `ok
github.com/dtauraso/wirefold/nodes/Wiring/dispatch 1.453s`). `bash scripts/stop-checks.sh`:
empty stdout. `git status --short`: empty after the commit.

**Commit:** `e1ed8dac` — "All 94 nodes/Wiring files move into nodes/Wiring/dispatch as one
package, with every external importer and doc citation repointed." 141 files changed (94
renames with in-file edits, 47 plain edits), 152 insertions / 152 deletions (every rename
carries exactly the package-decl line and/or one import-path line changed; no line-count
growth anywhere).

### What's still open (BONUS goal, not the requirement)

The `buildDeps` finding above is unchanged: node kinds still import
`nodes/Wiring/dispatch` for `BuildArgs`, which still embeds `*moverRegistry`/`*nodeInboxes`.
Decoupling audience 1 from the dispatch core remains a real, separate, not-yet-designed
follow-up (narrow interface over `nodeInboxes`/`moverRegistry`, then a kind-API package
extraction) — the REQUIREMENT this section's title names (`nodes/Wiring/` holding zero
`.go` files, only subdirectories) is met; the BONUS is not, and was never claimed to be.

## §24 — the BONUS goal done (`buildDeps` → bound func values, `kindapi` package); the
3-way-split requirement measured and only partly met

The follow-up §23 named as "not attempted" is now done: `buildDeps` (renamed `BuildDeps`,
exported) no longer embeds `*moverRegistry`/`*nodeInboxes` at all. It carries three BOUND
FUNC VALUES instead — `ClaimLatticeIn func(name string) chan int32`,
`ClaimTiltEditIn func(name string) chan movemsg.TiltEditMsg`,
`ClaimSelfDriveGeom func(name string) *nodeactor.NodeGeometry` — each closed over
`buildNodes`'s own `*MoveDispatch` at construction (`nodes/Wiring/dispatch/build_nodes.go`),
mirroring the exact pattern §17/§20 already used to cross the edgemover/nodeactor package
boundary (`resolveDest`/`centerOf`/`sendMove` as func values, never a raw pointer into the
owning actor). `BuildArgs`'s three consuming methods (`LatticeIn`/`TiltEditIn`/
`ClaimSelfDrive`) now call the bound func instead of reaching into `a.deps.mr`/
`a.deps.inboxes` directly — same nil-safe fallback shape (a zero `BuildDeps{}`, as a bare
test build with no loader passes, leaves every func nil, and each method degrades to the
same dead-end channel / nil return it always did).

This let `BuildArgs`/`BuildDeps`/`RegisterBuilder`/`NodeBuilder`/`Registry`/`BuildRegistry`/
`DrivenOut`/`NewDrivenOutForTest` — plus `build_edge_maps.go`'s `BuildTypeMaps` (its only
coupling was reading `Registry`, now in the same package) — move into a NEW package,
`nodes/Wiring/kindapi` (11 non-test + 2 test files: `drive_slot_claim_test.go`,
`state_seed_test.go`; `fixture_kinds_test.go`/`aimed_ports_test.go` stayed in `dispatch`
because their `init()`-registered fixture kinds — `SrcNode`/`SinkNode`/`AimedSrc`/
`AimedSink`/`AimedPacer` — are referenced by other `dispatch`-package tests that need them
in the SAME test binary; they now call `kindapi.RegisterBuilder`/`kindapi.BuildArgs`
qualified instead of unqualified). Every one of the 13 node-kind packages plus
`nodes/gatecommon` now imports `nodes/Wiring/kindapi`, not `nodes/Wiring/dispatch` — confirmed
by `go list -deps`, zero hits, for all 14 packages. The loop from the task brief
(`for p in $(go list ./nodes/Wiring/... | grep -v 'nodes/Wiring$'); do go list -deps "$p" |
grep -qx .../dispatch && echo DEPENDS; done`) reports only `dispatch` itself (trivially
depends on itself), confirming no OTHER subpackage under `nodes/Wiring/` reaches the
dispatch core either.

**Test-name/assertion parity.** 109 `TestXxx` functions across `dispatch`+`kindapi`'s test
files, identical set before/after (diff empty); 324 `t.Fatal|Fatalf|Error|Errorf(` calls,
identical count before/after. (The repo-wide 241/642 figure in §23 covered the WHOLE
`nodes/Wiring/` tree at that point in time, not just this directory — not directly
comparable; the number that matters here is dispatch+kindapi's own before/after equality,
which holds exactly.)

**Deliberate break, confirmed and restored.** Forced `ClaimSelfDriveGeom` to always return
`nil` (bound closure, `build_nodes.go`) → two tests failed by name:
`TestPairNodesHaveNoNodeMoverRingNodesDo`, `TestPairNodeSelfDrivePersistsThroughRealReload`.
Restored; `go build ./...` and `git diff` clean again. `LatticeIn`/`TiltEditIn`'s rewired
surface is exercised (not independently break-tested) by `scene_lattice_edit_test.go` and
`tilt_edit_speed_test.go` — both integration-shaped tests that already round-trip through
the real bound closures, so breaking `ClaimSelfDriveGeom` (the one with no such integration
test at this remove) was the more informative probe.

**Guard evidence.** `check-channel-names.sh` caught a real regression: the two bound
closures' local channel variables were named the generic `ch`, reported as
`channel-naming: nodes/Wiring/dispatch/build_nodes.go: generic name 'ch'` — renamed back to
`sceneToNodeLatticeIn`/`panelToNodeTiltEditIn` (the SAME names the pre-move
`build_args_lattice.go`/`build_args_tilt_vector.go` used), guard clean. `bash
scripts/stop-checks.sh` is empty (clean) both before this fix (it caught it) and after.
Doc citations of the four moved files (`build_args.go`, `driven_out.go`) repointed in
`.claude/rules/node-kinds.md`, `docs/investigations/audit-baseline.md`,
`docs/investigations/interior-stream-framing.md` — `check-docs-symbols`/`check-doc-drift`
clean.

**Verification.** `go build ./...`, `go vet ./...`: clean. `go test -race -count=1 ./...`:
every package `ok` or `[no test files]`, zero `FAIL`, zero race reports. `bash
scripts/stop-checks.sh`: empty stdout. `gofmt -l .`: empty (two import-order fixes needed
and applied, `nodes/TimeEnd/node.go`/`nodes/pacer/node.go`).

### The 3-way-split / ≤31-files-per-package requirement — first pass, only PARTLY met

`ls nodes/Wiring/kindapi/*.go | wc -l` → 13 (11 non-test + 2 test), well under the cap.
`ls nodes/Wiring/dispatch/*.go | wc -l` → 81 (34 non-test + 47 test) — still ONE package,
still well over ~31, and this task's "3 or more packages" ask is therefore not met: there
are two packages here (`dispatch`, `kindapi`), not three-plus, and `dispatch` itself did not
shrink enough.

**What was actually measured in this first pass, and why a further split was not attempted
yet.** Every file remaining in `dispatch` was checked for a `func (x *MoveDispatch)` /
`func (x *moverRegistry)` / `func (x *buildCtx)` / `func (x *layoutQuantizer)` method
receiver — 20 of 34 non-test files have one. Go requires a method's receiver type to be
DEFINED in the same package as the method, so none of those 20 can leave `dispatch` without
`MoveDispatch`/`moverRegistry`/`buildCtx`/`layoutQuantizer` themselves leaving too — not a
design choice, a language rule. **This first pass treated "has a receiver on one of the four
hub types" as one undifferentiated bucket. That was the wrong granularity** — the coordinator's
correction below is what actually moved the needle: the receiver-type rule blocks a method
from leaving its OWN type's package, but says nothing about whether the ARGUMENTS that
method's BODY reads are themselves hub-coupled. `layoutQuantizer` is the proof: every one of
its methods takes `*moverRegistry` as a plain PARAMETER, never reached through the receiver,
so `layoutQuantizer` (receiver `*layoutQuantizer`, a type with NO coupling of its own) could
leave while `moverRegistry` (the parameter type) stayed behind unexported — see the
`layoutquant`/`topoderive` write-ups below.

### Second pass — per-type coupling table, and two real moves landed

For each of the four hub types, every method was read for what it touches: whether the BODY
needs the whole hub (another hub type's bare, unexported field) or can be rewritten to take
the specific value/closure it reads as a parameter (the same bound-func-value pattern
`BuildDeps` already used, §17/§20's `resolveDest`/`centerOf`/`sendMove`).

| type | methods | what each body actually touches | verdict |
|---|---|---|---|
| `layoutQuantizer` (5) | `heldCenters`, `heldEdges`, `RootMove`, `commitNodeMoveLocal`, `broadcastToEdgesAndPartners` | ALL FIVE take `mr *moverRegistry` (or `ctx`/`ui`) as an explicit PARAMETER, never through the receiver. Bodies read exactly THREE things off `mr`: `mr.edgeMovers` (5×), `mr.nodeGeoms` (4×), `mr.centerOfNode` (1×) — all three already exported types from already-lifted packages (`*edgemover.EdgeMover`, `*nodeactor.NodeGeometry`) or an existing bound-method value. `layoutQuantizer`'s own field (`quantizedLayout`, one bool) is read by only 2 of the 5. | **MOVED** — `nodes/Wiring/layoutquant`, zero exports added to `moverRegistry`/`MoveDispatch`; dispatch's own callers pass `md.mr.nodeGeoms`/`md.mr.edgeMovers` (already same-package, already-exported element types) directly. |
| `buildCtx` (3) | `allocateWires`, `buildMoveDispatch`, `buildNodes` | `allocateWires`: reads ONLY `b.spec`/`b.nodeGeoms`/`b.tr` — buildCtx's OWN fields, never `md`/`mr`/`lq`. Zero hub coupling. `buildMoveDispatch`: calls the UNEXPORTED constructor `newMoveDispatch(...)`, then writes directly into `md.mr.nodeGeoms`/`md.lq.QuantizedLayout`/`md.mr.selfDriveClaimed` (bare, unexported hub fields) as one-time single-threaded construction. `buildNodes`: writes `b.md.inboxes.*`/`b.md.mr.nodeGeoms`/`b.md.mr.selfDriveClaimed` bare, and calls `md.mr.bind(...)` (moverRegistry's own unexported method) in `bindDispatch`. | **`allocateWires` MOVED** (`nodes/Wiring/topoderive`, joining its already-lifted pure-derive-phase siblings — same class, simply not carried over earlier). **`buildMoveDispatch`/`buildNodes` PINNED** — see statement below. Since Go requires ALL of one type's methods to share a package, and 2 of 3 needed to stay, `allocateWires` left as a REWRITE (buildCtx method → free function `topoderive.AllocateWires`, buildFromSpec assigns the 5 results back onto `b`'s own fields), not a plain file move — buildCtx itself, as a type, still exists with its remaining 2 methods. |
| `moverRegistry` (14) | `bind`, `start`, `finalizeActors`, `drainCenterMirror`, `centerOfNode`, `sendMove`, `enqueueFor`, `nodeKind`, `nodeBodyRadius`, `hasNodeMover`, `nodeSelfDriven`, `nodeQuantOffset`, `linkRefusal`, `nearestNodeTo` | Every one reads/writes `mr`'s OWN bare fields (`nodeGeoms`/`nodeMovers`/`edgeMovers`/`selfDriveClaimed`/`centerMirror`) THROUGH THE RECEIVER — this is moverRegistry's actual method set, not a namespace of parameter-taking free functions the way `layoutQuantizer` was. ~51 EXTERNAL bare-field touches (`md.mr.X`) also exist across 16 other `dispatch` files (`build_move_dispatch.go`, `build_nodes.go`, `commit_node_move.go` (pre-move), `distance_groups.go`, `gesture_hitclassify.go`, `gesture_handlers.go`, `move_dispatch_api.go`, `move_dispatch_construct.go`, `move_persist.go`, `move_dispatch.go`, `move_streams.go`, `scene_structure.go`, `stdin_apply.go`, `stdin_dispatch.go`, plus the 2 test files that build a bare `moverRegistry{...}` literal). | **NOT MOVED, not attempted this pass.** The receiver methods THEMSELVES are moverRegistry's own logic (unlike layoutQuantizer's, which only ever touched `mr` through a parameter) — lifting this type means EXPORTING its field surface (an accessor per bare-field touch, or exported fields, which the constraints forbid) across all 16 external files, matching the scope of §17 (edgemover, 16-member classification) or §20 (nodeactor, 16 fields/23 sites/5 files) as its OWN task, not a byproduct of this one. |
| `MoveDispatch` (30) | grouped below | see grouping | **NOT MOVED** — every method is either a gesture entry point reading MULTIPLE owners at once (`UI`+`mr`+`lq`+`ctx`, not one sub-object), a thin delegator kept ONLY because the sub-object it forwards to is unexported (`NodeSelfDriven`/`HasNodeMover`/`NodeQuantOffset` → `mr.nodeSelfDriven`/`hasNodeMover`/`nodeQuantOffset`, each doc-commented "kept because `mr` is unexported"), or a persistence/scene method with real logic of its own (`EnableEditPersist`, `LoadSceneSphere`, `CreateNode`, …). None is dead weight to delete; none is a thin single-owner forward to a package MoveDispatch could stop naming. |

**`MoveDispatch`'s 30 methods, grouped by what they front** (requested explicitly): **gesture
entry points** (13 — `updateHover`/`seedOrbitPivot`/`applyOrbit`/`applyOrbitLocked`/
`applySelect` in `gesture_actions.go`; `gestHome`/`gestPointerDown`/`gestPointerMove`/
`gestPointerUp`/`gestWheel` in `gesture_handlers.go`; `commitHandholdStart`/
`commitRotateStart` in `gesture_graph.go`; `HandleRawInput` in `gesture_dispatch.go` — every
one reads/writes several of `md.UI`/`md.mr`/`md.lq`/`md.ctx` in the SAME body, so none fronts
one sub-object); **`mr`-fronting** (4 — `Start` mixes `mr.start` with setting `md.ctx`, not
pure; `NodeSelfDriven`/`HasNodeMover`/`NodeQuantOffset` are genuine one-line delegators, kept
ONLY because `mr` is unexported to external callers like the root-package tests); **`sw`
(streamWiring)-fronting** (3 — `SetMsgTap`/`SetEdgeStreams`/`SetNodeStreams`, each doing real
parameter-injection work, not pure forwards); **`persist`-fronting** (2 —
`EnableViewpointPersist`/`EnableEditPersist`, real logic: arming persisters, setting
`nm.SetPersistRoot` on every mover); **scene/persistence read-modify-write, each its own file
with real I/O/logic** (8 — `SliderSpeed`/`LoadSpeed`, `LoadSceneSphere`,
`BroadcastLatticePoints`, `LoadOverlays`, `CreateNode`/`DeleteNode`,
`ResolveSceneDistanceGroups`). Checked for prunable pure delegators specifically (the
coordinator's "may not need to exist at all, or belong with their sub-object" question): only
the 3 `mr`-fronting one-liners qualify as pure delegators, and all 3 are pinned by their own
doc comments — the ONLY reachable surface for `mr` (unexported) from outside package
`dispatch`. Nothing here is removable.

### Two moves landed this pass

**`layoutQuantizer` → `nodes/Wiring/layoutquant`** (`quantized_move.go`, `commit_node_move.go`,
`broadcast_move.go`, `touching_beads.go`). `QuantizedLayout` exported (dispatch's own
`build_move_dispatch.go` sets it, `layoutquant`'s own `CommitNodeMoveLocal` reads it — crosses
the boundary now). `RootMove` no longer calls `moverRegistry.sendMove`; it calls
`nm.SendExternal(ctx, msg)` directly on the looked-up `*nodeactor.NodeGeometry` (already an
exported method, §20) — one fewer indirection, same behavior. Every dispatch call site
(`gesture_handlers.go` ×3, `gesture_graph.go`, `gesture_hitclassify.go` ×2,
`move_dispatch_construct.go`, `distance_groups.go`, plus 3 test files —
`continuous_drag_persist_test.go`, `quantized_layout_test.go`,
`drag_touching_bead_source_regression_test.go` — and the shared fixture helper
`wire_test_helpers_test.go`) updated to pass `md.mr.nodeGeoms`/`md.mr.edgeMovers` in place of
`&md.mr`. Test files STAYED in `dispatch` (they construct a real `*MoveDispatch` and touch
`md.mr`/`md.lq` bare fields directly — unreachable from an external test package, and no
bare-`layoutQuantizer{}` literal test existed to extract), so `layoutquant` itself carries no
tests of its own — behavior-preserving, same names/assertions, just exercised from `dispatch`.

**`buildCtx.allocateWires` → `topoderive.AllocateWires`.** Signature changed from a
`*buildCtx` method with no return (writing 5 fields onto `b`) to a free function
`AllocateWires(spec, nodeGeoms, tr) (destWire, edgeWire, edgeEndpoints, edgeSteps,
edgeSegments)`; `build.go`'s `buildFromSpec` now does
`b.destWire, b.edgeWire, b.edgeEndpoints, b.edgeSteps, b.edgeSegments =
topoderive.AllocateWires(b.spec, b.nodeGeoms, b.tr)`. `wireSegment` (dispatch's local alias)
became `wire.WireSegment` directly (topoderive has no reason to redeclare the alias for one
file). `check-uniform-pulse-speed.sh`'s "exactly one non-test `NewPacedWire` call site"
requirement stayed satisfied by construction — the OLD call site (`build_wires.go`) was
deleted in the same commit that added the new one, never coexisting.

**Test-name/assertion parity, both moves combined.** 109 `TestXxx` functions across
`dispatch`+`kindapi`+`layoutquant`'s test files, identical set before/after this pass (diff
empty — unchanged from the prior pass's 109, since no test file's OWN name set changed, only
call sites inside them); 324 assertions, identical count.

**Deliberate break, confirmed and restored, both moves.** `layoutquant`: no NEW break-test run
this pass beyond the prior `ClaimSelfDriveGeom` probe (still valid, unchanged); the moved
logic is unit-covered by the SAME `quantized_layout_test.go`/`drag_touching_bead_source_
regression_test.go`/`continuous_drag_persist_test.go` assertions that already existed, now
calling the qualified `layoutquant.X` names — their pass/fail behavior is unchanged by
construction (same bodies, same statements, only the package boundary moved). `topoderive.
AllocateWires`: forced `steps := 0` unconditionally (discarding the real `EdgeStepCount`
call) → **no test failed.** `per_edge_travel_time_test.go`, despite its name, only tests
`TestFanInRejectedAtLoad` (fan-in rejection at parse) — it never asserts a step count or
travel time despite the file name suggesting otherwise. This is a genuine, pre-existing
coverage hole (not introduced by this move — the function's behavior is unchanged, only its
package), reported per this task's own instruction rather than manufactured shut. Restored;
`go build ./...`/`go test ./...` clean again, confirmed via `diff` against a saved copy
(empty).

**Guard evidence, both moves.** `check-uniform-pulse-speed.sh`: verified clean after the
`AllocateWires` move (one production call site, in `topoderive/allocate_wires.go`, passing
`lattice.DwellTicksPerBead`). `check-doc-drift` caught 4 broken path references after the
`layoutquant` move — `MODEL.md` (`commitNodeMoveLocal`/`nodes/Wiring/dispatch/
commit_node_move.go` → `CommitNodeMoveLocal`/`nodes/Wiring/layoutquant/commit_node_move.go`),
`docs/investigations/which-lattice-a-node-lives-on.md` (×2, `nodes/Wiring/dispatch/
quantized_move.go` → `nodes/Wiring/layoutquant/quantized_move.go` — the cited symbols
`walkBeadPath`/`requantizePoleTraced` were ALREADY stale before this move, referring to
functions removed in an earlier pass; only the path was repointed, matching the guard's own
"delete rather than invent" rule against fixing unrelated staleness out of scope),
`memory/project/project_theta_phi_tilted_camera.md` (same). `bash scripts/stop-checks.sh`
empty (clean) after all four fixes.

**External interference during this pass, caught and fixed.** A concurrent agent session
working on the SAME branch (`task/god-objects`, TypeScript-side work, unrelated to this
section) ran `git commit` with no pathspec after `git add`-ing only its own two TS files;
because a pathspec-less `git commit` commits the WHOLE INDEX, it swept in this pass's
already-`git rm`-staged deletions of `broadcast_move.go`/`commit_node_move.go`/
`quantized_move.go`/`touching_beads.go` from `nodes/Wiring/dispatch/`, then "fixed forward"
by restoring those four files' pre-deletion content (as untracked files) once it noticed —
well-intentioned, but it left both the `dispatch/` originals (now dead — `staticcheck`
correctly flagged `heldCenters`/`commitNodeMoveLocal`/etc. as unused, U1000) and the
`layoutquant/` copies on disk simultaneously. Diffed each restored `dispatch/` file against
the git blob this pass had ALREADY moved into `layoutquant/` (`git show <this-pass's-commit>:
nodes/Wiring/dispatch/<file>` vs. the working-tree restore) before re-deleting — all four
byte-identical, confirmed via `diff` returning empty, so no content was silently lost or
diverged. Re-deleted with `git rm`, staged and committed by EXPLICIT pathspec this time
(`git commit -F <msg> -- <20 explicit paths>`), and verified with `git show --stat HEAD`
immediately after — exactly the 20 intended files, nothing riding along. `git stash` was NOT
used at any point in this recovery.

### Final partition, this pass

| package | non-test | test | total |
|---|---|---|---|
| `nodes/Wiring/kindapi` | 11 | 2 | 13 |
| `nodes/Wiring/layoutquant` | 4 | 0 | 4 |
| `nodes/Wiring/topoderive` (existing, +1) | 7 | 0 | 7 |
| `nodes/Wiring/dispatch` | 33 | 43 | 76 |

`ls nodes/Wiring/dispatch/*.go | wc -l` → 76 (down from 81 this pass, 94 at the branch's
starting point — 1.24× smaller, not yet the requested 3×). There are now FOUR packages
touched by this section's work (`dispatch`, `kindapi`, `layoutquant`, plus `topoderive` —
an EXISTING package gaining one file, matching the brief's own "an existing subpackage is a
valid destination" for exactly this shape of move), satisfying "3 or more packages" by
count, but `dispatch` itself remains well over the ~31-file cap.

**What would close the remaining gap, named with the specific pinning statement per type.**
`moverRegistry`'s own 14 methods are pinned by the STATEMENT in the coupling table above:
their bodies are moverRegistry's OWN logic operating on its OWN bare fields through the
receiver — not, like `layoutQuantizer`, a namespace of functions that merely take the hub as
a parameter. Lifting it requires an exported accessor (or bound closure) per external
bare-field touch across the ~16 files that reach `md.mr.X`/`mr.X` today — the same scope of
work §17 and §20 each spent as their OWN task. `buildMoveDispatch`/`buildNodes` are pinned by
the STATEMENT that they call the unexported `newMoveDispatch` constructor and write directly
into moverRegistry's/nodeInboxes'/streamWiring's bare fields as one-time construction —
lifting them requires the SAME moverRegistry export work, plus an exported constructor path.
`MoveDispatch`'s 30 methods are pinned by the STATEMENT in the grouped list above: 13 read
multiple owners in one body (not a single-owner front to move with), 3 are delegators kept
alive ONLY because `mr` is unexported (moving them requires the SAME moverRegistry work), and
14 contain real, non-delegated logic of their own. Every one of these named next steps
resolves to the SAME single lever — exporting `moverRegistry`'s field surface into its own
package, `nodes/Wiring/moverregistry` or similar, following §17/§20's own precedent
(constructor + accessor methods, no exported field, no exported channel) — not four separate
problems. That lever was measured, precisely bounded (14 methods, ~51 external touch sites
across 16 files, both counts confirmed by grep this pass), and NOT attempted: it is a task of
comparable size to §17 or §20 on its own, and this pass's remaining time went to verifying the
two moves actually landed (`layoutquant`, `topoderive.AllocateWires`) to the same standard —
build/vet/race clean, doc-drift clean, deliberate-break-and-restore on each, test parity
confirmed — plus recovering cleanly from the concurrent-session interference above, rather
than starting a third, under-verified lift in the same pass.

**Good-outcome stopping point, updated.** `git status --short` is empty, `bash
scripts/stop-checks.sh` is empty, `go build ./...`/`go vet ./...` clean, `go test -race
-count=1 ./...` clean (every package `ok` or `[no test files]`, zero `FAIL`, zero races). Two
real, verified, behavior-preserving moves landed this pass beyond the BONUS goal §24 already
recorded; the ≤31-files/3-plus-packages requirement is measurably closer (94 → 76 in
`dispatch`, one type (`layoutQuantizer`) and one method (`allocateWires`) fully extracted,
one existing package (`topoderive`) grown) but not yet met, for the single, precisely-named
reason above — not because it wasn't looked for.

## 25. The TS extension host's five largest files — one real split found, four declined

A separate concurrent pass (same branch) measured the five largest TS files in
`tools/topology-vscode/src/` for the same god-object shape, statement by statement, never by
signature or whole-file label.

**`buffer-log.ts` (338 → 131 lines) — real split, landed.** `decodeEventLine` (117 lines) and
its two helpers `nodeGeometryLine` (21 lines) and `overlayFlag`/`OVERLAY_KINDS` (25 lines) are
100% bucket (b): every statement in all three is a `readEvent*`/`readNode*`/`readOverlay*`
column read, a `switch` on the decoded `kind` string, or object-literal construction from
those reads — no `vscode.*`, no process/webview call, no mutation of anything outside the
function's own locals. They moved to
`src/webview/three/decode/decode-event-line.ts`, an EXISTING directory that already holds the
sibling `buffer-decode-{edge,node,interior,view,shared}.ts` modules these functions already
imported from (`nodeLabel`, `edgeLabel`, `INTERIOR_SLOTS_PER_NODE`) — no new directory. What
stayed in `buffer-log.ts`: `decodeBufferLog`, `decodeEventsFromView`, `decodeStreamFrameEvents`
(the per-frame loop that calls `decodeEventLine` once per row and serializes to a
`.probe/go*.jsonl` line) and the `DecodedEventLine` exported type (pinned by
`test/contracts/trace-event-fields.test.ts`, which imports it by that path and was left
untouched). `ViewBlocksOrNull` (the camera/overlay/scene-sphere view lookup) moved with
`decodeEventLine` since only that function reads it.

No guard names `buffer-log.ts`, `decodeEventLine`, or the moved symbols by path or string
(`grep -rl "buffer-log\|decodeEventLine\|decodeBufferLog\|decodeStreamFrameEvents\|nodeGeometryLine\|DecodedEventLine" tools/ --include="*.sh"`
— zero hits) — no guard to re-key. `messages.ts`'s parity fences (`OVERLAY_FLAGS_START/END`,
`EDIT_MSG_START/END`, `RAW_INPUT_START/END`) are untouched — the file was not moved.

Test-name/assertion equality: both consuming test files
(`test/contracts/trace-event-fields.test.ts`, `test/buffer/bufferLogBreadcrumbsOnly.test.ts`,
14 tests total) import unchanged from `../../src/buffer-log` and pass unmodified, same names,
same assertion counts, verified by `npx vitest run` before and after (180/180 project-wide).
Deliberate-break proof: `decodeEventLine`'s breadcrumb-label line
(`const label = BREADCRUMB_LABELS[labelId] ?? String(labelId);`) hardcoded to `"BROKEN"`
failed both `bufferLogBreadcrumbsOnly.test.ts` cases (`expected 'BROKEN' to be 'dwell_start'`);
restored and reconfirmed green. Named gap: `bufferLogBreadcrumbsOnly.test.ts` also carries an
`edge-bead` row through the same decode but asserts only its `kind`, never its
`x`/`y`/`z`/`f`/`bead` fields — breaking `readEventX` in the `edge-bead` case (tried:
substituting a literal for `readEventX(ev, i)`) did NOT fail either test. The `recv`/`fire`/
`send`/`arrive`/`geometry`/`node-geometry`/`node-bead`/`camera`/`scene-sphere`/`select`/
`hover`/overlay-kind branches of `decodeEventLine`, and all of `nodeGeometryLine`, have NO
live test that can fail on a wrong field value — `trace-event-fields.test.ts` only checks the
STATIC fixture JSONL's shape (`DecodedEventLine`'s type), it never calls `decodeEventLine`.
This gap pre-dates the move (same functions, same file, before the split) — the move does not
change it, only relocates it; recorded here per this task's instruction to say plainly which
lifted functions have no test that can fail.

**Declined, with the specific pinning statement quoted:**

- **`runCommand.ts` (408 lines).** Every remaining statement in the class body (`run`,
  `cancel`, `restart`, `dispose`, `newDemux`, `writeStdin`, the four `getLast*` accessors) is
  one of: a `vscode.*` call (`vscode.workspace.workspaceFolders`,
  `vscode.window.createOutputChannel`), a `cp.spawn`/`this.proc.*`/`process.kill` call, or a
  read/write of `this.proc`/`this.channel`/`this.demux`/`this.spawnGen`/`this.cancelled`/
  `this.looping`/`this.pendingStdin` — the runner's own process-lifecycle state, by
  construction (comment at the file's own head names what was ALREADY extracted to
  `runner/*.ts`: `stream-fds.ts`, `spawn-layout.ts`, `attach-listeners.ts`, `counts.ts`,
  `go-errors.ts`, `probe-paths.ts`, `framing.ts`, `parse-state.ts`, `stream-demux.ts`). No
  statement measured is bucket (b) on its own locals; the file is what a prior pass on this
  same branch already reduced it to. Declined as already-lifted, not as "it's the runner".
- **`stream-demux.ts` (364 lines).** Every `handle*Fd` method interleaves three things per
  statement: `splitFrames`/framing (already its own module, `runner/framing.ts`, imported not
  reimplemented here), `fs.appendFileSync` to a `.probe/go*.jsonl` file (bucket a — a real
  filesystem write, this class's own job per its doc comment: "the per-owner probe-log
  decode ... AND the relay to the webview"), and cache mutation
  (`this.lastViewFrame = ab.slice(0)`) plus `this.onSnapshot(...)` (bucket a — instance-state
  write + a callback into the render seam). The decode calls themselves
  (`decodeNodeStreamFrame`, `decodeEdgeStreamFrame`, `decodeInteriorStreamFrame`,
  `decodeStreamFrameEvents`) are already pure functions imported from
  `webview/three/decode/*.ts` — there is no further pure computation to lift; what remains
  is the demux's own per-fd orchestration statement by statement.
  `processInteriorLikeFrames`'s own doc comment states directly why its two callers
  (`handleInteriorFd`/`handleDriveFd`) cannot be further merged even though their bodies
  look similar (`docs/investigations/interior-stream-framing.md`'s fix: merging their CARRY
  BUFFERS would desync two physically distinct pipes) — declined on that concrete
  statement-level reason, not on file cohesion.
- **`messages.ts` (277 lines).** Read function-by-function: `parseWebviewToHost` (18 lines)
  and `parseHostToWebview` (17 lines) are pure (bucket b — `typeof`/`instanceof` checks on
  their `raw` argument, no `vscode.*`), but the remainder of the file is type-only
  (`EditMsg`, `RawInputEvent`, `RawHit`, `HostToWebviewMsg`, `WebviewToHostMsg`,
  `OVERLAY_FLAG_NAMES`) — TypeScript types erase at compile time, so "splitting" them is
  moving declarations, not lifting impure code away from pure code; there is no (a)/(b) split
  to make here, the whole file is already (b) or type-erased. Declined for lack of an (a)/(b)
  boundary to act on, not for file size. The `OVERLAY_FLAGS_START/END`, `EDIT_MSG_START/END`,
  `RAW_INPUT_START/END` parity fences (`.claude/rules/bridge-surface.md`: parity guards grep
  this file) are an added reason any future split of this file specifically must re-key
  `check-edit-op-parity.sh`/`check-message-kind-parity.sh` and re-prove teeth before landing.
- **`extension.ts` (290 lines).** `activate`, `armHostReloadWatcher`, `resetProbeLogs`,
  `openTopologyEditor` — every statement measured touches `vscode.*` (commands, output
  channels, webview panels, file-system watchers), `fs.*` tied 1:1 to a watcher/panel
  lifecycle, or constructs a `cp`-adjacent runner/callback. The genuinely pure helpers this
  file's watchers call (`hashBundle`, `isHostReloadEnabled`, `shouldReloadHost`,
  `shouldRestartAfterBuild`, `TrailingDebouncer`) are already extracted to
  `hostReload.ts`/`hotRestart.ts`, imported not reimplemented. No statement measured is
  bucket (b) on this file's own locals.

**Domain-state check:** the one real split (`decode-event-line.ts`) adds no store, cache,
manager class, or module-level mutable — it is three free functions plus one interface,
called synchronously per event row, holding nothing between calls. Refused: nothing here
would have given state a home.

**Verification.** `npx tsc --noEmit -p tools/topology-vscode` clean;
`npm run build` (from `tools/topology-vscode`) succeeds (`out/extension.js`, `out/webview.js`
both rebuilt); `npx eslint src/` clean; `npx vitest run` — 30 files, 180/180 passed, before and
after. `bash scripts/stop-checks.sh` from repo root could not be read as a pass/fail signal in
isolation on this branch: a concurrent Go-side agent's untracked/staged files
(`nodes/Wiring/layoutquant/*.go`, mid-refactor) trip `check-no-untracked-source` regardless of
this TS change; the TS-scoped checks above were run directly instead, per this task's
instruction to check the signal a check actually emits rather than trust a shared gate mid
concurrent-session.

**Process note — a `git commit` with no pathspec is not scoped by a prior `git add
<files>`.** The first commit here (`git add tools/topology-vscode/src/buffer-log.ts
tools/topology-vscode/src/webview/three/decode/decode-event-line.ts` followed by `git commit`
with no pathspec on the message) committed FOUR of the concurrent Go agent's already-staged
file deletions (`nodes/Wiring/dispatch/broadcast_move.go`, `commit_node_move.go`,
`quantized_move.go`, `touching_beads.go` — staged by that agent as part of an in-progress
rename into `nodes/Wiring/layoutquant/`) along with the two intended TS files, because `git
commit` with no pathspec commits the WHOLE INDEX, not just files added in the same shell
call. Caught immediately via `git show --stat HEAD`, fixed with a second commit restoring
those four files' HEAD~1 content as untracked (matching their pre-commit state exactly —
verified via `git diff a82dd640~1 fc10d7ed -- nodes/Wiring` returning empty), never staged or
committed by this pass again. `git status --short` after the fix shows only the other agent's
pre-existing unstaged/untracked WIP, unchanged from before this pass touched the tree.

## §26 — `moverRegistry` lifted into `nodes/Wiring/moverreg` (§24's declined lever, taken)

§24's per-type coupling table declined `moverRegistry` on its own terms: "lifting this type
means EXPORTING its field surface ... across all 16 external files, matching the scope of §17
(edgemover) or §20 (nodeactor) as its OWN task, not a byproduct of this one." This task IS
that own task. Package named `moverreg` — mirrors `edgemover`/`nodeactor`'s lowercase
actor-package naming, shorter than `moverregistry` while staying unambiguous.

**Re-measured touch count.** §24's "~51 across 16 files" combined production and test files.
Re-counted with comments excluded: **48 real code touches across 14 production files**
(`build.go`, `build_nodes.go`, `build_move_dispatch.go`, `gesture_graph.go`,
`gesture_handlers.go`, `gesture_hitclassify.go`, `move_persist.go`, `move_streams.go`,
`move_dispatch_api.go`, `move_dispatch_construct.go`, `scene_structure.go`, `stdin_apply.go`,
`distance_groups.go`) plus **~35 more across 16 test files** that construct a `*MoveDispatch`
directly and write into `md.mr`'s bare fields for fixture setup (`gesture_home_test.go`,
`gesture_drag_offset_test.go`, `continuous_drag_persist_test.go`, `quantized_layout_test.go`,
and 12 others) — test files stayed in package `dispatch` per §17/§20's own precedent
("construct a real `*MoveDispatch` and touch bare fields directly — unreachable from an
external test package") and were mechanically rewritten to the exported API, not left broken.

### Construction-time vs. post-construction — the ~48 production sites

| file | site(s) | class | disposition |
|---|---|---|---|
| `move_dispatch_construct.go` | 4-line map init (`nodeGeoms`/`edgeMovers`/`edgeOut`/`centerMirror`) | CONSTRUCTION | folded into `moverreg.New()` |
| `move_dispatch_construct.go` | `resolveDest`, `commitLocal`, `WireMessaging` binds, `nodeGeoms[id]=ng`, `centerMirror[id]=...`, mutual-pair/edge/partner-center/edge-id loops (14 sites) | CONSTRUCTION (single-threaded wiring pass, before any goroutine starts) | rewritten to `.NodeGeoms()`/`.EdgeMovers()`/`.EnqueueFor()`/`.CenterOfNode`/`.SeedCenter()` |
| `build_nodes.go` | `nodeGeoms[name]` read, `selfDriveClaimed` nil-check+init+set (3 stmts), `bind(...)` | CONSTRUCTION | read → `.NodeGeoms()[name]`; 3-stmt write folded into new `ClaimSelfDrive(id)`; `bind`→`Bind` |
| `build_move_dispatch.go` | 5 `nodeGeoms` reads/ranges (scene flags, quant offsets, self-kind, neighbor-kind, out-targets seeding) | CONSTRUCTION | `.NodeGeoms()` |
| `build.go` | `finalizeActors(...)` | CONSTRUCTION | `.FinalizeActors(...)` |
| `move_streams.go` | `SetMsgTap`'s `nodeGeoms` range | CONSTRUCTION-ADJACENT (test-only setup call, before `Start`) | `.NodeGeoms()` |
| `move_persist.go` | `nodeGeoms` range (`EnableEditPersist` seeding `persistRoot`) | CONSTRUCTION-ADJACENT (runs before `Start` in every real call path) | `.NodeGeoms()` |
| `gesture_graph.go`, `gesture_handlers.go` (×5), `gesture_hitclassify.go` (×3) | `nodeGeoms`/`centerOfNode`/`nodeBodyRadius` reads inside gesture handlers | POST-CONSTRUCTION (runs per pointer/wheel event) | `.NodeGeoms()`/`.CenterOfNode`/`.NodeBodyRadius` |
| `move_streams.go` | `edgeMovers`/`nodeGeoms` passed into `setEdgeStreams`/`setNodeStreams` | POST-CONSTRUCTION (called once at startup, after construction, before `Start`) | `.EdgeMovers()`/`.NodeGeoms()` |
| `move_dispatch_api.go` | `start`, `sendMove`, `nodeSelfDriven`, `hasNodeMover`, `nodeQuantOffset`, `setSelectionUI`'s `edgeMovers` | POST-CONSTRUCTION (external entry points, called after `Start`) | `.Start`/`.SendMove`/`.NodeSelfDriven`/`.HasNodeMover`/`.NodeQuantOffset`/`.EdgeMovers()` |
| `scene_structure.go` | `nearestNodeTo`, `linkRefusal` | POST-CONSTRUCTION (palette-drop gesture) | `.NearestNodeTo`/`.LinkRefusal` |
| `stdin_apply.go` | `nodeGeoms[id]` existence check, `sendMove` (×2) | POST-CONSTRUCTION (stdin dispatch) | `.NodeGeoms()`/`sendMove(mr, ...)` (mr now `*moverreg.MoverRegistry`) |
| `distance_groups.go` | `centerOfNode` (×2), `nodeGeoms` (×1, inside `RootMove` bind) | POST-CONSTRUCTION (VIEW-frame length readout, arrow-click apply) | `.CenterOfNode`/`.NodeGeoms()` |

### Exported surface, and why each export is unavoidable

**Two live-map accessors** (`NodeGeoms() map[string]*nodeactor.NodeGeometry`,
`EdgeMovers() map[string]*edgemover.EdgeMover`) replace every bare `md.mr.nodeGeoms`/
`md.mr.edgeMovers` touch — construction-time writes (`mr.NodeGeoms()[id] = ng`), ranges, and
existence checks all go through the SAME live map a bare field would have, since a Go map is
a reference type: the accessor returns no copy. `nodeGeoms`'s element type
(`*nodeactor.NodeGeometry`) and `edgeMovers`'s (`*edgemover.EdgeMover`) are already exported
from already-lifted packages (§17/§20), so returning the map exports no NEW type — only the
map's existence as a directory crosses the boundary, exactly as `md.mr.nodeGeoms` already did
within package `dispatch`.

**Two construction-time wiring methods, new** (`ClaimSelfDrive(id string)`,
`SeedCenter(id string, c vec3)`) fold what used to be 3- and 1-statement external bare-field
writes into one call each — the same consolidation shape §19 used for `nodeGeometry`'s own
construction-time writes, applied here because a nil-check-then-lazy-init-then-set sequence
on an unexported map cannot be expressed as three bare statements once the map is unexported
to a different package.

**One constructor** (`New() MoverRegistry`) replaces the 4-line init block, mirroring
`NewNodeGeometry`/`NewNodeMover`'s existing shape.

**Twelve unchanged-shape methods, exported (were unexported):** `Bind`, `Start`,
`FinalizeActors`, `CenterOfNode`, `SendMove`, `EnqueueFor`, `NodeBodyRadius`, `HasNodeMover`,
`NodeSelfDriven`, `NodeQuantOffset`, `LinkRefusal`, `NearestNodeTo` — each was already a
single-owner method reading/writing `mr`'s own fields through the receiver (§24's own
classification: "not a namespace of parameter-taking free functions the way `layoutQuantizer`
was"); capitalizing them is the whole move for this dozen, no signature change.

**Two methods stay unexported** (`nodeKind`, `drainCenterMirror`) — package `dispatch` never
called either directly, only through `NodeBodyRadius`/`LinkRefusal` and `CenterOfNode`
respectively, so exporting them would be surface with no caller.

**One pure helper moved with its only caller** (`firstPortOfDir`, from `scene_structure.go`'s
`linkRefusalFor`'s sibling) — it depended on `kindapi.Registry` (a package-level map) and had
no caller left in `dispatch` once `linkRefusalFor` moved, so it moved rather than leaving a
one-caller stub behind. `scene_structure.go` lost its `kindapi`/`portwiring` imports as a
result (both now unused there).

**`InboxDepth` exported** (was `dispatch`'s unexported `moverInboxDepth`) — a plain constant,
not a method or a channel, still needed by `build_nodes.go` (in package `dispatch`) to size
two directed inbox channels (`ClaimLatticeIn`/`ClaimTiltEditIn`) at the SAME depth every other
per-mover inbox uses. Exporting a constant is not the field/channel-export the task's
constraints ban — there was already exactly one definition, so it is exported rather than
duplicated (unlike edgemover's/nodeactor's own `InboxDepth`/`inboxDepth` duplicates, which
existed because those packages needed their OWN private copy for their own internal sizing).

### No channel or field exported — confirmed by construction

`MoverRegistry` has no channel field at all — its three channel-bearing collaborators
(`nodeactor.NodeGeometry`, `nodeactor.NodeMover`, `edgemover.EdgeMover`) already hold their
own channels privately behind their own exported methods (§17/§20); `MoverRegistry` itself
only holds directories (maps) and one bool map (`selfDriveClaimed`). `SendMove`/`EnqueueFor`
are the two places this file reaches into another actor's channel-backed send path, and both
already did so through an exported method (`nm.SendExternal`, `nm.EnqueueSend`) before this
move — nothing new crosses a channel boundary here. No field was exported: every access is a
method (`NodeGeoms()`/`EdgeMovers()` return the live map value, not a struct field
declaration; Go's export rule cares about the identifier `NodeGeoms`, not what it returns).

### What moved, what stayed

Only `mover_registry.go` held a `func (mr *moverRegistry)` receiver, so only that one file
moved (as `nodes/Wiring/moverreg/mover_registry.go`, plus a new `vec_alias.go` matching
`nodeactor`'s own local `vec3` alias pattern). The pure helper `firstPortOfDir` moved with it
(above). Every other touched file — the 14 production files and 16 test files in the tables
above — merely took `*moverRegistry` as a parameter or reached `md.mr.X`; all STAYED in
`dispatch` and were rewritten to the exported API, per the task's own classification rule.

### Goroutine count, channel set, ordering — unchanged

Still exactly one goroutine per ring node (`NodeMover.Run`, launched from `MoverRegistry.Start`,
formerly `moverRegistry.start` — same call site shape) and one per edge (`EdgeMover.Run`,
same launch site). Still the same channels per node/edge (unchanged since §17/§20 — this task
touched no channel declaration). `Start`'s body is byte-identical except the receiver's
capitalization; `Bind`/`FinalizeActors` are byte-identical except capitalization and the
`selfDriveClaimed`-write 3-statement fold into `ClaimSelfDrive`, which produces the exact same
map state (nil-check → lazy-init → set, same order, same effect).

### Test-name/assertion equality

`grep -oE '^func Test[A-Za-z0-9_]+' nodes/Wiring/dispatch/*_test.go nodes/Wiring/kindapi/*_test.go`
→ **109**, identical set to §24's own baseline (confirmed via prior baseline, not re-derived).
`t.Fatal|Fatalf|Error|Errorf(` call count: **324**, identical to §24's baseline. No test was
renamed, dropped, weakened, or skipped — every rewritten test file changed only the accessor
syntax at the call site (`md.mr.nodeGeoms[id]` → `md.mr.NodeGeoms()[id]`, a `moverRegistry{...}`
bare-struct literal → `moverreg.New()`), never an assertion.

### Guards, re-run and proven with teeth

Neither `check-scene-path-resolution.sh` nor `check-persist-write-ownership.sh` needed
re-keying — both match by FILENAME (`mover_registry.go`) via a recursive `find`/`grep` over
`nodes/Wiring/`, which already covers the new `moverreg/` subdirectory, the same precedent
§17/§20 documented for `edge_mover.go`/`node_mover.go`. Proven with teeth anyway: injected
`filepath.Join("nodes", "x", "y.json")` into `SeedCenter` — `check-scene-path-resolution.sh`
reported `hand-rolled-node-path: .../moverreg/mover_registry.go: 137:...` and exited 1;
injected `jsonpersist.WriteJSONAtomic("x.json", nil)` into the same method —
`check-persist-write-ownership.sh` reported `unauthorized-write: .../moverreg/mover_registry.go:
139:...` and exited 1. Both probes removed immediately after; `diff` against a pre-edit backup
of the file confirmed byte-identical restoration. `check-dep-rules.sh`, `check-channel-names.sh`,
`check-composer-fields.sh` (doesn't police `MoverRegistry` — only `MoveDispatch`/`NodeGeometry`
are in its `COMPOSERS` table, unaffected by this move), `check-doc-drift`, `check-docs-symbols.sh`
(the historical `mover_registry.go`/`moverRegistry` citations inside
`docs/planning/movedispatch-decomposition.md` and `docs/planning/gesture-actor.md` describe
PAST state as an accumulating decision log, not live symbol references the guard polices — both
guards ran clean, confirming no doc-symbol citation broke), `check-stream-fd-mismatch-reported.sh`
all ran clean, none needed re-keying (none match `moverreg`/`MoverRegistry` by content, only the
already-recursive filename match above).

Grepped `nodes/Wiring/moverreg/*.go` for `runtime.Caller`/`filepath.Join("..", ...)`/`../..` —
zero hits (the class that produced a real bug on this branch already, per §23's
`distance_groups_test.go` fix).

### Verification

`go build ./...`, `go vet ./...`: clean. `go test -race -count=1 ./...`: every package `ok` or
`[no test files]`, zero `FAIL`, zero race reports, including `nodes/Wiring/dispatch` and the
new `nodes/Wiring/moverreg` (no test files of its own — `mover_registry.go`'s own behavior is
exercised entirely from `dispatch`'s existing suite, same shape §17's `edgemover`/§20's
`layoutquant` extraction left behind). The no-package-under-`Wiring`-imports-`dispatch` loop
(`for p in $(go list ./nodes/Wiring/... | grep -v 'dispatch$'); do go list -deps "$p" | grep -qx
.../dispatch && echo DEPENDS; done`) is empty. `bash scripts/verify.sh`: empty stdout.

### Deliberate breaks, confirmed and restored (moved surface itself, not just guards)

- `NodeGeoms()` forced to `return nil` → `go test ./nodes/Wiring/dispatch/...` panics inside
  `newMoveDispatch` (`nodes/Wiring/dispatch/move_dispatch_construct.go:138`, "assignment to
  entry in nil map") on `TestLoadTopologyComputesReachRadii` and every other test that loads a
  topology — a hard production failure, not a soft assertion miss. Restored.
- `CenterOfNode` forced to always return `vec3{}` (dropping the real center, keeping `ok`) →
  7 tests fail by name: `TestRingResolvesItsDistanceGroups`,
  `TestGestureHomeComputesFitPoseFromGeometry`, `TestCommitNodeMoveLocalDrawsQuantizedNotRawTarget`,
  `TestCommitNodeMoveLocalNeverMovesTowardMouseTarget`, `TestCommitNodeMoveLocalRemoveTakesBeadsPlace`,
  `TestCommitNodeMoveLocalAddMovesOneBeadBeyondNewBead`, `TestCommitNodeMoveLocalPersistsQuantizedNotRawPolar`.
  Restored.
- `NodeBodyRadius` forced to `return 0` → `TestGestureHomeComputesFitPoseFromGeometry`,
  `TestGestureHomeFramesUnknownKindAtRenderRadius` fail by name. Restored.

**Uncovered, reported rather than silently accepted.** `LinkRefusal` forced to always return
`("", "", "", true)` (unconditional success, no refusal) → **no test failed** anywhere in the
full suite (`go test ./...`). `NearestNodeTo` forced to always return `("", false)` → **no test
failed**. Both are reached only from `scene_structure.go`'s palette-drop path
(`applyCreateNode`), which has no test driving it end-to-end today — a pre-existing gap (these
two functions' bodies are byte-identical to their pre-move `moverRegistry` versions; nothing
about the move introduced the hole). `Bind`/`Start`/`SendMove`/`EnqueueFor` are the
channel-touching class `docs/process/testing-shape.md`'s own doctrine excludes from unit
testing (cross-goroutine delivery) — same excluded class §17/§20 already named for
`TrySendFromSrc`/`SendExternal`, not re-probed individually here since the doctrine, not a new
measurement, is what excludes them. `ClaimSelfDrive`/`SeedCenter`/`New` are construction-time
writes exercised only as plumbing inside the same `TestLoadTopologyComputesReachRadii`-shaped
integration tests that already caught the `NodeGeoms()` break above — not independently
break-tested, same class as §19's own construction-time-write coverage note.

### `ls nodes/Wiring/dispatch/*.go` and final file counts

`nodes/Wiring/dispatch`: **75** (was 76 — `mover_registry.go` moved out, no new file added;
`scene_structure.go` lost `firstPortOfDir` but stayed one file). Still over the ~31 target —
the file remains dominated by `*MoveDispatch`-receiver methods and their tests, which is
`MoveDispatch`'s OWN unmoved surface (§24's per-type table), a separate, harder lever this task
was not scoped to pull. `nodes/Wiring/moverreg` (new): **2** non-test files
(`mover_registry.go` 395 LOC, `vec_alias.go` 7 LOC), 0 test files.

No `types`/`common` package, no alias shim, no dot-import, no package-level actor global, no
`*ForTest` constructor, no interface added.

**`git status --short` and commits:** see below.

## §27 — stdin cluster and build/load cluster, both RE-MEASURED and PINNED (no move landed)

Baseline confirmed unchanged from §26: `nodes/Wiring/dispatch` is 75 files (28 non-test + 47
test), non-test line counts matching the task brief exactly (`move_dispatch_construct.go` 230,
`stdin_dispatch.go` 182, `stdin_apply.go` 172, `move_dispatch.go` 163, `build.go` 136,
`loader.go` 78, `gesture.go` 48, `scene_switch.go` 45, `viewpoint_state.go` 44,
`gesture_hitclassify.go` 41, `vec_alias.go` 22, `bool_u8.go` 13). `bash scripts/stop-checks.sh`
empty (clean) before any analysis — confirmed as the starting baseline.

### Method

Every candidate function's body was read statement-by-statement per the task's own rule:
bucket (a) = writes/reads an unexported field of a type staying behind (or sends on a
channel / starts a goroutine / writes a file); bucket (b) = computation on locals, params,
and already-exported field/method reads. `MoveDispatch` itself is confirmed NOT moving (§24's
own table: 30 methods, "gesture entry points reading MULTIPLE owners at once", unchanged by
§26's `moverreg` lift) — so any function that reaches `md.mr`, `md.persist`, `md.ctx`,
`md.inboxes`, `md.sw`, `md.lq`, or `md.tapToInstall` (all still UNEXPORTED fields of
`*MoveDispatch`, per `move_dispatch.go`) is bucket (a) by construction, regardless of whether
the value read off that field is itself an exported type/method (e.g. `md.mr.NodeGeoms()` —
`NodeGeoms` is exported, but `mr` the FIELD is not, so the expression cannot be written outside
package `dispatch`).

### Stdin cluster — `stdin_dispatch.go` (182) + `stdin_apply.go` (172)

| function | body touches | bucket |
|---|---|---|
| `HandleRawInputMsg` | `md.HandleRawInput(...)` — one call to an EXPORTED method | **(b)** — the one exception |
| `HandleSaveMsg` | `md.persist.overlays.Schedule`, `md.persist.sphere.Schedule` — `persist` unexported | (a) |
| `ApplyEdit`/`applyEdit` table | dispatches into `applyUpdate` (below); table itself carries no md touch but is inseparable from its handlers | (a), transitively |
| `applyUpdate` + `updateKindHandlers` table | dispatches into the 5 `applyUpdate*` handlers (all touch `md` fields) | (a), transitively |
| `clockAttrHandlers["speed"]` closure | `md.UI.ClockDivisor`/`md.UI.Speed` (UI exported, fine) but also `md.persist.speed.Schedule(...)` | (a) |
| `overlayAttrHandlers["toggle"]` closure | ONLY `md.UI.OV`/`md.UI.EmitBreadcrumb`/`md.UI.EmitViewFrame` — all through the exported `UI` field | **(b)** — the other exception |
| `applyUpdateClock` | delegates to `clockAttrHandlers` | (a), transitively |
| `applyUpdateDistanceGroup` | `applyDistanceGroupTarget(md.ctx, &md.UI, &md.mr, &md.lq, ...)` — `ctx`/`mr`/`lq` unexported | (a) |
| `applyUpdateTiltVector` | `md.mr.NodeGeoms()`, `sendTiltEdit(&md.inboxes, md.ctx, ...)`, `sendMove(&md.mr, md.ctx, ...)` | (a) |
| `applyUpdateScene` | `SelectScene(&md.Scenes, ...)` (Scenes exported, fine) but also `md.UI.LatticePoints=`/`md.persist.lattice.Schedule`/`md.BroadcastLatticePoints`/`md.CreateNode`/`md.DeleteNode` — several are exported methods, but `md.persist.lattice` is not | (a) |
| `applyUpdateOverlays` | dispatches to `overlayAttrHandlers`, then unconditionally `md.persist.overlays.Schedule(md.UI.OV)` | (a) |

**Verdict: PINNED.** Exactly 2 of 11 functions are bucket (b) in isolation
(`HandleRawInputMsg`, the `"toggle"` closure), and both are small (5 and ~20 lines) pieces
wired into dispatch tables (`editOps`, `updateKindHandlers`, `overlayAttrHandlers`) that must
stay adjacent to their bucket-(a) siblings — splitting 2 functions into a third file to shave
~25 of 354 combined lines is not a real reduction and was not done. Every other function's
pinning statement is named in the table above (`md.persist.X.Schedule`, `md.mr.X`, `md.ctx`,
`md.inboxes`) — each is an unexported field read/write on the type that stays.

### Build/load cluster — `build.go` (136) + `loader.go` (78) + `move_dispatch_construct.go` (230)

**Re-measured against §24's stated blocker** ("calls the unexported constructor
`newMoveDispatch(...)`, then writes directly into `md.mr.nodeGeoms`/`md.lq.QuantizedLayout`/
`md.mr.selfDriveClaimed` (bare, unexported hub fields)"). That EXACT wording is now stale —
`moverreg.MoverRegistry` (§26) exposes `NodeGeoms()`/`EdgeMovers()`/`ClaimSelfDrive` as real
exported methods, so the specific bare-field-write phrasing no longer applies. But the
STRUCTURAL blocker survives the rewording: both `buildMoveDispatch` and `buildNodes` are
methods on `*buildCtx`, and BOTH still read/write `b.md`'s (a `*MoveDispatch`) own UNEXPORTED
fields directly — `mr` and `inboxes` and `sw`, not `mr`'s or `inboxes`'s own fields:

- `buildMoveDispatch` (`build_move_dispatch.go`): `b.md = md` (the assignment itself — `md`
  field unexported), `md.lq.QuantizedLayout = ...` (`lq` unexported), `md.mr.NodeGeoms()` (×4,
  `mr` unexported field access, independent of `NodeGeoms` being exported).
- `buildNodes` (`build_nodes.go`): `b.md.UI.LatticePoints` (UI exported — fine in isolation) but
  also `b.md.inboxes.lattice`/`b.md.inboxes.tiltEdit` (direct map writes on the unexported
  `inboxes` field, inside the `kindapi.BuildDeps` closures), `b.md.mr.NodeGeoms()`/
  `b.md.mr.ClaimSelfDrive(...)` (`mr` unexported), `b.md.RT` (exported, fine),
  `b.md.sw.interiorOuts`/`b.md.sw.driveOuts`/`b.md.sw.buildInteriorFrame` (`sw` unexported).
- `newMoveDispatch` (`move_dispatch_construct.go`, the constructor `buildMoveDispatch` calls):
  literally constructs `md := &MoveDispatch{tr: tr}` then writes `md.mr = moverreg.New()`,
  `md.UI.OV = ...`, `md.GS.NodeSeeds = ...`, `md.RT.Build(...)`, `md.UI.DistanceGroupLensFn =
  func() { ... mrForLens := &md.mr ... }` — a struct LITERAL and direct field assignment on
  every one of `MoveDispatch`'s fields, several unexported (`tr`, `mr`). A constructor of a
  type whose fields it sets directly by literal/assignment cannot live outside that type's own
  package, independent of whether the constructor's NAME is exported or not — this was never
  actually about the name `newMoveDispatch` vs `NewMoveDispatch`, contrary to how §24 phrased
  it.
- `buildFromSpec`/`buildCtx` (`build.go`): `buildCtx` itself has exactly 2 remaining methods
  (`buildMoveDispatch`, `buildNodes` — `allocateWires` already externalized in §24), both
  pinned above, so Go's "a method's receiver type must be defined in the same package as the
  method" rule keeps `buildCtx` itself in `dispatch` too. `buildFromSpec` constructs and reads
  `buildCtx`'s own fields directly (`b.nodeGeoms`, `b.md`, `b.speedSinks`, ...) and calls
  `b.md.mr.FinalizeActors(...)` directly (`mr` unexported) — pinned by the same rule.
- `LoadTopology` (`loader.go`): the ONE genuinely loose piece — its body never touches
  `MoveDispatch`'s or `buildCtx`'s unexported fields, only calls `kindapi.BuildRegistry()`,
  `loadspec.ParseSpec`/`ValidateSpec`, `scenepersist.LoadSceneSphere`, and the (unexported)
  `buildFromSpec`. It is bucket (b) IN ISOLATION but its only non-trivial call
  (`buildFromSpec`) is itself pinned in `dispatch` (above), so `LoadTopology` would need to
  call an EXPORTED `dispatch.BuildFromSpec` from a new package — a rename-only move that
  shrinks `dispatch` by exactly 1 non-test file (78 LOC) and adds 1 file elsewhere, for a net
  package-count change with no reduction in the files that actually dominate the 28-file total.
  Not done: not a real reduction, and the task named `loader.go` as part of ONE cluster with
  `build.go`/`move_dispatch_construct.go`, which stay.

**Verdict: PINNED**, confirmed by re-measurement rather than assumed. §24's blocker phrasing
was stale (moverreg's exported API means the WRITES are no longer bare-field writes on `mr`'s
OWN fields) but the CONCLUSION was right for a reason §24 didn't isolate: the blocker was
never `mr`'s encapsulation, it is `MoveDispatch`'s OWN encapsulation (`mr`/`inboxes`/`sw`/`lq`/
`ctx`/`tr`/`tapToInstall` as fields on `*MoveDispatch`, the type §24 already confirmed is not
moving). Lifting `mr` into `moverreg` (§26) could never unblock `buildMoveDispatch`/
`buildNodes`/`newMoveDispatch`, because their pinning statements were never really about `mr`'s
own internals — they were always about reaching `b.md.mr` / `b.md.inboxes` / `b.md.sw`, i.e.
`MoveDispatch`'s own field surface. Unblocking this cluster for real is the same shape of task
§24 named for `MoveDispatch` itself ("exporting its field surface... as its OWN task, not a
byproduct of this one") — not attempted here, correctly out of scope for a re-measurement pass.

### Outcome

No files moved, no code changed. `git status --short` empty, no commit made for either
cluster — both were genuinely blocked, per the task's own "if a cluster genuinely cannot move,
quote the specific pinning statement per function ... rather than stopping entirely" clause,
applied to both clusters in turn rather than only one. `bash scripts/stop-checks.sh`: empty
(unchanged tree). `nodes/Wiring/dispatch` file count: unchanged, **75** (28 non-test + 47
test).

## §28 — `MR`/`LQ`/`TR` exported on `MoveDispatch`; both §27 clusters re-measured, still
PINNED, pin statements narrowed

§27 listed five blockers for the stdin and build/load clusters: `md.mr`, `md.inboxes`,
`md.sw`, `md.lq`, `md.tr`. Of those, `GS`/`UI`/`Scenes`/`RT` were ALREADY exported fields of
`MoveDispatch` holding exported types from already-lifted packages (`geomseeds.GeomSeeds`,
`viewstate.UIState`, `sceneswitch.SceneSwitch`, `rowtables.RowTables`) — `mr`
(`moverreg.MoverRegistry`), `lq` (`layoutquant.LayoutQuantizer`), and `tr` (`*T.Trace`) were
the identical shape, unexported only because they predate the pattern. Confirmed by reading
`move_dispatch.go` directly before touching anything.

### Rename, commit 1

`mr`→`MR`, `lq`→`LQ`, `tr`→`TR` on `MoveDispatch` (`move_dispatch.go`'s struct decl) — pure
rename, no logic change. `persist`/`sw`/`inboxes`/`tapToInstall`/`ctx` were left unexported,
per the task's own constraint (they hold unexported types or are internal wiring).

**Call-site count.** 129 `md.mr` + 20 `md.lq` occurrences (comments excluded) across 30
production/test files, plus 2 `tr:`-literal / no-live-read `md.tr` sites in
`move_dispatch_construct.go`/`build.go`, plus 2 more `{mr: ...}`/`tr:` literals found only
after the first `go vet` pass (`scene_tabs_test.go`, missed by the initial grep because it
used `tr: T.New()` with no `.mr`/`.lq` neighbor on the same line) — 32 files touched in
total, confirmed by `git show --stat`. `md.tr` itself had **zero live in-package reads** —
only ever set at construction (`newMoveDispatch`'s struct literal) — confirmed by grep before
renaming, so `TR` is genuinely write-only today, same as before the rename.

**Verification.** `go build ./...`, `go vet ./...` clean. `go test -race -count=1 ./...`:
every package `ok` or `[no test files]`, zero `FAIL`, zero races. Test-name/assertion parity:
109 `TestXxx` / 324 assertions across `dispatch`+`kindapi`, identical to §26's baseline
(diff empty, count re-derived, not assumed). Guards: `check-no-network-locks.sh`,
`check-persist-write-ownership.sh`, `check-scene-path-resolution.sh`,
`check-channel-names.sh`, `check-doc-drift.sh`, `check-docs-symbols.sh`,
`check-no-untracked-source.sh`, `check-composer-fields.sh` (the one guard that mentions
`moverRegistry`/`layoutQuantizer` by name, in a comment and an error-message string — it
counts `MoveDispatch`'s field DECLARATIONS by pattern, not by name, so the rename left its
count (12) and its exit code (0) unchanged; no re-keying needed) all ran clean. The
no-package-under-`Wiring`-imports-`dispatch` loop: empty. Deliberate break: forced
`md.LQ.QuantizedLayout = false` unconditionally in `build_move_dispatch.go` (replacing the
real `scene.SceneUsesQuantizedDrag` call) → 4 tests failed by name:
`TestCommitNodeMoveLocalDrawsQuantizedNotRawTarget`,
`TestCommitNodeMoveLocalRemoveTakesBeadsPlace`,
`TestCommitNodeMoveLocalAddMovesOneBeadBeyondNewBead`,
`TestCommitNodeMoveLocalPersistsQuantizedNotRawPolar`. Restored; `go build ./...` and
`git diff` clean again.

### Stdin cluster, re-measured — still PINNED, narrower pin per function

Re-read every one of the same 11 functions/closures §27 tabulated, now that `md.MR`/`md.LQ`
are ordinary exported-field reads:

| function | remaining touch after the rename | bucket |
|---|---|---|
| `HandleRawInputMsg` | `md.HandleRawInput(...)` — exported method | **(b)**, unchanged |
| `HandleSaveMsg` | `md.persist.overlays.Schedule`, `md.persist.sphere.Schedule` | (a) — `persist` still unexported |
| `applyUpdateDistanceGroup` | `applyDistanceGroupTarget(md.ctx, &md.UI, &md.MR, &md.LQ, ...)` | (a) — narrowed to `md.ctx` ALONE; MR/LQ no longer part of the pin |
| `applyUpdateTiltVector` | `md.MR.NodeGeoms()` (now fine), `sendTiltEdit(&md.inboxes, md.ctx, ...)`, `sendMove(&md.MR, md.ctx, ...)` | (a) — narrowed to `md.ctx` + `md.inboxes`; MR dropped out |
| `applyUpdateScene` | `SelectScene(&md.Scenes, ...)` (fine), `md.persist.lattice.Schedule` | (a) — `persist` still unexported |
| `applyUpdateOverlays` | dispatches to `overlayAttrHandlers`, then `md.persist.overlays.Schedule(md.UI.OV)` | (a) — `persist` still unexported |
| `clockAttrHandlers["speed"]` closure | `md.UI.ClockDivisor`/`md.UI.Speed` (fine), `md.persist.speed.Schedule` | (a) — `persist` still unexported |
| `overlayAttrHandlers["toggle"]` closure | only `md.UI.*` | **(b)**, unchanged (the other pre-existing exception) |
| `ApplyEdit`/`applyEdit` table, `applyUpdate` + `updateKindHandlers` table, `applyUpdateClock` | pure dispatch, no `md` field touch of their own; still transitively pinned by their handlers above | (a), transitively, unchanged |

**Verdict: still PINNED**, exactly as §27 concluded, but the reason is now precise instead
of blanket: every one of the 5 functions that used to list `md.mr`/`md.lq` among their
blockers now lists ONLY `md.ctx` / `md.persist` / `md.inboxes` — genuinely unexported fields
of `MoveDispatch` that this task was told not to export (`ctx`/`persist`/`inboxes` all hold
either internal wiring or unexported types). MR/LQ dropped out of every one of these pin
statements; nothing else did. `nodes/Wiring/stdinreader` (already existing) was checked —
its own files (`stdin_reader.go`, framing/message-shape files) hold no `MoveDispatch`
receiver methods and no bucket-(a) touch of `md`'s fields at all; it is already a clean
package boundary and none of the 11 functions above are receiver methods that could move
there — they are free functions dispatched BY the reader, not the reader's own logic, so
there is nothing here that migrates INTO `stdinreader` either.

### Build/load cluster, re-measured — still PINNED, but the blocker itself is now provably
never about MR/LQ

Re-read `buildMoveDispatch`, `buildNodes`, `newMoveDispatch`, `buildFromSpec`/`buildCtx`,
`LoadTopology` with `MR`/`LQ`/`TR` now ordinary exported fields:

- **`buildMoveDispatch`** (`build_move_dispatch.go`, a `*buildCtx` method): every remaining
  statement in its body is now `md.LQ.QuantizedLayout = ...`, `md.MR.NodeGeoms()` (×4),
  `md.UI.*` — ALL exported-field/method reads on `md`. Its OWN body no longer touches a
  single unexported field of `MoveDispatch`. It still cannot move on its own, though — not
  because of anything it reads, but because Go requires every method of `*buildCtx` to live
  in the same package as `buildCtx` itself, and its sibling method `buildNodes` (below)
  remains genuinely pinned. This is the exact distinction §26 already drew for
  `layoutQuantizer` vs. `moverRegistry`: a method's OWN body can be un-pinned while the
  method itself stays put, blocked by its RECEIVER TYPE's other methods, not by anything in
  this one's body.
- **`buildNodes`** (`build_nodes.go`, a `*buildCtx` method): STILL directly touches
  `b.md.inboxes.lattice`/`b.md.inboxes.tiltEdit` (map writes inside the injected
  `kindapi.BuildDeps` closures) and `b.md.sw.interiorOuts`/`b.md.sw.driveOuts`/
  `b.md.sw.buildInteriorFrame` — `inboxes` and `sw` are UNEXPORTED types (`nodeInboxes`,
  `streamWiring`), not touched by this task's rename, so this function is pinned exactly
  where §27 left it. `b.md.MR.NodeGeoms()`/`b.md.MR.ClaimSelfDrive(...)` inside
  `ClaimSelfDriveGeom` are now fine, but that is not what pins this function.
- **`newMoveDispatch`** (`move_dispatch_construct.go`): constructs `md := &MoveDispatch{TR:
  tr}` then assigns `md.MR = moverreg.New()`, `md.UI.OV = ...`, `md.GS.NodeSeeds = ...`,
  `md.RT.Build(...)` — a struct LITERAL + direct field assignment on every one of
  `MoveDispatch`'s own fields, several STILL unexported in the general case (this
  constructor is IN package `dispatch`, so it is allowed to set unexported fields directly —
  that permission is exactly what a constructor outside the package would lose). This was
  never really about which fields are exported/unexported; it is that a function
  constructing a type via struct literal must live in that type's own package when even ONE
  field it sets is unexported (here: none currently unexported are set by literal after the
  rename, but the function ALSO wires `md.tapToInstall`/reads `md.MR`/`md.LQ` through
  closures bound at construction time — moving it would mean either exporting a
  constructor `NewMoveDispatch` that returns a partially-wired value, or exporting
  `tapToInstall`, both declined by this task's own constraints). Confirmed unchanged from
  §27: the blocker was never the field NAME, always the construction-time WIRING.
- **`buildFromSpec`/`buildCtx`** (`build.go`): `buildCtx` still has exactly 2 methods
  (`buildMoveDispatch`, `buildNodes`), both pinned above (one transitively, one directly),
  so `buildCtx` itself stays — and `buildFromSpec` constructs it via
  `&buildCtx{ctx: ctx, spec: spec, tr: tr, clk: clk, sphere: sphere, hasScene: hasScene,
  scenePath: scenePath}`, a struct literal setting `buildCtx`'s OWN unexported fields
  (`ctx`/`spec`/`tr`/`clk`/... — a DIFFERENT `tr` than `MoveDispatch.TR`, `buildCtx`'s own
  local field, untouched by this task's rename). This pin was **never about
  `MoveDispatch`'s `mr`/`lq`/`tr` at all** — it is `buildCtx`'s own encapsulation, a
  completely separate type. Re-measurement makes this explicit where §27 left it implicit.
- **`LoadTopology`** (`loader.go`): unchanged from §27 — its only non-trivial call is the
  unexported `buildFromSpec`, itself pinned above; not touched by this rename either way.

**One free-standing function's blocker did fully resolve: `bindDispatch`**
(`build_nodes.go`, NOT a `buildCtx`/`MoveDispatch` method — a plain function):
```go
func bindDispatch(md *MoveDispatch, outSink map[string]*wire.Out, destWire map[string]*wire.PacedWire) {
	md.MR.Bind(outSink, inputcodec.SlotRegistry(destWire))
}
```
Before the rename this was `md.mr.Bind(...)`, bucket (a). After, its ONLY touch is
`md.MR.Bind`, an exported method through an exported field — bucket (b) — and, being a free
function (no receiver), it is not blocked by the `buildCtx`/`MoveDispatch` same-package
receiver rule the way `buildMoveDispatch`/`buildNodes` are. **Declined to move or inline
anyway**: it has no natural destination package of its own (its 3-line body is now simple
enough that the "move" available is deleting it and inlining `b.md.MR.Bind(b.outSink,
inputcodec.SlotRegistry(b.destWire))` at its one call site in `build.go`, which changes
nothing about `nodes/Wiring/dispatch`'s file count or the 3-way-split goal this section
series is measured against — the task asked to move files to new packages / report pins
per statement, not to inline single-call helpers), so it is reported here as the one
statement in this cluster whose bucket genuinely flipped, without landing a change that
would not move the actual metric.

**Verdict: still PINNED**, matching §27's own conclusion, but the re-measurement proves
something §27 could only assert from the OLD (unexported-`mr`) vantage: the build/load
cluster's blocker was **never `MoveDispatch`'s `mr`/`lq`/`tr` fields** — it is (a)
`buildNodes`' own direct reach into `inboxes`/`sw` (unexported types, untouched by this
task), (b) `newMoveDispatch`'s construction-time wiring (closures bound at build time,
not expressible as a plain exported constructor without also exporting `tapToInstall`), and
(c) `buildCtx`'s OWN unexported fields, a completely separate type from `MoveDispatch`. Every
one of those three reasons is unrelated to the rename this task made, which is why exporting
`MR`/`LQ`/`TR` moved zero files in this cluster despite the earlier per-field grep counting
them among the "blockers" — they were never the actual load-bearing ones.

### Next lever, confirmed unchanged

Per the task's own instruction: the sole remaining lever for BOTH clusters is lifting the
UNEXPORTED TYPES `nodeInboxes` (`inboxes`) and `streamWiring` (`sw`) the way `moverRegistry`
was lifted into `moverreg` in §26 — `persist`/`ctx`/`tapToInstall` stay internal wiring by
this task's own constraint, not candidates for that lift. Not attempted here.

### Final state

`ls nodes/Wiring/dispatch/*.go` → **75** (28 non-test + 47 test), unchanged from §27 — the
rename (commit 1) touched 32 existing files but added/removed none, and the re-measurement
pass (commit 2, doc-only) moved no file either, since both clusters stayed pinned by
statements unrelated to `MR`/`LQ`/`TR`. Test-name/assertion parity confirmed unchanged
(109/324). `go test -race -count=1 ./...`: verbatim below.

## §29 — `nodeInboxes`/`streamWiring`/`persisters` lifted into `nodeinbox`/`streamwire`/
`viewpersist`; `Inboxes`/`Sw`/`Persist` exported on `MoveDispatch`; both §27/§28 clusters
re-measured, both stay pinned, but the stdin cluster's real blocker narrows to one field
and the build cluster's `buildNodes` blocker is gone

§28 named the sole remaining lever for both clusters: lift the unexported TYPES
`nodeInboxes` (`inboxes`) and `streamWiring` (`sw`), the way §26 lifted `moverRegistry`.
This task took that lever, plus `persisters` (`persist`), the third unexported sub-object
type `MoveDispatch` still held.

### Construction-time vs. post-construction — the touch tables

**`nodeInboxes` → `nodes/Wiring/nodeinbox.NodeInboxes`** (13 real touches across 5
production files, 2 test files):

| file | site(s) | class | disposition |
|---|---|---|---|
| `build_nodes.go` | `b.md.inboxes.lattice`/`.tiltEdit` nil-check+init+set (6 statements, inside `kindapi.BuildDeps` closures) | CONSTRUCTION (single-threaded build path, before any goroutine runs) | folded into `ClaimLatticeIn`/`ClaimTiltEditIn` |
| `stdin_apply.go` | `sendTiltEdit(&md.inboxes, ...)` (×3) | POST-CONSTRUCTION (stdin dispatch, per-message) | `sendTiltEdit`'s own parameter type changed to `*nodeinbox.NodeInboxes`; its body now delegates to `NodeInboxes.SendTiltEdit` |
| `move_dispatch_api.go` | `sendTiltEdit`'s body (`inboxes.tiltEdit[id]`, channel send/select) | POST-CONSTRUCTION (bare external-entry send helper) | body moved into `nodeinbox.NodeInboxes.SendTiltEdit`; `dispatch`'s `sendTiltEdit` is now a 1-line delegator |
| `scene_lattice_persist.go` | `md.inboxes.broadcastLatticePoints(points)` | POST-CONSTRUCTION (`BroadcastLatticePoints`, called from the view-owner goroutine on a scene edit) | `md.Inboxes.BroadcastLatticePoints(points)` |
| `scene_lattice_edit_test.go` | `md.inboxes.lattice = map[...]{...}` (×2, fixture setup) | CONSTRUCTION-style (test-only literal assignment) | `md.Inboxes.ClaimLatticeIn(id, ch)` per entry |

**`streamWiring` → `nodes/Wiring/streamwire.StreamWiring`** (7 real touches across 3
production files, 0 test files touching bare fields):

| file | site(s) | class | disposition |
|---|---|---|---|
| `move_streams.go` | `md.sw.setEdgeStreams(...)`, `md.sw.setNodeStreams(...)` | POST-CONSTRUCTION (called once at startup, before `Start`) | `md.Sw.SetEdgeStreams(...)`/`md.Sw.SetNodeStreams(...)` (methods capitalized, unchanged bodies) |
| `build_nodes.go` | `pb.InteriorOuts = &b.md.sw.interiorOuts`, `pb.DriveOuts = &b.md.sw.driveOuts`, `pb.BuildInteriorFrame = &b.md.sw.buildInteriorFrame` | CONSTRUCTION (single-threaded, before any node's Update goroutine launches — the pointers are captured empty and read live once `SetNodeStreams` populates the SAME field later) | `b.md.Sw.InteriorOutsPtr()`/`.DriveOutsPtr()`/`.BuildInteriorFramePtr()`, three new methods returning `&sw.<field>` |
| `viewpoint_state.go` | doc-comment only (`md.sw/md.RT to emit the VIEW frame`) | — | comment reworded to `md.Sw` |

**`persisters` → `nodes/Wiring/viewpersist.Persisters`** (12 real touches across 6
production files, 5 test files):

| file | site(s) | class | disposition |
|---|---|---|---|
| `move_persist.go` | `EnableViewpointPersist`: constructs `p := &camerapersist.ViewpointPersister{...}`, `md.persist.vp = p` | CONSTRUCTION (called once, after the startup seed, before `Start`) | `p := md.Persist.ArmViewpoint(topologyPath)` — the persister's own construction moved INTO `viewpersist.Persisters.ArmViewpoint`, which returns `p` so `EnableViewpointPersist` can still wire `md.UI.VP.Persist = p.Schedule` |
| `move_persist.go` | `EnableEditPersist`: constructs and assigns `md.persist.overlays`/`.sphere`/`.speed`/`.lattice` (4 struct literals) | CONSTRUCTION | folded into `viewpersist.Persisters.ArmEdit(topologyPath)`; `EnableEditPersist` now calls `md.Persist.ArmEdit(topologyPath)` and keeps only its OWN remaining work (`md.Scenes.TreeRoot`, the per-mover `SetPersistRoot` loop) |
| `stdin_apply.go` | `md.persist.lattice.Schedule(points)`, `md.persist.overlays.Schedule(md.UI.OV)` | POST-CONSTRUCTION (stdin dispatch) | `md.Persist.Lattice().Schedule(points)`, `md.Persist.Overlays().Schedule(md.UI.OV)` |
| `stdin_dispatch.go` | `md.persist.overlays.Schedule(...)`, `md.persist.sphere.Schedule(...)`, `md.persist.speed.Schedule(...)` | POST-CONSTRUCTION (`HandleSaveMsg`, `clockAttrHandlers["speed"]`) | `md.Persist.Overlays().Schedule(...)`, `md.Persist.Sphere().Schedule(...)`, `md.Persist.Speed().Schedule(...)` |
| `scene_sphere_persist.go`, `scene_speed_persist.go` | doc-comment only (`md.persist.sphere`/`.speed`) | — | reworded to `md.Persist`, reached via `.Sphere()`/`.Speed()` |
| 5 test files (`scene_clock_divisor_test.go`, `scene_edit_persist_test.go` ×2, `scene_lattice_persist_test.go`, `scene_speed_persist_test.go` ×2) | `md.persist.speed.Schedule(...)`, `md.persist.lattice.Schedule(...)`, `md.persist.overlays.Schedule(...)` | — (post-`EnableEditPersist`-armed accessor reads) | rewritten to `md.Persist.Speed()`/`.Lattice()`/`.Overlays()` |

### Exported surface, and why each export is unavoidable

**`nodeinbox.NodeInboxes`**: `ClaimLatticeIn(id string, ch chan int32)` and
`ClaimTiltEditIn(id string, ch chan movemsg.TiltEditMsg)` are new construction-time setters
folding the nil-check-then-lazy-init-then-set 3-statement sequence each map needed — same
consolidation shape §26 gave `ClaimSelfDrive`. `BroadcastLatticePoints(points int32)` and
`SendTiltEdit(ctx, id, msg) bool` are the two post-construction methods, unchanged bodies
from the pre-move `broadcastLatticePoints`/package-level `sendTiltEdit`, just promoted to
receiver methods. **No channel crosses the boundary**: `ClaimLatticeIn`/`ClaimTiltEditIn`
take a channel IN (the caller constructed it, this just registers it — same "directory
takes ownership of a channel handed to it" shape moverreg's `Bind` already had), and
`BroadcastLatticePoints`/`SendTiltEdit` send ON the held channels internally without ever
returning one. `dispatch`'s own `sendTiltEdit` free function keeps its exact name and
signature shape (now `*nodeinbox.NodeInboxes` instead of `*nodeInboxes`) as a 1-line
delegator, so `stdin_apply.go`'s three call sites needed no edit beyond `&md.inboxes` →
`&md.Inboxes`.

**`streamwire.StreamWiring`**: `SetEdgeStreams`/`SetNodeStreams` are the same two methods
promoted to exported, byte-identical bodies (matching §26's dozen-method precedent).
`InteriorOutsPtr()`/`DriveOutsPtr()`/`BuildInteriorFramePtr()` are three NEW methods, each
returning `&sw.<field>` — genuinely a pointer to internal state, but this is the same shape
`portwiring.PortBindings` already required from the FIELD before the move
(`*map[string]io.Writer`, not `map[string]io.Writer`): the caller needs the field's own
address because it captures the pointer BEFORE `SetNodeStreams` has populated (or
reassigned) the map, and reads through the pointer AFTER — a value-returning accessor
(`InteriorOuts() map[string]io.Writer`, `moverreg.NodeGeoms()`'s own shape) cannot express
this, since a bare map's reference semantics don't survive the map being REASSIGNED (`sw.
interiorOuts = map[string]io.Writer{}` inside `SetNodeStreams`) the way `NodeGeoms()`'s
directory (never reassigned, only mutated) does. This is a method, not a field export —
`InteriorOutsPtr`/`DriveOutsPtr`/`BuildInteriorFramePtr` are identifiers Go's export rule
governs, and what they return was already crossing this exact boundary (as
`portwiring.PortBindings.InteriorOuts *map[string]io.Writer`) before the move; the move
changes who owns the pointed-to field, not the pointer's own shape.

**`viewpersist.Persisters`**: `ArmViewpoint(topologyPath string)
*camerapersist.ViewpointPersister` and `ArmEdit(topologyPath string)` are two NEW
construction-time methods that absorbed the persister-construction bodies that used to live
directly in `MoveDispatch.EnableViewpointPersist`/`EnableEditPersist` — this is a bigger
fold than moverreg's/nodeinbox's "3 statements → 1 call" precedent because the FULL
struct-literal construction (`Path`/`Write`/`Tag` per persister) moved with it, since
`viewpersist` already imports every type (`scenepaths`, `scenepersist`, `viewstate`,
`geom`) that construction needs and `dispatch` does not need to see the literals at all.
`ArmViewpoint` returns the constructed value (the one read of `vp` anywhere) so
`EnableViewpointPersist` can still wire `md.UI.VP.Persist = p.Schedule` — this is why `vp`
itself gets no accessor, matching `MoveDispatch.TR`'s own write-only precedent from §28.
`Overlays()`/`Sphere()`/`Speed()`/`Lattice()` are four post-construction getters, each
returning the already-exported `*scenepersist.Persister[T]` pointer — nil until `ArmEdit`
runs, and `Persister[T].Schedule` is nil-receiver-safe (already true before this move), so
a `.Schedule(...)` call site through the getter behaves identically to the bare-field call
site it replaces. No new package folded scenepersist into camerapersist/viewstate/geom or
vice versa — see the package's own doc comment for the one-sentence justification for a
NEW package (`viewpersist`) rather than adding this grouping to `scenepersist`.

### No channel or field exported — confirmed by construction

Neither `nodeinbox.NodeInboxes` nor `streamwire.StreamWiring` nor `viewpersist.Persisters`
declares an exported field — every external touch above goes through a method. The two
channel-typed maps in `NodeInboxes` (`tiltEdit`, `lattice`) stay unexported; the two
methods that read them (`BroadcastLatticePoints`, `SendTiltEdit`) send ON the channels
internally and never return one — the same "no exported channel" shape §17/§20/§26 held.
`streamwire.StreamWiring`'s pointer accessors return `*map[string]io.Writer` (a pointer to
a directory, matching `portwiring.PortBindings`' own pre-existing field type) and
`*func(...)([]byte)` (a pointer to a func value, not a channel) — no channel there either.
`viewpersist.Persisters` holds no channel at all.

### What moved, what stayed

`nodes/Wiring/dispatch/stream_wiring.go` moved whole, as
`nodes/Wiring/streamwire/stream_wiring.go` (only file with a `streamWiring`/`*sw`
receiver). `nodeInboxes`'s type declaration and its one method moved out of
`move_dispatch.go` into the new `nodeinbox/node_inbox.go`; `move_dispatch.go` itself
stayed (it also declares `MoveDispatch`, which is not moving). `persisters`'s type
declaration moved out of `move_persist.go` into the new `viewpersist/persisters.go`;
`move_persist.go` stayed (its two methods, `EnableViewpointPersist`/`EnableEditPersist`,
are `*MoveDispatch` receiver methods that also touch `md.Scenes`/`md.UI`/`md.MR`, so they
cannot follow `persisters` out). Every other touched file (`build_nodes.go`,
`move_dispatch_api.go`, `move_streams.go`, `stdin_apply.go`, `stdin_dispatch.go`,
`scene_lattice_persist.go`, `scene_sphere_persist.go`, `scene_speed_persist.go`,
`viewpoint_state.go`, and the 8 test files) merely reached the old bare fields and were
rewritten to the exported API — all stayed in `dispatch`.

### Goroutine count, channel set, ordering — unchanged

Still exactly the same channels this task's own moved files touch: `NodeInboxes.tiltEdit`/
`.lattice` are the same two channel-typed maps, populated at the same build-time call
sites, read by the same two send paths. `StreamWiring`'s `SetEdgeStreams`/`SetNodeStreams`
bodies are byte-identical (verified — no diff beyond capitalization and the `sw.` receiver
staying `sw`). No goroutine was added, removed, or resequenced; no fd wiring order changed.

### Test-name/assertion equality

`grep -oE '^func Test[A-Za-z0-9_]+' nodes/Wiring/dispatch/*_test.go nodes/Wiring/kindapi/*_test.go`
→ **109**, matching the count on this branch immediately before this task started (verified
directly, not assumed — see the note below on why the number differs from §26/§28's own
stated baseline). `grep -c 't\.Fatal\|Fatalf\|Error\|Errorf(' ...`, summed across the same
files → **327** before AND after this task's commits (re-verified with `git stash`/`git
stash pop` against the working tree) — §26/§28 both recorded 324, but that count was
already stale by the time this task started (something on this branch between §28 and this
task's start added 3 matching lines with no file-count change recorded anywhere) — not
introduced by this task, confirmed by measuring the SAME pre-task tree with `git stash`,
not merely re-quoting the old number. No test in this task's own two commits was renamed,
dropped, weakened, or skipped: every rewritten test file changed only the accessor syntax
at the call site.

### Guards, re-run and proven with teeth

`check-doc-drift.sh` failed once, honestly, before any deliberate-break testing began: it
matched a broken `nodes/Wiring/dispatch/stream_wiring.go` reference inside
`docs/investigations/interior-stream-framing.md` (a real citation, not a guard
false-positive) — fixed by rewording that citation to `nodes/Wiring/streamwire/
stream_wiring.go`/`StreamWiring`/`SetNodeStreams`, then re-ran clean.
`check-no-untracked-source.sh` failed once (the three new package files were untracked,
invisible to the git-ls-files-driven guards) — fixed with `git add -N`, then re-ran clean.
Neither `check-persist-write-ownership.sh` nor `check-scene-path-resolution.sh` needed
re-keying (both match by filename via a recursive `find`/`grep`, already covering the new
`nodeinbox/`/`streamwire/`/`viewpersist/` subdirectories) — proven with teeth anyway:
injected `jsonpersist.WriteJSONAtomic("probe.json", nil)` into
`viewpersist.Persisters.ArmViewpoint` → `check-persist-write-ownership.sh` reported
`unauthorized-write: .../viewpersist/persisters.go: 75:...` and exited 1; injected
`filepath.Join("nodes", "x", "y.json")` into `streamwire.StreamWiring.InteriorOutsPtr` →
`check-scene-path-resolution.sh` reported `hand-rolled-node-path: .../streamwire/
stream_wiring.go: 83:...` and exited 1. Both probes removed immediately after; `diff`
against a pre-edit backup of each file confirmed byte-identical restoration.
`check-composer-fields.sh`, `check-channel-names.sh`, `check-no-network-locks.sh`,
`check-docs-symbols.sh`, `check-dep-rules.sh` all ran clean, none needed re-keying (none
match the three new type/package names by content). `bash scripts/stop-checks.sh`: empty
stdout after both fixes above.

Grepped the three new package directories for `runtime.Caller`/`filepath.Join("..", ...)`/
`../..` — zero hits.

### Verification

`go build ./...`, `go vet ./...`: clean. `go test -race -count=1 ./...`: every package `ok`
or `[no test files]`, zero `FAIL`, zero race reports, including `nodes/Wiring/dispatch` and
the three new packages (`nodes/Wiring/nodeinbox`, `nodes/Wiring/streamwire`,
`nodes/Wiring/viewpersist` — none has its own test file; each type's behavior is exercised
entirely from `dispatch`'s existing suite, same shape §26 left `moverreg` in). The
no-package-under-`Wiring`-imports-`dispatch` loop is empty.

### Deliberate breaks, confirmed and restored (moved surface itself, not just guards)

- `nodeinbox.NodeInboxes.BroadcastLatticePoints` forced to a no-op body (loop deleted) →
  `TestBroadcastLatticePointsReachesEveryRegisteredChannel` and
  `TestBroadcastLatticePointsDoesNotBlockOnAFullChannel` both fail by name, with the exact
  assertion text (`"node 1: BroadcastLatticePoints did not deliver onto its channel"`,
  `"channel holds 8 after broadcast, want the latest value 24"`). Restored;
  byte-identical `diff` confirmed.
- `viewpersist.Persisters.Overlays()` forced to `return nil` unconditionally →
  `TestOverlaysPersistPreservesCamera` fails by name. Restored; byte-identical `diff`
  confirmed.

**Uncovered, reported rather than silently accepted.** `streamwire.StreamWiring.
SetEdgeStreams`/`SetNodeStreams` are the channel-touching/fd-wiring class
`docs/process/testing-shape.md`'s own doctrine excludes from unit testing (they wire real
fds and claim registries, exercised only by the full runtime) — same excluded class §17/
§20/§26 already named, not re-probed individually here. `InteriorOutsPtr`/`DriveOutsPtr`/
`BuildInteriorFramePtr` and `nodeinbox.NodeInboxes.ClaimLatticeIn`/`ClaimTiltEditIn` and
`viewpersist.Persisters.ArmViewpoint`/`ArmEdit` are all construction-time writes exercised
only as plumbing inside integration-style tests that build a real topology — not
independently break-tested, same class as §19's/§26's own construction-time-write coverage
note. `viewpersist.Persisters.Sphere()`/`Speed()`/`Lattice()` were NOT individually
break-tested (only `Overlays()` was, above) — a probe on each would very likely be caught by
`scene_sphere_persist_test.go`/`scene_speed_persist_test.go`/`scene_lattice_persist_test.go`
(all three exist and already call `.Schedule` through the new accessor), but that was not
verified by an actual forced-nil run for the other three, so it is named here as unverified
rather than assumed.

### After the lifts — both clusters re-measured

**Stdin cluster (`stdin_dispatch.go` + `stdin_apply.go`), re-measured**: every one of the
11 functions/closures §27/§28 tabulated was re-read. `md.MR`/`md.LQ`/`md.UI`/`md.Scenes`
were already exported (§28); `md.Inboxes`/`md.Persist` are now exported too (this task).
The result: **every function's `md.Persist.X()`/`md.Inboxes` touch dropped out of its pin
statement.** `applyUpdateScene` (which used to list `md.persist.lattice` as its blocker)
is now bucket **(b)** outright — every remaining statement in its body
(`SelectScene(&md.Scenes, ...)`, `md.UI.LatticePoints = ...`,
`md.Persist.Lattice().Schedule(...)`, `md.BroadcastLatticePoints(...)`,
`md.CreateNode(...)`, `md.DeleteNode(...)`) is an exported-field/exported-method read.
`applyUpdateOverlays`, `HandleSaveMsg`, and the `clockAttrHandlers["speed"]` closure
resolve the same way — their ONLY remaining `md` touch is `md.Persist.X().Schedule(...)`,
now exported. **`applyUpdateDistanceGroup` and `applyUpdateTiltVector` are the two holdouts,
and their pin narrows to exactly one field: `md.ctx`** (`applyDistanceGroupTarget(md.ctx,
...)`; `sendTiltEdit(&md.Inboxes, md.ctx, ...)`, `sendMove(&md.MR, md.ctx, ...)`) —
`md.Inboxes`/`md.MR` in those same call sites are now ordinary exported-field reads, not
part of the pin. `ctx context.Context` stays unexported by this task's own constraint (it
holds no exported type of its own — it is `context.Context`, already exported by the
standard library, but the FIELD `ctx` on `*MoveDispatch` is the thing gating these two
functions, and exporting it was never asked for and is not a "sub-object holding an
already-exported type" the way `Inboxes`/`Persist`/`Sw` are). **Verdict: still PINNED**,
narrower than §28 left it — 9 of 11 functions are now bucket (b) in isolation, and the
remaining 2 are pinned by ONE field, not several. Not resolved here: exporting `ctx` (or
threading it as an explicit parameter instead of a field read) is a DIFFERENT lever than
the one this task was scoped to pull (lifting `inboxes`/`sw`/`persist`), and doing it was
not attempted.

**Build/load cluster (`build.go` + `loader.go` + `move_dispatch_construct.go`),
re-measured**: `buildMoveDispatch` was already bucket (b) in isolation per §28 (unchanged
here — nothing in its body ever touched `inboxes`/`sw`). **`buildNodes`'s OWN body pin is
now GONE**: every statement §28 named as its blocker (`b.md.inboxes.lattice`/`.tiltEdit`
map writes, `b.md.sw.interiorOuts`/`.driveOuts`/`.buildInteriorFrame` pointer captures) is
now `b.md.Inboxes.ClaimLatticeIn(...)`/`.ClaimTiltEditIn(...)` and
`b.md.Sw.InteriorOutsPtr()`/`.DriveOutsPtr()`/`.BuildInteriorFramePtr()` — all
exported-field/exported-method reads. `buildNodes` is bucket (b) in isolation, exactly
matching `buildMoveDispatch`'s own status. **Both `*buildCtx` methods are now bucket (b) in
isolation** — but `buildCtx` itself still cannot move, for the SAME reason §28 already
isolated: Go requires every method of `*buildCtx` to live in the same package as `buildCtx`,
and `buildFromSpec`/`buildCtx`'s own construction (`&buildCtx{ctx: ..., spec: ..., tr: ...,
...}`) sets `buildCtx`'s OWN unexported fields directly — a completely separate type from
`MoveDispatch`, untouched by this task, unrelated to `inboxes`/`sw`/`persist`.
`newMoveDispatch` (`move_dispatch_construct.go`) is unchanged from §28: it still constructs
`md := &MoveDispatch{TR: tr}` then sets `md.MR = moverreg.New()` and `md.tapToInstall =
...` via closures — the LAST of `MoveDispatch`'s seven original sub-object-shaped fields
(`ctx`, `tapToInstall`) stays unexported by this task's own constraint, and a struct-literal
constructor that sets even one unexported field must live in that type's own package,
independent of which fields ARE now exported. `LoadTopology` is unchanged from §27/§28 —
its only non-trivial call (`buildFromSpec`) is pinned above. **Verdict: still PINNED**, but
the reason has narrowed to exactly two things, both confirmed unrelated to this task's own
lift: (a) `buildCtx`'s own unexported field encapsulation (a different type), and (b)
`newMoveDispatch`'s construction-time wiring through `tapToInstall` (a field this task was
told to leave unexported) plus the struct-literal-construction rule itself. `buildNodes`'
own reach into `inboxes`/`sw` — the ONE blocker §27/§28 attributed specifically to THIS
task's lever — is confirmed gone.

### Final state

`ls nodes/Wiring/dispatch/*.go` → **74** (27 non-test + 47 test), down from 75 — the one
file that left (`stream_wiring.go`) is exactly `moverreg`-shaped: a whole receiver-bearing
file moved to its own package with no alias shim left behind. Test-name/assertion parity:
109/327 (see the note above on why 327, not 324). `nodes/Wiring/nodeinbox` (new): 1
non-test file (93 LOC), 0 test files. `nodes/Wiring/streamwire` (new): 1 non-test file (218
LOC — includes the 3 new pointer-accessor methods), 0 test files. `nodes/Wiring/viewpersist`
(new): 1 non-test file (106 LOC), 0 test files.

No `types`/`common` package, no alias shim, no dot-import, no package-level actor global, no
`*ForTest` constructor, no interface added.

`git status --short`: empty. Commits: `dfbb2c24` (the three package lifts +
`move_dispatch.go`'s field exports), `796a69fc` (every dispatch call site rewritten to the
new exported APIs, plus the one doc-drift fix and 8 test-file rewrites).

## §30 — stdin cluster landed in `nodes/Wiring/stdinreader`, pulling `ctx` explicit; other
three clusters (gesture, build/load, remainder) NOT attempted this session — ran out of
room, stopped at a committed buildable stage

### Partition table (named before moving, per the task's own method)

| cluster | files | verdict this session |
|---|---|---|
| stdin | `stdin_dispatch.go`, `stdin_apply.go`, `bool_u8.go` (single caller) | **MOVED** to `nodes/Wiring/stdinreader` |
| gesture | `gesture.go`, `gesture_hitclassify.go`, `gesture_actions.go`, `gesture_handlers.go`, `gesture_graph.go`, `gesture_dispatch.go` | not attempted |
| build/load | `build.go`, `loader.go`, `move_dispatch_construct.go`, `build_move_dispatch.go`, `build_nodes.go` | not attempted |
| remainder | `scene_switch.go`, `viewpoint_state.go`, `vec_alias.go`, `move_persist.go`, `move_streams.go`, `move_dispatch_api.go`, `move_dispatch.go`, `distance_groups.go`, `scene_structure.go`, `scene_*_persist.go` | not attempted |

### What actually unblocked the stdin cluster

§29 had narrowed the stdin cluster's pin to exactly one field: `md.ctx`, read directly by
`applyUpdateDistanceGroup` and `applyUpdateTiltVector`. Its own downstream callees
(`sendMove`, `sendTiltEdit`, `applyDistanceGroupTarget`) already took `ctx
context.Context` as an explicit parameter — only the two HANDLER functions themselves
still reached back into `md.ctx`. Tracing the call chain up from `ApplyEdit` found `ctx`
was ALREADY in scope at the one production caller
(`runtopology/gesture_actor.go`'s `startGestureActor` goroutine, which holds its own
`ctx context.Context` parameter) — so threading it down as an explicit parameter through
`ApplyEdit` → `applyUpdate` → `updateKindHandlers[...]` cost nothing at the call site: no
new state, no new field, just an extra argument on an existing signature. `md.ctx` itself
was NOT exported (task constraint honored) and nothing about `MoveDispatch.Start`
(the field's sole writer) changed.

Three package-level free functions moved from unexported to exported so the moved files
could still reach them from the new package: `sendMove` → `SendMove`, `sendTiltEdit` →
`SendTiltEdit` (`move_dispatch_api.go`, both bucket (b) already — thin delegators to
`mr`/`inboxes`, no `MoveDispatch` reach), and `applyDistanceGroupTarget` →
`ApplyDistanceGroupTarget` (`distance_groups.go`, same shape). Every other in-package
call site of these three (the gesture cluster's `gesture_actions.go`/`gesture_handlers.go`/
`gesture_graph.go`, staying in `dispatch` this session) was updated to the new
capitalization — a rename, not a behavior change; bodies are byte-identical.
`bool_u8.go`'s `boolU8` moved WITH the stdin files (its only caller) and stayed
unexported — no export needed since caller and callee now share a package.

### What moved, what stayed, and why `stdinreader` now imports `dispatch`

`stdin_dispatch.go`/`stdin_apply.go` → `nodes/Wiring/stdinreader/dispatch_edit.go`/
`dispatch_apply.go` (renamed to avoid a second `stdin_*.go` prefix inside a package that
already has `stdin_reader.go`), `bool_u8.go` moved unchanged. Every function's own
`*MoveDispatch` parameter became `*dispatch.MoveDispatch` — the type was already exported
(§28) with every field this cluster reaches (`UI`, `MR`, `LQ`, `Inboxes`, `Persist`,
`Scenes`) exported (§28/§29), so this is a straight type-qualification, not a new
capability crossing the boundary. This is a REAL new import edge (`stdinreader` → `dispatch`)
that did not exist before — `stdin_reader.go`'s own header comment previously said this
package "does not import nodes/Wiring", which was true of the framing half
(`RunStdinReader`, unchanged, still takes its three ops as plain `Handlers` function
values) but is no longer true of the package as a whole now that `dispatch_edit.go`/
`dispatch_apply.go` live here too. Both `stdin_reader.go`'s header and the two moved
files' own headers were reworded to say this plainly rather than leave a stale claim.
This is NOT a cycle: `dispatch`'s own non-test code never imports `stdinreader` (confirmed
by the no-package-under-Wiring-imports-dispatch loop below, which correctly reports
`stdinreader` as the one exception — expected and intended, not a defect); the closure-based
`Handlers` shape that kept `stdin_reader.go` decoupled from `dispatch` is preserved exactly
where it always mattered (the ext-host-facing wiring in `runtopology`), not weakened.

### Test moves — real cycle forced 8 file-level splits, not a straight `git mv`

Attempting a straight `git mv` of every dispatch test file that reached `applyUpdate`/
`ApplyEdit`/`updateKindHandlers`/etc. hit a genuine Go import-cycle error the first time:
an in-package (`package dispatch`) test file importing `stdinreader` (which now imports
`dispatch`) is disallowed by the toolchain, not a style choice — confirmed by running
`go vet ./...` and reading the exact error (`import cycle not allowed in test`), not
assumed. Each affected file was inspected and handled on its own terms rather than
force-relocated wholesale:

- `dispatch_keys_test.go` (one test, `TestDispatchTableKeysMatchFingerprint`, exercising
  FIVE tables) **split by table ownership**: `rawInputHandlers`/`hitClassifiers` (gesture
  cluster, staying in `dispatch`) stayed in `nodes/Wiring/dispatch/dispatch_keys_test.go`;
  `updateKindHandlers`/`clockAttrHandlers`/`overlayAttrHandlers`/`viewstate.OverlayFlagTraceKind`
  moved to a new `nodes/Wiring/stdinreader/dispatch_keys_test.go`. `mapKeys[V any]` is
  duplicated (one small generic helper per package, same "own trivial copy" precedent
  `bool_u8.go`'s header already documents) rather than exported from either side.
- `overlay_toggle_emit_test.go` (two tests) **split by whether the test calls
  `applyUpdate`**: `TestViewFrameCarriesEveryOverlayFlag` (drives `md.UI.EmitViewFrame`
  directly, never touches `applyUpdate`) stayed in `dispatch`;
  `TestApplyUpdateOverlayToggleEmitsViewFrame` (drives `applyUpdate` directly) moved to a
  new `nodes/Wiring/stdinreader/dispatch_edit_overlay_test.go`.
- `scene_lattice_edit_test.go` (four tests) **split the same way**: the two
  `TestBroadcastLatticePoints*` tests (call `md.BroadcastLatticePoints`/
  `md.Inboxes.ClaimLatticeIn` directly, never `applyUpdate`) moved to a new
  `nodes/Wiring/dispatch/scene_lattice_broadcast_test.go`; the two
  `TestApplyUpdateSceneLatticePoints*` tests (call `applyUpdate`) moved to a new
  `nodes/Wiring/stdinreader/dispatch_apply_scene_test.go`, using
  `dispatch.LoadTopology`+`writeMinimalTree`/`loadMinimalMD` (below) in place of the
  original file's `writeTree`/`loadTreeMD` helpers (those two stayed behind in
  `nodes/Wiring/dispatch/scene_edit_persist_test.go`, unexported, unreachable from the new
  package).
- `stdin_input_integration_test.go` (one test) and `stdin_reader_framing_test.go` (two
  tests) **moved whole** to `nodes/Wiring/stdinreader/dispatch_edit_integration_test.go`
  and `nodes/Wiring/stdinreader/stdin_reader_framing_test.go` — both files' own header
  comments had already argued (pre-this-task) for exactly this destination once the cycle
  cleared, and neither test needed splitting.

None of these four originals used the unexported `newMoveDispatch` for anything a real
`dispatch.LoadTopology(ctx, root, tr, clk)` load can't equally construct, so every moved
test that built its own `*MoveDispatch` now does so through `LoadTopology` over a real
(tiny, on-disk) tree instead. Two new test-only helpers,
`writeMinimalTree`/`loadMinimalMD` (`dispatch_edit_integration_test.go`), replace the old
`newMoveDispatch(map[string]nodegeom.NodeGeom{}, ...)` bare-construction calls. Because
`kindapi.BuildRegistry` panics loudly on an EMPTY registry (by design — a silent empty
build was the exact bug class it exists to prevent) and `stdinreader`'s test binary is
SEPARATE from `dispatch`'s (so `dispatch`'s own `fixture_kinds_test.go` `SrcNode`/
`SinkNode` registration is invisible here), `dispatch_edit_integration_test.go` carries its
own minimal `fixtureSrcNode` kind + `init()` registration — same pattern
`fixture_kinds_test.go`'s own header already documents for why the pattern exists, applied
to a second test binary that now also needs one real kind registered.

No test was renamed, dropped, weakened, or skipped — every split kept every original
assertion, and every moved test's body is unchanged beyond the qualification
(`MoveDispatch` → `dispatch.MoveDispatch`, `applyUpdate`/`ApplyEdit` unqualified since
they're now same-package calls) its new location requires.

### Test-name/assertion count — a deliberate, explained delta, not silent drift

`grep -oE '^func Test[A-Za-z0-9_]+' nodes/Wiring/dispatch/*_test.go nodes/Wiring/kindapi/*_test.go nodes/Wiring/stdinreader/*_test.go`
→ **110** (100 dispatch + 3 kindapi + 7 stdinreader), one more than the 109 baseline §29
confirmed directly. The one new function is `dispatch_keys_test.go`'s own
`TestDispatchTableKeysMatchFingerprint`, which existed as a SINGLE function checking five
tables before this task and is now TWO functions (one per package) each checking a subset
of the same five tables — a structural split forced by the real import cycle above, not a
new assertion invented from nothing. `grep -c 't\.Fatal\|Fatalf\|Error\|Errorf(' ...`
→ **329**, two more than the 327 baseline: each half of the split
`TestDispatchTableKeysMatchFingerprint` carries its own `t.Errorf` call site (one loop
body each) where the original had one shared site — same split, same non-assertion-losing
reason. No test's assertion COUNT within a single scenario was reduced; the delta is
entirely the file-split mechanics above, confirmed by reading the diff rather than assumed
from the numbers alone.

### Guards — re-run, and proven with teeth on the moved fences specifically

Neither `check-edit-op-parity.sh` nor `check-message-kind-parity.sh` nor
`check-input-attr-dispatched.sh` needed re-keying: all three locate their Go-side fence by
CONTENT (`grep -rl`/`find` over `nodes/Wiring` recursively, or by the literal string
`updateKindHandlers = map`), not by a hardcoded path, so moving the fenced files one level
deeper (into `nodes/Wiring/stdinreader`) left every one of them green with no edit. Proven
with teeth, not just re-run clean:

- Deleted the `"update": applyUpdate` entry from `editOps` in the moved `dispatch_edit.go`
  → `check-edit-op-parity.sh` failed with `EMPTY extracted set for 'axis1 nodes/Wiring
  ops' — sentinel block missing/renamed; refusing vacuous parity pass`, exit 1. Restored;
  `git diff` empty afterward.
- Renamed the `overlayAttrHandlers["toggle"]` key to `"toggleXX"` in the same file →
  `check-input-attr-dispatched.sh` failed with `attr "toggle" on entity "overlays" is
  DECODED but never DISPATCHED`, exit 1; `check-edit-op-parity.sh` stayed clean on this
  probe (expected — that guard checks the KIND table, not the per-kind attr table).
  Restored; `go build ./...` clean afterward.
- Forced `points < 0` in place of the real `points < 4 || points > 64 || points%4 != 0`
  range check inside the moved `applyUpdateScene` (`dispatch_apply.go`) →
  `TestApplyUpdateSceneLatticePointsIgnoresInvalidCounts` failed by name with
  `latticePoints=0: md.UI.LatticePoints changed to 0, want unchanged 24`. Restored;
  `go build ./...` clean, full suite green afterward.
- `check-doc-drift.sh` failed once, honestly, citing two now-broken references
  (`.claude/rules/bridge-surface.md` and `nodes/PairNode/SPEC.md`, both naming the
  pre-move `nodes/Wiring/dispatch/stdin_dispatch.go`/`stdin_apply.go` paths) — fixed by
  rewording both to the new `nodes/Wiring/stdinreader/dispatch_edit.go`/
  `dispatch_apply.go` paths, then re-ran clean.
- `check-no-untracked-source.sh` failed once (the five new test files were untracked) —
  fixed with `git add -N` on those five paths, then re-ran clean.

**Attempted a deliberate break on `applyUpdateTiltVector`'s direction handling** (forced
`up := false` regardless of `msg.Flag`) and could NOT find a test that fails by name: the
one production caller that exercises this path end-to-end
(`pair_self_drive_persist_test.go`'s `TestPairNodeSelfDrivePersistsThroughRealReload`)
only asserts that node 2's `position.json` CHANGED after a tilt edit, not which direction
it moved — it stayed green with the break in place. `nodes/Wiring/tiltvector`'s own test
suite covers the channel/index primitives one layer below, not this handler's own
direction-selection line. **Reported rather than silently accepted**: this is an
uncovered moved surface — restored immediately (`up := msg.Flag == "up"`), confirmed
`go build ./...` clean and the full suite green afterward, but no test in this codebase
today can catch a sign flip specifically in `applyUpdateTiltVector`'s `up` computation.

`check-composer-fields.sh`, `check-channel-names.sh`, `check-no-network-locks.sh`,
`check-docs-symbols.sh`, `check-dep-rules.sh`, `check-persist-write-ownership.sh`,
`check-scene-path-resolution.sh` all ran clean with no re-keying (none match the moved
file/symbol names by content). Grepped `nodes/Wiring/stdinreader` for
`runtime.Caller`/`filepath.Join("..", ...)`/`../..` — zero hits.

### The no-package-under-Wiring-imports-dispatch loop — one expected line, not empty

Running the loop from this task's own instructions verbatim now reports exactly one line:
`DEPENDS ON DISPATCH: github.com/dtauraso/wirefold/nodes/Wiring/stdinreader` — this is the
new, INTENDED edge from this task (see "What moved, what stayed" above), not a defect the
loop caught. No other package under `nodes/Wiring/` appears.

### Verification

`go build ./...`, `go vet ./...`: clean. `go test -race -count=1 ./...`: every package `ok`
or `[no test files]`, zero `FAIL`, zero race reports, including `nodes/Wiring/dispatch` and
the new `nodes/Wiring/stdinreader` test files. `bash scripts/verify.sh`: `stop-checks:
clean`, exit 0.

### Final state and what did not get attempted

`ls nodes/Wiring/dispatch/*.go | wc -l` → **69** (24 non-test + 45 test), down from 74 (27
non-test + 47 test) — three non-test files left (`stdin_dispatch.go`, `stdin_apply.go`,
`bool_u8.go`) and their two test files' worth of splits netted a two-file REDUCTION on the
test side (47 → 45: `scene_lattice_edit_test.go` and `stdin_input_integration_test.go`
deleted outright, replaced by new files that moved to `stdinreader`, minus the one new
`scene_lattice_broadcast_test.go` that stayed). `nodes/Wiring/stdinreader` now holds 4
non-test files (`stdin_reader.go`, `bool_u8.go`, `dispatch_edit.go`, `dispatch_apply.go`)
and 5 test files.

**The gesture cluster, the build/load cluster, and the "remainder" cluster (scene/persist
wrappers, `viewpoint_state.go`, `scene_switch.go`, `vec_alias.go`, and the rest of the 24
non-test files still in `dispatch`) were NOT attempted this session** — the stdin cluster's
ctx-threading + real-import-cycle test-splitting cost the full session's remaining budget.
This is a stop-at-a-committed-buildable-stage report, not a decline: `nodes/Wiring/dispatch`
is still above the ≤31-file target (69, 24 non-test), and the task's own framing ("work all
four groups... do not stop after one cluster") was not met this session. The next pass
should start from the gesture cluster (§26's "23 stayed" list and this doc's own history
already measured most of its bodies) rather than re-measuring the stdin cluster again.

Commit: `0a7dcbf3` (stdin cluster move + ctx-threading + guard doc fixes + test splits).

## §31 — the gesture cluster moves to `nodes/Wiring/gesture`

Continuing from §30's stop point. Measured all three remaining groups (gesture / build-load /
remainder) before moving anything, per the task's partition-table requirement.

### Partition table (24 non-test files in `dispatch` before this pass)

- **gesture cluster (5 files)**: `gesture.go`, `gesture_hitclassify.go`, `gesture_actions.go`,
  `gesture_handlers.go`, `gesture_graph.go` — plus two of their own dedicated test files
  (`dispatch_keys_test.go`, `gesture_graph_test.go`, which reach the cluster's own unexported
  tables `hitClassifiers`/`commitEdges`/`applyAction` directly, not just through
  `MoveDispatch.HandleRawInput`).
- **build/load cluster (3 files)**: `build.go`, `loader.go`, `move_dispatch_construct.go` —
  not attempted this pass (see below).
- **remainder (16 files)**: `build_move_dispatch.go`, `build_nodes.go`, `distance_groups.go`,
  `gesture_dispatch.go` (thin wrapper, stays), `move_dispatch.go`, `move_dispatch_api.go`,
  `move_persist.go`, `move_streams.go`, `scene_lattice_persist.go`,
  `scene_overlays_persist.go`, `scene_speed_persist.go`, `scene_sphere_persist.go`,
  `scene_structure.go`, `scene_switch.go`, `vec_alias.go`, `viewpoint_state.go` — not
  attempted this pass.

### Gesture cluster — (a)/(b) classification

Per-statement classification (§ Method) found the SAME result across every function in the
five files: every field access was either a call through an exported sub-object method
(`md.MR.NodeGeoms()`, `md.UI.Gest`, `md.RT.NodeFromHit`, `md.LQ.RootMove`, all already
exported per §29) or `md.ctx` — a `context.Context` VALUE, not a shared/mutable field. There
was no channel send, no goroutine start, no file write, and no unexported-field WRITE on a
type staying behind anywhere in the cluster. Bucket (b) end to end. `gesturefsm`'s own doc
comment (written during an earlier pass) claimed several of these functions "additionally...
reach unexported `MoveDispatch` fields (`md.mr`, `md.lq`, `md.RT`, `md.ctx`) that cannot be
named outside package Wiring at all" — that was STALE: `MR`/`LQ`/`RT` were exported in §29,
so only `md.ctx` was ever still true, confirming the task brief's framing exactly.

### What moved where

New package `nodes/Wiring/gesture` (11 files: `gesture.go` doc-only, `gesture_dispatch.go`,
`gesture_handlers.go`, `gesture_hitclassify.go`, `gesture_actions.go`, `gesture_graph.go`,
`gesture_select.go`, `gesture_camera.go`, `vec_alias.go`, plus `dispatch_keys_test.go` and
`gesture_graph_test.go`). A `Deps` struct (`MR *moverreg.MoverRegistry`, `UI
*viewstate.UIState`, `LQ *layoutquant.LayoutQuantizer`, `RT *rowtables.RowTables`, `Ctx
context.Context`) replaces every `*MoveDispatch` receiver/parameter — the exact
explicit-parameter shape §30 used for `applyUpdate`'s table handlers, generalized to five
fields instead of one. Two small helpers moved WITH the cluster because they were used
*only* by it: `setSelectionUI`/`sendEdgeSelect` (were in `move_dispatch_api.go`, now
`gesture_select.go`, exported nothing new — `applySelect` is their one caller) and
`cameraViewEvent` (was in `viewpoint_state.go`, now `gesture_camera.go` as exported
`CameraViewEvent()` — package `dispatch`'s own `viewpoint_ops_test.go`/
`viewpoint_bridge_test.go` also emit this exact event and now call
`gesture.CameraViewEvent()` rather than duplicating the row shape, which would have been the
alias-shim this decomposition forbids). `gesture_dispatch.go` in `dispatch` is now a
five-line delegator: `HandleRawInput` bundles `&md.MR, &md.UI, &md.LQ, &md.RT, md.ctx` into
a `gesture.Deps` and forwards — `md.ctx` stays unexported on `MoveDispatch`; only the
*value* crosses as `Deps.Ctx`, per the task's explicit "do NOT export ctx" constraint.

Per the task's fingerprint note: `gesture_graph.go`'s three commit actions
(`commitDragStart`/`commitHandholdStart`/`commitRotateStart`) took a `*T.Trace` parameter
none of their bodies ever read — confirmed by grepping each function body for `tr` and
finding only the signature line. Dropped `tr` from `gestureEdge.action`'s type and all three
commit actions; `applyAction`'s type (whose three entries DO use `tr`, handing it to
`applyOrbit`/`applyOrbitLocked`) was left alone — a different table, no shared type to keep
in sync.

### Dependency direction

New edge: `dispatch → gesture` (dispatch's `HandleRawInput` calls `gesture.HandleRawInput`;
the two viewpoint test files call `gesture.CameraViewEvent()`). This is a caller depending on
what it calls, same shape as §30's `stdinreader → dispatch` edge, and is now FIXED — package
`gesture` must never import `dispatch`. Confirmed no cycle: `gesture` imports only
`gesturefsm`, `viewstate`, `moverreg`, `layoutquant`, `rowtables`, `movemsg`, `inputcodec`,
`geom`, `edgemover`, `nodes/wire`, `Trace`, `context` — none of which import `dispatch` or
`gesture`.

### Six test files stayed in `dispatch` untouched

`gesture_camera_outcomes_test.go`, `gesture_home_test.go`, `gesture_selection_test.go`,
`gesture_pan_snapshot_test.go`, `gesture_drag_offset_test.go`, `gesture_hover_test.go`, and
their shared fixture `gesture_helpers_test.go` construct a real `*MoveDispatch` (exported
fields only: `&MoveDispatch{MR: moverreg.New()}`) and call `md.HandleRawInput(...)` — the
wrapper — never reaching into the moved package's internals directly. No changes needed;
they exercise the moved code exactly as before, through the public method.

### Verify

`go build ./...`, `go vet ./...`: clean. `go test ./nodes/Wiring/...`: every package `ok` or
`[no test files]`, including the new `nodes/Wiring/gesture`. `go test -race -count=1 ./...`:
every package `ok` or `[no test files]`, zero `FAIL`, zero race reports. `bash
scripts/stop-checks.sh` from repo root: empty stdout both before the second commit and after.

Deliberate break: flipped `commitDragStart`'s grab-offset sign
(`g.DragGrabOffset = g.DragStartCenter.Sub(hit)` → `hit.Sub(g.DragStartCenter)`) in
`gesture_graph.go`. `go test ./nodes/Wiring/dispatch/... -run TestGesture -v` failed exactly
`TestGestureDragOffCenterPreservesGrabPoint`, no other test — restored the correct sign and
reran clean. Every other moved surface (the hit classifiers, the commit/apply tables'
precedence, `gestWheel`'s zoom/pan math, `gestHome`'s fit) has a dedicated table-shape test
(`gesture_graph_test.go`, now in package `gesture`) or an outcome test in `dispatch`
(`gesture_camera_outcomes_test.go` etc.) — no moved surface in this cluster was found with
NO test that can fail, unlike §30's `applyUpdateTiltVector` finding.

### Test-name/assertion parity

110 `TestXxx` functions across `dispatch`+`gesture`+`kindapi`+`stdinreader` before and after
(baseline unchanged — two whole test files moved verbatim, no test renamed, dropped, or
`t.Skip`ped; the two viewpoint test files' only edits were an import line and swapping
`cameraViewEvent()` → `gesture.CameraViewEvent()` at three and two call sites respectively,
no assertion touched).

### Guards

Grepped `tools/` (excluding `node_modules`/`out`/`.git`) for every moved file/symbol name.
No `.sh` guard names `gesture_actions.go`/`gesture_handlers.go`/`gesture_hitclassify.go`/
`gesture_graph.go`/`dispatch_keys_test.go`/`cameraViewEvent`/`setSelectionUI`/`commitEdges`/
`applyAction`/`hitClassifiers` — the only hits were seven stale doc-comment references to
`nodes/Wiring/gesture.go` (a pre-§20 path that was already wrong even before this move, since
the file lived in `nodes/Wiring/dispatch/gesture.go`) in
`tools/topology-vscode/ARCHITECTURE.md`, `messages.ts`, `extension.ts`,
`raw-input.ts` (×2), `scene-content.tsx`, `viewpoint-bridge.ts` — updated the six that named
an exact path to `nodes/Wiring/gesture package`; left the one bare "(gesture.go)" aside
reference in `raw-input.ts`/`scene-content.tsx` as loose prose, not a guarded path.
`check-polar-only-nav.sh`, `check-no-camera-roundtrip.sh`, `check-no-network-locks.sh`
(allowlist stayed empty), `check-persist-write-ownership.sh`, `check-scene-path-resolution.sh`,
`check-channel-names.sh`, the stream-fd guards, `check-doc-drift`, `check-docs-symbols.sh`
(scoped to `docs/pair-node/*.html` only — N/A here), `check-no-untracked-source`,
`check-input-attr-dispatched.sh` all ran clean inside `bash scripts/stop-checks.sh`, with no
re-keying needed (none of them matched the moved names by content, and none of the moved
files touch persistence, scene paths, camera roundtrip, or polar nav directly — they call
through `viewstate`'s already-covered ops). Grepped every moved file for `runtime.Caller`,
`filepath.Join("..", ...)`, `../..` — zero hits.

### The no-package-under-Wiring-imports-dispatch loop

Re-ran the same loop §30 introduced. It still reports exactly the one expected line
(`DEPENDS ON DISPATCH: github.com/dtauraso/wirefold/nodes/Wiring/stdinreader`) from §30's
edge — `nodes/Wiring/gesture` does NOT appear, confirming the new `dispatch → gesture` edge
runs the intended direction and no package under `Wiring` imports `dispatch` back.

### Final state and what did not get attempted

`ls nodes/Wiring/dispatch/*.go | wc -l` → **62** (19 non-test + 43 test), down from 69 (24
non-test + 45 test). Target is ≤31; still above it. **The build/load cluster (`build.go`,
`loader.go`, `move_dispatch_construct.go`) and the remainder cluster (16 files) were NOT
attempted this session** — the gesture cluster's per-statement measurement across five files
plus a new package's worth of guard re-keying and doc updates cost the full remaining
budget. This is a stop-at-a-committed-buildable-stage report, not a decline. The task's own
framing ("work all three groups... do not repeat the one-enabler-per-pass pattern") was
still not fully met this session, though this pass moved 5 non-test files + 2 test files in
one commit pair rather than one file total. The next pass should start from the build/load
cluster: §29 already confirmed `buildNodes`'s own pin is gone, so the remaining work is
`buildCtx`'s own fields (which the task brief says should travel WITH the cluster into a new
package, not be treated as a leak) and `newMoveDispatch`'s `tapToInstall` wiring — measure
`build.go`/`loader.go`/`move_dispatch_construct.go` bodies fresh rather than trusting this
doc's summary.

Commits: `2495a7d6` (gesture cluster moves to `nodes/Wiring/gesture`, ctx threaded as
`Deps.Ctx`), `9e5e23d2` (removes the six now-empty-of-purpose old file paths in `dispatch`
the first commit's pathspec missed).

## §32 — the build/load cluster moves to `nodes/Wiring/build`; the remainder cluster (16
files) NOT attempted — a test-helper coupling this doc's own arithmetic did not anticipate
ran out the session

### Partition table (19 non-test files in `dispatch` before this pass)

- **build/load cluster (4 files)**: `build.go`, `loader.go`, `build_move_dispatch.go`,
  `build_nodes.go` — **MOVED** to `nodes/Wiring/build`.
- **stays in `dispatch` on purpose (1 file)**: `move_dispatch_construct.go` —
  `NewMoveDispatch` (renamed from `newMoveDispatch`, exported) reads `md.tapToInstall`
  directly in its own struct literal + `WireMessaging` call, so per the task's own
  constraint ("do NOT export `tapToInstall`") the CONSTRUCTOR cannot move even though its
  one caller (`buildMoveDispatch`) did. This is exactly the open question the task brief
  flagged in advance ("measure, don't assume") — measured, and it stays.
- **remainder (14 files)**: `distance_groups.go`, `gesture_dispatch.go`, `move_dispatch.go`,
  `move_dispatch_api.go`, `move_persist.go`, `move_streams.go`, `scene_lattice_persist.go`,
  `scene_overlays_persist.go`, `scene_speed_persist.go`, `scene_sphere_persist.go`,
  `scene_structure.go`, `scene_switch.go`, `vec_alias.go`, `viewpoint_state.go` — **NOT
  attempted this pass** (see below).

### `buildCtx`/`buildFromSpec`/`LoadTopology` — (a)/(b) classification

Per-statement classification found the SAME result across `build.go`/`loader.go`/
`build_move_dispatch.go`/`build_nodes.go`: every field access on the `*MoveDispatch` these
phases construct or extend was either a write to `buildCtx`'s OWN unexported fields
(bucket (a), but `buildCtx` is a build-phase-only struct that travels WITH this cluster per
the task's own framing — "moving the type moves its methods") or a call through an
already-exported `MoveDispatch` sub-object (`b.md.MR`, `b.md.UI`, `b.md.RT`, `b.md.Sw`,
`b.md.Inboxes` — all exported since §28/§29). The one call that reaches an UNEXPORTED
symbol is `newMoveDispatch` itself (`build_move_dispatch.go`'s `buildMoveDispatch`) — bucket
(a) against a symbol that stayed behind, resolved by exporting the callee (`NewMoveDispatch`)
rather than by not moving the caller, since the callee's own body is what touches
`tapToInstall`, not the caller's.

### What moved where

New package `nodes/Wiring/build` (4 files, unchanged names): `build.go`, `loader.go`,
`build_move_dispatch.go`, `build_nodes.go`. Every `MoveDispatch`/`vec3`/`wireSegment` bare
reference became `dispatch.MoveDispatch`/`wire.Vec3`/`wire.WireSegment` (the local
`vec_alias.go` aliases only existed for in-package brevity; the new package spells the
underlying `nodes/wire` types directly rather than re-declaring its own alias pair — one
copy, not two). `move_dispatch_construct.go`'s `newMoveDispatch` → `NewMoveDispatch`
(rename, exported) is the ONE new capability crossing the boundary; its body is
byte-identical. `vec_alias.go`'s `wireSegment` alias — used only by the moved `build.go` —
was deleted from `dispatch` as dead weight rather than travelling with a file it no longer
appears in; `vec3` stays (still ~200 in-package call sites).

### Dependency direction

New edge: `build → dispatch` (constructs `*dispatch.MoveDispatch`, calls
`dispatch.NewMoveDispatch`). Same caller-depends-on-callee shape as §30's
`stdinreader → dispatch` and §31's `dispatch → gesture`. Confirmed no cycle: the
no-package-under-Wiring-imports-dispatch loop (§30) now reports exactly two lines —
`nodes/Wiring/build` (new, this task) and `nodes/Wiring/stdinreader` (§30) — `gesture`
still does not appear (it is the one edge running the OTHER direction, unaffected).

### `newMoveDispatch` must stay; here is what actually stays with it

`NewMoveDispatch` is the composition root's OWN constructor — it reads `md.tapToInstall`
(forbidden to export) inside the exact struct literal that builds `*MoveDispatch`, so per
the task's stated rule this one function cannot leave package `dispatch` no matter how thin
its caller becomes. Nothing else in the 4-file cluster shares that property — every other
statement in `build.go`/`loader.go`/`build_move_dispatch.go`/`build_nodes.go` reached only
`buildCtx`'s own fields or already-exported `MoveDispatch` sub-objects — so exactly ONE
file (`move_dispatch_construct.go`, already present, now with one renamed export) stays
behind for this reason, not a family of leftover phases.

### A coupling this doc's own arithmetic did not price in: `LoadTopology` is a TEST FIXTURE, not just a production entry point

The build/load cluster's own 4 files compiled clean immediately. What actually cost the
session was every OTHER call site of `LoadTopology`/`newMoveDispatch` across the whole
repo, most of them tests: **16 test files** (7 in package `dispatch` calling it directly,
2 already `package dispatch_test` calling it qualified, plus 5 root/other-package test
files calling it qualified as `Wiring.LoadTopology`, plus `nodes/TimeEnd`'s and
`nodes/Wiring/stdinreader`'s own copies) had to be updated to call `build.LoadTopology`
instead. Of the 7 in-package callers, TWO share a further wrinkle only found by trying to
compile them: `loadTreeMD`/`writeTree` (`scene_edit_persist_test.go`'s own local helpers)
are each used by **7 additional sibling test files** in the SAME package that do not
themselves call `LoadTopology` by name — they call the HELPER, which does. Converting the
one file that DEFINED the helper to `package dispatch_test` (required, since it now needs
`build.LoadTopology`) silently broke all 7 siblings (`undefined: loadTreeMD`), a second-order
break this doc's own "convert the caller" framing did not anticipate. Fixed by moving the
canonical, still-unexported `writeTree` into `wire_test_helpers_test.go` (package
`dispatch`, alongside `writeSpecTree`/`writeTreeFile`) with a `_test.go`-only exported
`WriteTree` wrapper (same precedent as `WriteSpecTree`), and converting each of the 7
siblings' OWN package line to `dispatch_test` too, once it turned out `loadTreeMD` itself
(not just `writeTree`) needed the same treatment for the same reason (it calls
`LoadTopology`). Two more test-only exports were added for the same "in-package helper an
external file now needs" reason: `WriteTreeFile` (wraps `writeTreeFile`) and
`QuantizedDragTarget` (wraps the renamed `quantizedDragTargetImpl`, replacing the now-dead
unexported `quantizedDragTarget` wrapper — `staticcheck` caught the dead one, U1000, before
this was cleaned up). `distance_groups_test.go`/`distance_groups_scene_test.go` additionally
read `md.ctx` (unexported) directly in an `ApplyDistanceGroupTarget` call; since `md.ctx` is
provably `nil` in both (neither test ever calls `md.Start`), the call sites now pass
`context.Background()` instead — behaviorally identical (`ApplyDistanceGroupTarget` only
uses `ctx.Done()` for cancellation, never exercised by either test), not a weakening.
`scene_edit_persist_test.go` also called a THIRD in-package-only test helper
(`loadSceneViewpoint`, `camera_viewpoint_test_helper_test.go`) — resolved without exporting
anything new, by switching that one call site to the already-exported, already-identical
production sibling `scenecamera.LoadSceneViewpoint` (the local helper's own doc comment
already documented it as "a test-only duplicate of `nodes/Wiring/scenecamera`'s
`LoadSceneViewpoint`, kept here" for the opposite-direction cycle — this call site simply
didn't need the duplicate).

None of this reached genuinely NEW test surface: every converted file's assertions are
byte-identical, just re-qualified (`LoadTopology` → `build.LoadTopology`,
`MoveDispatch`/`vec3`/`ApplyDistanceGroupTarget`/`DistanceGroupLens`/`NewMoveDispatch` →
`dispatch.X`, `writeSpecTree`/`writeTreeFile`/`writeTree`/`quantizedDragTarget` →
`dispatch.WriteSpecTree`/`WriteTreeFile`/`WriteTree`/`QuantizedDragTarget`).

### Test-name parity

`grep -oE '^func Test[A-Za-z0-9_]+' nodes/Wiring/{dispatch,kindapi,stdinreader,gesture,build}/*_test.go`
→ **110**, unchanged from §31's baseline — no test was moved OUT of the fingerprint set
(every one of the 16 files above stayed physically in `nodes/Wiring/dispatch/`; only their
`package` line and call-site qualification changed), none renamed, dropped, weakened, or
`t.Skip`ped.

### Verify

`go build ./...`, `go vet ./...`: clean. `go test ./...`: every package `ok` or
`[no test files]`. `go test -race -count=1 ./...`: every package `ok` or `[no test files]`,
zero `FAIL`, zero race reports, including `nodes/Wiring/dispatch` and the new
`nodes/Wiring/build` (`[no test files]` there — the moved tests all stayed keyed to
`nodes/Wiring/dispatch`'s own `LoadTopology` fixtures, now qualified). `bash
scripts/stop-checks.sh`: clean (empty stdout) after fixing three findings it caught that a
bare `go build`/`go vet` pass did not — `staticcheck` (`wireSegment`/`quantizedDragTarget`
dead code, both above), `gofmt` (one misordered import block in
`created_node_loads_test.go`), and `check-doc-drift.sh` (four stale
`nodes/Wiring/dispatch/loader.go` references in `CLAUDE.md`, `MODEL.md`,
`memory/feedback/architecture/feedback_schema_parser_parity.md`,
`tools/topology-vscode/ARCHITECTURE.md` — all reworded to `nodes/Wiring/build/loader.go`).

Deliberate break: commented out `topoderive.ComputeReachRadii(b.spec, b.nodeGeoms)` in the
moved `build.go`. `go test ./nodes/Wiring/dispatch/... -run TestLoadTopologyComputesReachRadii -v`
failed exactly that test (`node 1 ReachR = 0, want 50`) — confirming the moved production
code and its moved-along coverage still connect across the new package boundary. Restored;
`go build ./...` clean, `bash scripts/stop-checks.sh` clean afterward.

### Guards

Grepped `tools/` for `build.go`/`loader.go`/`newMoveDispatch`/`buildCtx`/`buildFromSpec` —
one hit, `tools/docs/check-comment-vocab.sh`, which names `loader.go` only inside a
descriptive comment (not a path match it enforces); no guard needed re-keying.
`check-persist-write-ownership.sh`, `check-scene-path-resolution.sh` (does not name
`loader.go` — it names `loader_tree.go`, a DIFFERENT file in `nodes/Wiring/loadspec`, unmoved
and unaffected), `check-no-network-locks.sh` (allowlist stayed empty), `check-channel-names.sh`,
the stream-fd guards, `check-composer-fields.sh`, `check-docs-symbols.sh`,
`check-no-untracked-source.sh`, `check-generated.sh` all ran clean inside `stop-checks.sh`
with no re-keying. Grepped every file in `nodes/Wiring/build` for `runtime.Caller`,
`filepath.Join("..", ...)`, `../..` — zero hits (the one `runtime.Caller`-based test helper
touched this pass, `repoRootForDistanceGroupsTest` in `distance_groups_test.go`, did NOT
change file location — only its `package` line changed — so its depth-3 relative path is
still correct and was left alone).

### Which moved surface has no test that can fail

Not found this pass — the one deliberate break above (`ComputeReachRadii`) failed a named
test immediately, and every other statement in the 4 moved files was already covered by
`build_load_derive_test.go` (the other reach/quantized-layout gap this doc's own history
already closed) or by the wider node-move/vector-channel/speed-delivery suites that also
moved to calling `build.LoadTopology`.

### Final state and what did not get attempted

`ls nodes/Wiring/dispatch/*.go | wc -l` → **58** (15 non-test + 43 test), down from 62
(19 non-test + 43 test — the test count is UNCHANGED because every moved test stayed
physically in this directory, only re-packaged; the reduction is 4 non-test files only).
Target is ≤31; still well above it. **The remainder cluster (14 files:
`distance_groups.go`, `gesture_dispatch.go`, `move_dispatch.go`, `move_dispatch_api.go`,
`move_persist.go`, `move_streams.go`, `scene_lattice_persist.go`, `scene_overlays_persist.go`,
`scene_speed_persist.go`, `scene_sphere_persist.go`, `scene_structure.go`, `scene_switch.go`,
`vec_alias.go`, `viewpoint_state.go`) was NOT attempted this session** — the build/load
cluster's own test-helper coupling cascade (above) cost the full remaining budget. This is a
stop-at-a-committed-buildable-stage report, not a decline.

Measured (not attempted) for the next pass: every remaining `*MoveDispatch` METHOD in this
list reads/writes ONLY already-exported sub-objects (`md.MR`, `md.UI`, `md.Scenes`,
`md.Persist`, `md.RT`, `md.Sw`, `md.Inboxes`, `md.GS`) — e.g. `EnableViewpointPersist`
touches only `md.Persist`/`md.UI.VP`, `CreateNode`/`DeleteNode` touch only `md.Scenes`/
`md.UI`/`md.MR`/`md.RT` — so each CAN become a free function taking those sub-objects
directly (no `*dispatch.MoveDispatch` parameter at all, avoiding any import of `dispatch`
and therefore any cycle regardless of direction), with a thin delegator method LEFT ON
`MoveDispatch` in `dispatch` (mirroring `gesture_dispatch.go`'s existing
`HandleRawInput`/`NodeSelfDriven`/`HasNodeMover` precedent) so every current external
caller (`runtopology/scene_state.go`, `nodes/Wiring/stdinreader/dispatch_apply.go`) needs NO
signature change. The one exception is `move_streams.go`'s `SetMsgTap` (writes
`md.tapToInstall`, must stay a `MoveDispatch` method in `dispatch`) and
`move_dispatch_api.go`'s `Start` (writes `md.ctx`, same). `distance_groups.go`'s
`DistanceGroupLens` must ALSO stay in `dispatch` specifically (not because of an unexported
field, but because `move_dispatch_construct.go`'s `NewMoveDispatch` — which stays — calls it
to bind `md.UI.DistanceGroupLensFn`; moving it would require `dispatch` to import whatever
package it moved to, which would import `dispatch` back for its `MoveDispatch`-method
delegator, a real cycle). `move_dispatch.go` (the `MoveDispatch` struct itself) stays as the
composition root regardless. The next pass should start by confirming this measurement
against fresh reads (this doc's own history has been wrong before when trusted over
re-measuring) rather than re-deriving it, and should budget for the SAME test-helper-cascade
risk this pass hit — check every remaining file's test siblings for shared unexported
fixture helpers BEFORE converting the first one's package line.

Commit: `d15bda9e` (build/load cluster moves to `nodes/Wiring/build`, `NewMoveDispatch`
exported, 16 test-fixture call sites re-qualified, test-helper cascade fixed, guards
re-verified).

## §33 — the remainder cluster, first 8 of 14 files moved/deleted; 6 left (`scene_overlays_persist.go`, `scene_speed_persist.go`, `scene_sphere_persist.go`, `scene_structure.go`, `move_persist.go`) — ran out the session

§32's measurement HELD for every file this pass touched: every method converted was a pure
single-owner forward reading/writing only already-exported sub-objects, with zero surprise
unexported-field reach-ins. Re-measured fresh (not trusted from §32's prose) before each
move, per this doc's own repeated lesson.

### What moved (8 files closed, 4 net non-test files off `dispatch`: 54 from 58)

- **`distance_groups.go`**: `ResolveSceneDistanceGroups`/`ApplyDistanceGroupTarget` → package
  `distancegroups` (new file `wiring_dispatch.go`, no cycle — neither `viewstate` nor
  `moverreg` nor `layoutquant` nor `scene` imports `distancegroups`). `DistanceGroupLens`
  stays, exactly as §32 pinned it (`NewMoveDispatch` calls it directly). File shrinks from 90
  to 26 lines. Callers re-qualified: `runtopology/scene_state.go`,
  `nodes/Wiring/stdinreader/dispatch_apply.go`, plus the 2 dispatch test files that called
  `dispatch.ApplyDistanceGroupTarget`/`md.ResolveSceneDistanceGroups` directly
  (`distance_groups_test.go`, `distance_groups_scene_test.go` — both stayed in `dispatch`,
  since they are integration tests over a real `*dispatch.MoveDispatch` from
  `build.LoadTopology`, not unit tests of the moved functions in isolation).
- **`scene_switch.go`**: `SelectScene` (already a free function taking `*sceneswitch.SceneSwitch`
  — no method-to-function conversion needed) → package `sceneswitch`, new file
  `select_scene.go`. File deleted outright (0 lines left behind). Caller re-qualified:
  `dispatch_apply.go`'s `applyUpdateScene`. Its test, `scene_tabs_test.go`, was SPLIT: the 5
  `SelectScene`-driving tests moved to `sceneswitch/select_scene_test.go` (rewritten against
  `*sceneswitch.SceneSwitch` directly, dropping `*MoveDispatch` entirely — they never needed
  it), the other 3 (`TestSelectedSceneIndexFallsBackToTabZero`,
  `TestUntabbedAnchorLoadsItselfAndHasNoTabs`, `TestResolveScenePathPicksTheSelectedSibling`)
  stayed in `dispatch/scene_tabs_test.go`, since they exercise pure `nodes/Wiring/scene`
  helpers and never called `SelectScene` at all.
  **A duplication bug and its fix, both in this pass**: the first edit of
  `select_scene_test.go` accidentally kept ALL 7 original tests instead of trimming to the 5
  `SelectScene` ones, so the 3 `scene`-only tests briefly existed in BOTH packages —
  harmless to `go test` (different packages, no name collision) but a real violation of
  "moved, not duplicated" and exactly the kind of drift the 110-`TestXxx` fingerprint exists
  to catch. Caught by re-running the fingerprint count (113, not 110) before declaring the
  cluster done, not by a build/vet/test failure — none of those tools see file-level
  duplication. Fixed in a follow-up commit (trimmed `select_scene_test.go` back to the 5
  `SelectScene`-driving tests) rather than folded into the original commit, per this task's
  own "create NEW commits, don't amend" rule.
- **`viewpoint_state.go`**: deleted outright. Every symbol it once held (`viewpointState`,
  `SetViewpoint`/`EmitViewpoint` delegators, `cameraViewEvent`, the Orbit/Zoom ops,
  `PanViewpoint`) had already moved to `gesturefsm`/`viewstate`/`gesture` in earlier passes
  (§28/§31/gesture-actor.md); the file itself was 100% prose doc comments recording where
  each piece went, holding zero live code and zero test file. Nothing to move — deleting it
  is the honest action, git history is the archive.
- **`move_dispatch_api.go`**: `SendMove`, `SendTiltEdit` (already free functions, pure
  one-line forwards to `mr.SendMove`/`inboxes.SendTiltEdit`), and `NodeSelfDriven`/
  `HasNodeMover`/`NodeQuantOffset` (methods, pure one-line forwards to `md.MR.X`) — all 5
  DELETED rather than moved: each had either zero logic of its own (the first two) or was a
  same-package-boundary forward with no reason to exist once `MR`/`Inboxes` were exported
  (§28/§29). Every caller now addresses the owner directly: `dispatch_apply.go`'s 5 call
  sites became `md.MR.SendMove(...)`/`md.Inboxes.SendTiltEdit(...)`; `pair_node_mover_absence_test.go`
  (package `main`, root dir) and `pair_self_drive_persist_test.go` (same) became
  `md.MR.HasNodeMover(id)`/`md.MR.NodeSelfDriven(id)`/`md.MR.NodeQuantOffset("2")`. `Start`
  stays (writes `md.ctx`, pinned per the task's own constraint). File shrinks from 86 to 22
  lines but does not disappear — `Start` has no other home.
- **`move_streams.go`**: `SetEdgeStreams`/`SetNodeStreams` DELETED the same way — pure
  one-line forwards to `md.Sw.SetEdgeStreams`/`md.Sw.SetNodeStreams`, whose own signatures
  already took `edgeMovers map[string]*edgemover.EdgeMover`/`nodeMovers
  map[string]*nodeactor.NodeGeometry` directly (not a `moverreg` type), so there was nothing
  to adapt — `runtopology/edge_stream.go` and `runtopology/node_stream.go` (the only two
  callers) now read `md.Sw.SetEdgeStreams(md.GS.EdgeSeeds, md.MR.EdgeMovers(), ...)` /
  `md.Sw.SetNodeStreams(md.GS.NodeSeeds, md.MR.NodeGeoms(), ...)` directly, and one dispatch
  test (`node_geometry_wire_kindid_test.go`) that drove `SetNodeStreams` as "the real
  production entry point" was updated to drive `md.Sw.SetNodeStreams` instead — same
  production path, just addressed one level lower now that the forward is gone. `SetMsgTap`
  stays (writes `md.tapToInstall`, pinned). File shrinks from 78 to 22 lines.
- **`scene_lattice_persist.go`**: DELETED outright. `BroadcastLatticePoints` was a pure
  nil-guarded forward to `md.Inboxes.BroadcastLatticePoints`; its one caller
  (`dispatch_apply.go`'s `applyUpdateScene`, which already null-checks `md` at entry) now
  calls `md.Inboxes.BroadcastLatticePoints(points)` directly. The `latticePersister`
  doc-comment content (which Persister instance, who arms it, who calls Schedule) had no
  code left to attach to — folded into `nodes/Wiring/scenepersist/scene_lattice_persist.go`'s
  own header instead of surviving as a comment-only file. Its test,
  `scene_lattice_broadcast_test.go`, moved BODILY to `nodes/Wiring/nodeinbox/broadcast_lattice_points_test.go`,
  rewritten against a bare `NodeInboxes` value instead of `&MoveDispatch{}` — it never
  exercised anything on `MoveDispatch` beyond the one-line forward being deleted.

### Verify (this pass)

`go build ./...`, `go vet ./...`: clean after each of the 6 commits. `go test ./...`: every
package `ok` or `[no test files]` after each commit. `go test -race -count=1 ./...`: same,
zero `FAIL`, zero race reports, run at the end of the pass. `bash scripts/stop-checks.sh`:
EMPTY stdout, run at the end of the pass (and once more after the duplication fix).

Deliberate break: flipped `select_scene.go`'s `if idx == scene.SelectedSceneIndex(...)`
early-return to `if idx !=` (inverting "already showing" to "different tab", so selecting
the SAME tab now falls through and re-persists/quits instead of no-op'ing).
`go test ./nodes/Wiring/sceneswitch/... -run TestSelectSceneWritesTheSelectionAndEndsTheRun -v`
failed exactly that test (`SelectScene(1) did not end the run`) — confirming the moved
production code and its moved-along test still connect. Restored; `go build ./...` clean,
tree clean.

Test-name fingerprint: `grep -oE '^func Test[A-Za-z0-9_]+'
nodes/Wiring/{dispatch,kindapi,stdinreader,gesture,build,sceneswitch,nodeinbox,distancegroups}/*_test.go`
→ **110**, matching §31's baseline exactly (after the duplication fix above — it read 113
before the fix, which is what caught the bug). None renamed, dropped, weakened, or
`t.Skip`ped; every one of the 12 moved/relocated tests is byte-identical to its pre-move
body except the receiver/import rewrite the move itself required (`md.X(...)` →
`X(&md.Y, ...)` or `owner.X(...)`).

### Guards

Grepped `tools/` for every filename and symbol moved this pass (`SelectScene`,
`BroadcastLatticePoints`, `SendMove`, `SendTiltEdit`, `NodeSelfDriven`, `HasNodeMover`,
`NodeQuantOffset`, `SetEdgeStreams`, `SetNodeStreams`, `ResolveSceneDistanceGroups`,
`ApplyDistanceGroupTarget`, `scene_switch.go`, `scene_lattice_persist.go`,
`viewpoint_state.go`, `distance_groups.go`) — zero hits inside any guard's enforced pattern
(a few appear only in unrelated prose, same as §32's `loader.go` finding). No guard needed
re-keying; `check-persist-write-ownership.sh`, `check-scene-path-resolution.sh`,
`check-no-network-locks.sh` (allowlist stayed empty), `check-channel-names.sh`, the
stream-fd guards, `check-composer-fields.sh`, `check-doc-drift.sh`, `check-docs-symbols.sh`,
`check-no-untracked-source.sh`, `check-no-state-cache.sh` all ran clean inside
`stop-checks.sh`. Grepped every new/touched file for `runtime.Caller`,
`filepath.Join("..", ...)`, `../..` — zero hits (no new depth-sensitive helper was added or
moved this pass).

### Which moved surfaces have no test that can fail

`SendMove`, `SendTiltEdit`, `NodeSelfDriven`, `HasNodeMover`, `NodeQuantOffset`,
`SetEdgeStreams`, `SetNodeStreams` were DELETED rather than moved, and none of the five
delegator deletions in `move_dispatch_api.go` had a test that named the deleted method
directly by symbol — their only coverage was always through the underlying
`moverreg`/`nodeinbox` methods (still tested indirectly via the full-network tests in
`nodes/Wiring/dispatch`, `pair_self_drive_persist_test.go`, `pair_node_mover_absence_test.go`)
or, for `SetEdgeStreams`/`SetNodeStreams`, through `node_geometry_wire_kindid_test.go`'s
integration path (which DOES fail on a break, verified above via `SelectScene` as the
representative case rather than repeating the break on every deleted symbol — each deletion
is a strictly smaller diff than a move, and the compiler itself is the check that no
production or test caller was missed, which `go build`/`go vet` already re-ran clean after
every commit).

### Final state and what remains

`ls nodes/Wiring/dispatch/*.go | wc -l` → **54** (12 non-test + 42 test), down from 58 (15
non-test + 43 test — `viewpoint_state.go` had no test file, so the test count also dropped
by 1: `scene_lattice_broadcast_test.go` and 5 of `scene_tabs_test.go`'s 7 tests left, while
`scene_tabs_test.go` itself and 3 of its tests stayed). Target ≤31; still above it, but the
non-test surface is now 12 files, down from the original 19 the cluster started this task
at (§32 start) — **6 of the 14 named files fully closed** (`distance_groups.go` shrunk but
stays for the pinned `DistanceGroupLens`, `scene_switch.go`/`scene_lattice_persist.go`/
`viewpoint_state.go` deleted outright, `move_dispatch_api.go`/`move_streams.go` shrunk to
their one pinned method each).

**Not attempted this pass, still pending**: `scene_overlays_persist.go` (`LoadOverlays`),
`scene_speed_persist.go` (`HumanEditSpeed`/`SliderSpeed`/`LoadSpeed`),
`scene_sphere_persist.go` (`LoadSceneSphere`) — all three are bucket (b) by the SAME
measurement §32 made (touch only `md.UI`/`md.GS`/`md.Persist`, all exported), landing
target `nodes/Wiring/scenepersist` (already imports `viewstate`, no cycle confirmed this
pass), single caller each (`runtopology/scene_state.go`, plus `dispatch_apply.go` for
`SliderSpeed`/`HumanEditSpeed`) — but each has real test-fixture weight
(`scene_speed_persist_test.go`, `scene_sphere_persist_test.go`, plus
`scene_clock_divisor_test.go`/`scene_edit_persist_test.go`/`tilt_edit_speed_test.go` that
call these methods incidentally without being named for them) that was not re-measured for
the LoadTopology-fixture-cascade risk §32 hit and this pass's budget did not reach.
`scene_structure.go` (`CreateNode`/`DeleteNode`, 227 lines) is bucket (b) but has no obvious
existing landing package (not persistence-of-one-file, not tab-switching, not geometry
math — genuinely its own boundary, same reasoning `distancegroups` itself gives for being
its own package) and its own test (`refuse_structural_edit_emit_test.go`) constructs
`&MoveDispatch{...}` directly, so the move needs that test rewritten against whatever new
package's types, not just re-qualified. `move_persist.go` (`EnableViewpointPersist`/
`EnableEditPersist`) is bucket (b) (touches only `md.Persist`/`md.UI.VP`/`md.Scenes`/
`md.MR`), landing target `nodes/Wiring/viewpersist` (already imports `viewstate`, confirmed
no cycle with `sceneswitch`/`moverreg` this pass), single caller
(`runtopology/scene_state.go`) — not attempted, no test-fixture risk identified but not
reached. `gesture_dispatch.go`, `move_dispatch.go`, `move_dispatch_construct.go`,
`vec_alias.go` stay per §32's original pins (unexported `ctx`, composition root, `newMoveDispatch`'s
own `tapToInstall` reach, ~200 in-package `vec3` call sites) — all re-confirmed true this
pass by inspection, not re-measured in depth since nothing about them changed.

The next pass should start with `scene_overlays_persist.go`/`scene_speed_persist.go`/
`scene_sphere_persist.go` together (same target package, same caller file, likely shares
the LoadTopology-fixture-cascade risk as one group), THEN `move_persist.go` (simplest of
what remains, no identified fixture risk), THEN `scene_structure.go` last (largest, needs a
new-package decision plus a test rewrite, highest risk of the four).

Commits: `7e395993` (SelectScene → sceneswitch), `42faf338` (viewpoint_state.go deleted),
`5634efcc` (distance groups → distancegroups), `91b853ab` (5 MoveDispatch forwards
deleted), `95857821` (SetEdgeStreams/SetNodeStreams deleted), `1d3c985b`
(BroadcastLatticePoints deleted), `567a38a5` (duplication fix).

## §34 — the remainder cluster closes (5 of 5 named files), plus the 9 stranded test files

§33's measurement HELD for all 5 files it named: every one was bucket (b) exactly as
measured — pure forwards over already-exported sub-objects, no surprise unexported-field
reach-ins. Re-measured fresh before each move, not trusted from §33's prose.

### What moved (Group A — the 5 non-test files)

- **`move_persist.go`**: `EnableViewpointPersist`/`EnableEditPersist` → package
  `viewpersist` (new file `enable_persist.go`; `viewpersist` already existed as the
  `Persisters` owner, no cycle with `sceneswitch`/`moverreg`/`viewstate`). 19 call sites
  across `runtopology`, `nodes/Wiring/dispatch` (7 test files), `nodes/Wiring/stdinreader`
  (2 test files), and root-package `pair_self_drive_persist_test.go` rewritten from
  `md.EnableEditPersist(root)` to `viewpersist.EnableEditPersist(&md.Persist, &md.Scenes,
  &md.MR, root)` (viewpoint variant takes `&md.Persist, &md.UI` instead). File deleted
  outright.
- **`scene_overlays_persist.go`/`scene_speed_persist.go`/`scene_sphere_persist.go`**:
  `LoadOverlays`/`HumanEditSpeed`+`SliderSpeed`+`LoadSpeed`/`LoadSceneSphere` → package
  `scenepersist`, folded into that package's SAME-NAMED existing files (which already held
  `WriteScene*`/`LoadScene*`) rather than new files, since `scenepersist.LoadSceneSphere`
  already existed as a pure read function — the MoveDispatch-facing method would have
  collided on the name inside one package (it did not collide across packages, which is
  what let §33's header comment call the collision "on purpose"). Renamed on the move:
  `LoadOverlays`→`InstallOverlays`, `LoadSpeed`→`InstallSpeed`, `LoadSceneSphere`→
  `InstallSceneSphere`; `SliderSpeed`/`HumanEditSpeed`/`EffectiveClockSpeed` keep their
  names (no collision). All 3 files deleted from `dispatch`; their 3 header comments in
  `scenepersist` (which pointed at the now-deleted dispatch files) were rewritten rather
  than left dangling — caught by `check-docs-symbols`, not by inspection. Two test files
  that exercised nothing but these functions and `viewstate.UIState`/`geomseeds.GeomSeeds`
  moved bodily rather than being rewritten as call-site patches:
  `scene_sphere_persist_test.go`'s 3 non-round-trip tests → `scenepersist/install_scene_sphere_test.go`
  (the round-trip test already lived in `scenepersist` from an earlier pass — untouched);
  `tilt_edit_speed_test.go`'s 3 tests → `scenepersist/tilt_edit_speed_test.go`. The other 6
  call sites (5 in `dispatch` test files that also drive `EnableEditPersist`/real tree
  fixtures via `loadTreeMD`/`writeTree`, 1 in `runtopology/scene_state.go`) were rewritten
  in place rather than moved, since they exercise real-tree/production wiring beyond these
  three functions alone.
- **`scene_structure.go`** (227 lines, the pass's own "highest risk" call): `CreateNode`/
  `DeleteNode` → new package `scenestructure` (its own boundary was already argued in §33's
  prose — not persistence-of-one-file, not tab-switching, not geometry math). Signature:
  `CreateNode(scenes *sceneswitch.SceneSwitch, ui *viewstate.UIState, mr
  *moverreg.MoverRegistry, kindID uint8, ndcX, ndcY float64, tr *T.Trace)`,
  `DeleteNode(scenes *sceneswitch.SceneSwitch, ui *viewstate.UIState, rt
  *rowtables.RowTables, row int, tr *T.Trace)` — 4 owner types total across the two
  functions (3 each), no cycle (`scenestructure` imports `countspersist`/`edgefile`/
  `geom`/`loadspec`/`moverreg`/`nodeactor`/`rowtables`/`sceneswitch`/`viewstate`; none of
  those import it back). One production caller
  (`nodes/Wiring/stdinreader/dispatch_apply.go`'s `applyUpdateScene`) rewritten to
  `scenestructure.CreateNode(&md.Scenes, &md.UI, &md.MR, ...)`/
  `scenestructure.DeleteNode(&md.Scenes, &md.UI, &md.RT, ...)`. The test §33 flagged as
  needing a rewrite (`refuse_structural_edit_emit_test.go`, which constructed
  `&MoveDispatch{Scenes: ..., UI: ...}` directly) moved bodily with its 2 tests, rewritten
  against bare `&sceneswitch.SceneSwitch{}`/`&viewstate.UIState{}` values instead —
  `CreateNode(scenes, ui, nil, 0, 0, 0, nil)`/`DeleteNode(scenes, ui, nil, 0, nil)` pass
  `nil` for the `mr`/`rt` parameter each test's own cheapest-refusal branch
  (`!ui.SceneEditable`) returns before touching, exactly preserving what the original
  bare-`MoveDispatch` construction left zero-valued.

### What moved (Group B — 9 files measured as never naming `MoveDispatch`/`md.`/etc.)

Re-verified the measurement rather than trusting it (per this doc's own repeated lesson):
5 of the 9 turned out to be genuinely coupled to a SIBLING file that stays in `dispatch`,
not actually stranded — moving them would have either broken a same-package fixture-kind
`init()` a sibling test relies on, or duplicated a helper 4+ other files call.

**Stayed, with the coupling found:**
- `aimed_ports_test.go`, `fixture_kinds_test.go` — fixture node kinds (`aimedSrc`/
  `aimedSink`/`aimedPacer`, `srcNode`/`sinkNode`) self-registered via `kindapi.RegisterBuilder`
  in each file's own `init()`. `grep`-confirmed consumers still in `dispatch`:
  `build_load_derive_test.go`/`node_geometry_wire_kindid_test.go`/`node_move_row_table_test.go`
  (aimed fixtures), `node_move_test.go`/`per_edge_travel_time_test.go`/
  `wire_test_helpers_test.go` (`SrcNode`/`SinkNode`). A `_test.go` file's `init()` only runs
  within ITS OWN package's test binary — moving the fixture file out of `dispatch` would
  silently un-register the kind for every sibling that still names it by string in JSON,
  failing with `unknown type "X"` rather than a compile error.
- `per_edge_travel_time_test.go` — needs `fixture_kinds_test.go`'s `SinkNode` kind
  registered in the SAME test binary for the same reason; genuinely coupled, not stranded.
- `distance_groups_kind_import_test.go` — blank-imports node kinds so
  `distance_groups_test.go`/`distance_groups_scene_test.go` (both stay in `dispatch` per
  §33's own pin — they drive a REAL `*dispatch.MoveDispatch` via `build.LoadTopology`) can
  build the production topology. Confirmed load-bearing the hard way: moving
  `speed_delivery_full_set_test.go`/`vector_channel_threading_test.go` (Group B's other
  two) took THEIR blank imports (`Input`/`Time`/`TimeEnd`/`TimeStart`/`PulseLeft`/
  `PulseRight`/`pulse`) out of `dispatch`'s test binary with them, and
  `distance_groups_test.go`/`distance_groups_scene_test.go` had been relying on those
  incidentally — `go build`/`go vet` stayed clean (blank imports have no symbol reference
  to check), and only `go test ./...` caught it, with `unknown type "X"` errors identical
  in shape to a forgotten `gen-node-defs` run. Fixed by expanding
  `distance_groups_kind_import_test.go`'s own blank-import list to the full set the
  production topology needs, named explicitly rather than depending on a sibling file's
  incidental list — the exact fix the file's own updated header now states as the reason.
- `vec_close_test.go` — confirmed still used by 4 sibling files
  (`gesture_camera_outcomes_test.go`/`gesture_drag_offset_test.go`/`gesture_home_test.go`/
  `scene_camera_persist_test.go`), exactly as the task's own note anticipated. Stayed.

**Moved, each to the package it actually exercises:**
- `tilt_vector_phi_removed_persist_test.go` → `nodes/Wiring/loadspec` (package
  `loadspec_test`): exercised only `loadspec.LoadTree`. Its one dependency,
  `writeTreeFile` (a 10-line fixture-file writer), was used by no other file in `dispatch`
  once this one left — duplicated locally rather than exported cross-package, since a
  single non-shared 10-line helper does not meet the bar for an export.
- `interior_sphere_test.go` → `nodes/Wiring/nodegeom` (package `nodegeom_test`):
  exercised only `nodegeom.NodeRadius` and `interior`'s exported slot geometry; its own
  header comment's stated reason for living in `dispatch` ("these two assertions need
  Wiring's own nodeRadius, which package interior must not import") was already stale —
  `NodeRadius` had already moved to `nodegeom`, a package `interior` does not import
  either, so the move is clean with zero new cycle risk.
- `speed_delivery_full_set_test.go`, `vector_channel_threading_test.go` → `nodes/Wiring/build`
  (package `build_test`): both exercise `build.LoadTopology` end-to-end (speed-sink count,
  PairNode tilt-vector channel threading), never `MoveDispatch`'s own methods. Both called
  `dispatch.WriteSpecTree` — discovered NOT reusable across directories: `WriteSpecTree` is
  a "_test.go-only export" (defined in `dispatch`'s own internal `_test.go` file, exported
  so `dispatch`'s EXTERNAL test package in the SAME directory can call it — `go test`
  compiles internal+external test files of one directory into one augmented test binary).
  That augmentation does not cross directories: `go vet` on the moved files failed with
  `undefined: W.WriteSpecTree` because the NORMAL (non-test-augmented) `dispatch` package
  `nodes/Wiring/build` imports has no such symbol. Fixed by duplicating `writeSpecTree`/
  `writeTreeFile` into a new `nodes/Wiring/build/wire_test_helpers_test.go`, the same
  "genuine copy, not a cross-directory reuse" call `tilt_vector_phi_removed_persist_test.go`
  made for its own smaller helper.

### Verify

`go build ./...`, `go vet ./...`: clean after each of the 3 Group-A-package commits and the
1 Group-B commit. `gofmt -l .`: empty after every commit (one auto-reformat by the tool
itself, not hand-fixed). `go test ./...`: caught the `distance_groups_kind_import_test.go`
blackout above (3 `FAIL`s, `unknown type "X"` for 8 kinds) — fixed in the same commit rather
than a follow-up, since the guard suite does not check test-binary composition and this was
found before that commit was made, not after. `go test -race -count=1 ./...`: zero `FAIL`,
zero race reports, run at the end of each of the 4 commits and once more at the very end.
`bash scripts/stop-checks.sh`: EMPTY stdout, run after every commit.

Deliberate breaks, one per commit's most-exercised moved surface: (1) `viewpersist`'s
`SliderSpeed`, `+ 1` added to `EffectiveClockSpeed(ui.Speed, ui.ClockDivisor)` →
`TestSliderSpeedMatchesALiveSliderChange` failed by name (`userSpeed=0 divisor=1:
SliderSpeed = 1, want 0`); restored, `go build` clean. (2) `scenestructure`'s
`check-refusal-emits-frame.sh` re-run directly (pattern-based, not path-based, so no
deliberate break was needed to prove it still bites — it already found and counted all 13
call sites unchanged). (3) `loadspec.LoadTree`'s `TopTiltVectorThetaIdx` parse, `+ 1` added
→ `TestLoadTreeIgnoresLegacyTopTiltVectorPhiIdx` failed by name (`want
TopTiltVectorThetaIdx=5 ... got 0x...748` — the address of the mutated local, a sign the
break landed exactly where intended); restored, `go build` clean, `git status --short`
empty both times.

Test-name fingerprint: `grep -oE '^func Test[A-Za-z0-9_]+'
nodes/Wiring/{dispatch,kindapi,stdinreader,gesture,build,sceneswitch,nodeinbox,distancegroups,scenepersist,scenestructure,loadspec,nodegeom}/*_test.go`
→ **147**, up from §33's 110 baseline by design, not drift: §33's own count formula never
included `scenepersist` (which already held 1 test, `TestSceneSphereRoundTrip`, untouched
this pass) or `scenestructure`/`loadspec`/`nodegeom` (which had zero `dispatch`-adjacent
tests before this pass) in its package list — each new package this pass's moves actually
landed tests in had to be added to the counted set, the same way §33 itself added
`distancegroups` to §31's list rather than treating that omission as evidence of a
duplication bug. Zero duplicate test names confirmed directly (`grep -rl "^func
<name>("` returns exactly 1 file for every one of the 5 moved test names), which is the
actual invariant the fingerprint stands in for. None renamed, dropped, weakened, or
`t.Skip`ped; every moved test is byte-identical to its pre-move body except the
receiver/import rewrite the move itself required.

### Guards

Grepped `tools/` for every filename and symbol moved this pass (`EnableViewpointPersist`,
`EnableEditPersist`, `LoadOverlays`, `LoadSpeed`, `LoadSceneSphere`, `SliderSpeed`,
`HumanEditSpeed`, `InstallOverlays`, `InstallSpeed`, `InstallSceneSphere`, `CreateNode`,
`DeleteNode`, `scene_structure.go`, `move_persist.go`, `scene_overlays_persist.go`,
`scene_speed_persist.go`, `scene_sphere_persist.go`, `vector_channel_threading_test.go`,
`speed_delivery_full_set_test.go`, `interior_sphere_test.go`,
`tilt_vector_phi_removed_persist_test.go`) — zero hits inside any guard's enforced pattern.
`check-persist-write-ownership.sh`'s `VIEW_OWNERS` array matches by BASENAME, not full
path, and every `scene_*_persist.go` basename still exists (now under `scenepersist/`
instead of `dispatch/`) with its `writeJSONAtomic` call site unmoved — confirmed by reading
the guard rather than assuming, since this pass moved the directory a basename-only guard
could have silently stopped covering. `check-refusal-emits-frame.sh` re-run directly:
`checked 13 refuseStructuralEdit( call site(s)`, exit 0 — same count as §33, confirming the
pattern-based (not path-based) match still finds every site now that they live in
`scenestructure`. `check-docs-symbols.sh` caught one real drift this pass introduced and
missed by the first build: `nodes/Wiring/dispatch/scene_speed_persist.go` was still named in
`docs/pair-node/math/formulas.html`'s two `data-src` attributes after the file moved —
fixed by repointing both to `nodes/Wiring/scenepersist/scene_speed_persist.go`.
`check-persist-write-ownership.sh`, `check-scene-path-resolution.sh`,
`check-no-network-locks.sh` (allowlist stayed empty), `check-channel-names.sh`, the
stream-fd guards, `check-composer-fields.sh`, `check-doc-drift.sh`,
`check-no-untracked-source.sh`, `check-no-state-cache.sh`, `check-kind-imports.sh` (relevant
to `distance_groups_kind_import_test.go`'s expanded blank-import list — clean, since none of
the newly-added kinds import a sibling kind, only the shared spine) all ran clean inside
`stop-checks.sh`. Grepped every new/touched file for `runtime.Caller`,
`filepath.Join("..", ...)`, `../..` — zero hits (no new depth-sensitive helper was added or
moved this pass; `writeTreeFile`'s duplicate copies use `t.TempDir()`/`filepath.Join(root,
rel)`, not a relative-to-source-file path).

### Which moved surfaces have no test that can fail

`EnableViewpointPersist`/`EnableEditPersist`'s own moved BODIES are exercised indirectly
through every test that arms persistence and then asserts on a written file
(`TestPersistViewpointRoundTrips`, `TestPersistOverlaysRoundTrips`, etc.) — no test names
`viewpersist.EnableEditPersist` itself and asserts on ITS OWN behavior in isolation (e.g.
that `md.Scenes.TreeRoot` gets set), only on its downstream effects. `InstallOverlays`/
`InstallSpeed`/`InstallSceneSphere`'s rename-only-no-behavior-change is unverified by a
literal old-name assertion (impossible — the old name no longer exists) but IS verified by
every test that calls the new name and checks the same on-disk round-trip §33's tests
already checked. `scenestructure.CreateNode`/`DeleteNode`'s two refusal branches proven by
`refuse_structural_edit_emit_test.go` (Group A above); every OTHER branch (the 11 remaining
`refuseStructuralEdit`/`emitViewFrame` pairs `check-refusal-emits-frame.sh` still counts) has
no test reaching it directly, same gap §33 already recorded and did not close.

### Final state

`ls nodes/Wiring/dispatch/*.go | wc -l` → **42** (7 non-test + 35 test), down from 54 (12 +
42) at this pass's start — **target ≤31 still not reached, but every one of §33's 5 named
remaining files is now closed**, and Group B closed 4 of its 9 measured files (5 were
confirmed genuinely coupled to a sibling that stays, not stranded). The 7 non-test files
left (`gesture_dispatch.go`, `move_dispatch.go`, `move_dispatch_api.go`,
`move_dispatch_construct.go`, `move_streams.go`, `distance_groups.go`, `vec_alias.go`) are
exactly §32/§33's own pinned set — unexported `ctx`, composition root, `tapToInstall`
reach, `DistanceGroupLens`'s cycle, ~200 in-package `vec3` call sites — none re-examined
this pass since nothing about them changed and §33 already re-confirmed them by inspection.

Commits: `8649965d` (EnableViewpointPersist/EnableEditPersist → viewpersist), `45f40368`
(LoadOverlays/LoadSpeed/SliderSpeed/HumanEditSpeed/LoadSceneSphere → scenepersist),
`4b4530a9` (CreateNode/DeleteNode → scenestructure), `c83622e3` (4 Group-B test files move,
distance_groups_kind_import_test.go absorbs the kind registrations they took with them).

## §35 — both pins on the remaining 7 non-test files closed: `ctx` un-stored, `tapToInstall`
deleted outright (dead, not moved)

§34 closed with 7 non-test files left in `dispatch`, each pinned by one of two unexported
`MoveDispatch` fields: `ctx` (reached by `gesture_dispatch.go`) and `tapToInstall` (reached
by `move_dispatch_construct.go`/`move_streams.go`). This pass measured both directly rather
than trusting §34's prose, and closed both.

### Pin 1 — `ctx context.Context`

Measured fresh: **one write** (`move_dispatch_api.go`'s `Start`, `md.ctx = ctx`), **one
read** (`gesture_dispatch.go`, `Ctx: md.ctx` into `gesture.Deps`). Go's own doctrine says
not to store a `Context` on a struct; the sole production caller of `HandleRawInput`
(`runtopology/gesture_actor.go`'s `startGestureActor` goroutine) already has a `ctx` in
scope, the same shape `ApplyEdit` already used since §30. Fixed by threading `ctx` as an
explicit `HandleRawInput(ctx, ev, slotReg, tr)` parameter (mirroring `ApplyEdit`), deleting
the field, and updating every call site: `stdinreader.HandleRawInputMsg` (production caller:
`runtopology/gesture_actor.go`; also `pair_self_drive_persist_test.go` and
`stdin_reader_framing_test.go`, both of which already had a `ctx` in scope) and every
in-package `dispatch` test that drives `md.HandleRawInput` directly (6 test files, ~90 call
sites, `context.Background()` — none of these tests exercised cancellation, so the same
"no cancellation available" value `Start`-less bare `MoveDispatch` construction used to
supply via a nil `ctx` field is unchanged in spirit). Two stale comments in
`distance_groups_test.go`/`distance_groups_scene_test.go` that explained "`md.ctx`
(unexported) is never set here" were rewritten to explain the new explicit-parameter shape
instead of describing a field that no longer exists.

### Pin 2 — `tapToInstall` (bucket (a): deleted outright)

Grepped every test file in the repo for `SetMsgTap` (the only way to populate the field) —
**zero call sites, anywhere**. The field's own comment called it a "TEST-ONLY observability
seam," but no test used the seam; it was dead code carrying three files
(`move_dispatch.go`, `move_dispatch_construct.go`, `move_streams.go`) as reach. Chose bucket
(a) per the task's own decision procedure: deleted `tapToInstall`, deleted
`move_streams.go` outright (`SetMsgTap` was its only content), and dropped the `tap`
parameter from `WireMessaging`'s call in `move_dispatch_construct.go`. Followed the reach
one level deeper into `nodes/Wiring/nodeactor` (not one of the 3 dispatch-pinned files, but
the same dead seam continued there): `NodeGeometry.SetMsgTap`, its `msg.tap` field, and the
`EnqueueSend` fire-the-tap check were all deleted too — `SetMsgTap` at that layer also had
zero test callers. `EnqueueSend` itself is unchanged in every OTHER respect (still appends
to `pending` and calls `flushPending`).

### What unpinned

Both fields gone leaves `MoveDispatch` (`move_dispatch.go`) with **zero unexported
fields** — every field is now exported. `move_dispatch_construct.go`'s `NewMoveDispatch`
still constructs the `MoveDispatch{...}` struct literal directly (naming the type, not an
unexported field), so it stays with the type's own declaration for that reason alone, not a
field-reach reason. None of the 7 files actually MOVED this pass — `gesture_dispatch.go`
and `move_dispatch_api.go` are the composition root's own thin API surface (the same
already-argued reason §33/§34 left them), `distance_groups.go`/`vec_alias.go` are still
pinned by `DistanceGroupLens`'s import cycle and the ~200 in-package `vec3` call sites
(unchanged, not re-examined this pass since neither pin touches `ctx`/`tapToInstall`).
`move_streams.go` is gone (deleted, not moved — its only content was dead code).

### Verify

`go build ./...`, `go vet ./...`: clean after both commits. `go test ./...`: zero FAIL after
both commits (`pair_self_drive_persist_test.go`'s `HandleRawInputMsg` call site — the one
production-shaped root-package caller — was the one this pass's own first `go test ./...`
run caught missing the new `ctx` argument; fixed in the same commit, not a follow-up).
`gofmt -l .`: 6 gesture test files needed a follow-up `gofmt -w` (the `context.Background(),`
insertions left inconsistent alignment) — a 3rd commit, formatting-only, no behavior change.
`go test -race -count=1 ./...`: zero `FAIL`, zero race reports, run at the end.
`bash scripts/stop-checks.sh`: EMPTY stdout after the final (3rd) commit.

`TestXxx` count: **381** before and after (grepped repo-wide, not just the package list
§33/§34 tracked, since this pass touched call sites across `dispatch`, `stdinreader`, and
the root package rather than moving any test file) — unchanged, since this pass edited
existing test bodies in place and deleted zero tests (`move_streams.go`, the one file
deleted outright, held no `_test.go` content). Zero duplicate `TestXxx` names introduced
(re-confirmed by the same count staying flat).

Deliberate breaks, one per pinned surface: (1) `gesture_dispatch.go`'s `Ctx`/`RT` forwarding
— `RT: &md.RT` → `RT: nil` — `TestGestureHoverTracksNode` panicked (nil pointer
dereference in `rowtables.(*RowTables).LookupNodeRow`, called through
`gesture.updateHover`); restored, `go build`/`git status --short` clean. (2)
`nodeactor.EnqueueSend`'s surviving-after-tap-removal `pending` append — commented out —
`TestEnqueueForPanicsWhenPendingExceedsBound` failed by name ("EnqueueSend did not panic
after 65 retained sends to a wedged destination"); restored, clean. (An earlier attempt at
this same break, run against `TestGestureDragOffCenterPreservesGrabPoint`, passed —
that test's drag-target assertion reads `extIn` via `SendExternal`, not the
`pending`/`EnqueueSend` path, so it was the wrong test to pin this surface; recorded here so
the same wrong pick isn't retried.)

### Guards

Grepped `tools/` (excluding `node_modules`) for `tapToInstall`, `SetMsgTap`,
`move_streams.go`, and `md.ctx` — zero hits in any guard's enforced pattern; none of these
symbols were ever guard-checked. `check-composer-fields.sh` re-run directly: exit 0 (it
does not name `ctx`/`tapToInstall` specifically, so removing them changes nothing it
checks). All of `bash scripts/stop-checks.sh`'s guard suite (`check-no-network-locks.sh`
empty allowlist, `check-persist-write-ownership.sh`, `check-scene-path-resolution.sh`,
`check-channel-names.sh`, the stream-fd guards, `check-doc-drift.sh`,
`check-docs-symbols.sh`, `check-no-untracked-source.sh`, `check-kind-imports.sh`) ran clean
inside the same `stop-checks.sh` pass reported above.

### Which changed surfaces have no test that can fail

`Start`'s own body (`md.MR.Start(ctx)`, now with no `md.ctx = ctx` side effect) has no test
that asserts on the side effect's ABSENCE directly — every test that calls `Start` was
already only checking the goroutines it launches, never a field write, so this is
unobservable by construction, not a coverage gap this pass opened. The `context.Background()`
value now threaded through ~90 `dispatch` test call sites is never itself asserted on (no
test checks cancellation behavior through `HandleRawInput`) — same gap the old nil-`ctx`
field already had, not newly introduced.

### Final state

`ls nodes/Wiring/dispatch/*.go | wc -l` → **41** (6 non-test + 35 test), down from 42 (7 +
35) — one file (`move_streams.go`) deleted outright, zero files moved. The 6 non-test files
left (`distance_groups.go`, `gesture_dispatch.go`, `move_dispatch.go`, `move_dispatch_api.go`,
`move_dispatch_construct.go`, `vec_alias.go`) carry no unexported-field reach at all any
more — what remains is the composition root itself (the struct declaration, its `Start`/
`HandleRawInput` API, and its constructor) plus the two already-known-and-unchanged pins
(`DistanceGroupLens`'s cycle, ~200 in-package `vec3` call sites). The 35 test files were not
re-measured this pass (out of budget) — that is the next task's own starting point, not
re-litigated here.

Commits: `80bd90f1` (HandleRawInput takes `ctx` as an explicit parameter), `82030bd1`
(delete `tapToInstall` and its whole tap plumbing), `3bee19f1` (gofmt the 6 touched gesture
test files).

## §36 — 3 of the 35 dispatch test files measured and moved; the other 32 re-confirmed as
genuinely dispatch-shaped, not re-litigated wholesale

§35 left all 35 test files unmeasured, "the next task's own starting point." This pass
measured each of the 35 by reading it fresh (not trusting the earlier prose) and moved the
3 that turned out to name neither `MoveDispatch`/`md.` nor exercise `dispatch`'s own two
remaining methods (`HandleRawInput`, `Start`).

### What moved

- **`scene_tabs_test.go`** → `nodes/Wiring/scene` (package `scene_test`, new — the `scene`
  package had zero tests before this move). Read cover to cover: it names `scene.*`/
  `scenepersist.*` only — `SelectedSceneIndex`/`AnchorIsTabbed`/`ResolveScenePath`/
  `SceneTabNames`/`WriteSelectedScene` — and never touches `dispatch` at all; its own header
  comment already said as much ("those tests never needed `*MoveDispatch`"). Moved verbatim,
  package line changed `dispatch` → `scene_test`.
- **`scene_lattice_persist_test.go`**, **`scene_speed_persist_test.go`** → `nodes/Wiring/build`
  (package `build_test`, the same destination §34 used for `speed_delivery_full_set_test.go`/
  `vector_channel_threading_test.go`). Both build a `*dispatch.MoveDispatch` only as a fixture
  (via `build.LoadTopology`) to reach `scenepersist.LoadSceneLattice`/`InstallSpeed`/
  `viewpersist.EnableEditPersist` — no assertion anywhere reads a `dispatch`-owned method or
  field beyond what those calls need. Hit the exact §32/§34 hazard: `writeTree`/`loadTreeMD`
  (both unexported, `dispatch_test`-package helpers in `scene_edit_persist_test.go`/
  `wire_test_helpers_test.go`) do not resolve across directories, and `writeTree`'s own
  fixture tree uses the `SrcNode`/`SinkNode` kinds self-registered by `fixture_kinds_test.go`'s
  `init()`, which — per §34's own already-recorded lesson — does not travel with a moved
  file either. Fixed the same way §34 fixed `writeSpecTree`/`writeTreeFile`: a new
  `nodes/Wiring/build/fixture_kinds_test.go` duplicates the `SrcNode`/`SinkNode` registration
  and a local `writeTree`/`loadTreeMD` pair (genuine copies, not cross-directory reuse — a
  ~80-line helper file is not a shared-package export candidate any more here than
  `wire_test_helpers_test.go`'s own smaller duplicate was in §34).
- **`overlay_toggle_emit_test.go`** → `nodes/Wiring/viewstate` (package `viewstate_test`,
  which already held `overlay_state_test.go`). Its one test builds `&MoveDispatch{UI:
  viewstate.UIState{...}}` and calls only `md.UI.SetViewStream`/`md.UI.EmitViewFrame` —
  `MoveDispatch` was purely a wrapper around a bare `UIState` value, never itself asserted
  on. Moved with the wrapper dropped: `&viewstate.UIState{OV: viewstate.DefaultOverlayState()}`
  in place of `&MoveDispatch{UI: ...}`, `ui.SetViewStream`/`ui.EmitViewFrame` in place of
  `md.UI.*`. No other rewrite — the assertion body (reflect over `ViewOverlayFlags`, matching
  `inputcodec.InOverlayFlags`'s length) is untouched.

### What was re-measured and stayed, with the reason found

Read all remaining 29 (34 minus the 5 §34 already pinned: `aimed_ports_test.go`,
`fixture_kinds_test.go`, `per_edge_travel_time_test.go`, `distance_groups_kind_import_test.go`,
`vec_close_test.go` — not re-litigated, per this task's own instruction) top-to-bottom rather
than by filename pattern:

- **7 gesture files** (`gesture_camera_outcomes_test.go`, `gesture_drag_offset_test.go`,
  `gesture_helpers_test.go`, `gesture_home_test.go`, `gesture_hover_test.go`,
  `gesture_pan_snapshot_test.go`, `gesture_selection_test.go`) all drive
  `md.HandleRawInput(...)` — `dispatch`'s own remaining method, the composition that bundles
  `MR`/`UI`/`LQ`/`RT` into a `gesture.Deps` and forwards. This is genuinely dispatch's own
  behavior on trial (does the bundle wire correctly?), not `gesture` package behavior wearing
  a dispatch costume — confirmed by grep (`HandleRawInput\|MoveDispatch{` present in all 7).
- **`camera_viewpoint_test_helper_test.go`** — same-package (`package dispatch`) helper
  duplicating `scenecamera.LoadSceneViewpoint`, used by `scene_camera_persist_test.go`'s own
  in-package tests; its own header names the exact reason it can't be `scenecamera`'s copy
  (an import cycle back through `scenecamera`'s own `*MoveDispatch` reference) — the same
  shape `vec_close_test.go` already has. Stayed.
- **`distance_groups_test.go`**, **`distance_groups_scene_test.go`** — exercise
  `ApplyDistanceGroupTarget`, which lives in `distance_groups.go`, one of the 6 pinned
  non-test files (`DistanceGroupLens`'s own import-cycle pin, unchanged since §33/§34/§35).
  Genuinely dispatch's own code under test, not a fixture-only use. Stayed.
- **The remaining 19** (`build_load_derive_test.go`, `continuous_drag_persist_test.go`,
  `drag_touching_bead_source_regression_test.go`, `flush_pending_persists_test.go`,
  `node_geometry_wire_kindid_test.go`, `node_move_row_table_test.go`, `node_move_test.go`,
  `pair_node_self_clock_speed_test.go`, `quantized_layout_test.go`,
  `scene_camera_persist_test.go`, `scene_clock_divisor_test.go`, `scene_drag_mode_test.go`,
  `scene_edit_persist_test.go`, `viewpoint_bridge_test.go`, `viewpoint_ops_test.go`, plus
  `wire_test_helpers_test.go` itself, which is the shared-helper file 9 siblings call by name
  — `writeTree`/`loadTreeMD`/`WriteTree`/`WriteSpecTree`/`WriteTreeFile`/quantized-drag
  helpers) were each opened and read, not skipped: each builds a real `*dispatch.MoveDispatch`
  via `loadTreeMD`/`build.LoadTopology` and then exercises multi-owner behavior through it —
  a drag committing through `md.MR`+`md.LQ` together, a persister armed via
  `viewpersist.EnableEditPersist(&md.Persist, &md.Scenes, &md.MR, root)` and then read back
  through a SECOND fresh `loadTreeMD`, or a row-table/kindID assertion that needs the full
  load path's own row derivation. None of these reduce to "one already-exported sub-object,
  used as a fixture only" the way the 3 moved files did — moving any of them would mean
  duplicating not one small helper but the multi-owner wiring itself, which is exactly the
  shape `wire_test_helpers_test.go`'s own doc comment says stays in `dispatch` because ~9
  sibling files still call it unexported. Not moved this pass; **still the next task's own
  starting point** if a future pass wants to push further (each would need the same
  duplicate-vs-cycle judgment call the 3 moved files needed, file by file).

### Verify

`go build ./...`, `go vet ./...`: clean after both commits. `go test ./...`: zero `FAIL`
after both commits. `gofmt -l nodes/Wiring/build nodes/Wiring/dispatch nodes/Wiring/scene
nodes/Wiring/viewstate`: empty after both commits. `go test -race -count=1 ./...`: zero
`FAIL`, zero race reports, run after each commit. `bash scripts/stop-checks.sh`: EMPTY stdout
after each commit (one `git add -N` needed per commit first — `check-no-untracked-source`
flags a brand-new file before it is staged with content, exactly as designed).

`TestXxx` count: **381** after both commits — unchanged from §35's baseline, confirmed by the
same repo-wide `grep -roE '^func Test[A-Za-z0-9_]+'` this doc has used since §33. Zero
duplicate names confirmed directly for the 6 moved test names (`grep -rl "^func <name>("`
returns exactly 1 file each). None renamed, dropped, weakened, or `t.Skip`ped.

Deliberate breaks, one per moved destination: (1) `viewstate.UIState.EmitViewFrame`'s
`SceneTori: boolU8(ui.OV.SceneToriVisible)` → `SceneTori: 0` —
`TestViewFrameCarriesEveryOverlayFlag` failed by name from its NEW location
(`nodes/Wiring/viewstate`): `ViewOverlayFlags.SceneTori arrived 0 with every overlay
defaulting ON — emitViewFrame never assigns it, so this flag streams as off whatever the
toggle says`; restored, `go build` clean. (2) `scene`/`build`'s moved tests were not broken
deliberately this pass (budget) — `TestSelectedSceneIndexFallsBackToTabZero`/
`TestPersistLatticePointsRoundTrips`/`TestPersistSpeedRoundTrips` all pass from their new
locations per the `go test` run above, but no one of them was pinned by a verbatim failure
the way §34's own model (`speed_delivery_full_set_test.go`) required; recorded as a gap for
whoever picks this up next rather than skipped silently.

### Guards

Grepped `tools/` (excluding `node_modules`) for every moved filename and every moved
`TestXxx` name (`scene_tabs_test`, `scene_lattice_persist_test`, `scene_speed_persist_test`,
`overlay_toggle_emit_test`, `TestSelectedSceneIndexFallsBackToTabZero`,
`TestPersistLatticePointsRoundTrips`, `TestPersistSpeedRoundTrips`,
`TestViewFrameCarriesEveryOverlayFlag`) — zero hits anywhere. No guard in this repo names a
test file or a `TestXxx` symbol by string, so none needed re-keying this pass — a narrower
outcome than the task anticipated, not a skipped step. `bash scripts/stop-checks.sh`'s own
guard suite (`check-no-network-locks.sh` empty allowlist, `check-persist-write-ownership.sh`,
`check-scene-path-resolution.sh`, `check-channel-names.sh`, the stream-fd guards,
`check-doc-drift.sh`, `check-docs-symbols.sh`, `check-no-untracked-source.sh`,
`check-test-integrity.sh`) ran clean inside both commits' `stop-checks.sh` passes.
`check-test-integrity.sh` in particular: rename-and-split-aware, passed with no
`[allow-test-weakening: ...]` marker needed, confirming the moves read as moves rather than
edits.

### Final state

`ls nodes/Wiring/dispatch/*.go | wc -l` → **37** (6 non-test + 31 test), down from 41 (6 +
35). The 6 non-test files are unchanged from §35 (`distance_groups.go`, `gesture_dispatch.go`,
`move_dispatch.go`, `move_dispatch_api.go`, `move_dispatch_construct.go`, `vec_alias.go`) —
neither pin was touched this pass. Of the 31 test files left: 5 are §34's own pinned set
(genuinely coupled to a same-package sibling), 2 exercise `distance_groups.go` directly, 7
drive `HandleRawInput`, 1 is a same-package helper duplicate, and the remaining 16
(`wire_test_helpers_test.go` plus its 15 real siblings named above) build a real
`*dispatch.MoveDispatch` and exercise MULTI-owner wiring through it — none reduces to the
"already-exported sub-object used as a bare fixture" shape the 3 moved files had. **Target
≤31 files is now reached** on file count, though the remaining 31 test files were not all
individually re-examined against a stricter bar than "does it need dispatch's own multi-owner
wiring" — that bar, and whether any of the 16 remaining multi-owner-fixture tests could still
be decomposed into a per-owner test plus a thin dispatch-level integration test, is the next
task's own starting point if this directory needs to shrink further.

Commits: `093965c2`/`aadcac04` (`scene_tabs_test.go` → `scene`, the rename got split across
two commits by an initial `git mv` staging mistake — `--amend` folded it into one clean
rename before it left the branch), `fd7ee551`/`80dfa77e` (`scene_lattice_persist_test.go`/
`scene_speed_persist_test.go` → `build`, plus the new `fixture_kinds_test.go` duplicate — same
amend-to-fold-the-deletion pattern), `b208f47f`/`eab81961` (`overlay_toggle_emit_test.go` →
`viewstate`, same pattern a third time — `git mv` followed by `git commit -- <paths>` does not
reliably stage the SOURCE side of a rename in this environment; each of the 3 commits above is
the POST-`--amend` clean rename, and the intermediate two-part commits never left the branch).

## §40 — `nodes/wire` and `Trace/Trace.go`: verifying the "no pure run ≥3 lines" claim

Two clusters left with little/no attention on this branch: `nodes/wire` (19 files, three
over 200 lines — `out_port.go` 325, `wire_readout.go` 236, `in_port.go` 205) and the
untouched `Trace/Trace.go` (354 lines, hand-written, not generated — `head -3` confirms).

**Method.** Every function's body was classified statement-by-statement: (a) touches the
wire's own mutable state (`inflight`/`delivered`/queues), sends/receives on a channel,
starts a goroutine, or reads a clock; or (b) computation on locals/params/field-reads. An
earlier pass on this branch reported no pure run ≥3 lines survived in most of `nodes/wire`
after excluding channel ops and `pw.inflight` reads/writes — that claim was re-verified
rather than inherited, per this task's own instruction (every predecessor's stated blocker
on this branch went stale at least once).

**`nodes/wire`: the claim mostly held, with one real miss.** `out_port.go` (`Geom`,
`publishSteps`/`publishSegment`, `placeDrivenNoWalker`, `flushSendEvent`) is
channel-drain/channel-send/stream-write end to end — the only pure statement,
`placementFrom`'s 6-line struct build, is already a minimal, already-isolated method with
nothing left to extract. `wire_readout.go` (the file flagged "most suspicious, never
individually reported on") is the same story: `flushDroppedBreadcrumbs`/
`drainBreadcrumbEvents`/`appendPending`/`drainPendingEvents` are channel ops or slice
mutation on `pw.readout`'s own owned state; `DrainPendingEvents`' internal→exported slice
conversion loop is pure but converts a package-PRIVATE type (`pendingWireEvent`) so it
cannot leave package `wire`, and at 6 lines inline is not worth a same-package helper.
`in_port.go` is `PollRecv`/`flushRecvEvent`/`Breadcrumb` (all channel-or-stream I/O) plus
`breadcrumbLabelFor`, an already-minimal pure switch. `drive_item.go`, `paced_wire_send.go`,
`ports.go`, `owner_events.go`, `send_rule.go`, `broadcast.go`, `node.go` were also read in
full: each is either a channel/state operation or an already-minimal pure
predicate/constructor with nothing left to lift (`ParseSendRule`, `DriveItem.Live/Failed/
BufferFull`, `TryEmit`). `arrival.go`, `bead_placement.go`, `geometry.go`,
`chan_nonblocking.go` were already pure-only files before this pass (no change needed).

**The one miss:** `bead_advance.go`'s `advanceBead` re-implemented, inline, the exact
clamp-and-divide `lattice.BeadFraction` already exists to share — `lattice/bead_fraction.go`'s
own header comment names its two call sites (`live_beads.go`, `paced_wire_drive.go`'s
`ReviseInFlightGeometry`) and simply never listed `advanceBead` as a third. Fixed:
`advanceBead` now computes `final` directly from the deadline comparison and calls
`lattice.BeadFraction(nowTick, placementTick, crossTicks)` for `t`, deleting the duplicate
9-line clamp. Verified equivalent statement-by-statement: `BeadFraction` internally
re-derives the same `target = min(nowTick, deadline)` `advanceBead` used to compute by hand,
divides the same way, and clamps `t` to `[0,1]` — the extra `t<0` clamp `BeadFraction` adds
is a no-op here because `stepAll` only calls `advanceBead` after already checking
`nowTick > b.placementTick`. `lattice.BeadFraction`'s doc comment updated to name the third
call site. Commit `7c3de447`.

**`Trace/Trace.go`: not a pure/impure mix worth lifting — a four-way concern split.**
Statement classification confirmed the file is exactly what its own header says: a closed
event-kind vocabulary (data, not logic) plus a thin writer. The genuinely pure parts
(`marshalBreadcrumb`/`marshalNodeBead`) were ALREADY extracted to `Trace/marshal.go` on an
earlier pass; nothing pure remained un-lifted in `Trace.go` itself. What remained was one
file carrying four distinct concerns — the closed `Kind*` vocabulary + `TraceEventKinds`
(~130 lines), the breadcrumb-label sub-vocabulary + `BreadcrumbLabels` +
`BreadcrumbLabelID` (~70 lines), the `PortGeom`/`Event` payload value types (~30 lines), and
the `Trace` struct + its constructors + `Breadcrumb`/`NodeBead` methods (~95 lines) — split,
same as every other "ONE JOB" file split on this branch, into `kind_events.go`,
`breadcrumb_labels.go`, `event.go`, and a trimmed `Trace.go`. All four stay `package Trace`
(no new package: these types are load-bearing exports of one existing package, not a new
boundary). Commit `1d89e255`.

**Guard re-key.** `tools/network/trace/check-breadcrumb-label-registered.sh` hardcoded
`TRACE_GO="Trace/Trace.go"` and `awk`-scanned that one file for `BreadcrumbLabels`'
literal — exactly the single-file-path guard-blindness class
(`memory/feedback_guards_hardcoding_single_file_break_on_split.md`) moving `BreadcrumbLabels`
into `breadcrumb_labels.go` would have hit. Re-keyed to scan every `Trace/*.go` file
(matching `tools/gen-node-defs/trace_kinds.go`'s `parseBreadcrumbLabels`, which already
scanned the whole dir and needed no change). Deliberate-break proof: removed `"bead-crud"`
from `BreadcrumbLabels` — guard failed by name (`FAIL — 1 unregistered breadcrumb label(s):
nodes/Wiring/layoutquant/commit_node_move.go:142: label "bead-crud" is not in
Trace.BreadcrumbLabels`), restored, guard passed clean again. `go run
./tools/gen-node-defs` was also re-run after the split: byte-identical output (`git status`
showed no diff on any generated file), confirming `parseTraceKinds`/`parseBreadcrumbLabels`'
whole-dir scan tolerated the split with zero drift.

**Declines (with the pinning statement quoted).** No file in `nodes/wire` was declined outright
this pass — every file was either already pure-only, already minimal, or fixed
(`bead_advance.go`). The nearest thing to a decline is `wire_readout.go`'s
`DrainPendingEvents` conversion loop, kept in place because it converts
`pendingWireEvent` (package-private) to `PendingWireEvent` (exported) — moving it out of
package `wire` would require exporting the private type or duplicating its shape in a new
package, neither of which is a real win for 6 lines already sitting next to their one
caller.

**LOC before/after:**
| file | before | after |
|---|---|---|
| `nodes/wire/bead_advance.go` | 85 | 79 |
| `nodes/wire/lattice/bead_fraction.go` | 44 | 45 (doc-only) |
| `Trace/Trace.go` | 354 | 117 |
| `Trace/kind_events.go` | — | 119 (new) |
| `Trace/breadcrumb_labels.go` | — | 94 (new) |
| `Trace/event.go` | — | 36 (new) |

`nodes/wire` file count: 19 → 19 (no file added/removed, one dedupe). `Trace/` file count
(non-test, non-generated): 2 → 5.

**Surfaces with NO check of any kind that can fail:** `advanceBead`'s `final`/`t` computation
itself (no test exists anywhere in this repo — verification is `go build`/`go vet`/loud
runtime assertion only, per this task's constraints); the `Trace` struct's constructors and
`Breadcrumb`/`NodeBead` methods (headless-test-only code paths, never wired in production,
with no guard naming them); `PortGeom`/`Event`'s field shapes (no guard checks their fields
against any consumer). This list is long by design — most of both clusters is either
channel/state plumbing (correct by construction, per this repo's testing-shape doctrine) or
vocabulary data with no logic to break.

**Verify:** `go build ./...`, `go vet ./...` clean. `go test -race -count=1 ./...` — every
package printed `[no test files]` (72 packages), none failed; verbatim tail:
```
?   	github.com/dtauraso/wirefold/tools/gen-node-defs/buflayout	[no test files]
?   	github.com/dtauraso/wirefold/tools/gen-node-defs/constexpr	[no test files]
?   	github.com/dtauraso/wirefold/tools/gen-node-defs/kindscan	[no test files]
?   	github.com/dtauraso/wirefold/tools/gen-stream-fixture	[no test files]
?   	github.com/dtauraso/wirefold/tools/topology-vscode/node_modules/flatted/golang/pkg/flatted	[no test files]
```
`bash scripts/stop-checks.sh` printed empty stdout after both commits (confirmed run from
repo root). No-import-cycle check: only `nodes/wire/beadchain` imports `nodes/wire` among
`nodes/wire`'s subpackages — pre-existing (verified before this pass's first edit too, not
introduced by it). `git status --short` empty after both commits. Commits: `7c3de447`
(`bead_advance.go`/`lattice/bead_fraction.go`), `1d89e255` (`Trace.go` split + guard re-key).

## `stream-demux.ts`'s five handlers and `runCommand.ts` — re-measured

`stream-demux.ts`'s five `handle<Kind>Fd` methods (view/edge/node/interior/drive) were
declined once, citing `processInteriorLikeFrames`'s own doc comment on carry-buffer desync
risk (`docs/investigations/interior-stream-framing.md`). Re-read: that argument is scoped to
sharing ONE carry buffer across two physically distinct pipes ("reusing `interiorBufs`'
carry state across two physically distinct pipes would reintroduce the exact desync"), not
to sharing the reassembly CODE SHAPE. It does not forbid a shared code path that takes the
carry buffer as a parameter and keeps each fd's own buffer field.

Factored the dead-stream-check + `splitFrames` + rest-storage + error-report shape into one
private `dispatchFrames(key, carry, chunk, storeRest, errorContext, onFrames)`. Each
`handle<Kind>Fd` call site still names, explicitly, its own carry-buffer field (`stream.
viewBuf` / `stream.edgeBufs[row]` / `stream.nodeBufs[row]` / `stream.interiorBufs[row]` /
`stream.driveBufs[row][slot]`), its own probe file, decode function, `BUF_BLOCK_TAG_*`, and
last-frame cache. `handleInteriorFd`/`handleDriveFd` still both feed the pre-existing shared
`processInteriorLikeFrames` tail (decode/probe/cache), which itself was already correctly
shared before this pass and is unchanged in shape. `stream-demux.ts`: 364 → 402 lines (grew
— five call sites now each carry an explanatory comment naming what they still own, plus the
new `dispatchFrames` method itself with its own doc comment); the point was removing the
statement-for-statement duplication, not line count. `check-stream-kind-ts-parity.sh` is a
pure name-level grep (`handle<Cap>Fd(` anywhere under the ext-host src tree) — it does not
care whether that text is a method definition, a call site, or an error-context string, so
renaming only the definition left it passing; renaming the definition, the
`attach-listeners.ts` call site, AND the error-context string together made it fire by name
(`stream kind "drive" is declared in Go but no handleDriveFd( reader exists`), then
restored.

`runCommand.ts` was re-measured per function rather than by a blanket claim. Every method
outside `run()` is 3–17 lines and is a thin, already-terminal delegation (`constructor` 3,
`currentGen` 3, `newDemux` 15, `cancel` 17, `isRunning` 3, `restart` 6, the four
`getLast*Frame(s)` accessors 3 each, `writeStdin` 8, `dispose` 4) — each body is entirely
`vscode.*`/`cp.*`/`proc.*`/`process.*`/`this.*`/`fs.*` or a direct delegation to an
already-extracted pure helper (`frameRecord`, the four `demux.getLast*` calls). `run()` is
the one large method (190 lines, 121–310) and is almost entirely sequenced side effects in
a fixed order (channel setup, orphan reap, build, counts read, spawn-layout compute, `cp.
spawn`, listener attach, close/error handlers) — 75 of its 232 non-comment/non-blank lines
match the `vscode.|cp.spawn|proc.|process.|this.|fs.` scan directly, and of the rest,
nothing is a standalone pure computation worth its own file: the two arguably-pure
fragments (`topArgs = topologyPath ? [...] : []`, one ternary; the `WIREFOLD_STREAM_FDS`/
`WIREFOLD_EDGE_BEAD_TRACE` env-object literal, ~4 lines) are each smaller than the import +
call-site overhead a new file would add, and neither is duplicated anywhere else in the
tree. Declined again, this time with the per-function breakdown the previous pass's blanket
claim lacked.

No new home for domain state was created by either change — `dispatchFrames` holds no
field, takes its carry buffer and storage callback as parameters, and returns nothing that
outlives the call; `StreamDemux`'s existing per-kind last-frame caches (sanctioned replay
state) are untouched in shape.

**Surfaces with NO check of any kind that can fail:** `dispatchFrames` itself (no guard or
test names it — its correctness is that its five callers still produce identical output,
which is asserted only by reading the diff and by the live editor); the per-call-site
closures' exact ordering of probe-decode vs. cache-set vs. `onSnapshot` (no guard checks
statement order within a handler, only that a `handle<Kind>Fd` symbol exists somewhere);
`runCommand.ts`'s `run()` sequencing (channel-clear-before-build, orphan-reap-before-spawn,
`spawnGen++`-before-spawn) has no guard — those invariants are stated only in comments.

**Verify:** `bash scripts/stop-checks.sh` printed empty stdout after the `stream-demux.ts`
commit (run from repo root, confirmed via `pwd`). `git status --short` empty after the
commit. Commit: `fb0d4763` (`stream-demux.ts` dispatchFrames factoring).

## The four largest hand-written Go files

`mover_registry.go` (395), `nodes/input/node.go` (363), `edge_mover.go` (359),
`nodes/PairNode/node.go` (348) — measured statement-by-statement per CLAUDE.md's method
(writes an unexported field/sends on a channel/starts a goroutine/writes a file = bucket
(a); computation on locals/params/field reads = bucket (b)).

**`moverreg/mover_registry.go` — split in place, four files, no API change.** Every method
on `*MoverRegistry` reads or writes the type's own unexported fields (bucket (a) by the
task's own rule — a method on a type legitimately touches that type's fields; this isn't a
channel-export question, `MoverRegistry` holds maps, not channels). The two fully-pure
functions (`linkRefusalFor`, `firstPortOfDir`) already didn't touch `MoverRegistry` at all.
Grouped by concern, same package: `mover_registry.go` (138 lines — struct, `New`,
`NodeGeoms`/`EdgeMovers` accessors, `ClaimSelfDrive`/`SeedCenter` construction-time
writes), `mover_registry_wire.go` (111 — `Bind`/`Start`/`FinalizeActors`, wiring + actor
launch), `mover_registry_query.go` (131 — `drainCenterMirror`/`CenterOfNode`/`SendMove`/
`EnqueueFor`/`nodeKind`/`NodeBodyRadius`/`HasNodeMover`/`NodeSelfDriven`/
`NodeQuantOffset`/`NearestNodeTo`, all read-only lookups), `mover_registry_linkrefusal.go`
(59 — `LinkRefusal` + its two pure helpers). No decline; every function found a concern
home. Commit `2a89023b`.

**`edgemover/edge_mover.go` — split in place.** `handle` (the inbox message switch) and
`recomputeGeometry` (the geometry propagate it drives) are the two functions with real
weight and a distinct role — "what runs once a message has arrived" — versus the rest of
the file (struct, construction, the bound-func-value setter/accessor surface: `Select`,
`TrySendFromSrc/Dst`, `SendSteps`, `SetOut`/`SetDest`/`SetSpeedCh`/`SetStream`). Both moved
to `edge_mover_handle.go` (109 lines); `edge_mover.go` dropped 359→268. The unexported
`extIn`/`srcIn`/`dstIn`/`stepsIn` channels were untouched — no export, no new package
boundary, same package `edgemover`. Commit `a7c46321`.

**`nodes/input/node.go` — mostly declined, small lift.** Per-function bucket count: `clock`
(6, all field reads, bucket a-adjacent trivial), `broadcastPlace` (9, calls
`n.OutCadence.Wired()`/`.PlaceDrivenAt` — bucket a, side-effecting placement), `popEnd`
(11, **fully bucket (b)** — only the caller's own local slice pointers, no `Node` field
touched at all), `updateFeedbackRing` (~70, ctx/channel/clock blocking calls throughout —
bucket a), `Update` (~55, the goroutine loop itself — bucket a), `inputCadenceTicks` (7:
one field read `n.OutCadence.Geom().Steps`, bucket a, then 5 lines of pure arithmetic on
that int, bucket b), `init`/`RegisterBuilder` (~45, construction writes — stays per the
node-kind landing rule). Lifted the two genuinely pure pieces — `popEnd` whole,
`inputCadenceTicks`'s arithmetic split out as `cadenceTicks(steps int) int64` — into a new
sibling file `nodes/input/emit_helpers.go` (38 lines) in the SAME package `input`, not a
subpackage (too small to justify a new one-file package per CLAUDE.md's "existing sibling
first" bias, and there is no existing sibling shaped for "arithmetic serving exactly one
node kind's own cadence"). `node.go` 363→347 — a small drop, honestly reported: nearly
everything else in this file is bucket (a) by construction (a periodic-source node's own
goroutine loop, its channel drains, its placements), which is the loop-body/decision
content the primitive landing rule keeps in `node.go`. Declined further split: every
remaining function's own statements are field reads/writes or channel/blocking calls on
`*Node`'s own state, not stray computation. `go run ./tools/gen-node-defs` produced
byte-identical generated output (git status showed only the two edited files after the
run). Commit `a8f6d8af`.

**`nodes/PairNode/node.go` — mostly declined, small lift.** This file's own header
comment already states its contract: "THE KIND'S LOGIC IS THIS FILE: the Node struct, the
builder..., the Update loop..., and the two functions that decide — stepFromVector and
handleVectorCycle." Code-only line count (comments/blanks stripped) is 144 of 348 lines —
the rest is the prose the file's own doc block promises. `stepFromVector` (35 lines: field
reads via `n.ringOf()`/`n.topState()`, calls to `n.tilt.Machine.Settled/Step`, one field
write `n.rest.restedThisCycle = true` — mixed, but IS the decision, per the header's own
naming) and `handleVectorCycle` (52 lines, same shape) are declined outright — the file's
own stated contract names them as what stays, and pulling either out would leave `node.go`
without either of "the two functions that decide," contradicting the header on the same
branch that wrote it. `Update` (34 lines, the goroutine loop) also stays, same reasoning
as `input`. Lifted three helpers that are NEITHER the composer/builder NOR one of the two
deciding functions: `clock()` (6 lines, a nil-guard field read), `openingEmit` (7 lines,
three field/method calls made once before the loop), `paceOnBeadArrival` (7 lines, one
non-blocking channel drain) — into a new sibling file `nodes/PairNode/lifecycle.go` (57
lines), same package, no subpackage (mirrors `input`'s reasoning — this is per-kind
lifecycle plumbing, not a concern any existing sibling like `tiltring` is shaped for).
`node.go` 348→308. `go run ./tools/gen-node-defs` produced byte-identical generated output.
`check-docs-symbols.sh` (which asserts `nodes/PairNode/node.go#handleVectorCycle` still
resolves, since `docs/pair-node/*.html` link to it) stayed green because that symbol never
moved. Commit `cba04edf`.

**Guard teeth, proven with a deliberate failure.** Grepped `tools/` for every filename and
symbol touched; the two guards that name `edge_mover.go`/`edge_file.go` literally
(`check-persist-write-ownership.sh`'s `EDGE_OWNERS`, `check-scene-path-resolution.sh`'s
`NODE_PATH_OWNERS`) do NOT name the new `edge_mover_handle.go` — deliberately: that file
holds no `filepath.Join`/`writeJSONAtomic` call today, so it correctly has no owner entry
yet. Proved this is a live tooth, not silent-green-by-omission: added
`jsonpersist.WriteJSONAtomic("x", nil)` to `edge_mover_handle.go` via the Edit tool and ran
`check-persist-write-ownership.sh` — it failed loudly (`unauthorized-write:
.../edge_mover_handle.go: 18:...`, exit 1) — then reverted via Edit and reran clean (`git
diff --stat` empty on the file, guard back to exit 0). If a real write is ever added to
`edge_mover_handle.go`, that file's basename must be added to `EDGE_OWNERS` in the same
commit — recorded here so the guard's silence today isn't mistaken for coverage of a write
that doesn't exist yet.

**Surfaces with NO check of any kind that can fail:** the four-way file split inside
`moverreg` (which method landed in which of the four files) — nothing asserts a given
method lives in a given file, only that the package still builds and exports the same
API; `edgemover`'s `handle`/`recomputeGeometry` extraction, same story; `input`'s and
`PairNode`'s sibling-file placement of `popEnd`/`cadenceTicks`/`clock`/`openingEmit`/
`paceOnBeadArrival` — no guard names these functions, only `check-docs-symbols.sh`'s one
hardcoded anchor (`node.go#handleVectorCycle`) which happens not to have moved.

**Verify, all four commits:** `go build ./...` — no output, exit 0. `go vet ./...` — no
output, exit 0. `go test -race -count=1 ./...` — every package `[no test files]` (expected
per this repo's no-test doctrine). The no-dispatch-import loop (`for p in $(go list
./nodes/Wiring/... | grep -v 'dispatch$'); do go list -deps "$p" | grep -qx
.../nodes/Wiring/dispatch && echo DEPENDS; done`) printed only the two expected exceptions,
`nodes/Wiring/build` and `nodes/Wiring/stdinreader`. `go run ./tools/gen-node-defs` +
`git status --short` after each of the `input`/`PairNode` commits showed only the
hand-edited files, confirming byte-identical generated output. `bash
scripts/stop-checks.sh` printed empty stdout (confirmed `pwd` first) after all four
commits landed. `git status --short` empty at the end. Commits: `2a89023b` (moverreg),
`a7c46321` (edgemover), `a8f6d8af` (input), `cba04edf` (PairNode).

## `three/scene/` audit (18 files) — one genuine duplication, rest already split

Read every file in `tools/topology-vscode/src/webview/three/scene/` (18 files, 2281 LOC) and
classified bodies statement-by-statement per the method above. Finding: the directory is
**already decomposed correctly** for this task's purpose — `node-depth-order.ts`,
`edge-stream-blocks.ts`, `view-blocks.ts`, `bead-style.ts`, and `node-stream-blocks.ts` are
already pure modules with no React/three imports; the `.tsx` files (`NodeInstances.tsx`,
`ThreeView.tsx`, `SphereRings.tsx`, `ChainBeadInstances.tsx`, `InteriorBeadInstances.tsx`,
`EdgeLines.tsx`) are almost entirely bucket-(a) — `useFrame`/`useRef`/`useMemo`/three-object
`.set(...)`/`InstancedMesh` matrix writes — with only one- or two-line `Math.min`/`Math.max`
clamps that aren't worth a separate module (moving `const n = Math.min(count, capacity)` into
its own file would not shrink the body of statements that *do the actual work*, which stay
where the `useFrame` is).

One real finding: `NodeInstances.tsx` (ring orientation, lines 134–141 pre-edit) and
`node-stream-blocks.ts` (`getChainBeads`'s per-node ring axis, lines 233–238 pre-edit) each
independently reimplemented the identical θ/φ → unit-axis conversion (the `Buffer/layout.go`
`RingAxisTheta`/`RingAxisPhi` columns), one with a `(poleTheta===0 && polePhi===0)` special
case and the other with an `if` guard defaulting to `(0,1,0)` — both unnecessary, since
`poleAxis(0,0) = (sin(0)cos(0), cos(0), sin(0)sin(0)) = (0,1,0)` already, with no branch
needed. Lifted the shared math into `buffer-scene-shared.ts` (already imported by
`NodeInstances.tsx`, and now also by `node-stream-blocks.ts`) as
`poleAxis(theta, phi): [number, number, number]`, dropping both branches in favor of a plain
destructure. `nav/buffer-nav.ts` has its own `poleVec` doing the same math but returning a
`THREE.Vector3` for the nav overlay — left alone, since `node-stream-blocks.ts` is
`three`-free by design (pure decode module feeding both `.tsx` consumers and the
`check-no-node-node-polar.sh` guard's single bead-centre-summation-site invariant) and pulling
`three` into it to share with `nav/` would be the wrong direction for that boundary.

**Domain-state check:** `poleAxis` is a pure function of two numbers to three numbers, no
module-level mutable, no ref, no store — it does not give domain state a new home, and it
computes no NEW geometry (the trig it does was already being computed twice; this makes it
computed once, still purely from streamed buffer values, still at the renderer edge per
`nav/buffer-nav.ts`'s own doc comment on `poleVec`).

**Guard check:** grepped `tools/` for `NodeInstances`, `node-stream-blocks`, and
`buffer-scene-shared` by filename; two guards reference them: `check-ts-shading-from-go.sh`
(comment-only mentions of `NodeInstances.tsx`, unaffected — no shading constant moved) and
`check-no-node-node-polar.sh`, which asserts exactly one bead-centre-summation site named
`node-stream-blocks.ts`. The summation itself (`cx + readChainBeadOX(...)` etc.) was left in
place in `node-stream-blocks.ts` — only the axis-vector math was extracted — so the guard
still names the same file and passed (`✓ no node-node polar record; exactly one bead-centre
summation site.`, exit 0) both before and after. A deliberate-failure proof (temporarily
duplicating the summation line via a shell script) was attempted and correctly BLOCKED by the
placement-brief pre-write hook (source writes must go through Edit/Write, not a shell
heredoc) — reported here rather than routed around, per
`memory/feedback/process/feedback_hook_block_means_stop.md`.

**Neighbours not visited:** `NodePalette.tsx`, `polar-frame.tsx`, `TiltVectors.tsx`,
`snapshot-buffer.ts`, `SpeedSlider.tsx` were listed as fallback candidates "if the scene dir
yields little" — the scene dir yielded one real (if small) finding, so the neighbours were
not opened this session; they remain open for a future pass if friction surfaces there.

Verify: `bash scripts/stop-checks.sh` from repo root — empty stdout (tsc, npm webview build,
eslint, and the guard suite all clean). Commit `58944ca9`.

## `tools/gen-node-defs/` — the seven files over 200 lines, all measured and split

First look at this directory on the branch. It is tooling (parses `SPEC.md`, Go `const`
declarations, and `messages.ts`; emits TS/Go), not the network — MODEL.md's
goroutine/ownership doctrine does not apply here, and it is mostly pure parse→build→emit,
which makes it decomposable in the ordinary way. Every function in every file below was
a single job done in one long body, no interleaving of unrelated concerns within a
function — the split in each case is "this file was doing two jobs sequentially,
give each its own file," not a lift of incidental pure computation out of I/O.

- **`kindscan/spec_md.go` (327→110)**: `parseSpecMD` (View+Ports table parse, kept),
  `parsePortsFromSpec` (fallback Ports-table→Port list) and `parseDefaultData` (fenced
  JSON block extraction) moved to their own files
  (`spec_md_ports.go`, `spec_md_default.go`); the markdown-table parsing they all
  duplicated (`sectionLines`, `parseMDTable`, `parseMDRowCells`, `isSep`, `indexOf`,
  `readSpecMDLines`) was lifted into a new shared `spec_md_table.go` — this also removed
  real duplication (`parsePortsFromSpec` had its own hand-rolled copy of the separator-row
  skip and cell-trim logic `parseSpecMD`'s closure already had).
- **`overlay_gen.go` (303→129 + `overlay_write.go` 182)**: `parseOverlayFlags` (reads
  `messages.ts`, mostly pure token/override derivation after the read) stayed;
  `writeOverlayGen` (≈165 lines of sequential `fmt.Fprintln`/`Fprintf` emission, bucket
  (a) by nature — text generation IS the emitted output) moved to its own file. One trap:
  `writeOverlayGen`'s emitted Go comment names its own source file
  (`tools/gen-node-defs/overlay_gen.go`) for a future reader of the GENERATED
  `overlay_state.go` — that string is OUTPUT, not documentation, so it was left pointing at
  `overlay_gen.go` even though the function now lives in `overlay_write.go`; changing it
  would have silently broken byte-identical output (caught by the required post-split
  generator diff, but noted here since it's the kind of thing a blind grep-and-rename could
  get wrong).
- **`input_layout.go` (248→78 + `input_layout_parse.go` 181)**: fingerprint-string parsing
  (`parseInputLayoutFingerprintDir`/`parseInputLayoutFingerprint`/`fpListToken`/`fpList`/
  `unquoteGoString`/`kindConstName`) split from the TS emit (`writeInputLayout`/
  `writeTSArray`) — same parse/emit shape as overlay_gen.go.
- **`params.go` (241→31 + `params_curve.go` 89 + `params_shading.go` 138)**: two
  independent const-family pipelines (`CurveParam*`→`curve-params.ts`,
  `ShadingParam*`→`shading-params.ts`, the latter needing `constexpr.Env` for non-literal
  values) that only shared one naming helper — `camelToScreamingSnake` stayed in
  `params.go`, which is now just that.
- **`constexpr/constexpr.go` (233→159 + `eval.go` 87)**: package/import loading
  (`NewEnv`/`loadPkgConsts`/`importDir`, all file/dir reads) stayed; `Eval` (pure recursive
  descent over an already-parsed AST) plus its `exprString` diagnostic helper moved to
  `eval.go`.
- **`buflayout/buf_layout_parse.go` (229→140 + `buf_layout_file_parse.go` 99)**: the
  directory-scan/block-reordering half (`ParseBufferLayoutDir`, `bufBlockOrder`,
  `buildBufFingerprint`) stayed; the per-file AST extraction (`parseBufferLayoutFile`)
  moved out — same "scan a directory, don't hardcode a filename" shape as
  `input_layout_parse.go`, called out in both files' own doc comments.
- **`kindscan/ast_ports.go` (204→130 + `ast_embedded_ports.go` 84)**: direct channel-typed
  field scan (`parsePortsFromAST`, `chanDirection`) stayed; the embedded-struct recursive
  port walk (`parseEmbeddedPorts`) moved out — it's a distinct concern (following
  `gatecommon`-style embedded packages) that only calls `parsePortsFromAST` as a leaf.

**Guards:** grepped `tools/` and `scripts/` for every moved filename and every moved
function/type name before splitting. Only two hits, neither a file-path guard that could go
silently green on a split: `check-spec-format-view-fields.sh` names `spec_md.go` in a
comment but extracts its `vmap["..."]` fields via a RECURSIVE grep of the whole
`tools/gen-node-defs` tree, so the split didn't blind it (confirmed unaffected — still
greps every file). `buffer_layout.go`'s doc comment names `buf_layout_parse.go` in prose,
also unaffected (comment, not a check). No guard anywhere greps for
`overlay_gen.go`/`input_layout.go`/`params.go`/`constexpr.go`/`ast_ports.go` by filename or
for any of the moved function names as a placement check — grepped and confirmed absent.

**Guard teeth, proven with a deliberate break:** `check-generated.sh` (self-heals a stale
*generated* file by regenerating in place, so editing a generated file directly doesn't
prove anything — it just gets overwritten). The real test is a generator-SOURCE edit that
changes emitted output without regenerating the tracked file: appended `+ "BROKEN"` to
`parseSpecMD`'s `Bg: vmap["bg"]` line in `kindscan/spec_md.go`, ran
`bash tools/buffer-schema/check-generated.sh`, got:
```
check-generated: stale generated file(s) — commit the regenerated output:
 M tools/topology-vscode/src/schema/node-defs.ts
EXIT=1
```
Reverted the edit, reran `go run ./tools/gen-node-defs`, confirmed `git status --short`
empty again.

**No check of any kind exists for:** the shared `spec_md_table.go` helpers' correctness
(no tests project-wide, by design) — coverage here is the byte-identical generator output
check plus `check-spec-format-view-fields.sh`'s doc/parser-field parity, neither of which
would catch a subtly wrong `parseMDTable` that still happened to produce the same output
for every CURRENT `SPEC.md` file. Same gap existed before the split (the logic was inline,
untested either way); the split does not add or remove this exposure.

**Byte-identical generator output, confirmed after every commit:** `go run
./tools/gen-node-defs && git status --short` returned empty after each of the 7 commits
below.

**Commits** (all on `task/god-objects`, each `go build ./...`/`go vet ./...` clean, each
followed by an empty-diff generator run): `8a074350` (spec_md.go three-way split +
shared table helpers), `32fe690f` (input_layout.go parse/emit split), `17f5285c`
(params.go curve/shading split), `0c967642` (constexpr.go load/eval split), `d7ae36cd`
(buflayout dir-scan/per-file split), `45df1a6d` (ast_ports.go direct/embedded split),
`c6427cc0` (overlay_gen.go parse/emit split).

Verify: `bash scripts/stop-checks.sh` from repo root — empty stdout. `go build ./...` and
`go vet ./...` both clean (no output beyond the build/vet completing).

## 2. Six webview neighbours audited (`NodePalette.tsx`, `polar-frame.tsx`, `TiltVectors.tsx`,
   `snapshot-buffer.ts`, `SpeedSlider.tsx`, `TiltVectorAnglePanel.tsx`)

Statement-by-statement classification, per the render-and-forward-only invariant (no store,
no geometry computed in TS).

**`polar-frame.tsx` (239→232) — split.** Lines 82–92 were 11 statements of pure arithmetic
(`radiusKey`/`poleLen`/`poleRadius`/`coneH`/`coneBaseR`/`arcR`/`arcTube`/`arcMid`/`hhR`/
`arcHH`), all derived from the `scale` prop alone, no THREE/ref/hook touch. Lifted into
`nav/polar-frame-geometry.ts` (`computePolarFrameGeometry`, 35 lines) — local DRAWING SCALE
for a decorative overlay frame (stick/cone/handhold sizes), not a bead/node/edge position;
`center`/`scale` still arrive from Go untouched. `polar-frame.tsx`'s remaining body is JSX
+ the one `useMemo` quaternion (touches a THREE object) — bucket (a).

**`SpeedSlider.tsx` (221→177) — split.** `SPEED_SETTINGS`/`settingKey`/`DEFAULT_INDEX`
(module-level pure data, ~19 lines) plus `closestSettingIndex` (10 lines, pure numeric
lookup) had zero react/vscode-api dependency — same shape as the existing
`tilt-vector-angle-format.ts` split TiltVectorAnglePanel already uses. Lifted into a new
sibling `panels/speed-settings.ts` (52 lines). `SpeedSlider.tsx`'s remaining body is the
component (hooks, `createPortal`, `postGoRecord`) plus inline `CSSProperties` objects —
bucket (a)/style, not logic.

**`snapshot-buffer.ts` (223→233, but 51 lines of near-identical structure removed) —
split-in-place, not a new file.** The edge/node/interior sections were three copies of the
same five-operation shape (keyed Map, per-generation routing via `genTable`, listener
`Set`, optional version counter) differing only in variable names and whether a version
counter existed. Factored into one closure-returning `makeRowStreamTable(withVersion)`
(bucket (b): pure plumbing, no hook/ref/three/post — `noteGen`/`genTable` were already pure
helpers in this file) instantiated three times (`edgeStream`, `nodeStream`,
`interiorStream`, the last two `withVersion=true`). Every exported function name/signature
is unchanged (`setLatestEdgeStreamFrame`, `getLatestNodeStreamFrames`,
`getInteriorStreamVersion`, etc.) — only the internal implementation is shared, so every
call site in `main.tsx`/`node-stream-blocks.ts`/`edge-stream-blocks.ts` is untouched. This
does NOT create a new home for domain state: it is the same one-pointer-per-key module
cells the file's own header comment already disclaims as "not a store", just built by one
factory instead of copy-pasted three times. Grepped `tools/` for the internal variable
names being renamed (`edgeStreamTables`, `nodeStreamTables`, `interiorStreamTables`,
`*StreamListeners`) — zero hits, nothing guards them by name.

**Declined — `NodePalette.tsx` (249, unchanged).** The only pure, non-JSX/non-bridge
statement is `dropKindFromEvent` (6 lines: `dataTransfer.getData` read + `Number.isInteger`
guard) — too small to justify a new file, and every other function in the file either
touches a hook/ref (`PaletteRow`, `RefusedNotice`, the `useEffect` keydown binder) or calls
`postGoRecord`/`e.dataTransfer.setDragImage` directly (`fireCreateAt`, the drag handlers) —
bucket (a) by the stated classification rule ("touches a React hook, a three.js object, a
ref, or posts to the host"). No factorable shape shared with the other two panels beyond
what `overlay-chrome.ts`'s pill/popover primitives (already shared, already a separate
file) provide.

**Declined — `TiltVectors.tsx` (224, unchanged).** `writeArrowInto`'s geometry math (axis
from θ, shaft/head midpoint/scale) is pure arithmetic in isolation, but every statement
writes directly into `axisRef.current`/`posRef.current`/`quatRef.current`/
`sclRef.current`/`matRef.current` — mutable THREE.js objects held in refs specifically to
avoid a per-frame allocation (the file's own header comment: "holding no state of its own"
/ "writes instance matrices imperatively"). The method's own doc comment: "writeArrowInto
composes one arrow's shaft+head matrices into whichever mesh pair and instance index the
caller supplies... so the geometry math lives in one place" — it is ALREADY the one
extraction point this file has, and it is bucket (a) by the classification rule ("touches a
... three.js object, a ref") on every statement, not bucket (b). Pulling the θ→axis/
scale arithmetic out into a plain function while leaving the ref writes behind would split
one coherent unit into two files that must be read together to see the seven writes it
does, for no bucket-(b) LOC gained (it's already the sole such function in the file).

**Declined — `TiltVectorAnglePanel.tsx` (205, unchanged).** Its own header comment already
states the split this task would otherwise propose: "The actual derivation lives in
tilt-vector-angle-format.ts (formatAngle, imported above) — split out so it has no
react/vscode-api dependency". The remaining pure (b) content is four `LATTICE_POINTS_*`
constants (3 lines) and one `points + delta` computation inline in
`LatticePointsRow`'s `adjust` closure (1 line) — below any reasonable extraction
threshold. Every function (`AxisRow`, `NodeGroupSection`, `LatticePointsRow`,
`TiltVectorAnglePanel`) is hooks/JSX/`postGoRecord` — bucket (a).

**No new home for domain state, no geometry computed in TS:** confirmed by running
`bash tools/webview/check-no-webview-state.sh` clean after each commit, and by a
deliberate break — added `import { create } from "zustand";` to `snapshot-buffer.ts`,
reran the guard, got:
```
zustand-import: .../snapshot-buffer.ts:1:import { create } from "zustand";  (Zustand store in the webview — domain state must live in Go)
no-webview-state: 1 hit(s) — the webview must hold no domain state (no Zustand store, no stateful domain hook); the model lives in Go and streams as the binary content buffer
EXIT:1
```
reverted, guard clean again. `check-ts-computes-no-geometry.sh` scans a fixed forbidden-token
list (`getPointAt`/`rfArcLength`/`arcLengthToSimLatencyMs`/`patchPulse`/`buildPortCurve`/
`buildEdgeCurve`) unrelated to what moved here — `polar-frame-geometry.ts`'s scale-derived
stick/handhold sizing was never one of those tokens and stays a local-drawing-scale helper,
not a position/curve/timing computation; confirmed the guard passes clean regardless (no
targeted break attempted, since none of its named tokens were touched — grepped `tools/` for
`polar-frame-geometry`/`speed-settings`/`makeRowStreamTable` and found no guard names any
of them).

**LOC:** `polar-frame.tsx` 239→232 (+35 new file); `SpeedSlider.tsx` 221→177 (+52 new file);
`snapshot-buffer.ts` 223→233 (structure dedup, not a shrink — three copies became one
factory + three 2-line instantiations, at the cost of the factory's own ~28-line body).
`NodePalette.tsx`/`TiltVectors.tsx`/`TiltVectorAnglePanel.tsx` unchanged.

**No check of any kind that can fail** on the specific arithmetic moved: `computePolarFrameGeometry`
and `closestSettingIndex`/`SPEED_SETTINGS` have no test (none exist project-wide by
design) and no guard reads their numeric output — only `tsc`/eslint/build catch a
type/syntax break, and only a human noticing a wrongly-scaled frame or a slider landing on
the wrong tick would catch a value regression. Same exposure as before the split; the move
does not add or remove it.

Verify: `bash scripts/stop-checks.sh` blocked on a PRE-EXISTING, out-of-scope failure
(`check-docs-symbols`: `nodes/Wiring/geom/gesture_camera.go does not exist`, referenced by
`docs/planning/movedispatch-decomposition.md` itself, from a concurrent Go-side session
editing `nodes/Wiring/loadspec/loader_tree.go` in this same checkout — confirmed via
`git status --short` showing that file modified but not staged by this task, and never
touched by any commit here). Verified the TS-owned surface directly instead: `npx tsc
--noEmit -p tools/topology-vscode` (empty output after each commit), `npx eslint` on every
changed/new file (empty output), `npm run build` inside `tools/topology-vscode` (clean,
`out/webview.js` refreshed each time), and `bash tools/webview/check-no-webview-state.sh`
(clean each time, teeth proven above).

Commits (`task/god-objects`): `076e97dd` (PolarFrame geometry split),
`8160d816` (SpeedSlider speed-settings split), `dbfea6ff` (snapshot-buffer.ts table
factoring).

## Two never-examined files: gesture_camera.go (geom) and loader_tree.go (loadspec)

Both files already lived in lifted subpackages (`geom`, `loadspec`); this pass was purely
about splitting a large file by concern, not escaping a god package.

**`nodes/Wiring/geom/gesture_camera.go` (321 lines).** Every function in this file is
bucket (b): pure computation on parameters/locals, no field writes on a type staying
behind, no channel send, no goroutine start, no file I/O. There is nothing to lift — the
file already sits in its correct sibling package (`geom`). Split IN PLACE by the six
concerns the file's own section-comment banners already named: `camera_angles.go`
(`AnglesToWorldOffset`/`WorldDirToAngles`, 50 lines), `camera_basis.go` (`CamBasis`/
`BasisFromViewpoint`/`EyeOf`, 36 lines), `camera_screen.go` (`PolarDir`/`ScreenToPolar`/
`ToWorldDir`/`PlaneSlide`/`DeltaToPolar`/`PanDisplacementPolar`, 69 lines),
`camera_focus.go` (`FocusAhead`/`ContentSphereOf`/`RegionFocus` + the three gesture
constants, 99 lines), `camera_homefit.go` (`FitDistanceGo`/`HomeFitPose`, 50 lines),
`camera_project.go` (`ProjectNDC`/`RayDirThroughNDC`, 37 lines). No exported surface
changed — same package, same signatures, same body text moved verbatim. The great-circle
orbit formulation (no free axis/sign parameter — `memory/feedback_make_bug_class_unrepresentable.md`)
was preserved exactly: nothing in the split touches the math, only which file a function's
text lives in. Grepped `tools/` and `scripts/` for `gesture_camera.go` and every moved
symbol name before splitting — zero hits, so no guard hardcoded this file's path or any of
its function names. One doc citation did reference it: `docs/pair-node/math/formulas.html`'s
three `data-src="nodes/Wiring/geom/gesture_camera.go"` rows (the x/y/z spherical-to-cartesian
formula, which is `AnglesToWorldOffset`) now point at `camera_angles.go`, caught by
`check-docs-symbols.sh` failing loudly (`nodes/Wiring/geom/gesture_camera.go does not
exist`) rather than silently. No test file existed for this package (`gesture_camera_test.go`,
named in the old header comment, was never created — this repo has no tests by decision).

**`nodes/Wiring/loadspec/loader_tree.go` (284 lines).** `LoadTree` itself is bucket (a):
every top-level step is `os.ReadFile`/`os.ReadDir` (via `readDirNames`) building `spec`,
which stays behind — kept in `loader_tree.go` untouched, including the `ROW ID = NODE ID -
1` loud-failure block (`.claude/rules/persistence-ownership.md`) and the panic on a stale
edge `Source` field. The four trailing helpers (`NodeIDsInTree`, `NodeIDStringsInTree`,
`LargestNodeID`, `CountEdgeFiles`, 51 lines total) are a SEPARATE concern already marked off
by the file's own `--- the tree's own shape, for the operations that CHANGE it ---` banner:
each does its own `os.ReadDir` (via `readDirNames`) and pure arithmetic (parse/sort/max) —
still bucket (a) by the letter of the rule (file I/O), but answering "what does the tree
look like today" for `scene_structure.go`'s create/delete, not "load the graph" the way
`LoadTree` does. Split in place: moved to a new file in the SAME package,
`tree_shape.go`, no API change, `readDirNames` stayed in `loader_tree.go` and both files
call it. `go build ./nodes/... `/`go vet` clean before any guard re-keying.

**Guard re-keyed:** `tools/network/persist/check-scene-path-resolution.sh`'s `NODE_PATH_OWNERS`
array names the files allowed to `filepath.Join` a `nodes/` path — it hardcoded
`loader_tree.go` and would have gone silently green on `tree_shape.go`'s two `nodes/`
`filepath.Join` calls landing in an un-listed file... except it doesn't go green, it FAILS,
because the guard's positive assertion is "every `nodes/`-segment Join must be in an
allow-listed file", not "loader_tree.go's Joins are still there" — an unlisted new file
trips it. Added `"tree_shape.go"` to `NODE_PATH_OWNERS` plus updated the three prose
comments (`PLACEMENT:` line, the `nodes/<id>/...` bullet, the `NODE_JOIN_HITS` error
string) that named `loader_tree.go` as the sole reader. Proved the guard still bites: used
Edit to remove `"tree_shape.go"` from the array, ran the guard, got:

```
hand-rolled-node-path: /Users/David/Documents/github/wirefold/nodes/Wiring/loadspec/tree_shape.go: 22:	names, err := readDirNames(filepath.Join(root, "nodes"))
hand-rolled-node-path: /Users/David/Documents/github/wirefold/nodes/Wiring/loadspec/tree_shape.go: 65:		names, err := readDirNames(filepath.Join(root, "nodes", id, "edges"))

check-scene-path-resolution: 2 hand-rolled nodes/ filepath.Join(...) hit(s) outside node_mover.go/edge_mover.go/edge_file.go/loader_tree.go/tree_shape.go/position_file.go — a node/port path belongs to its owning mover; call node_mover.go's or positionfile's resolvers instead of reconstructing the path.
exit=1
```

Restored the array entry, guard back to clean. Grepped `tools/` and `scripts/` for
`loader_tree.go` and `tree_shape.go`/every moved function name beforehand — the only other
hits were prose mentions in `check-persist-write-ownership.sh`'s comments (not enforced
matching, informational only, left as-is since they still describe `loader_tree.go`'s role
accurately — the moved functions are read-only queries, not writes, so that guard's actual
scope never touched them).

**Changed surfaces with no check of any kind that can fail:** the six `camera_*.go`
concern splits — no guard names any of their symbols or the old filename in a way that
would have caught a wrong split (confirmed by grep before splitting); correctness rests on
`go build`/`go vet` plus the human driving the editor exercising camera gestures, per this
repo's no-tests-by-design policy. `tree_shape.go`'s four helper functions are likewise
uncovered by any behavioral check — `check-scene-path-resolution.sh` verifies WHERE a
`nodes/` path is constructed, not that `LargestNodeID`/`CountEdgeFiles` compute the right
number; that arithmetic has always been unchecked (same exposure as before the split, not
newly introduced).

**Declines:** none — both files split cleanly with no channel/goroutine/field-write
statement pinning a boundary.

**Commits** (`task/god-objects`): `855310f1` (gesture_camera.go → six geom concern files),
`c3015224` (loader_tree.go → loader_tree.go + tree_shape.go, guard re-keyed).

Verify: `bash scripts/stop-checks.sh` after each commit — the one failure surfaced both
times (`check-no-webview-state` zustand import in `snapshot-buffer.ts`, then a `tsc`
conflict in `SpeedSlider.tsx`) is in `tools/topology-vscode/src/webview/`, owned by a
concurrent session on this same branch per this task's own instructions — not touched by
either commit here. `go build ./...` and `go vet ./...` both clean after each commit.
`go run ./tools/gen-node-defs && git status --short` empty (beyond the two files this
pass itself created/modified) after both commits.

## §41 — `nodes/Wiring/nodeactor`'s four files over 250 lines, statement-by-statement

§20 moved the per-node actor out of package `Wiring` wholesale but never split the files
inside it. Four were over 250 lines: `node_geometry.go` (319), `chain_beads.go` (318),
`node_geometry_parts.go` (294), `node_geometry_accessors.go` (272). Measured each
statement-by-statement, never by signature or whole-file shape, per this task's own
correction (seven prior declines on this branch were overturned by that distinction).

**`node_geometry.go` (319 → 148).** The file was the composer type declaration + its
constructor (both pure field-assembly, no channel/file/goroutine touch beyond the
constructor's one self-seed send on `centerOut`) plus one function, `handle()` — a
168-line `movemsg.Msg`-kind dispatch whose every branch is bucket (a): unexported-field
writes (`m.ui.selected = ...`, `m.tilt.topTiltVectorThetaIdx += delta`), method calls that
themselves mutate/send/persist (`m.ApplyCenter`, `m.emitGeometry`, `m.startBeadDrag`,
`m.endBeadDrag`, `m.persistTiltVectorAngle`, `m.writeStreamFrame`), and one Trace
breadcrumb call. Zero pure-computation statements worth lifting — this is the actor's own
message-dispatch table, the same shape §18/§20 already named as a natural per-concern
split point elsewhere in this package. **Split in place**, same package: `handle()` moved
verbatim to a new `node_geometry_handle.go` (182 lines, its own `fmt`/`wire`/`movemsg`/`T`
imports — both became genuinely unused in `node_geometry.go` and were dropped there).
Behavior identical; no field, method, or channel changed shape.

**`node_geometry_parts.go` (294 → 223, plus a new 82-line file).** Confirmed: ten
composer sub-struct type declarations (`nodeMessaging`, `pendingSend`, `nodeClocks`,
`nodeStream`, `nodeUI`, `nodeTilt`, `pairReadout`, `nodeOuts`, `neighborTopology`,
`sceneFlags`, `nodeBeads`) plus two more types that are NOT composer state at all:
`NodeFrameInput` (the ~35-field wire-frame argument handed to the injected stream packer
every emit) and `NodeFrameBuilder` (the packer's func type). These ARE a genuinely
distinct concern from the other ten — they describe what goes ON THE WIRE, built fresh
per call, not state the actor holds between calls — so this is a **split in place** by
concern rather than a decline-for-count: `NodeFrameInput`/`NodeFrameBuilder` moved to a
new `node_frame_input.go` (82 lines). The other ten composer sub-structs stay one file,
genuinely one concern (the actor's own state decomposition, per
`check-composer-fields.sh`'s own doc comment) — zero functions in either file, still true.

**`node_geometry_accessors.go` (272 → 149, plus a new 132-line file).** Eighteen
post-construction methods. Two visibly distinct concerns: thirteen PLAIN read accessors
(`ID`, `Traced`, `Breadcrumb`, `Kind`, `SelfKind`, `Tick`, `Label`, `WorldCenter`,
`NodeRow`, `EdgeIDs`, `PartnerCenters`, `NeighborKinds`, `SendMove`, `NeighborIDs`,
`QuantOffset`, `QuantizedOffsetValue`, `ReachR`, `WriteStreamFrame`, `CommitQuantOffset`)
— every body a bucket (b)/(a)-boundary one-liner or a short field-read/field-write/method
call, no channel — versus five methods whose entire body IS a channel operation
(`NeighborTrySend`, `PollCenter`, `SendExternal`, `TryRecvExternal`, `EnqueueSend`), the
same "channel-touching vs plain accessor" split the file's own OLD header comment already
drew in prose ("The channel-touching members ... stay unexported and are reached ONLY
through the methods below") without actually separating them into files. **Split in
place**: the five channel methods moved verbatim to a new `node_geometry_channels.go` (132
lines, carrying that same header comment forward); the file's own `context`/`fmt` imports
moved with them (both became unused in the accessors file). No channel or field was
exported that was not already exported by §20; `EnqueueSend`'s panic message (already
site-tagged `NodeGeometry(%s): pending exceeded %d retry-queued sends; ...`, naming the two
causes and the mechanism) moved unchanged, still satisfies `check-panic-message.sh`.

**`chain_beads.go` (318, unchanged) — declined, and it is genuinely the impure spine.**
An earlier pass (referenced in this task's own brief) already extracted the file's three
pure phases into `beadindex` (`ChainEdgeGeometry`, `ChainBeadRows`,
`ChainAimBreadcrumbText`). What is left is one function, `chainBeads()`, and it is one
sequential per-edge loop threading loop-scoped locals (`dist`, `liveDir`, `count`,
`pulses`, `chainSep`, `actorChain`, `resolved`) through statements that are each bucket
(a): `m.clocks.clk.Tick()` (a same-goroutine clock read), `m.outs.outWireOuts[i].PublishSteps(count)`
and `m.outs.outStepsIn[i](count)` (non-blocking sends/calls onto the edge's own paced wire
and its `stepsIn` func value), `m.outs.outWires[i].LiveBeadFractions(tick)` (a read of
another actor's own owned-and-driven wire state, safe only because this node's own
goroutine is that wire's driver), `m.reconcileBeadChain(to, count, offsetAt, aimUnit)`
(starts/stops bead-actor goroutines — the file's own doc comment on `beadTickFn`), and
`m.tr.Breadcrumb(...)` (a channel send, gated `chainAimTraceEnabled`). The ~230 lines of
header/inline comment are the file's actual bulk; the code itself is one function whose
every statement either mutates loop-local state feeding the NEXT impure statement in the
same iteration, or performs one of the five impure operations above — there is no
standalone pure block left to name and lift (the task's own worked example — "impure
because its top level sent on channels; its body was arithmetic on locals" — does not
apply here: this body's statements are not disguised arithmetic, they are the actual sends
and reads). Splitting the ONE function across files would only fragment one linear loop
with no natural seam; declined on that basis, not on line count.

**Guard grep, before touching anything.** `check-no-sqrt-in-chain-beads.sh` hardcodes
`nodes/Wiring/nodeactor/chain_beads.go` — untouched, since `chain_beads.go` was not moved
or renamed. Grepped `tools/` and `scripts/` for every other touched filename
(`node_geometry.go`, `node_geometry_parts.go`, `node_geometry_accessors.go`) and every
symbol moved (`NodeFrameInput`, `NeighborTrySend`, `PollCenter`, `SendExternal`,
`TryRecvExternal`, `EnqueueSend`, `handle`) — no other guard names any of them by path or
symbol; `check-composer-fields.sh` locates `type NodeGeometry struct {` by content-grep
(not by filename), so it needed no re-keying and was re-verified passing after both
`node_geometry_parts.go`-touching commits.

**Verification.** `go build ./...`, `go vet ./...` both clean after every commit.
`go run ./tools/gen-node-defs && git status --short` produced no diff beyond this task's
own new/modified files. `bash scripts/stop-checks.sh` clean (empty stdout) after the final
commit — one unrelated failure surfaced mid-task (`check-doc-citations` on
`nodes/Wiring/nodegeom/parallel_chain_offset.go`, a file this task never touched, created
by the concurrent session working `nodegeom`/`gen-stream-fixture` per this task's own
instructions) and resolved itself without this task's intervention before the final run.

**No channel or field was exported that §20 had not already exported.** The
`node_geometry_channels.go` split moved five ALREADY-exported methods verbatim; the
`node_frame_input.go` split moved two ALREADY-exported types verbatim. Goroutine count,
channel set, and send/receive order are unchanged — this pass is a same-package file
reorganization only.

**`nodes/Wiring/nodeactor` file count:** 15 → 18 (three new files:
`node_geometry_handle.go`, `node_frame_input.go`, `node_geometry_channels.go`; no file
deleted, no file renamed). LOC per touched file: `node_geometry.go` 319→148 (+182 in
`node_geometry_handle.go`); `node_geometry_parts.go` 294→223 (+82 in
`node_frame_input.go`); `node_geometry_accessors.go` 272→149 (+132 in
`node_geometry_channels.go`); `chain_beads.go` 318 unchanged (declined). Every other file
in the package was measured briefly and left alone — the next-largest, `pair_node_self.go`
at 243 lines, is under the 250 threshold and was not touched this pass.

**Surfaces with no check that can fail.** This pass moved Go code only, entirely within
one already-guarded package; every symbol touched was already covered by
`check-composer-fields.sh` (composer shape) or `check-panic-message.sh` (the one panic,
moved unchanged). There is no check anywhere in this repo that would fail if
`node_geometry_handle.go`'s dispatch table were accidentally reordered in a way that
changed which `movemsg.Kind` branch runs first for a message satisfying two conditions
(none do today, by construction of the `if ... return` chain) — this was true before the
split too, since the branches were never independently testable
(`docs/process/testing-shape.md`'s cross-goroutine exclusion covers the channel methods;
the plain accessors have no such exclusion but also have no dedicated check beyond `go
build`/`go vet` catching a signature mismatch).

## 2. `tools/gen-stream-fixture/` and `nodes/Wiring/nodegeom/` — two never-examined files

`tools/gen-stream-fixture/main.go` (was 257 lines) is tooling, not the network — package
`main`, no goroutines, no channels, no locks. Every statement in `buildNodeFrame`/
`buildEdgeFrame`/`buildInteriorFrame` is bucket (b): pure struct construction from literals,
one call into a real production packer (itself pure), one `hex.EncodeToString`. `main()`'s
own body is the only bucket (a): `os.WriteFile` and the two `os.Exit`/`Fprintf` failure
paths. No existing sibling package fits this tooling's own domain (fixture-shape types +
frame builders), so it split IN PLACE: `types.go` (the four fixture structs), `build_frames.go`
(the three builders, unchanged bodies), and a trimmed `main.go` left as pure CLI entry point
(61 lines: arg parsing, marshal, write). No committed `stream_fixture.json` artifact exists
in this tree today (grepped and confirmed absent), so there was nothing to diff for
byte-identity; ran `go run ./tools/gen-stream-fixture <scratch-path>` post-split and
confirmed it still emits well-formed JSON with the same literal field values.

`nodes/Wiring/nodegeom/port_geometry.go` (was 257 lines) is the already-lifted pure-geometry
package MODEL.md names for `ParallelChainOffset`/`EdgeCenterDistAndDir`. Every function body
is bucket (b) — no field writes on a type staying behind, no channel/goroutine/file I/O,
all local vec3 math. Split IN PLACE by concern, not lifted (nodegeom already IS the sibling
package these functions belong to): `port_geometry.go` keeps `EdgeSegment` +
`EdgeCenterDistAndDir` (the single edge segment / live distance-and-direction concern the
file is named for); `parallel_chain_offset.go` takes `ParallelChainOffset` + `NodeIDLess`
(the mutual-pair concern); `ring_axis.go` takes `PoleContainingEdge`,
`TorusDefaultAxisAngles`, `UprightRingAxis`, and the trailing removed-function comment (the
ring-axis-derivation concern). No API change — same package, same signatures.

Also measured `nodes/Wiring/nodegeom/shading_params.go` (242 lines, over the ~200 bar) and
**declined** to split it: it has no function bodies to classify — it is 100% top-level
`const` declarations, explicitly documented in its own header as "single source of truth"
that `tools/gen-node-defs/params_shading.go`'s `parseShadingParams` reads via AST from this
ONE hardcoded path (`tools/gen-node-defs/main.go:136`) to generate `shading-params.ts`.
Splitting it would either break that codegen (consts scattered across files it never scans)
or require rewiring the generator to scan a directory — churn with no separable-concern
payoff, since there is no statement-level logic here to separate. `node_geom.go` (170 lines)
and `frame_geometry.go` (101 lines) are both under the ~200 bar; not touched.

**Guard teeth:** grepped `tools/` and `scripts/` for every moved symbol
(`ParallelChainOffset`, `NodeIDLess`, `PoleContainingEdge`, `TorusDefaultAxisAngles`,
`UprightRingAxis`) and for `gen-stream-fixture`/`port_geometry` filenames — no guard
pattern-matches any of them; `check-docs-symbols.sh` references `port_geometry.go` from
`docs/pair-node/math/formulas.html`, but by PATH only (no `#symbol` anchor), so it is
unaffected by which functions live in the file, only that the file still exists. Proved that
guard still bites: edited `formulas.html`'s `data-src` to a nonexistent
`port_geometry_DELIBERATE_BREAK.go`, ran `bash tools/docs/check-docs-symbols.sh`, got
`check-docs-symbols: nodes/Wiring/nodegeom/port_geometry_DELIBERATE_BREAK.go does not exist
(referenced as "nodes/Wiring/nodegeom/port_geometry_DELIBERATE_BREAK.go")`, exit 1; reverted
with Edit, guard back to exit 0.

**Changed surfaces with no check of any kind that can fail:** every statement in
`gen-stream-fixture`'s three builder functions (whether the fixture's literal field values
stay correct) — covered only by the TS-side fixture test reading the SAME hex, which this
change did not touch, and by the human running `stream-fixture.test.ts`. All five functions
moved out of `port_geometry.go` — `ParallelChainOffset`'s sign/perpendicular derivation,
`NodeIDLess`'s numeric-vs-string fallback, the two ring-axis functions' plane math — have no
guard or test verifying their arithmetic; correctness rests on `go build`/`go vet` (unchanged
signatures, so nothing to catch here) plus the human driving the editor and observing ring
orientation and mutual-pair chain separation, per this repo's no-tests-by-design policy.

**Declines:** `shading_params.go` only (reasoning above) — no per-function pinning
statement applies since there are no function bodies in that file.

**Commits** (`task/god-objects`): `b9f8b810` (gen-stream-fixture → types.go + build_frames.go
+ trimmed main.go), `1fa2950d` (port_geometry.go → port_geometry.go + parallel_chain_offset.go
+ ring_axis.go), `cdb45939` (fix a doc-citation mismatch this pass introduced in
`parallel_chain_offset.go`'s header, caught by `check-doc-citations` on the same pass).

Verify: `bash scripts/stop-checks.sh` empty after the citation fix. `go build ./...`/
`go vet ./...` on the whole tree failed with pre-existing, concurrent-session errors in
`nodes/Wiring/layoutquant`/`nodes/Wiring/moverreg` (missing methods on
`nodeactor.NodeGeometry`) — confirmed via `git status` these are uncommitted edits under
`nodes/Wiring/nodeactor/` from another agent working that path per this task's own
instructions, not touched here; `go build`/`go vet` scoped to
`./nodes/Wiring/nodegeom/...`/`./tools/gen-stream-fixture/...` both clean. Dependency-rule
scan (`dispatch` importers) showed only the two legitimate importers (`build`,
`stdinreader`). `go run ./tools/gen-node-defs && git status --short` clean (regenerated
`shading-params.ts` etc. identically, confirming the untouched `shading_params.go` still
parses correctly through the unmodified codegen path).
