package Wiring

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"testing"
)

// gesture_drag_offset_test.go — pins the grab-offset fix: dragging a node from a point OFF
// its center must not recenter the node under the cursor on the very first move past the
// slop. Before the fix, applyNodeDragTarget called RootMove(g.dragNode, hit) directly, so
// the node's CENTER jumped to wherever the pointer's plane-hit landed — reported live as
// "if I'm not clicking on the center of the node it moves itself so the mouse is in the
// center of it." The fix captures the grab offset once at commitDragStart (the ONE place a
// drag begins) and reapplies it on every move (dragPlaneHit + dragGrabOffset, gesture.go /
// gesture_actions.go / gesture_graph.go).

// dragOffsetMD builds a MoveDispatch with a single node "n" centered at the origin, camera
// looking straight down -Z at it (canonicalViewpoint), so a pointer-down/move at screen
// center (400,300) projects EXACTLY onto the node's center — the deterministic "center
// grab" case the companion test below relies on.
func dragOffsetMD() *MoveDispatch {
	md := &MoveDispatch{mr: moverRegistry{nodeMovers: map[string]*nodeMover{}, edgeMovers: map[string]*edgeMover{}}}
	md.ui.vp.viewpoint = canonicalViewpoint()
	g := nodeGeom{nodeIdentity: nodeIdentity{Kind: "TimeEnd"}}
	setNodeWorld(&g, vec3{X: 0, Y: 0, Z: 0})
	nm := newNodeMover("n", g, nil, wire.NewRealClock())
	md.mr.nodeMovers["n"] = nm
	// No goroutine started (mirrors gesture_home_test's homeMD): extIn is a buffered
	// channel (moverInboxDepth), so sendMove's writes land there for the test to drain
	// without a live mover loop committing them.
	md.rt.nodeRowTable = []string{"n"}
	return md
}

func nodeHit() rawHit { return rawHit{Kind: "node", NodeRow: 0} }

// drainDrag reads off nm.extIn until it sees a moveMsgKindDrag, returning its Target. Fails
// the test if none arrives (commitDragStart's DragStart message precedes it on the same
// channel for the first move of a drag).
func drainDrag(t *testing.T, nm *nodeMover) vec3 {
	t.Helper()
	for {
		select {
		case msg := <-nm.extIn:
			if msg.Kind == moveMsgKindDrag {
				return msg.Target
			}
		default:
			t.Fatal("no moveMsgKindDrag arrived on extIn")
			return vec3{}
		}
	}
}

// TestGestureDragOffCenterPreservesGrabPoint grabs the node well off its center and
// verifies the FIRST post-slop move places the node back at (approximately) its own
// pre-drag center — the grabbed point stays under the cursor instead of the center jumping
// to the plane-hit. It then drags further and checks the SAME offset is preserved on a
// second, genuine move.
func TestGestureDragOffCenterPreservesGrabPoint(t *testing.T) {
	md := dragOffsetMD()
	nm := md.mr.nodeMovers["n"]

	down := rawEvent("pointerdown", 450, 300) // OFF the node's screen-center projection
	down.Hit = nodeHit()
	md.HandleRawInput(down, nil, nil)
	if md.ui.gest.dragNode != "n" {
		t.Fatalf("pointerdown on node did not arm dragNode: got %q", md.ui.gest.dragNode)
	}

	hit1, ok := md.dragPlaneHit(rawEvent("pointermove", 480, 300))
	if !ok {
		t.Fatal("dragPlaneHit(ev1) reported not-ok; test setup assumption (non-parallel ray) broken")
	}

	move1 := rawEvent("pointermove", 480, 300) // any displacement from (450,300) commits
	move1.Hit = nodeHit()
	md.HandleRawInput(move1, nil, nil)
	if md.ui.gest.phase != gestDragging {
		t.Fatalf("after slop-cross move: phase=%v want dragging", md.ui.gest.phase)
	}

	// commitDragStart captured offset = dragStartCenter - hit1 = (0,0,0) - hit1 = -hit1.
	// applyNodeDragTarget's own hit for the SAME event is also hit1 (same ev, same plane),
	// so the very first target is hit1 + offset == (0,0,0): the node stays put at
	// engagement instead of jumping to hit1 under the (off-center) cursor.
	target1 := drainDrag(t, nm)
	if !vecClose(target1, vec3{X: 0, Y: 0, Z: 0}, 1e-9) {
		t.Fatalf("first drag target=%v want node's own center (0,0,0) — grab offset was not preserved (recentered on cursor at %v)", target1, hit1)
	}

	// A second, genuine move: target must track hit2 PLUS the same offset, not hit2 alone.
	move2 := rawEvent("pointermove", 520, 340)
	md.HandleRawInput(move2, nil, nil)
	hit2, ok := md.dragPlaneHit(move2)
	if !ok {
		t.Fatal("dragPlaneHit(ev2) reported not-ok; test setup assumption broken")
	}
	wantTarget2 := hit2.Add(vec3{X: 0, Y: 0, Z: 0}.Sub(hit1)) // hit2 + (dragStartCenter - hit1)
	target2 := drainDrag(t, nm)
	if !vecClose(target2, wantTarget2, 1e-9) {
		t.Fatalf("second drag target=%v want %v (hit2 + grab offset)", target2, wantTarget2)
	}
	if vecClose(target2, hit2, 1e-6) {
		t.Fatalf("second drag target=%v equals the raw plane-hit %v — grab offset was dropped", target2, hit2)
	}
}

// TestGestureDragCenterGrabUnchanged is the companion: the offset is captured from the
// COMMITTING move event (the one that crosses the slop — see commitDragStart), so "center
// grab" here means that event's plane-hit lands EXACTLY on the node's center, yielding a
// zero offset and drag targets equal to the raw plane-hit, same as before the fix.
func TestGestureDragCenterGrabUnchanged(t *testing.T) {
	md := dragOffsetMD()
	nm := md.mr.nodeMovers["n"]

	down := rawEvent("pointerdown", 410, 300) // off-center; only its distance to move1 matters for the slop check
	down.Hit = nodeHit()
	md.HandleRawInput(down, nil, nil)

	move1 := rawEvent("pointermove", 400, 300) // screen center → projects exactly onto (0,0,0)
	move1.Hit = nodeHit()
	md.HandleRawInput(move1, nil, nil)
	if md.ui.gest.phase != gestDragging {
		t.Fatalf("after slop-cross move: phase=%v want dragging", md.ui.gest.phase)
	}
	hit1, ok := md.dragPlaneHit(move1)
	if !ok {
		t.Fatal("dragPlaneHit(move1) reported not-ok")
	}
	target1 := drainDrag(t, nm)
	if !vecClose(target1, hit1, 1e-9) {
		t.Fatalf("center-grab first target=%v want raw hit %v (zero offset)", target1, hit1)
	}
}
