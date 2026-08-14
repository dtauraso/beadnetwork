package nodeactor

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/framegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/nodeframe"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
	"github.com/dtauraso/wirefold/nodes/rowevent"

	T "github.com/dtauraso/wirefold/Trace"
)

func (m *NodeGeometry) emitGeometry() {

	m.writeStreamFrame([]rowevent.RowEvent{{
		Kind: T.KindNodeGeometry, NodeRow: m.stream.NodeRow(),
		PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1,
	}})
}

func (m *NodeGeometry) postSelfEvents(events []rowevent.RowEvent) {
	m.stream.PostSelfEvents(events)
}

func (m *NodeGeometry) drainSelfEvents() []rowevent.RowEvent {
	return m.stream.DrainSelfEvents()
}

func (m *NodeGeometry) writeStreamFrame(events []rowevent.RowEvent) {
	if !m.stream.Ready() {
		return
	}

	row := m.stream.NodeRow()
	for _, e := range events {
		if e.NodeRow != row {
			panic(fmt.Sprintf(
				"NodeGeometry.writeStreamFrame: node %q (row %d) is carrying a %s event for row %d on its OWN dedicated stream — NodeRow is an ownership claim, not a reference; a foreign node belongs in TargetRow",
				m.id, row, e.Kind, e.NodeRow))
		}
	}

	coplanarEdges, upAxis := m.flags.Flags()
	topIdx, bottomIdx, normalIdx, receivedIdx, receivedSet, latticePoints := m.tilt.FrameGeometryFields()
	fg := framegeom.DeriveFrameGeometry(framegeom.FrameGeometryInputs{
		Geom:                   m.geom,
		UpAxis:                 upAxis,
		CoplanarEdges:          coplanarEdges,
		PartnerDeltas:          m.deltas.All(),
		TopTiltVectorThetaIdx:  topIdx,
		BottomThetaIdx:         bottomIdx,
		NormalThetaIdx:         normalIdx,
		ReceivedVectorThetaIdx: receivedIdx,
		ReceivedVectorSet:      receivedSet,
		LatticePoints:          latticePoints,
		DefaultLatticePoints:   tiltvector.FullTurnThetaIdx,
	})
	center := fg.Center
	sphereR := fg.SphereR
	polePhi, poleTheta := fg.PolePhi, fg.PoleTheta
	ringAxisPhi, ringAxisTheta := fg.RingAxisPhi, fg.RingAxisTheta
	points := fg.LatticePoints
	topTiltVectorLen := fg.TopTiltVectorLen
	receivedVectorLen := fg.ReceivedVectorLen
	label := m.geom.Label
	if label == "" {
		label = m.id
	}
	selected, hovered, latchedSel := m.ui.Flags()
	kindID := m.stream.KindID()
	roundsToParallel, msgsToParallel := m.readout.RoundsToParallel()

	m.stream.WriteFrame(nodeframe.NodeFrameInput{
		Tick:                  uint32(m.clocks.Tick()),
		NodeRow:               row,
		NodeID:                row + 1,
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
		PolePhi:               float32(polePhi),
		PoleTheta:             float32(poleTheta),
		RingAxisPhi:           float32(ringAxisPhi),
		RingAxisTheta:         float32(ringAxisTheta),
		TopTiltVectorLen:      float32(topTiltVectorLen),
		TopTiltVectorIdx:      fg.TopTiltVectorIdx,
		TopTiltVectorPhi:      float32(fg.TopTiltVectorPhi),
		TopTiltVectorTheta:    float32(fg.TopTiltVectorTheta),
		BottomTiltVectorPhi:   float32(fg.BottomTiltVectorPhi),
		BottomTiltVectorTheta: float32(fg.BottomTiltVectorTheta),
		CoplanarNormalPhi:     float32(fg.CoplanarNormalPhi),
		CoplanarNormalTheta:   float32(fg.CoplanarNormalTheta),
		ReceivedVectorLen:     float32(receivedVectorLen),
		ReceivedVectorPhi:     float32(fg.ReceivedVectorPhi),
		ReceivedVectorTheta:   float32(fg.ReceivedVectorTheta),
		Selected:              selected,
		KindID:                kindID,
		Hovered:               hovered,
		LatchedSel:            latchedSel,
		LatticePoints:         uint8(points),
		RoundsToParallel:      roundsToParallel,
		MsgsToParallel:        msgsToParallel,
		Label:                 label,
		Events:                events,
	})
}
