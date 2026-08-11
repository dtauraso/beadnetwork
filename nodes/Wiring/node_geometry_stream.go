// node_geometry_stream.go — a nodeGeometry's re-emit trigger and its dedicated per-fd
// content-buffer frame packer, the two together making up the ONLY path any node writes
// to its own stream.
package Wiring

import (
	"encoding/binary"
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
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
	// Every derived geometry column below (pole, ring axis, tilt/received vector angles) is
	// pure arithmetic on this node's own already-held state — see
	// nodegeom.DeriveFrameGeometry's own doc comment for why it lives there rather than here.
	fg := nodegeom.DeriveFrameGeometry(nodegeom.FrameGeometryInputs{
		Geom:                   m.geom,
		UpAxis:                 m.flags.upAxis,
		CoplanarEdges:          m.flags.coplanarEdges,
		PartnerCenters:         m.topo.partnerCenters,
		TopTiltVectorThetaIdx:  m.tilt.topTiltVectorThetaIdx,
		BottomThetaIdx:         m.tilt.bottomThetaIdx,
		NormalThetaIdx:         m.tilt.normalThetaIdx,
		ReceivedVectorThetaIdx: m.tilt.receivedVectorThetaIdx,
		ReceivedVectorSet:      m.tilt.receivedVectorSet,
		LatticePoints:          m.tilt.latticePoints,
		DefaultLatticePoints:   tiltvector.FullTurnThetaIdx,
	})
	center := fg.Center
	sphereR := fg.SphereR
	poleTheta, polePhi := fg.PoleTheta, fg.PolePhi
	ringAxisTheta, ringAxisPhi := fg.RingAxisTheta, fg.RingAxisPhi
	points := fg.LatticePoints
	topTiltVectorLen := fg.TopTiltVectorLen
	topTiltVectorTheta := fg.TopTiltVectorTheta
	bottomTiltVectorTheta := fg.BottomTiltVectorTheta
	coplanarNormalTheta := fg.CoplanarNormalTheta
	receivedVectorLen := fg.ReceivedVectorLen
	receivedVectorTheta := fg.ReceivedVectorTheta
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
		VRX:                   loadspec.VerticalRingNormalX,
		VRY:                   loadspec.VerticalRingNormalY,
		VRZ:                   loadspec.VerticalRingNormalZ,
		FRX:                   loadspec.FlatRingNormalX,
		FRY:                   loadspec.FlatRingNormalY,
		FRZ:                   loadspec.FlatRingNormalZ,
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
