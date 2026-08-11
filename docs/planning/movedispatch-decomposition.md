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
