// node_geometry_stream.go — a nodeGeometry's re-emit trigger and its dedicated per-fd
// content-buffer frame packer, the two together making up the ONLY path any node writes
// to its own stream.
package Wiring

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

// emitGeometry re-emits this node's authoritative geometry (center, radius, ring
// normals — no port geometry: a port carries none, docs/bead-model/channels-not-ports.md).
// This method and applyCenter both run on this node's own driving goroutine only, so a
// plain field read here can never race a concurrent writer.
func (m *nodeGeometry) emitGeometry() {
	// Dedicated per-node stream (see streamOut's doc comment): write this node's own
	// combined frame immediately on a geometry change, in addition to the tick-driven
	// write in the driving loop's own per-cycle write. NodeGeometry rides THIS frame's
	// own EVENTS section (fully decentralized — it never rides the VIEW stream's
	// fallback bucket) — this node is the sole owner of its own geometry, so it resolves
	// its own NodeRow at the call site (owner_events.go) rather than routing through a
	// shared accumulator.
	m.writeStreamFrame([]wire.RowEvent{{
		Kind: T.KindNodeGeometry, NodeRow: m.stream.nodeRow,
		PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1,
	}})
}

// writeStreamFrame packs and writes this node's combined per-fd frame (center/radius/
// ring-normals + ports + label + selection-UI columns) to its OWN dedicated fd
// (streamOut). No-op when streamOut is nil (the fallback — see its doc comment) or
// buildFrame was never injected (bare test construction). Called only by this node's own
// driving goroutine, reading m.geom. events carries whatever this call's caller wants
// riding this frame's trailing EVENTS section (nil from a plain tick-driven write).
func (m *nodeGeometry) writeStreamFrame(events []wire.RowEvent) {
	if !m.stream.streamOut.Ok() || m.stream.buildFrame == nil {
		return
	}
	// INVARIANT: a node carries only its OWN events on its OWN dedicated stream. This is
	// the per-goroutine bridge stated in CLAUDE.md's "Bridge surface" and in
	// memory/feedback_no_single_writer_bridge.md + memory/feedback_per_goroutine_bridge.md,
	// and until now it was enforced by prose alone. NodeRow is the ownership column; a
	// FOREIGN node is referenced through TargetRow (see quantized_move.go's abc-drag
	// breadcrumb, which sets NodeRow: nm.stream.nodeRow and TargetRow: the other node). Violating
	// it produces a frame the TS side decodes onto the wrong row — a silently wrong scene
	// that still renders, which is the expensive failure this panic converts into a cheap
	// one. Placed AFTER the nil guard on purpose: bare geometries built in tests never
	// reach the pack path, and nodeRow is seeded alongside streamOut (stream_wiring.go),
	// so any frame that gets here has a real row.
	for _, e := range events {
		if e.NodeRow != m.stream.nodeRow {
			panic(fmt.Sprintf(
				"nodeGeometry.writeStreamFrame: node %q (row %d) is carrying a %s event for row %d on its OWN dedicated stream — NodeRow is an ownership claim, not a reference; a foreign node belongs in TargetRow",
				m.id, m.stream.nodeRow, e.Kind, e.NodeRow))
		}
	}
	center := nodegeom.NodeWorldPos(m.geom)
	sphereR := nodegeom.EffectiveRadius(m.geom)
	// This node's own local-frame pole: its own scene-polar direction reversed, so the frame
	// points back at the scene centre (Buffer/layout.go PoleTheta/PolePhi). Derived here from
	// m.geom.ScenePolar — this node's own coordinate, on this node's own goroutine, no
	// neighbour read. Before HasPos there is no direction yet, so the frame stays world +y.
	var poleTheta, polePhi float64
	if m.geom.HasPos {
		poleTheta, polePhi = geom.InwardPole(m.geom.ScenePolar)
	}
	// The DRAWN ring's axis, separate from the navigation pole above (Buffer/layout.go's
	// RingAxisTheta/RingAxisPhi). Default is the torus's own +Z normal, which draws exactly
	// as an unrotated ring did — so a scene that has not asked for anything looks unchanged.
	ringAxisTheta, ringAxisPhi := nodegeom.TorusDefaultAxisAngles()
	// topTiltVectorLen is this node's own drawn vector, along the SAME axis as its ring, and 0
	// where a scene draws none (Buffer/layout.go's TopTiltVectorLen). It runs from the node's
	// centre to its own top, so its length IS the node's radius.
	var topTiltVectorLen float64
	if m.flags.upAxis && m.geom.HasPos && len(m.topo.partnerCenters) == 1 {
		// UPRIGHT: the ring STANDS UP along its edge — its plane holds both the edge and
		// world +y, so the node's own up-vector lies IN the ring's plane rather than
		// sticking out of a flat disc. An axis of +y itself would lie the ring flat and
		// put the vector perpendicular to it, which is the opposite arrangement.
		for _, partner := range m.topo.partnerCenters {
			if t, p, ok := nodegeom.UprightRingAxis(nodegeom.NodeWorldPos(m.geom), partner); ok {
				ringAxisTheta, ringAxisPhi = t, p
			}
		}
		topTiltVectorLen = nodegeom.NodeRadius(m.geom.Kind)
	} else if m.flags.coplanarEdges && m.geom.HasPos && len(m.topo.partnerCenters) == 1 {
		// COPLANAR EDGES: swing the axis off the inward pole by the smallest amount that
		// puts the edge INSIDE the ring plane — the inward pole with its along-the-edge
		// component removed. The chain, this node's torus and the beads' own tori then
		// share one plane instead of the chain running through the holes. Only for a node
		// with exactly ONE neighbour: two non-collinear edges have no common plane.
		for _, partner := range m.topo.partnerCenters {
			if t, p, ok := nodegeom.PoleContainingEdge(poleTheta, polePhi, nodegeom.NodeWorldPos(m.geom), partner); ok {
				ringAxisTheta, ringAxisPhi = t, p
			}
		}
	}
	// latticeThetaStep is THIS node's own angle-per-index — 2π / latticePoints, not the
	// fixed CurveParamTiltVectorAngleStep (which stays π/12, the compile-time 24-point
	// default every OTHER conversion in this codebase still uses). A pair node's own
	// lattice size is a scene setting (Node.adoptLattice, nodes/PairNode/node.go), reported
	// here one-way via PairNodeSelf.SetLatticePoints, so the same index draws a different
	// angle depending on how many points that node's own ring currently has. Derived once
	// per frame; every conversion below reads this local rather than recomputing it.
	points := m.tilt.latticePoints
	if points == 0 {
		points = FullTurnThetaIdx
	}
	latticeThetaStep := 2 * math.Pi / float64(points)
	// topTiltVectorTheta is this node's OWN vector direction — separate from the ring
	// axis above, so a scene/user can aim a node's vector somewhere other than its ring.
	// Never a free float: index × latticeThetaStep (this node's own lattice step, above),
	// the streamed value is pure arithmetic on the integer state this node's own mover
	// holds and persists (m.tilt.topTiltVectorThetaIdx). There is no φ: every tilt vector in
	// this model is θ-only (task/drop-tilt-vector-phi).
	topTiltVectorTheta := float64(m.tilt.topTiltVectorThetaIdx) * latticeThetaStep
	// The BOTTOM TILT VECTOR: streamed straight from this node's own bottomThetaIdx,
	// decided by THIS node's OWN goroutine (a half turn in θ from its own top
	// tilt index, same rule run unmodified by both nodes of a pair — PairNode's bottomTilt)
	// and reported one-way
	// via PairNodeSelf.SetTiltIndex alongside the top and the normal. Pure mirror here, same
	// as every other index on this frame: this mover derives none of them.
	bottomTiltVectorTheta := float64(m.tilt.bottomThetaIdx) * latticeThetaStep
	// The COPLANAR NORMAL: streamed straight from this node's own normalThetaIdx,
	// which THIS node's OWN goroutine decided (a fixed +90° in θ from its
	// own tilt index, same rule run unmodified by both nodes of a pair — PairNode's
	// coplanarNormal) and reported one-way via PairNodeSelf.SetTiltIndex. This mover is a pure mirror here, same shape
	// as topTiltVectorTheta above — it derives nothing from the edge/partner.
	// Turning the tilt therefore visibly turns the drawn normal WITH it, staying 90° away,
	// instead of the normal staying fixed toward the partner while the tilt moves under it.
	coplanarNormalTheta := float64(m.tilt.normalThetaIdx) * latticeThetaStep
	// The THIRD vector: the direction last received on this node's tilt-vector channel
	// (receivedVectorThetaIdx, mirrored one-way from this node's own goroutine —
	// see the field's own doc comment). Same length-says-whether-and-how-far convention
	// as topTiltVectorLen: zero when nothing has been received yet (or a reset cleared it),
	// non-zero (this node's own radius, same as topTiltVectorLen) otherwise — so a node with
	// nothing received is distinguishable from one whose received direction happens to be
	// 0, which still streams a non-zero length.
	var receivedVectorLen float64
	var receivedVectorTheta float64
	if m.tilt.receivedVectorSet {
		receivedVectorLen = nodegeom.NodeRadius(m.geom.Kind)
		receivedVectorTheta = float64(m.tilt.receivedVectorThetaIdx) * latticeThetaStep
	}
	label := m.geom.Label
	if label == "" {
		label = m.id
	}
	selected, hovered, latchedSel, kindID := m.ui.selected, m.ui.hovered, m.ui.latchedSel, m.stream.kindID
	// This node's own placeholder chain beads, node-local (chain_beads.go). Computed here
	// on this node's own goroutine from its own center + its own partnerCenters map — no
	// cross-goroutine position read.
	chainOX, chainOY, chainOZ, chainLit, chainLitVal, chainBreadcrumbs := m.chainBeads()
	if len(chainBreadcrumbs) > 0 {
		// DIAGNOSTIC ONLY (task/log-node4-chain-aim): chainBeads' own "chain-aim" events,
		// appended here rather than sent via a nested writeStreamFrame call from inside
		// chainBeads (which would recurse back into chainBeads — see that function's doc
		// comment on its breadcrumbs return value).
		events = append(events, chainBreadcrumbs...)
	}
	// nodeID is this node's own numeric identity: ROW ID = NODE ID - 1 (enforced at load,
	// persistence-ownership.md), so it is m.stream.nodeRow+1 by construction — not re-derived by any
	// offline rule the decoder also has to apply, it travels with the frame.
	frame := m.stream.buildFrame(NodeFrameInput{
		Tick:                  uint32(m.clocks.clk.Tick()),
		NodeRow:               m.stream.nodeRow,
		NodeID:                m.stream.nodeRow + 1,
		CX:                    float32(center.X),
		CY:                    float32(center.Y),
		CZ:                    float32(center.Z),
		Radius:                float32(nodegeom.NodeRadius(m.geom.Kind)),
		SphereR:               float32(sphereR),
		VRX:                   verticalRingNormalX,
		VRY:                   verticalRingNormalY,
		VRZ:                   verticalRingNormalZ,
		FRX:                   flatRingNormalX,
		FRY:                   flatRingNormalY,
		FRZ:                   flatRingNormalZ,
		PoleTheta:             float32(poleTheta),
		PolePhi:               float32(polePhi),
		RingAxisTheta:         float32(ringAxisTheta),
		RingAxisPhi:           float32(ringAxisPhi),
		TopTiltVectorLen:      float32(topTiltVectorLen),
		TopTiltVectorTheta:    float32(topTiltVectorTheta),
		BottomTiltVectorTheta: float32(bottomTiltVectorTheta),
		CoplanarNormalTheta:   float32(coplanarNormalTheta),
		ReceivedVectorLen:     float32(receivedVectorLen),
		ReceivedVectorTheta:   float32(receivedVectorTheta),
		Selected:              selected,
		KindID:                kindID,
		Hovered:               hovered,
		LatchedSel:            latchedSel,
		LatticePoints:         uint8(points),
		RoundsToParallel:      m.readout.roundsToParallel,
		MsgsToParallel:        m.readout.msgsToParallel,
		Label:                 label,
		ChainBeadOX:           chainOX,
		ChainBeadOY:           chainOY,
		ChainBeadOZ:           chainOZ,
		ChainBeadLit:          chainLit,
		ChainBeadLitValue:     chainLitVal,
		Events:                events,
	})
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], uint32(len(frame)))
	// Fire-and-forget, same reasoning throughout this bridge: no delivery
	// guarantee on this channel, errors ignored.
	_, _ = m.stream.streamOut.Write(hdr[:])
	_, _ = m.stream.streamOut.Write(frame)
}
