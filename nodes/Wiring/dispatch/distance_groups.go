// distance_groups.go — the "distance home button" toolbar panel's Go-owned controller.
//
// Three named groups of node-pair distances (GATE definitions in
// nodes/Wiring/distancegroups, node ids are the string ids from topology/nodes/<id>). Each
// pair is (source, target) taken from the bead edge's direction; the TARGET node is the one
// that moves. Clicking a group's up arrow sets that group's target length L = currentMax*1.1
// (down: currentMax/1.1), where currentMax is the max over the group's pairs of
// |center(target)-center(source)| (mirrors reachRFromPolar's max-over-edges loop,
// quantized_move.go). Then EVERY pair in the group is repositioned so its length becomes L,
// in FLAT LIST ORDER — a target node that appears in two pairs (gate group: 8 is the target
// of both (3,8) and (5,8); 9 of both (5,9) and (7,9)) ends at the LAST pair's placement. This
// is intended and accepted (per the agreed model): there is no tree/graph solver, no
// averaging, no equal-radii resolve for a shared target.
//
// The repositioning itself reuses lq.RootMove exactly as the drag tests call it
// programmatically (abc_drag_count_target_node_test.go) — RootMove already routes the
// move to the target node's OWN goroutine, commits via commitNodeMoveLocal, and
// rebroadcasts geometry so incident edges' segments recompute and redraw. This file adds
// no new position/commit path.
//
// The GROUP TABLE and the MAX/LENS/APPLY math live in nodes/Wiring/distancegroups
// (god-object decomposition): those functions only ever needed "read a node's live center"
// and "move a node's target" through *moverreg.MoverRegistry/*layoutQuantizer, both actor
// types — so they now take centerOf/rootMove as plain func values, bound to the real
// actor methods here, the same bound-func-value pattern move_dispatch_construct.go already
// uses (`ng.msg.sendMove = md.MR.EnqueueFor(ng)`). This file is left holding exactly the
// parts that read/write MoveDispatch's own actor state: the ResolveSceneDistanceGroups
// writer, and the two thin wrappers that bind centerOf/rootMove and forward.
package dispatch

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/distancegroups"
	"github.com/dtauraso/wirefold/nodes/Wiring/layoutquant"
	"github.com/dtauraso/wirefold/nodes/Wiring/moverreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/scene"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

// ResolveSceneDistanceGroups records whether the tree being loaded owns the three named
// distance groups (SceneTab.DistanceGroups). Call it once at load, BEFORE anything emits a
// VIEW frame — the frame carries the three group lengths, and until this runs they read as
// "no groups", which is the safe direction: a scene that should have them shows them one
// frame later, whereas the reverse would flash the ring's lengths into another scene's tab.
func (md *MoveDispatch) ResolveSceneDistanceGroups(scenePath string) {
	md.UI.HasDistanceGroups = scene.SceneHasDistanceGroups(scenePath)
	// Resolved from the same path at the same moment, for the same reason: both are facts
	// about the tree being loaded, and both ride the VIEW frame. Until this runs the scene
	// reads as NOT editable, which is the safe direction — a palette that appears a frame
	// late costs nothing, one that appears in a scene that cannot take it invites a delete.
	md.UI.SceneEditable = scene.SceneIsEditable(scenePath)
	md.UI.SceneKinds = scene.SceneKindMask(scenePath)
}

// DistanceGroupLens returns the 3 groups' current max pair lengths, in
// distancegroups.GroupOrder (time, input, gate) — for the VIEW stream's Overlay
// GroupLenTime/GroupLenInput/GroupLenGate columns (read-only reflect; see view_stream.go's
// emitViewFrame). A group whose centers aren't resolvable yet reads 0. Binds
// mr.CenterOfNode once per call and forwards to distancegroups.Lens, which does the actual
// per-group max/scan.
func DistanceGroupLens(ui *viewstate.UIState, mr *moverreg.MoverRegistry) (timeLen, inputLen, gateLen float32) {
	return distancegroups.Lens(ui.HasDistanceGroups, mr.CenterOfNode)
}

// applyDistanceGroupTarget is the controller for one arrow click: groupIdx indexes
// distancegroups.GroupOrder (0/1/2, out of range = no-op); dir > 0 is the up arrow, dir < 0
// is down. Binds mr.centerOfNode and lq.RootMove and forwards to distancegroups.ApplyTarget,
// which does the actual max/target-length/reposition-loop math. Returns false if the group
// is unknown, has no resolvable pair, or groupIdx is out of range.
//
// The 3 GroupLen* Overlay columns are recomputed from live centers ONLY inside
// emitViewFrame (view_stream.go). RootMove emits NODE frames (the moved geometry), not the
// VIEW frame, so without a caller-side emit the panel's displayed lengths never refresh
// after a button press. This runs on the stdin/dispatch goroutine — the VIEW-stream owner —
// so emitting is safe once distancegroups.ApplyTarget's internal settle has returned.
//
// This function itself no longer emits: the caller (applyUpdateDistanceGroup,
// stdin_apply.go — the same stdin/dispatch goroutine) emits the VIEW frame when moved is
// true, per docs/planning/movedispatch-decomposition.md's write-then-emit split.
func applyDistanceGroupTarget(ctx context.Context, ui *viewstate.UIState, mr *moverreg.MoverRegistry, lq *layoutquant.LayoutQuantizer, groupIdx, dir int) bool {
	rootMove := func(ctx context.Context, target string, newPos vec3) bool {
		return lq.RootMove(ctx, mr.NodeGeoms(), target, newPos)
	}
	return distancegroups.ApplyTarget(ctx, ui.HasDistanceGroups, mr.CenterOfNode, rootMove, groupIdx, dir)
}
