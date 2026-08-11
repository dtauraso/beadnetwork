---
branch: task/god-objects
---

# Making the gesture FSM its own actor

The last 20 of `MoveDispatch`'s 37 methods are one subsystem — the gesture → view loop — and
it resists every decomposition tool used on this branch because it is a CYCLE, not a
hierarchy: read raw input → mutate UI state → emit a frame, with each step touching the same
state. Extract-type and extract-package pull a leaf off a tree; there is no leaf here.

A cycle is cut with a channel, not a function signature. That is what this change does, and it
is the shape MODEL.md already uses everywhere else.

## What is already true (measured, not assumed)

**The FSM is already actor-shaped with respect to the movers.** It reads mover-owned state
through exactly two accessors:

- `mr.centerOfNode(id)` — reads `centerMirror`, *"the DISPATCH goroutine's OWN mirror of every
  node's last-known world center, kept current by messages from each node's own goroutine"*
  (`mover_registry.go`), drained non-blockingly from each mover's `centerOut` channel.
- `mr.nodeBodyRadius(id)` — a static per-node dimension.

So there is no shared-memory read to remove. The mirror-fed-by-channels pattern this change
would need already exists and is already in use.

**`RunStdinReader` is already framing + routing only.** It takes a `stdinreader.Handlers`
struct of three func values (`ApplyEdit`, `HandleRawInput`, `HandleSave`) and imports nothing
from `Wiring`. The seam a channel would replace is already a seam.

## The one real problem: who owns the VIEW stream

MODEL.md pins one writer per stream. Today the view-owner is `RunStdinReader`'s goroutine, and
BOTH the gesture handlers and the stdin dispatch handlers emit on it — they are the same
goroutine, so that is currently safe. Split the FSM off and there are two writers unless
ownership moves with it.

Emit sites today: `viewpoint_state.go` (5), `gesture_actions.go` (2), `gesture_handlers.go`
(1), `gesture_graph.go` (1) — the FSM; plus `stdin_dispatch.go` (2), `scene_*_persist.go` (4),
`scene_structure.go` (1), `distance_groups.go` (1), `view_stream.go` (1) — the dispatch side.

**Decision: the gesture goroutine becomes the view owner.** It is the dominant emitter
(camera, selection, hover, drag are all gesture), and the alternative — FSM computes, sends a
result, dispatch emits — puts a channel round-trip inside every pointer-move, which is the
tightest loop in the editor.

`RunStdinReader` then does what its name says: frame bytes, decode, route. Raw input goes to
the FSM's channel. Non-raw edits (overlay toggle, clock speed, scene select, structural edit)
ALSO go to the FSM's channel, because they mutate state the frame carries. That is one inbox,
one owner, one writer — the same shape as a `nodeMover`.

## Order

1. **Give the FSM its own inbox and goroutine**, receiving a message union of {raw-input,
   edit, save}. `RunStdinReader`'s three `Handlers` funcs become three channel sends. No
   behaviour change yet — the FSM still calls the same code, just on its own goroutine.
2. **Move VIEW-stream ownership to it.** — **ALREADY DONE BY STEP 1, no separate work.**
   Ownership followed the entry points: all three (`ApplyEdit`, `HandleRawInputMsg`,
   `HandleSaveMsg`) moved to the actor as a unit, and every runtime `emitViewFrame` is
   reachable only from those three, so the actor became the sole runtime writer without
   `SetViewStream` changing at all.

   Verified rather than assumed. There is exactly ONE write site for the stream
   (`view_stream.go:174-175`, inside `writeViewFrame`). No mover emits — grep of
   `node_geometry*.go`, `node_mover.go`, `edge_mover*.go` for `emitViewFrame` is empty.

   The precise invariant is narrower than "one writer" and worth stating exactly:
   **startup seeds sequentially, then the actor owns it.** `runtopology/topology_run.go`
   calls `emitStartupBreadcrumbs` (line 63) and `loadSceneState` (line 65) — which reach the
   stream via `EmitViewpoint`/`LoadOverlays`/`LoadSpeed`/`LoadSceneSphere`/`EmitBreadcrumb` —
   BEFORE `startStdinReader` (line 80) creates the actor. Those writes are on the startup
   goroutine, sequenced before any other goroutine exists, so they never race the actor. Two
   writers in TIME, never concurrently.

   Anything that later adds a view-stream write between line 65 and line 80, or from a
   goroutine other than the actor, breaks this. It is program order doing the work, not a
   lock, and nothing checks it.
3. **Only then** re-measure `MoveDispatch`. NOTE what step 1 did and did not do: the actor
   lives in `runtopology` and calls into `Wiring` — the GOROUTINE moved, the CODE did not. So
   `MoveDispatch` is still 37 methods; nothing lifted yet.

   The payoff is now available but is its own change: with the actor owning all gesture + view
   work on one goroutine, the gesture files, `uiState`, `gestureState`, `viewpointState`,
   `overlayState` and `emitViewFrame` can move into ONE package together — which is exactly
   what the failed `uiState` probe needed. That probe declined because `gestureState`'s 18
   private fields are read at 49 sites across 5 files by handlers typed
   `func(md *MoveDispatch, g *gestureState, …)`; if those handlers move into the same package
   as the state, the field access is no longer a boundary violation.

   That is a large move (5 gesture files + 4 state types + the emitter) and it is NOT approved
   by the step-1 decision. Measure it first: if the resulting package needs a back-reference to
   `MoveDispatch` for anything, it is the same cycle under a new name and should not be built.

## Verification — the part that cannot be skipped

- **`go test -race ./...`** on every step. This is the only change on this branch that adds a
  goroutine; the existing suite does not exercise concurrency (per `docs/process/testing-shape.md`
  the repo deliberately does not test cross-goroutine behaviour), so the race detector is the
  only automated signal.
- **Exactly one writer per stream** must remain true. `memory/feedback/architecture/feedback_no_single_writer_bridge.md`
  and MODEL.md both pin it. State how it was checked, not that it holds.
- **Drive the editor.** Pointer-move is the tightest loop here and the suite asserts nothing
  about it. A regression shows up as lag, a dropped drag, or a stale frame — none of which any
  test will catch.
- The `.probe` logs (`.claude/rules/go-debugging.md`) are the diagnostic if it misbehaves.

## Risks

- **This adds a goroutine to the interaction path.** Every prior step on this branch was pure
  code motion with `no behaviour change` as the claim. This one changes when work happens.
  It is not reversible by `git revert` once the editor has been used against it.
- **Deadlock**: the FSM sends to movers and movers push to `centerOut`, which the FSM drains.
  If the FSM ever blocks on a send while a mover blocks on a push, both stop. Every send on
  this path must stay non-blocking (`SendLatestNonBlocking` / `select` with `default`), which
  is the existing convention — verify it, do not assume it.
- **If step 2 finds an `emitViewFrame` caller that cannot move to the FSM goroutine**, stop.
  That is a real second writer and the design is wrong as stated.

## Step 3 — probed and declined, then the decline corrected (owner-pointer binding done, lift itself still not attempted)

The re-measurement this doc calls for (`## The one real problem`'s closing paragraph) first
found the lift does not go through: `docs/planning/movedispatch-decomposition.md`'s "6a."
section has the original account. Short version of that original finding: the gesture
files' own package-level helpers (`beginSphereRotation`, `applyNodeDragTarget`,
`commitDragStart` in the cluster, plus `setSelectionUI`/`sendMove`/`DistanceGroupLens` in
`move_dispatch_api.go`/`distance_groups.go`, which stay in `Wiring`) take `*moverRegistry`/
`*layoutQuantizer` — unexported `Wiring` types — as parameters, not just through the three
read-only accessors (`centerOfNode`, `nodeBodyRadius`, `heldCenters`) the earlier measurement
named. Those two types cannot be named outside package `Wiring` at all (Go's export rule,
not a design choice), so the cluster could not lift while it still needed `mr`/`lq` by
pointer.

**That decline tested the signature, not the body — corrected in
`docs/planning/movedispatch-decomposition.md`'s "6." section.** None of the four functions
USES `mr`/`lq` for anything beyond passing them on to `sendMove`, `heldCenters`, and
`RootMove`. Replaced the owner-pointer parameters with bound func values
(`sendMove func(id string, msg movemsg.Msg)`, `heldCenters func() map[string]vec3`,
`rootMove func(id string, target vec3) bool`), the same treatment `nodeGeometry` already
gets (`ng.msg.sendMove = md.mr.enqueueFor(ng)`). `grep -nE "\*moverRegistry|\*layoutQuantizer"`
across the nine cluster files is now empty; `go test -race -count=1 ./...` passes.

**The package lift itself is still NOT attempted.** Binding was the whole scope of that
task, done deliberately so the lift can be measured fresh with the owner-pointer blocker
actually gone, rather than folded into the same change. `git status --short` was NOT empty
during this pass (four files touched, one commit); the original probe's "no file written"
claim applies only to the superseded decline, not to this correction.

## Step 3 — HALF lifted: the FSM's hub-free state moved, uiState still does not come free

Re-measured the cluster at method granularity per
`docs/planning/movedispatch-decomposition.md`'s "6b." section: 17 `MoveDispatch` entry
methods stay (they build closures reaching `Wiring`); `gestureState`/`gestureRect`
(gesture.go) and `viewpointState` (viewpoint_state.go), which reference nothing but their
own fields plus `geom`/`Trace`, moved into `nodes/Wiring/gesturefsm` as exported types, with
plain type aliases left behind in `Wiring` (same shape as `vec_alias.go`'s
`vec3 = wire.Vec3`) so every call site keeps its short name.

`uiState` itself is still blocked, confirmed rather than assumed: `view_stream.go` — out of
scope for this task by the task's own constraint — reads `md.ui.ov.*` (13 fields),
`md.ui.editRefused/sceneEditable/sceneKinds/speed/lastDraggedNode/sceneSphere` all by
unexported field, and `move_persist.go` writes `md.ui.vp.Persist` by field. Moving `uiState`
would force `overlayState`'s 13 booleans open too — a second type's surface forced by a
THIRD file, not the moving type's own API, which is the explicit revert condition. So
`beginSphereRotation`/`applyNodeDragTarget`/`commitDragStart`/`setHover`/`dragPlaneHit`
(the leaf actions that take `*uiState`) stay in `Wiring` unchanged, and the 7 export-blocked
`MoveDispatch` methods are still blocked — no second commit.

`go test -race -count=1 ./...` clean; the no-imports-`Wiring` loop empty;
`go run ./tools/gen-node-defs` no diff; `gesturefsm` has zero dependency on `Wiring` (no
cycle existed to route around). Full detail, coverage-injection result, and the one
uncovered gap found (Reset not asserting `DragNode` is cleared) are in
`docs/planning/movedispatch-decomposition.md`'s "6b." section.

## Step 4 — LIFTED: `uiState`/`overlayState`/the VIEW emitter into `nodes/Wiring/viewstate`

The blocker in Step 3 was never `uiState`'s own field count — it was that `view_stream.go`
(explicitly out of scope for that task) reached `md.ui.ov.*`/`md.ui.vp`/`md.ui.sceneSphere`
by field, from a THIRD file. This task named `view_stream.go` in scope and moved it
alongside `uiState`/`overlayState`, which dissolves that exact objection: the VIEW emitter's
direct field access is now intra-package.

**Lifted into `nodes/Wiring/viewstate`** (exported, matching `GS`/`RT`/`Scenes`'s "the
type's own surface" cost): `UIState` (was `uiState`, all 14 fields capitalized),
`OverlayState` (was `overlayState`, generated — `tools/gen-node-defs/overlay_gen.go` now
writes `nodes/Wiring/viewstate/overlay_state.go`, confirmed byte-identical against a
hand-written draft before wiring the generator to it), the VIEW frame packer
(`ViewFrameBuilder`/`ViewOverlayFlags`/`ViewSceneState`/`EmitViewFrame`/`EmitBreadcrumb`/
`SetViewStream`), and the VIEW stream's own fd/claim (`viewClaimedStream`, a small
independent copy of `stream_claim.go`'s `claimedStream` scoped to the VIEW stream's own
singleton claim — confirmed safe to split, per this doc's own earlier note that `sw.claims`
was already namespaced `"view:"`/`"node:<id>"`/`"edge:<label>"`). `streamWiring`
(`stream_wiring.go`) keeps only the node/edge fd directories; its `viewOut`/`viewBuildFrame`/
`viewTick` fields moved onto `UIState`.

**Two Wiring-only dependencies (`moverRegistry`'s edge-mover map, `*rowtables.RowTables`/
`DistanceGroupLens`'s `*moverRegistry`) could not follow `EmitViewFrame`/`SetSelectionUI`
into `viewstate`** — same class of blocker the gesture cluster hit, same fix: bound func
values. `UIState.SetSelectionUI` takes `sendMove`/`sendEdgeSelect` closures (the edge-send
logic itself, `sendEdgeSelect`, stays a Wiring package-level func, since it needs
`map[string]*edgeMover`); `UIState.NodeRowFor`/`UIState.DistanceGroupLensFn` are bound ONCE
in `newMoveDispatch` (mirroring `ng.msg.sendMove = md.mr.enqueueFor(ng)`), not re-resolved on
every emit.

**~15 small `*uiState`-typed methods that read only their own fields moved with the type**
(`SetHoverUI`, `DropPointFromNDC`, `DragPlaneHit` (was `dragPlaneHit`),
`OrbitViewpoint`/`OrbitLockedViewpoint`/`ZoomViewpoint`, `RefuseStructuralEdit`) — these were
never blocked by anything but living in the same file as the type; once the type moved, so
did they, as plain package-boundary moves (not method-syntax-preference conversions).

**Every other `*uiState`-taking function in `Wiring`** (the gesture cluster's
`beginSphereRotation`/`applyNodeDragTarget`/`commitDragStart`/`setHover`, `distance_groups.go`'s
three functions, `commitNodeMoveLocal`) had its parameter type updated to
`*viewstate.UIState` and stayed in `Wiring` — they are entry points/owner-crossing code, not
part of the moving type's own surface.

**Mechanical breadth, not a new blocker.** ~48 files touched (about half of
`nodes/Wiring`'s own file count) — every file that read `md.ui.X`/`ui.X`/`overlayState.X` by
field needed its selector capitalized and, where the value crossed a package boundary, its
receiver retyped to `*viewstate.UIState`/`viewstate.OverlayState`. No alias shim was used
(the ninth-shim-then-removed lesson from Step 3's own history, `movedispatch-decomposition.md`
section "6c."); every call site names the type directly.

**Second commit (the payoff): 3 of the 7 export-blocked `MoveDispatch` methods deleted.**
`SetViewpoint`, `EmitViewpoint`, `SetViewStream` were one-line forwards to `md.UI.VP.X`/
`md.UI.X` — with `UI` now exported, their only reason to exist (an unexported field an
external caller could not reach) was gone. Deleted; `runtopology/scene_state.go`,
`runtopology/view_stream.go`, and `nodes/Wiring/scenecamera`'s own external test now call
`md.UI.VP.SetViewpoint`/`md.UI.VP.EmitViewpoint`/`md.UI.SetViewStream` directly.
`ResolveSceneDistanceGroups`/`LoadOverlays`/`LoadSpeed`/`EnableSceneSwitch` were NOT
attempted — they live in files outside this task's four-things scope
(`distance_groups.go`/`scene_overlays_persist.go`/`scene_speed_persist.go`/
`scene_switch.go`) and each reaches at least one more owner (`persist`, `Scenes`, `scene`
package helpers) beyond `md.UI`, so deleting their delegators is a separate, unscoped
measurement. `MoveDispatch` method count: 37 → 35 (exported: 24 → 22).
`Viewpoint()`/`PanViewpoint()` were kept — neither is one of the 7, and both still have a
genuine external caller reason recorded earlier in this doc.

**VIEW-stream ownership verified, not assumed.** Read `runtopology/topology_run.go` after
the change: `wireViewStream` (→ `md.UI.SetViewStream`) still runs at the same line, before
`emitStartupBreadcrumbs`/`loadSceneState` (→ `md.UI.EmitBreadcrumb`/`md.UI.EmitViewFrame` via
existing delegators), all still before `startStdinReader` creates the gesture actor. Nothing
moved between those lines; only the receiver each call routes through changed name
(`md.SetViewStream` → `md.UI.SetViewStream`), not which goroutine calls it or when.

`go build ./...`, `go vet ./...`, `go test -race -count=1 ./...` clean on both commits; the
no-imports-`Wiring` loop empty; `go run ./tools/gen-node-defs` no diff.
`check-refusal-emits-frame.sh` and `check-overlay-flag-name-parity.sh` both had a hardcoded
path/symbol case (`refuseStructuralEdit(`/`emitViewFrame(` regex, `nodes/Wiring/overlay_gen.go`
path) that would have gone silently green post-rename — both fixed and each confirmed to
fail once on purpose (a deliberately dropped `emitViewFrame(nil)` after
`RefuseStructuralEdit` reproduced `check-refusal-emits-frame.sh`'s exact failure text) before
being restored clean.
