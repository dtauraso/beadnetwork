// build_wires.go — the wire-allocation phase of buildFromSpec: one *PacedWire per
// destination port, with each edge's own initial bead-step count and straight-segment
// endpoints computed alongside it.

package dispatch

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

// allocateWires allocates one *PacedWire per destination port (one edge per port —
// fan-in is rejected at parse) and computes each edge's own INITIAL bead-step count
// (docs/bead-model/bead-lattice.md "The count") and straight-segment endpoints.
//   - destWire: "destNode.destPort" → *PacedWire (owned by the destination).
//   - edgeWire: edge label → *PacedWire (same pointer; for stdin_reader lookup).
//   - edgeEndpoints: edge label → source/target node IDs + handles (for NodeMoveRegistry).
//   - edgeSteps: each edge's OWN bead-step count (per-edge geometry).
//   - edgeSegments: each edge's straight-segment endpoints (Start/End) so the
//     bead's position stream evaluates P(t)=Start+t*(End-Start).
//
// All keyed by edge label; consumed by buildNodes when binding the source Out.
func (b *buildCtx) allocateWires() {
	destWire := map[string]*wire.PacedWire{}
	edgeWire := loadspec.WireRegistry{}
	edgeEndpoints := map[string]inputcodec.EdgeEndpoints{}
	edgeSteps := map[string]int{}
	edgeSegments := map[string]wireSegment{}
	for _, e := range b.spec.Edges {
		destKey := e.Target + "." + e.TargetHandle
		// Segment: node SURFACE to node surface (docs/bead-model/channels-not-ports.md), the GPU
		// boundary the renderer draws from (nodegeom.EdgeSegment) — no port name feeds this any
		// more, a port contributes no geometry.
		srcG, tgtG := b.nodeGeoms[e.Source], b.nodeGeoms[e.Target]
		seg := nodegeom.EdgeSegment(srcG, tgtG)
		// Steps: the LIVE center-to-center distance between the two nodes, run through
		// EdgeStepCount (docs/bead-model/bead-lattice.md "The count") — the SAME function and the
		// SAME kind of distance the source node's own chainBeads pass (chain_beads.go)
		// will keep recomputing once running, so this load-time value and that first
		// live pass can never disagree.
		dist := nodegeom.NodeWorldPos(srcG).Sub(nodegeom.NodeWorldPos(tgtG)).Length()
		steps := nodegeom.EdgeStepCount(dist, srcG.Kind, tgtG.Kind)
		edgeSteps[e.Label] = steps
		edgeSegments[e.Label] = seg
		// One wire per destination input port, and — since fan-in is removed
		// (validateNoFanIn) — exactly one edge per port, so this is strictly one wire per
		// edge. A destKey already present means a fan-in spec slipped past the parser: that
		// is a build-invariant violation, not a topology to silently share a wire for.
		if _, exists := destWire[destKey]; exists {
			panic("allocateWires: two edges target " + destKey + " — validateNoFanIn should have rejected this fan-in at parse")
		}
		// lattice.DwellTicksPerBead is the ONE canonical dwell-per-step constant
		// (docs/bead-model/bead-lattice.md "Timing" — uniform pulse speed is now structural,
		// not a length divided by a speed); guarded as the sole non-test
		// NewPacedWire call site by tools/network/beads/check-uniform-pulse-speed.sh.
		pw := wire.NewPacedWire(steps, lattice.DwellTicksPerBead)
		pw.Target = e.Target
		pw.TargetHandle = e.TargetHandle
		pw.SetTrace(b.tr)
		destWire[destKey] = pw
		edgeWire[e.Label] = pw
		edgeEndpoints[e.Label] = inputcodec.EdgeEndpoints{
			Source: e.Source, Target: e.Target,
			SourceHandle: e.SourceHandle, TargetHandle: e.TargetHandle,
		}
	}
	b.destWire = destWire
	b.edgeWire = edgeWire
	b.edgeEndpoints = edgeEndpoints
	b.edgeSteps = edgeSteps
	b.edgeSegments = edgeSegments
}
