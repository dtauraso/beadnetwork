// build_wires.go — the wire-allocation phase of buildFromSpec: one *PacedWire per
// destination port, with each edge's own initial bead-step count and straight-segment
// endpoints computed alongside it.

package Wiring

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// allocateWires allocates one *PacedWire per destination port (one edge per port —
// fan-in is rejected at parse) and computes each edge's own INITIAL bead-step count
// (docs/bead-lattice.md "The count") and straight-segment endpoints.
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
	edgeWire := WireRegistry{}
	edgeEndpoints := map[string]EdgeEndpoints{}
	edgeSteps := map[string]int{}
	edgeSegments := map[string]wireSegment{}
	for _, e := range b.spec.Edges {
		destKey := e.Target + "." + e.TargetHandle
		// Segment: node SURFACE to node surface (docs/channels-not-ports.md), the GPU
		// boundary the renderer draws from (edgeSegment) — no port name feeds this any
		// more, a port contributes no geometry.
		srcG, tgtG := b.nodeGeoms[e.Source], b.nodeGeoms[e.Target]
		seg := edgeSegment(srcG, tgtG)
		// Steps: the LIVE center-to-center distance between the two nodes, run through
		// edgeStepCount (docs/bead-lattice.md "The count") — the SAME function and the
		// SAME kind of distance the source node's own chainBeads pass (chain_beads.go)
		// will keep recomputing once running, so this load-time value and that first
		// live pass can never disagree.
		dist := nodeWorldPos(srcG).Sub(nodeWorldPos(tgtG)).Length()
		steps := edgeStepCount(dist, srcG.Kind, tgtG.Kind)
		edgeSteps[e.Label] = steps
		edgeSegments[e.Label] = seg
		// One wire per destination input port, and — since fan-in is removed
		// (validateNoFanIn) — exactly one edge per port, so this is strictly one wire per
		// edge. A destKey already present means a fan-in spec slipped past the parser: that
		// is a build-invariant violation, not a topology to silently share a wire for.
		if _, exists := destWire[destKey]; exists {
			panic("allocateWires: two edges target " + destKey + " — validateNoFanIn should have rejected this fan-in at parse")
		}
		// wire.DwellTicksPerBead is the ONE canonical dwell-per-step constant
		// (docs/bead-lattice.md "Timing" — uniform pulse speed is now structural,
		// not a length divided by a speed); guarded as the sole non-test
		// NewPacedWire call site by tools/check-uniform-pulse-speed.sh.
		pw := wire.NewPacedWire(steps, wire.DwellTicksPerBead)
		pw.Target = e.Target
		pw.TargetHandle = e.TargetHandle
		pw.SetTrace(b.tr)
		destWire[destKey] = pw
		edgeWire[e.Label] = pw
		edgeEndpoints[e.Label] = EdgeEndpoints{
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

// allocateVectorChannels creates one dedicated node-to-node tilt-vector channel
// (tilt_vector_channel.go's TiltVectorMsg, buffered 1, latest-wins) per directed
// edge whose BOTH endpoint kinds ask for one (KindWantsVectorChannel — today only
// PairNode). A kind that never asks gets no entry in either map and is entirely
// unaffected. This travels ALONGSIDE the ordinary bead edge (allocateWires above),
// never replacing it — the source node keeps its existing *wire.Out for beads and
// additionally gets this channel's send end; the target node keeps its existing
// *wire.In and additionally gets this channel's receive end.
func (b *buildCtx) allocateVectorChannels() {
	kindByID := make(map[string]string, len(b.spec.Nodes))
	for _, n := range b.spec.Nodes {
		kindByID[n.ID] = n.Type
	}
	vectorOutByNode := map[string]chan TiltVectorMsg{}
	vectorInByNode := map[string]chan TiltVectorMsg{}
	for _, e := range b.spec.Edges {
		if !KindWantsVectorChannel(kindByID[e.Source]) || !KindWantsVectorChannel(kindByID[e.Target]) {
			continue
		}
		sourceToTargetVectorCh := make(chan TiltVectorMsg, 1)
		vectorOutByNode[e.Source] = sourceToTargetVectorCh
		vectorInByNode[e.Target] = sourceToTargetVectorCh
	}
	b.vectorOutByNode = vectorOutByNode
	b.vectorInByNode = vectorInByNode
}
