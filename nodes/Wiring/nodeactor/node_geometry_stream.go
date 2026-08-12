package nodeactor

import (
	"encoding/binary"
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/nodeframe"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
	"github.com/dtauraso/wirefold/nodes/rowevent"

	T "github.com/dtauraso/wirefold/Trace"
)

func (m *NodeGeometry) emitGeometry() {

	m.writeStreamFrame([]rowevent.RowEvent{{
		Kind: T.KindNodeGeometry, NodeRow: m.stream.nodeRow,
		PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1,
	}})
}

func (m *NodeGeometry) writeStreamFrame(events []rowevent.RowEvent) {
	if !m.stream.streamOut.Ok() || m.stream.buildFrame == nil {
		return
	}

	for _, e := range events {
		if e.NodeRow != m.stream.nodeRow {
			panic(fmt.Sprintf(
				"NodeGeometry.writeStreamFrame: node %q (row %d) is carrying a %s event for row %d on its OWN dedicated stream — NodeRow is an ownership claim, not a reference; a foreign node belongs in TargetRow",
				m.id, m.stream.nodeRow, e.Kind, e.NodeRow))
		}
	}

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

	chainOX, chainOY, chainOZ, chainLit, chainLitVal, chainBreadcrumbs := m.chainBeads()
	if len(chainBreadcrumbs) > 0 {

		events = append(events, chainBreadcrumbs...)
	}

	frame := m.stream.buildFrame(nodeframe.NodeFrameInput{
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

	_, _ = m.stream.streamOut.Write(hdr[:])
	_, _ = m.stream.streamOut.Write(frame)
}
