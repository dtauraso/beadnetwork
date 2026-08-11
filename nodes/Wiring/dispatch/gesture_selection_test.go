package dispatch

import (
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/gesturefsm"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

// gesture_selection_test.go — click resolution through the gesture FSM: click-select (node
// and edge), Go-owned, primary tap, secondary (two-finger) tap through drift, select-mode
// independence, and the click-vs-drag movement discriminator.

// Click-select is Go-owned for EDGES too: a click on an edge (new-system: a numeric buffer
// EDGE-ROW hit) resolves the row → edge label via the injected edge-row table and sets
// md.UI.Sel.SelectedEdge, clearing any node selection (exclusive). Selecting a node afterwards
// clears the edge selection, and an empty click clears both.
func TestGestureClickSelectsEdgeGoOwned(t *testing.T) {
	md := newGestureMD(canonicalViewpoint())
	md.RT.EdgeRowTable = []string{"e0", "e1"}
	md.RT.NodeRowTable = []string{"N7"}

	// First select a node so we can prove edge-select clears it.
	nd := rawEvent("pointerdown", 400, 300)
	nd.Hit = inputcodec.RawHit{Kind: "node", NodeRow: 0}
	md.HandleRawInput(nd, nil, nil)
	nu := rawEvent("pointerup", 400, 300)
	nu.Hit = inputcodec.RawHit{Kind: "node", NodeRow: 0}
	md.HandleRawInput(nu, nil, nil)
	if md.UI.Sel.Selected != "N7" {
		t.Fatalf("pre: selected=%q want N7", md.UI.Sel.Selected)
	}

	// Tap EDGE row 1 → selectedEdge=e1, node selection cleared.
	ed := rawEvent("pointerdown", 400, 300)
	ed.Hit = inputcodec.RawHit{Kind: "edge", EdgeRow: 1}
	md.HandleRawInput(ed, nil, nil)
	eu := rawEvent("pointerup", 400, 300)
	eu.Hit = inputcodec.RawHit{Kind: "edge", EdgeRow: 1}
	md.HandleRawInput(eu, nil, nil)
	if md.UI.Sel.SelectedEdge != "e1" {
		t.Fatalf("selectedEdge=%q want e1", md.UI.Sel.SelectedEdge)
	}
	if md.UI.Sel.Selected != "" {
		t.Fatalf("selected=%q want empty (edge select clears node)", md.UI.Sel.Selected)
	}

	// Selecting a node clears the edge selection (exclusive both ways).
	nd2 := rawEvent("pointerdown", 400, 300)
	nd2.Hit = inputcodec.RawHit{Kind: "node", NodeRow: 0}
	md.HandleRawInput(nd2, nil, nil)
	nu2 := rawEvent("pointerup", 400, 300)
	nu2.Hit = inputcodec.RawHit{Kind: "node", NodeRow: 0}
	md.HandleRawInput(nu2, nil, nil)
	if md.UI.Sel.SelectedEdge != "" {
		t.Fatalf("selectedEdge=%q want empty after node select", md.UI.Sel.SelectedEdge)
	}

	// Empty-space click clears the current selection (highlight is transient).
	md.HandleRawInput(rawEvent("pointerdown", 400, 300), nil, nil)
	md.HandleRawInput(rawEvent("pointerup", 400, 300), nil, nil)
	if md.UI.Sel.Selected != "" || md.UI.Sel.SelectedEdge != "" {
		t.Fatalf("after empty click: selected=%q selectedEdge=%q want empty,empty (cleared)", md.UI.Sel.Selected, md.UI.Sel.SelectedEdge)
	}
}

// Click-select is Go-owned: a click on a node sets md.UI.Sel.Selected to that node id; a click on
// empty space clears it. (No camera change — covered by TestGestureClickNoCameraChange.)
func TestGestureClickSelectsNodeGoOwned(t *testing.T) {
	md := newGestureMD(canonicalViewpoint())
	md.RT.NodeRowTable = []string{"N7"}

	down := rawEvent("pointerdown", 400, 300)
	down.Hit = inputcodec.RawHit{Kind: "node", NodeRow: 0}
	md.HandleRawInput(down, nil, nil)
	md.HandleRawInput(func() inputcodec.RawInputMsg {
		e := rawEvent("pointerup", 401, 300)
		e.Hit = inputcodec.RawHit{Kind: "node", NodeRow: 0}
		return e
	}(), nil, nil)
	if md.UI.Sel.Selected != "N7" {
		t.Fatalf("selected=%q want N7", md.UI.Sel.Selected)
	}

	// Empty-space click CLEARS the highlight (md.UI.Sel.Selected), even though the rule-builder's
	// sticky panel Center (md.ruleCenter) stays put — see TestGestureRuleCenterStickyOnEmptyClick.
	d2 := rawEvent("pointerdown", 400, 300) // Hit defaults to empty
	md.HandleRawInput(d2, nil, nil)
	md.HandleRawInput(rawEvent("pointerup", 401, 300), nil, nil)
	if md.UI.Sel.Selected != "" {
		t.Fatalf("selected=%q want empty (cleared) after empty-space click", md.UI.Sel.Selected)
	}
}

// A SECONDARY (two-finger trackpad tap, button 2) select is a tap-select that must survive
// finger drift PAST the move slop: two fingers don't land precisely, so the down→up path
// jitters more than the slop. It must stay gesturefsm.GestPending (never convert to drag/rotate) and
// still resolve to a node select on pointer-up. Empty-space two-finger tap preserves selection.
func TestGestureSecondaryTapSelectsThroughDrift(t *testing.T) {
	md := newGestureMD(canonicalViewpoint())
	md.RT.NodeRowTable = []string{"N7"}

	// Two-finger tap ON a node, with drift between down and up.
	down := rawEvent("pointerdown", 400, 300)
	down.Button = 2
	down.Hit = inputcodec.RawHit{Kind: "node", NodeRow: 0}
	md.HandleRawInput(down, nil, nil)
	if !md.UI.Gest.Secondary || md.UI.Gest.Phase != gesturefsm.GestPending {
		t.Fatalf("after secondary down: secondary=%v phase=%v", md.UI.Gest.Secondary, md.UI.Gest.Phase)
	}
	// Finger drift past the slop must NOT convert to drag/rotate — it stays a tap-select.
	drift := rawEvent("pointermove", 410, 300)
	drift.Button = 2
	drift.Hit = inputcodec.RawHit{Kind: "node", NodeRow: 0}
	md.HandleRawInput(drift, nil, nil)
	if md.UI.Gest.Phase != gesturefsm.GestPending {
		t.Fatalf("secondary tap converted out of pending: phase=%v", md.UI.Gest.Phase)
	}
	up := rawEvent("pointerup", 410, 300)
	up.Button = 2
	up.Hit = inputcodec.RawHit{Kind: "node", NodeRow: 0}
	md.HandleRawInput(up, nil, nil)
	if md.UI.Sel.Selected != "N7" {
		t.Fatalf("selected=%q want N7 after secondary tap-select through drift", md.UI.Sel.Selected)
	}

	// Two-finger tap on EMPTY space (with drift) clears the current selection.
	d2 := rawEvent("pointerdown", 400, 300) // Hit defaults to empty
	d2.Button = 2
	md.HandleRawInput(d2, nil, nil)
	m2 := rawEvent("pointermove", 410, 300)
	m2.Button = 2
	md.HandleRawInput(m2, nil, nil)
	u2 := rawEvent("pointerup", 410, 300)
	u2.Button = 2
	md.HandleRawInput(u2, nil, nil)
	if md.UI.Sel.Selected != "" {
		t.Fatalf("selected=%q want empty (cleared) after secondary empty-space tap", md.UI.Sel.Selected)
	}
}

// press+release with NO move in between still selects (click path intact after the
// distance-threshold removal — see TestGestureClickNoCameraChange for the camera-pose half
// of this, and TestGestureSelectModeOffStillHighlights below for a node-target click).
func TestGesturePressReleaseNoMoveSelects(t *testing.T) {
	md := newGestureMD(canonicalViewpoint())
	md.RT.NodeRowTable = []string{"N7"}
	// A real node geometry, so the "node" hit classifier's centerOfNode lookup succeeds and
	// actually arms g.DragNode="N7" at pointerdown — needed below to prove Reset clears it
	// even though this press never becomes a drag.
	md.MR.NodeGeoms()["N7"] = nodeactor.NewNodeGeometry("N7", nodegeom.NodeGeom{NodeIdentity: nodegeom.NodeIdentity{Kind: "TimeEnd"}}, nil, clock.NewRealClock())

	down := rawEvent("pointerdown", 400, 300)
	down.Hit = inputcodec.RawHit{Kind: "node", NodeRow: 0}
	md.HandleRawInput(down, nil, nil)
	up := rawEvent("pointerup", 400, 300)
	up.Hit = inputcodec.RawHit{Kind: "node", NodeRow: 0}
	md.HandleRawInput(up, nil, nil)

	if md.UI.Sel.Selected != "N7" {
		t.Fatalf("selected=%q want N7 after press+release with no move", md.UI.Sel.Selected)
	}
	if md.UI.Gest.Phase != gesturefsm.GestIdle {
		t.Fatalf("after click phase=%v want idle", md.UI.Gest.Phase)
	}
	// The pointerdown's "node" hit classifier armed DragNode="N7" even though this press
	// never crossed the slop into a drag; Reset (called from gestPointerUp) must clear it
	// back to "" along with the phase, or a later reader of DragNode (e.g. the Overlay
	// block's DragNodeRow column) would keep reporting a stale drag target.
	if md.UI.Gest.DragNode != "" {
		t.Fatalf("after click DragNode=%q want \"\" (Reset must clear it)", md.UI.Gest.DragNode)
	}
}

// A secondary (two-finger) press with movement still stays pending and tap-selects on
// release — it never converts to a drag/rotate no matter how much it moves, unlike the
// primary-button case pinned above. Regression guard for the movement-commits-a-drag change:
// the commit guard is `dist > 0 && !g.Secondary`, so the secondary check must still gate it.
func TestGestureSecondaryMoveStaysPendingAndTapSelects(t *testing.T) {
	md := newGestureMD(canonicalViewpoint())
	md.RT.NodeRowTable = []string{"N7"}

	down := rawEvent("pointerdown", 400, 300)
	down.Button = 2
	down.Hit = inputcodec.RawHit{Kind: "node", NodeRow: 0}
	md.HandleRawInput(down, nil, nil)

	move := rawEvent("pointermove", 401, 300) // 1px is enough to commit a PRIMARY press
	move.Button = 2
	move.Hit = inputcodec.RawHit{Kind: "node", NodeRow: 0}
	md.HandleRawInput(move, nil, nil)
	if md.UI.Gest.Phase != gesturefsm.GestPending {
		t.Fatalf("secondary press converted out of pending on move: phase=%v", md.UI.Gest.Phase)
	}

	up := rawEvent("pointerup", 401, 300)
	up.Button = 2
	up.Hit = inputcodec.RawHit{Kind: "node", NodeRow: 0}
	md.HandleRawInput(up, nil, nil)
	if md.UI.Sel.Selected != "N7" {
		t.Fatalf("selected=%q want N7 after secondary tap-select", md.UI.Sel.Selected)
	}
}

// A node click sets md.UI.Sel.Selected regardless of the selSpherePoles overlay state (the
// rule-builder authoring path that used to intercept it under selSpherePoles has been
// removed; click-select is now uniform).
func TestGestureSelectModeOffStillHighlights(t *testing.T) {
	md := newGestureMD(canonicalViewpoint())
	md.UI.OV.SelSpherePolesVisible = false
	md.RT.NodeRowTable = []string{"A"}

	down := rawEvent("pointerdown", 400, 300)
	down.Hit = inputcodec.RawHit{Kind: "node", NodeRow: 0}
	md.HandleRawInput(down, nil, nil)
	up := rawEvent("pointerup", 401, 300)
	up.Hit = inputcodec.RawHit{Kind: "node", NodeRow: 0}
	md.HandleRawInput(up, nil, nil)

	if md.UI.Sel.Selected != "A" {
		t.Fatalf("selected=%q after node click with select mode OFF, want A", md.UI.Sel.Selected)
	}
}

// A press-release with NO pointermove in between stays in pending (no move event ever
// arrives to evaluate the commit guard) and resolves as a click (recognized only); it must
// NOT change the camera pose.
func TestGestureClickNoCameraChange(t *testing.T) {
	md := newGestureMD(canonicalViewpoint())
	before := md.UI.VP.Viewpoint
	nodeHit := rawEvent("pointerdown", 400, 300)
	nodeHit.Hit = inputcodec.RawHit{Kind: "empty"}
	md.HandleRawInput(nodeHit, nil, nil)
	md.HandleRawInput(rawEvent("pointerup", 402, 301), nil, nil) // no move event → click
	if md.UI.VP.Viewpoint != before {
		t.Fatalf("click changed camera: %+v != %+v", md.UI.VP.Viewpoint, before)
	}
	if md.UI.Gest.Phase != gesturefsm.GestIdle {
		t.Fatalf("after click phase=%v want idle", md.UI.Gest.Phase)
	}
}

// A single pixel of movement past the press point now commits to dragging — the "click vs.
// drag = click-with-no-movement vs. drag" discriminator has no distance floor. Before this
// change (dist > gestureMoveSlopPx == 6px), this exact 1px move would have stayed
// gesturefsm.GestPending; this pins that it no longer does.
func TestGestureOnePixelMoveCommitsToDrag(t *testing.T) {
	md := dragOffsetMD() // real nodeMover for "n" so the dragNode commit guard's centerOfNode succeeds

	down := rawEvent("pointerdown", 400, 300)
	down.Hit = inputcodec.RawHit{Kind: "node", NodeRow: 0}
	md.HandleRawInput(down, nil, nil)
	if md.UI.Gest.Phase != gesturefsm.GestPending {
		t.Fatalf("after pointerdown: phase=%v want pending", md.UI.Gest.Phase)
	}

	move := rawEvent("pointermove", 401, 300) // 1px displacement, well under the old 6px slop
	move.Hit = inputcodec.RawHit{Kind: "node", NodeRow: 0}
	md.HandleRawInput(move, nil, nil)
	if md.UI.Gest.Phase != gesturefsm.GestDragging {
		t.Fatalf("after 1px move: phase=%v want dragging (movement itself commits)", md.UI.Gest.Phase)
	}
}

// A pointermove event reporting the SAME coordinates as the press is not movement — some
// input stacks emit a move at the press point — and must NOT commit. The guard is on actual
// displacement (dist > 0), not "a move event arrived".
func TestGestureMoveAtPressPointDoesNotCommit(t *testing.T) {
	md := newGestureMD(canonicalViewpoint())
	md.RT.NodeRowTable = []string{"n"}

	down := rawEvent("pointerdown", 400, 300)
	down.Hit = inputcodec.RawHit{Kind: "node", NodeRow: 0}
	md.HandleRawInput(down, nil, nil)

	same := rawEvent("pointermove", 400, 300) // identical to the press point → zero displacement
	same.Hit = inputcodec.RawHit{Kind: "node", NodeRow: 0}
	md.HandleRawInput(same, nil, nil)
	if md.UI.Gest.Phase != gesturefsm.GestPending {
		t.Fatalf("after zero-displacement move: phase=%v want still pending (no movement occurred)", md.UI.Gest.Phase)
	}
}
