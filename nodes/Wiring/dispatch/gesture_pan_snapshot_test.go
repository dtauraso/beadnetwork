package dispatch

import (
	"context"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
)

func TestUnseededViewpointPanIsDegenerate(t *testing.T) {
	// Zero viewpoint: pos = up = {Theta:0, Phi:0} → both map to +Y → parallel.
	md := newGestureMD(geom.Viewpoint{})
	ev := rawEvent("wheel", 400, 300)
	ev.DeltaX = 40
	md.HandleRawInput(context.Background(), ev, nil, nil)

	posW := geom.AnglesToWorldOffset(1, md.UI.VP.Pos.Theta, md.UI.VP.Pos.Phi)
	upW := geom.AnglesToWorldOffset(1, md.UI.VP.Up.Theta, md.UI.VP.Up.Phi)
	// Degenerate: |pos × up| ≈ 0 (parallel), the collapsed-basis condition.
	if cross := posW.Cross(upW).Length(); cross > 1e-9 {
		t.Fatalf("expected degenerate (parallel pos/up) from zero viewpoint, |pos×up|=%v", cross)
	}

	// A seeded viewpoint keeps a valid (non-degenerate) basis after the same pan.
	md2 := newGestureMD(canonicalViewpoint())
	md2.HandleRawInput(context.Background(), ev, nil, nil)
	posW2 := geom.AnglesToWorldOffset(1, md2.UI.VP.Pos.Theta, md2.UI.VP.Pos.Phi)
	upW2 := geom.AnglesToWorldOffset(1, md2.UI.VP.Up.Theta, md2.UI.VP.Up.Phi)
	if cross := posW2.Cross(upW2).Length(); cross < 1e-6 {
		t.Fatalf("seeded viewpoint should keep a valid basis, but |pos×up|=%v", cross)
	}
}
