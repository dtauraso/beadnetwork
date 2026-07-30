package Wiring

import (
	"math"
	"testing"
)

// refPortWorldPos is an independent reimplementation of the portWorldPos algorithm,
// used to lock the production code's output. The node center is provided directly.
func refPortWorldPos(kind string, center vec3, ports []portGeom, name string, _ bool) vec3 {
	if name == "" {
		return center
	}
	idx := -1
	for i, p := range ports {
		if p.Name == name {
			idx = i
			break
		}
	}
	if idx < 0 {
		return center
	}
	port := ports[idx]
	anchorIdx := 0
	if port.AnchorId != nil {
		anchorIdx = *port.AnchorId
	}
	R := nodeRadius(kind)
	dir := ringAnchorDir(R, anchorIdx)
	return vec3{X: center.X + dir.X*R, Y: center.Y + dir.Y*R, Z: center.Z + dir.Z*R}
}

func almostEqual(a, b, eps float64) bool { return math.Abs(a-b) <= eps }

func TestPortWorldPosMirrorsReference(t *testing.T) {
	anchorId1 := 1
	center := vec3{X: 46.5425, Y: 93.085, Z: -139.6275}
	g := nodeGeom{
		nodeIdentity: nodeIdentity{Kind: "HoldFlip"},
		HasPos:       true, ScenePolar: cart2polar(center),
		Inputs: []portGeom{
			{Name: "In", AnchorId: &anchorId1},
			{Name: "In2"},
		},
		Outputs: []portGeom{{Name: "Out"}},
	}
	got := portWorldPos(g, "In", true)
	want := refPortWorldPos(g.Kind, center, g.Inputs, "In", true)
	if !almostEqual(got.X, want.X, 1e-9) || !almostEqual(got.Y, want.Y, 1e-9) || !almostEqual(got.Z, want.Z, 1e-9) {
		t.Fatalf("portWorldPos = %+v, want %+v", got, want)
	}
}

// TestNodeTorusOuterR verifies nodeTorusOuterR = nodeRadius(kind) * (1 + ratio),
// the formula chain_beads.go's tangent placement depends on (docs/bead-lattice.md).
func TestNodeTorusOuterR(t *testing.T) {
	for _, kind := range []string{"Input", "Time"} {
		want := nodeRadius(kind) * (1 + ShadingParamNodeRingTubeRatio)
		if got := nodeTorusOuterR(kind); got != want {
			t.Fatalf("nodeTorusOuterR(%q) = %v, want %v", kind, got, want)
		}
	}
}

// TestPortRadiusPerPort verifies that two ports on the same node with DIFFERENT
// PortR values place each port (and its edge endpoint / arc length) at its own
// radius, not the shared nodeRadius(kind) value.
func TestPortRadiusPerPort(t *testing.T) {
	anchorId0 := 0
	r1, r2 := 15.0, 40.0
	g := nodeGeom{
		nodeIdentity: nodeIdentity{Kind: "HoldFlip"},
		Inputs: []portGeom{
			{Name: "InSmall", AnchorId: &anchorId0, PortR: &r1},
			{Name: "InBig", AnchorId: &anchorId0, PortR: &r2},
		},
	}
	center := nodeWorldPos(g)
	dir := ringAnchorDir(nodeRadius(g.Kind), 0)

	gotSmall := portWorldPos(g, "InSmall", true)
	wantSmall := center.Add(dir.Scale(r1))
	if math.Abs(gotSmall.X-wantSmall.X) > 1e-9 || math.Abs(gotSmall.Y-wantSmall.Y) > 1e-9 || math.Abs(gotSmall.Z-wantSmall.Z) > 1e-9 {
		t.Fatalf("portWorldPos(InSmall) = %v, want %v (r=%v)", gotSmall, wantSmall, r1)
	}

	gotBig := portWorldPos(g, "InBig", true)
	wantBig := center.Add(dir.Scale(r2))
	if math.Abs(gotBig.X-wantBig.X) > 1e-9 || math.Abs(gotBig.Y-wantBig.Y) > 1e-9 || math.Abs(gotBig.Z-wantBig.Z) > 1e-9 {
		t.Fatalf("portWorldPos(InBig) = %v, want %v (r=%v)", gotBig, wantBig, r2)
	}

	if got := portRadiusByName(g, "InSmall", true); got != r1 {
		t.Fatalf("portRadiusByName(InSmall) = %v, want %v", got, r1)
	}
	if got := portRadiusByName(g, "InBig", true); got != r2 {
		t.Fatalf("portRadiusByName(InBig) = %v, want %v", got, r2)
	}
	// A port with no PortR falls back to nodeRadius(kind).
	if got := portRadiusByName(g, "NoSuchPort", true); got != nodeRadius(g.Kind) {
		t.Fatalf("portRadiusByName(unknown) = %v, want fallback %v", got, nodeRadius(g.Kind))
	}
}

// TestPortAnchorIdRingPath verifies that AnchorId selects the correct ring slot and
// that a nil AnchorId falls back to ring slot 0.
func TestPortAnchorIdRingPath(t *testing.T) {
	var kind string
	for k := range kindDims {
		kind = k
		break
	}
	if kind == "" {
		t.Skip("no kinds in kindDims")
	}

	anchorId1 := 1
	g := nodeGeom{
		nodeIdentity: nodeIdentity{Kind: kind},
		Inputs:       []portGeom{{Name: "In", AnchorId: &anchorId1}},
		Outputs:      []portGeom{{Name: "Out"}}, // nil AnchorId → ring slot 0
	}

	// Anchored input: direction == ringAnchorDir(R, 1)
	R := nodeRadius(kind)
	dir, ok := portDir(g, "In", true)
	if !ok {
		t.Fatal("portDir(In) not found")
	}
	want := ringAnchorDir(R, 1)
	if math.Abs(dir.X-want.X) > 1e-9 || math.Abs(dir.Y-want.Y) > 1e-9 || math.Abs(dir.Z-want.Z) > 1e-9 {
		t.Fatalf("portDir(In, anchorId=1) = %v, want %v", dir, want)
	}

	// Output with nil AnchorId → ring slot 0
	dir0, ok := portDir(g, "Out", false)
	if !ok {
		t.Fatal("portDir(Out) not found")
	}
	want0 := ringAnchorDir(R, 0)
	if math.Abs(dir0.X-want0.X) > 1e-9 || math.Abs(dir0.Y-want0.Y) > 1e-9 || math.Abs(dir0.Z-want0.Z) > 1e-9 {
		t.Fatalf("portDir(Out, nil anchorId) = %v, want ring[0] %v", dir0, want0)
	}

	// World pos == center + dir*nodeRadius
	center := nodeWorldPos(g)
	wantPos := center.Add(want.Scale(R))
	gotPos := portWorldPos(g, "In", true)
	if math.Abs(gotPos.X-wantPos.X) > 1e-9 || math.Abs(gotPos.Y-wantPos.Y) > 1e-9 || math.Abs(gotPos.Z-wantPos.Z) > 1e-9 {
		t.Fatalf("portWorldPos(In) = %v, want %v", gotPos, wantPos)
	}
}
