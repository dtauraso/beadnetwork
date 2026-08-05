// Unit tests for Buffer/buffer_layout_gen.go — typed writer round-trips.
//
// Each test writes known values via the generated Set*Row helpers and asserts
// the correct bytes appear at the expected offsets in a raw []byte buffer.
// This is the trust frontier of the schema: it exercises column offset math
// (offset, stride, endianness) for every block.

package Buffer

import (
	"encoding/binary"
	"math"
	"testing"
)

// assertF32At asserts that buf[offset:offset+4] equals the LE encoding of want.
func assertF32At(t *testing.T, buf []byte, offset int, want float32, label string) {
	t.Helper()
	got := math.Float32frombits(binary.LittleEndian.Uint32(buf[offset:]))
	if got != want {
		t.Errorf("%s: got %v, want %v (at byte offset %d)", label, got, want, offset)
	}
}

// assertU32At asserts that buf[offset:offset+4] equals the LE encoding of want.
func assertU32At(t *testing.T, buf []byte, offset int, want uint32, label string) {
	t.Helper()
	got := binary.LittleEndian.Uint32(buf[offset:])
	if got != want {
		t.Errorf("%s: got %v, want %v (at byte offset %d)", label, got, want, offset)
	}
}

// assertI32At asserts that buf[offset:offset+4] equals the LE encoding of want.
func assertI32At(t *testing.T, buf []byte, offset int, want int32, label string) {
	t.Helper()
	got := int32(binary.LittleEndian.Uint32(buf[offset:]))
	if got != want {
		t.Errorf("%s: got %v, want %v (at byte offset %d)", label, got, want, offset)
	}
}

// assertU8At asserts that buf[offset] equals want.
func assertU8At(t *testing.T, buf []byte, offset int, want uint8, label string) {
	t.Helper()
	if buf[offset] != want {
		t.Errorf("%s: got %v, want %v (at byte offset %d)", label, buf[offset], want, offset)
	}
}

func TestSetNodeRow(t *testing.T) {
	buf := make([]byte, BufNodeStride)
	SetNodeRow(buf, 0,
		42,            // nodeId
		1.0, 2.0, 3.0, // cx, cy, cz
		0.5, 0.25, // radius, sphereR
		0.1, 0.2, 0.3, // vrx, vry, vrz
		0.4, 0.5, 0.6, // frx, fry, frz
		0.7, 0.8, // poleTheta, polePhi
		0.9, 1.1, // ringAxisTheta, ringAxisPhi — the DRAWN ring's axis, separate from the
		//           navigation pole above (Buffer/layout.go)
		1,    // selected
		3,    // kindID (Input = index 3 in NODE_DEFS_ARRAY)
		7, 4, // labelOff, labelLen
		1, // hovered
		1, // latchedSel
	)

	assertI32At(t, buf, BufNodeColNodeId, 42, "NodeId")
	assertF32At(t, buf, BufNodeColCX, 1.0, "CX")
	assertF32At(t, buf, BufNodeColCY, 2.0, "CY")
	assertF32At(t, buf, BufNodeColCZ, 3.0, "CZ")
	assertF32At(t, buf, BufNodeColRadius, 0.5, "Radius")
	assertF32At(t, buf, BufNodeColSphereR, 0.25, "SphereR")
	assertF32At(t, buf, BufNodeColVRX, 0.1, "VRX")
	assertF32At(t, buf, BufNodeColVRY, 0.2, "VRY")
	assertF32At(t, buf, BufNodeColVRZ, 0.3, "VRZ")
	assertF32At(t, buf, BufNodeColFRX, 0.4, "FRX")
	assertF32At(t, buf, BufNodeColFRY, 0.5, "FRY")
	assertF32At(t, buf, BufNodeColFRZ, 0.6, "FRZ")
	assertF32At(t, buf, BufNodeColPoleTheta, 0.7, "PoleTheta")
	assertF32At(t, buf, BufNodeColPolePhi, 0.8, "PolePhi")
	assertF32At(t, buf, BufNodeColRingAxisTheta, 0.9, "RingAxisTheta")
	assertF32At(t, buf, BufNodeColRingAxisPhi, 1.1, "RingAxisPhi")
	assertU8At(t, buf, BufNodeColSelected, 1, "Selected")
	assertU8At(t, buf, BufNodeColKindId, 3, "KindId")
	assertU32At(t, buf, BufNodeColLabelOff, 7, "LabelOff")
	assertU32At(t, buf, BufNodeColLabelLen, 4, "LabelLen")
	assertU8At(t, buf, BufNodeColHovered, 1, "Hovered")
	assertU8At(t, buf, BufNodeColLatchedSel, 1, "LatchedSel")
}

func TestSetEdgeRow(t *testing.T) {
	// Edge SX..EZ are the edge's own SEGMENT endpoints (docs/channels-not-ports.md — a
	// port has no row/geometry of its own any more, so the edge carries its own
	// node-surface-to-node-surface segment directly instead of referencing a Port block).
	buf := make([]byte, BufEdgeStride*2)
	SetEdgeRow(buf, 0, 1, 2, 3, 4, 5, 6, 1, 11, 22)
	SetEdgeRow(buf, 1, 7, 8, 9, 10, 11, 12, 0, 0, 0)

	assertF32At(t, buf, BufEdgeColSX, 1, "row0.SX")
	assertF32At(t, buf, BufEdgeColSY, 2, "row0.SY")
	assertF32At(t, buf, BufEdgeColSZ, 3, "row0.SZ")
	assertF32At(t, buf, BufEdgeColEX, 4, "row0.EX")
	assertF32At(t, buf, BufEdgeColEY, 5, "row0.EY")
	assertF32At(t, buf, BufEdgeColEZ, 6, "row0.EZ")
	assertU8At(t, buf, BufEdgeColSelected, 1, "row0.Selected")
	assertU32At(t, buf, BufEdgeColEdgeLabelOff, 11, "row0.EdgeLabelOff")
	assertU32At(t, buf, BufEdgeColEdgeLabelLen, 22, "row0.EdgeLabelLen")

	base := BufEdgeStride
	assertF32At(t, buf, base+BufEdgeColSX, 7, "row1.SX")
	assertF32At(t, buf, base+BufEdgeColEZ, 12, "row1.EZ")
}

func TestSetCameraRow(t *testing.T) {
	buf := make([]byte, BufCameraStride)
	SetCameraRow(buf, 1.0, 2.0, 3.0, 10.0, 0.5, 1.0, 0.25, 0.75)

	assertF32At(t, buf, BufCameraColPX, 1.0, "PX")
	assertF32At(t, buf, BufCameraColPY, 2.0, "PY")
	assertF32At(t, buf, BufCameraColPZ, 3.0, "PZ")
	assertF32At(t, buf, BufCameraColR, 10.0, "R")
	assertF32At(t, buf, BufCameraColPosTheta, 0.5, "PosTheta")
	assertF32At(t, buf, BufCameraColPosPhi, 1.0, "PosPhi")
	assertF32At(t, buf, BufCameraColUpTheta, 0.25, "UpTheta")
	assertF32At(t, buf, BufCameraColUpPhi, 0.75, "UpPhi")
}

func TestSetOverlayRow(t *testing.T) {
	buf := make([]byte, BufOverlayStride)
	SetOverlayRow(buf, OverlayRow{
		SceneTori:      1,
		ScenePoles:     0,
		NodePoles:      1,
		SelSpherePoles: 0,
		Handholds:      1,
		LabelsGlobal:   0,
		OverlaysVis:    0,
	})

	assertU8At(t, buf, BufOverlayColSceneTori, 1, "SceneTori")
	assertU8At(t, buf, BufOverlayColScenePoles, 0, "ScenePoles")
	assertU8At(t, buf, BufOverlayColNodePoles, 1, "NodePoles")
	assertU8At(t, buf, BufOverlayColSelSpherePoles, 0, "SelSpherePoles")
	assertU8At(t, buf, BufOverlayColHandholds, 1, "Handholds")
	assertU8At(t, buf, BufOverlayColLabelsGlobal, 0, "LabelsGlobal")
	assertU8At(t, buf, BufOverlayColOverlaysVis, 0, "OverlaysVis")
}

func TestNodeStrideIsPackedSize(t *testing.T) {
	// Node block: 1×i32 (nodeId) + 5×f32 + 6×f32 (vr/fr normals) + 2×f32 (poleTheta/polePhi)
	//           + 2×f32 (ringAxisTheta/ringAxisPhi — the DRAWN ring's axis, separate from the
	//             navigation pole)
	//           + 1×u8 (selected) + 1×u8 (kindID) + 2×u32 (label off/len) + 1×u8 (hovered) + 1×u8 (latchedSel)
	//           = 4 + (5+6+2+2)×4 + 1 + 1 + 8 + 1 + 1 = 76
	want := 1*4 + 5*4 + 6*4 + 2*4 + 2*4 + 1*1 + 1*1 + 2*4 + 1*1 + 1*1
	if BufNodeStride != want {
		t.Errorf("BufNodeStride = %d, want %d (packed size)", BufNodeStride, want)
	}
}

func TestEdgeStrideIsPackedSize(t *testing.T) {
	// Edge block: 6×f32 (SX..EZ, the edge's own node-surface-to-node-surface segment —
	// docs/channels-not-ports.md, there is no port row to reference any more) + 1×u8
	// (selected) + 2×u32 (edge-label off/len) = 33.
	want := 6*4 + 1 + 2*4
	if BufEdgeStride != want {
		t.Errorf("BufEdgeStride = %d, want %d (packed size)", BufEdgeStride, want)
	}
}

func TestCameraStrideIsPackedSize(t *testing.T) {
	// Camera block: 8×f32 = 32
	want := 8 * 4
	if BufCameraStride != want {
		t.Errorf("BufCameraStride = %d, want %d (packed size)", BufCameraStride, want)
	}
}

func TestOverlayStrideIsPackedSize(t *testing.T) {
	// Overlay block: 7×u8 + 1×i32 + 3×f32 = 23 (7 overlay flags — the 7 render gates —
	// DragNodeRow, and the "distance home button" panel's 3 GroupLen* columns)
	want := 7*1 + 1*4 + 3*4
	if BufOverlayStride != want {
		t.Errorf("BufOverlayStride = %d, want %d (packed size)", BufOverlayStride, want)
	}
}

func TestVersionGenerated(t *testing.T) {
	if BufLayoutVersionGenerated != BufLayoutVersion {
		t.Errorf("BufLayoutVersionGenerated (%d) != BufLayoutVersion (%d) — regenerate", BufLayoutVersionGenerated, BufLayoutVersion)
	}
}
