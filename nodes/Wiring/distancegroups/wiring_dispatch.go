// wiring_dispatch.go — the "distance home button" panel's Wiring-facing entry points:
// ResolveSceneDistanceGroups and ApplyDistanceGroupTarget. Moved from
// nodes/Wiring/dispatch's distance_groups.go (docs/planning/movedispatch-decomposition.md,
// the remainder cluster) — both already read/wrote only already-exported sub-objects
// (*viewstate.UIState, *moverreg.MoverRegistry, *layoutquant.LayoutQuantizer), the same
// bound-func-value shape distance_groups.go itself already used to reach Max/Lens/ApplyTarget
// above. DistanceGroupLens stays in package dispatch — move_dispatch_construct.go's
// NewMoveDispatch calls it to bind md.UI.DistanceGroupLensFn, and moving it here would make
// dispatch import this package for the delegator while this package would need to import
// dispatch back for MoveDispatch, a real cycle.
package distancegroups

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/layoutquant"
	"github.com/dtauraso/wirefold/nodes/Wiring/moverreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/scene"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// ResolveSceneDistanceGroups records whether the tree being loaded owns the three named
// distance groups (SceneTab.DistanceGroups). Call it once at load, BEFORE anything emits a
// VIEW frame — the frame carries the three group lengths, and until this runs they read as
// "no groups", which is the safe direction: a scene that should have them shows them one
// frame later, whereas the reverse would flash the ring's lengths into another scene's tab.
func ResolveSceneDistanceGroups(ui *viewstate.UIState, scenePath string) {
	ui.HasDistanceGroups = scene.SceneHasDistanceGroups(scenePath)
	// Resolved from the same path at the same moment, for the same reason: both are facts
	// about the tree being loaded, and both ride the VIEW frame. Until this runs the scene
	// reads as NOT editable, which is the safe direction — a palette that appears a frame
	// late costs nothing, one that appears in a scene that cannot take it invites a delete.
	ui.SceneEditable = scene.SceneIsEditable(scenePath)
	ui.SceneKinds = scene.SceneKindMask(scenePath)
}

// ApplyDistanceGroupTarget is the controller for one arrow click: groupIdx indexes
// GroupOrder (0/1/2, out of range = no-op); dir > 0 is the up arrow, dir < 0
// is down. Binds mr.centerOfNode and lq.RootMove and forwards to ApplyTarget,
// which does the actual max/target-length/reposition-loop math. Returns false if the group
// is unknown, has no resolvable pair, or groupIdx is out of range.
//
// The 3 GroupLen* Overlay columns are recomputed from live centers ONLY inside
// emitViewFrame (view_stream.go). RootMove emits NODE frames (the moved geometry), not the
// VIEW frame, so without a caller-side emit the panel's displayed lengths never refresh
// after a button press. This runs on the stdin/dispatch goroutine — the VIEW-stream owner —
// so emitting is safe once ApplyTarget's internal settle has returned.
//
// This function itself does not emit: the caller (applyUpdateDistanceGroup, dispatch_apply.go
// — the same stdin/dispatch goroutine) emits the VIEW frame when moved is true, per
// docs/planning/movedispatch-decomposition.md's write-then-emit split.
func ApplyDistanceGroupTarget(ctx context.Context, ui *viewstate.UIState, mr *moverreg.MoverRegistry, lq *layoutquant.LayoutQuantizer, groupIdx, dir int) bool {
	rootMove := func(ctx context.Context, target string, newPos wire.Vec3) bool {
		return lq.RootMove(ctx, mr.NodeGeoms(), target, newPos)
	}
	return ApplyTarget(ctx, ui.HasDistanceGroups, mr.CenterOfNode, rootMove, groupIdx, dir)
}
