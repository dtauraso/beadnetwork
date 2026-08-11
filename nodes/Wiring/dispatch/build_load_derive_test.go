// build_load_derive_test.go — closes the load-path coverage gap recorded in
// docs/planning/movedispatch-decomposition.md ("Coverage gap found, not assumed away"):
// two build phases (computeReachRadii, computeQuantizedLayout — now
// topoderive.ComputeReachRadii/topoderive.ComputeQuantizedLayout) were converted to
// explicit functions and NOTHING in the suite asserted on what they PRODUCE — dropping
// either call entirely passed `go test ./...` unmodified. Each test below was run against
// its own injected bug and confirmed to fail (see the commit history/task report for the
// verbatim failure text); this file is the closed hole, not a guess at one.
//
// The third test this file used to hold, TestAllocateVectorChannelsKeysSourceOutTargetIn,
// moved to nodes/Wiring/topoderive/vector_channels_test.go alongside
// topoderive.AllocateVectorChannels itself — it drove only that pure function plus
// encoding/json, no MoveDispatch/writeSpecTree/LoadTopology harness, so it did not need
// to stay in package Wiring.

package dispatch

import (
	"context"
	"math"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/quantoffset"
	"github.com/dtauraso/wirefold/nodes/wire/clock"

	T "github.com/dtauraso/wirefold/Trace"
)

// TestLoadTopologyComputesReachRadii pins that a load-time source node's own ReachR
// (nodegeom.NodeGeom.ReachR, "max distance to any node it outputs to") is the real
// polar distance to its edge partner, not the zero value a dropped computeReachRadii
// call would leave in place. Fails under a dropped computeReachRadii call: ReachR reads
// 0 instead of the computed distance.
func TestLoadTopologyComputesReachRadii(t *testing.T) {
	const topo = `{
	  "nodes": [
	    {"id":"1","type":"AimedSrc","scenePolarR":0,"scenePolarTheta":0,"scenePolarPhi":0},
	    {"id":"2","type":"AimedSink","scenePolarR":50,"scenePolarTheta":1.5707963267948966,"scenePolarPhi":0}
	  ],
	  "edges": [
	    {"label":"e0","kind":"data","source":"1","sourceHandle":"Out","target":"2","targetHandle":"In"}
	  ]
	}`
	root := writeSpecTree(t, t.TempDir(), topo)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, md, _, err := LoadTopology(ctx, root, T.New(), clock.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology: %v", err)
	}

	src, ok := md.mr.nodeGeoms["1"]
	if !ok {
		t.Fatalf("nodeGeoms missing source node %q", "1")
	}
	srcPolar := geom.Polar{R: 0, Theta: 0, Phi: 0}
	dstPolar := geom.Polar{R: 50, Theta: math.Pi / 2, Phi: 0}
	want := geom.PolarDist(srcPolar, dstPolar)

	if want <= 0 {
		t.Fatalf("test fixture produced a zero expected distance; fixture is degenerate")
	}
	if math.Abs(src.ReachR()-want) > 1e-6 {
		t.Fatalf("node 1 ReachR = %v, want %v (distance to its edge partner)", src.ReachR(), want)
	}
}

// TestLoadTopologyComputesQuantizedOffsets pins that every node's load-time
// nodeGeometry.quantOffset is a real measured offset, not the zero value a dropped
// computeQuantizedLayout call would leave in place: DeriveCenters reconstructs the
// offset back to (approximately) the node's own scene-polar world center. Fails under a
// dropped computeQuantizedLayout call: quantOffset reads the zero QuantizedOffset for
// every node, which derives every node back to the scene center regardless of its real
// position.
func TestLoadTopologyComputesQuantizedOffsets(t *testing.T) {
	const topo = `{
	  "nodes": [
	    {"id":"1","type":"AimedSrc","scenePolarR":0,"scenePolarTheta":0,"scenePolarPhi":0},
	    {"id":"2","type":"AimedSink","scenePolarR":50,"scenePolarTheta":1.5707963267948966,"scenePolarPhi":0}
	  ],
	  "edges": [
	    {"label":"e0","kind":"data","source":"1","sourceHandle":"Out","target":"2","targetHandle":"In"}
	  ]
	}`
	root := writeSpecTree(t, t.TempDir(), topo)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, md, _, err := LoadTopology(ctx, root, T.New(), clock.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology: %v", err)
	}

	dst, ok := md.mr.nodeGeoms["2"]
	if !ok {
		t.Fatalf("nodeGeoms missing node %q", "2")
	}
	// No sphere.json was written, so the scene center is the zero vector — the same
	// center computeQuantizedLayout measured node 2's offset about.
	wantCenter := geom.Polar2cart(geom.Polar{R: 50, Theta: math.Pi / 2, Phi: 0})
	got := quantoffset.DeriveCenters(map[string]quantoffset.QuantizedOffset{"2": dst.QuantizedOffsetValue()}, vec3{})["2"]

	// Quantization introduces lattice rounding error (1-degree angular cells, one bead
	// of radial cell) — a generous tolerance that still fails hard on a zero offset,
	// whose reconstructed center is (0,0,0), ~50 world units from wantCenter.
	const tol = 10.0
	if d := got.Sub(wantCenter).Length(); d > tol {
		t.Fatalf("node 2 quantOffset derives to %+v, want close to %+v (scene-polar center); off by %v > tol %v", got, wantCenter, d, tol)
	}
}
