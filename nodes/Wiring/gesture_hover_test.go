package Wiring

import (
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
)

// gesture_hover_test.go — hover tracking through the gesture FSM.

// Hover is Go-owned: a pointer-move over a node's TORUS ring records it as the hovered node
// (the concentric hover ring emphasizes the ring handle, so it lights only on a torus hit, not
// a body hit); a move over empty space — or over the node BODY — clears hover. There is no
// port hover any more (docs/bead-model/channels-not-ports.md — a port is never drawn or hit-testable).
// Drives moves and asserts md.UI.Sel.HoverNode tracks the hit.
func TestGestureHoverTracksNode(t *testing.T) {
	md := newGestureMD(canonicalViewpoint())
	md.RT.NodeRowTable = []string{"N7"}

	// Move over node N7's torus ring → hovered node.
	mv := rawEvent("pointermove", 400, 300)
	mv.Hit = inputcodec.RawHit{Kind: "torus", NodeRow: 0}
	md.HandleRawInput(mv, nil, nil)
	if md.UI.Sel.HoverNode != "N7" || md.UI.Sel.HoverPort != "" {
		t.Fatalf("torus hover: hoverNode=%q hoverPort=%q want N7,''", md.UI.Sel.HoverNode, md.UI.Sel.HoverPort)
	}

	// Move over the node BODY (kind "node") → hover clears (body does not light the ring).
	bodyMv := rawEvent("pointermove", 402, 300)
	bodyMv.Hit = inputcodec.RawHit{Kind: "node", NodeRow: 0}
	md.HandleRawInput(bodyMv, nil, nil)
	if md.UI.Sel.HoverNode != "" || md.UI.Sel.HoverPort != "" {
		t.Fatalf("body hover: hoverNode=%q hoverPort=%q want '',''", md.UI.Sel.HoverNode, md.UI.Sel.HoverPort)
	}

	// Move over empty space → hover cleared.
	md.HandleRawInput(rawEvent("pointermove", 500, 300), nil, nil)
	if md.UI.Sel.HoverNode != "" || md.UI.Sel.HoverPort != "" {
		t.Fatalf("empty hover: hoverNode=%q hoverPort=%q want '',''", md.UI.Sel.HoverNode, md.UI.Sel.HoverPort)
	}
}
