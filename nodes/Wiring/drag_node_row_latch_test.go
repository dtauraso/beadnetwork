// drag_node_row_latch_test.go — proves the in-editor "dragging <name>" line's Go-side
// source, Overlay.DragNodeRow, LATCHES to the last-dragged node instead of clearing to
// -1 on pointerup: it stays at the just-dragged node's row after that drag ends, and
// only moves when a DIFFERENT drag starts (uiState.lastDraggedNode, set at
// commitDragStart, gesture_graph.go; derived in emitViewFrame, view_stream.go).

package Wiring

import (
	"context"
	"encoding/binary"
	"testing"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	B "github.com/dtauraso/wirefold/Buffer"
	T "github.com/dtauraso/wirefold/Trace"
)

// lastViewStreamDragNodeRow decodes the LAST complete framed view-stream payload's
// Overlay.DragNodeRow column ([len:u32][payload] framing, no tag byte — mirrors
// lastNodeStreamDragMsg/lastNodeStreamSelectedHovered in ui_publish_propagation_test.go).
func lastViewStreamDragNodeRow(raw []byte) (row int32, ok bool) {
	off := 0
	var last []byte
	for off+4 <= len(raw) {
		n := int(binary.LittleEndian.Uint32(raw[off:]))
		off += 4
		if off+n > len(raw) {
			break
		}
		last = raw[off : off+n]
		off += n
	}
	const overlayOff = B.BufViewFrameHeaderSize + B.BufCameraStride
	if last == nil || len(last) < overlayOff+B.BufOverlayStride {
		return 0, false
	}
	overlay := last[overlayOff : overlayOff+B.BufOverlayStride]
	return int32(binary.LittleEndian.Uint32(overlay[B.BufOverlayColDragNodeRow:])), true
}

// TestDragNodeRowLatchesAcrossPointerupThenSwitches drives two real drags (x, then t)
// through HandleRawInput and asserts: (1) DragNodeRow starts at -1 (nothing ever
// dragged), (2) it resolves to x's row while x is being dragged AND stays at x's row
// after pointerup (does not clear to -1), and (3) starting a NEW drag on t switches it
// to t's row.
func TestDragNodeRowLatchesAcrossPointerupThenSwitches(t *testing.T) {
	root := writeXTN(t) // x --Out--> t (chain), x --Out--> n (data)
	tr := T.NewWithSinkHook(nil, nil)

	_, _, md, _, err := LoadTopology(context.Background(), root, tr, wire.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology: %v", err)
	}

	viewBuf := &uiPubLockedBuf{}
	md.SetViewStream(viewBuf, func(tick uint32,
		camPX, camPY, camPZ, camR, camPosTheta, camPosPhi, camUpTheta, camUpPhi float32,
		sceneTori, scenePoles, nodePoles, selSpherePoles, handholds, labelsGlobal, overlaysVis, doubleLinks uint8,
		dragNodeRow int32,
		sceneCX, sceneCY, sceneCZ, sceneRadius float32,
		events []wire.RowEvent,
	) []byte {
		return B.BuildViewStreamFrame(tick, camPX, camPY, camPZ, camR, camPosTheta, camPosPhi, camUpTheta, camUpPhi,
			B.OverlayRow{
				SceneTori: sceneTori, ScenePoles: scenePoles, NodePoles: nodePoles,
				SelSpherePoles: selSpherePoles, Handholds: handholds, LabelsGlobal: labelsGlobal,
				OverlaysVis: overlaysVis, DoubleLinks: doubleLinks,
				DragNodeRow: dragNodeRow,
			},
			sceneCX, sceneCY, sceneCZ, sceneRadius, nil)
	})

	xRow, ok := md.NodeRowFor("x")
	if !ok {
		t.Fatal("no NODE-ROW for x")
	}
	tRow, ok := md.NodeRowFor("t")
	if !ok {
		t.Fatal("no NODE-ROW for t")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	md.Start(ctx)

	rectFov := func(ev rawInputMsg) rawInputMsg {
		ev.RectLeft, ev.RectTop, ev.RectWidth, ev.RectHeight = 0, 0, 800, 600
		ev.Fov = 50
		return ev
	}

	if row, ok := lastViewStreamDragNodeRow(viewBuf.Bytes()); ok && row != -1 {
		t.Fatalf("DragNodeRow before any drag = %d, want -1", row)
	}

	// --- Drag 1: x, past the slop threshold to commit gestDragging, then release. ---
	md.HandleRawInput(rectFov(rawInputMsg{Kind: "pointerdown", X: 400, Y: 300, Button: 0,
		Hit: rawHit{Kind: "node", NodeRow: int(xRow), PortRow: -1, EdgeRow: -1}}), nil, tr)
	for i, dx := range []float64{20, 40, 60, 80, 100} {
		md.HandleRawInput(rectFov(rawInputMsg{Kind: "pointermove", X: 400 + dx, Y: 300 + float64(i)*5}), nil, tr)
	}
	if md.ui.gest.phase != gestDragging {
		t.Fatalf("phase after x's moves = %v, want gestDragging", md.ui.gest.phase)
	}
	if row, ok := lastViewStreamDragNodeRow(viewBuf.Bytes()); !ok || row != xRow {
		t.Fatalf("DragNodeRow while dragging x = %v (ok=%v), want %d", row, ok, xRow)
	}
	md.HandleRawInput(rectFov(rawInputMsg{Kind: "pointerup", X: 500, Y: 325, Button: 0}), nil, tr)
	if md.ui.gest.dragNode != "" {
		t.Fatalf("gest.dragNode after pointerup = %q, want \"\" (live field must still clear)", md.ui.gest.dragNode)
	}
	if row, ok := lastViewStreamDragNodeRow(viewBuf.Bytes()); !ok || row != xRow {
		t.Fatalf("DragNodeRow after x's pointerup = %v (ok=%v), want it to STAY at x's row %d (latched), not clear", row, ok, xRow)
	}

	// --- Drag 2: t, a DIFFERENT node — DragNodeRow must switch to t's row. ---
	md.HandleRawInput(rectFov(rawInputMsg{Kind: "pointerdown", X: 400, Y: 300, Button: 0,
		Hit: rawHit{Kind: "node", NodeRow: int(tRow), PortRow: -1, EdgeRow: -1}}), nil, tr)
	for i, dx := range []float64{20, 40, 60, 80, 100} {
		md.HandleRawInput(rectFov(rawInputMsg{Kind: "pointermove", X: 400 + dx, Y: 300 + float64(i)*5}), nil, tr)
	}
	if md.ui.gest.phase != gestDragging {
		t.Fatalf("phase after t's moves = %v, want gestDragging", md.ui.gest.phase)
	}
	if row, ok := lastViewStreamDragNodeRow(viewBuf.Bytes()); !ok || row != tRow {
		t.Fatalf("DragNodeRow while dragging t = %v (ok=%v), want %d (switched off x's latch)", row, ok, tRow)
	}
	md.HandleRawInput(rectFov(rawInputMsg{Kind: "pointerup", X: 500, Y: 325, Button: 0}), nil, tr)
	if row, ok := lastViewStreamDragNodeRow(viewBuf.Bytes()); !ok || row != tRow {
		t.Fatalf("DragNodeRow after t's pointerup = %v (ok=%v), want it to STAY at t's row %d", row, ok, tRow)
	}
}
