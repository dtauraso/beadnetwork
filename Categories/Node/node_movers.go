package Node

import (
	"context"

	NodeDrag "github.com/dtauraso/beadnetwork/Categories/Node/Drag"
	"github.com/dtauraso/beadnetwork/Categories/Node/Edge/edgetable"
	"github.com/dtauraso/beadnetwork/Categories/Node/Files"
	"github.com/dtauraso/beadnetwork/Categories/Node/Geom"
	"github.com/dtauraso/beadnetwork/Categories/Node/Messaging"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

type NodeMover struct{}

func HeldCenters(nodeGeoms map[string]*NodeGeometry, centerOf func(id string) (Vec3, bool)) map[string]Vec3 {
	out := make(map[string]Vec3, len(nodeGeoms))
	for id := range nodeGeoms {
		if c, ok := centerOf(id); ok {
			out[id] = c
		}
	}
	return out
}

func (mv *NodeMover) RootMove(ctx context.Context, nodeGeoms map[string]*NodeGeometry, nodeID string, target Vec3) bool {
	nm, ok := nodeGeoms[nodeID]
	if !ok {
		return false
	}

	sc := nm.Constants()
	targetIdx := polarindex.MeasureIndex(polar.Cart2polar(polar.Vec3(target.Sub(nm.SceneCenter()))), sc)

	nm.Msg().SendExternal(ctx, Messaging.Msg{NodeID: nodeID, Body: Messaging.Drag{Target: &targetIdx}})
	return true
}

func (mv *NodeMover) CommitNodeMoveLocal(nodeGeoms map[string]*NodeGeometry, edgeTable map[string]*edgetable.Edge, nm *NodeGeometry, committedIdx polarindex.Index) {
	mv.applyNodeMoveLocal(nodeGeoms, edgeTable, nm, committedIdx)
	nm.CommitIndex()
}

func (mv *NodeMover) ApplyDerivedNodeMove(nodeGeoms map[string]*NodeGeometry, edgeTable map[string]*edgetable.Edge, nm *NodeGeometry, committedIdx polarindex.Index) {
	nm.OutEdges().SuppressDeltaPersist()
	mv.applyNodeMoveLocal(nodeGeoms, edgeTable, nm, committedIdx)
}

func (mv *NodeMover) applyNodeMoveLocal(nodeGeoms map[string]*NodeGeometry, edgeTable map[string]*edgetable.Edge, nm *NodeGeometry, committedIdx polarindex.Index) {
	nodeID := nm.ID()

	deltaIdx := polarindex.Delta(committedIdx, nm.ComposedIndex())
	nm.Deltas().ShiftSelfBy(deltaIdx)

	nm.ApplyCenter(committedIdx)
	committedPos := Geom.WorldPosOf(nm.geom, committedIdx)
	BroadcastToPartners(edgeTable, nodeGeoms,
		map[string]Vec3{nodeID: Vec3(committedPos)},
		map[string]polarindex.Offset{nodeID: deltaIdx},
		nm.Msg().SendMove())
}

func BroadcastToPartners(edges map[string]*edgetable.Edge, nodeGeoms map[string]*NodeGeometry, newCenters map[string]Vec3, moveDeltas map[string]polarindex.Offset, enqueue func(id string, msg Messaging.Msg)) {

	partners := map[string]string{}
	for _, e := range edges {
		if _, moved := newCenters[e.SrcID()]; moved {
			if _, alsoMoved := newCenters[e.DstID()]; !alsoMoved {
				partners[e.DstID()] = e.SrcID()
			}
		}
		if _, moved := newCenters[e.DstID()]; moved {
			if _, alsoMoved := newCenters[e.SrcID()]; !alsoMoved {
				partners[e.SrcID()] = e.DstID()
			}
		}
	}
	for partnerID, movedID := range partners {
		if _, ok := nodeGeoms[partnerID]; !ok {
			continue
		}

		moved := Messaging.NeighborMoved{SenderID: movedID}
		if d, ok := moveDeltas[movedID]; ok {
			moved.Delta = &d
		}
		enqueue(partnerID, Messaging.Msg{NodeID: partnerID, Body: moved})
	}
}

func (m *NodeGeometry) ApplyCenter(idx polarindex.Index) {
	Geom.SetNodeWorld(&m.geom, idx)

	m.msg.PublishCenter(Vec3(Geom.NodeWorldPos(m.geom)))

	m.emitGeometry()
}

func (m *NodeGeometry) SetKindRule(trim NodeDrag.Trim, request NodeDrag.Request) {
	m.kindRule.SetKindRule(trim, request)
}

func (m *NodeGeometry) TrimOwnDrag(delta polarindex.Offset) polarindex.Offset {
	return NodeDrag.Apply(m.kindRule.Trim(), delta, m.dragState())
}

func (m *NodeGeometry) RequestedDrag(delta polarindex.Offset) map[string]polarindex.Offset {
	return NodeDrag.Requested(m.kindRule.Request(), delta, m.dragState())
}

func (m *NodeGeometry) dragState() NodeDrag.State {
	st := NodeDrag.State{
		Index:      m.ComposedIndex(),
		Constants:  m.geom.SceneConstants,
		OutTargets: m.outTargets,
		OutDelta:   make(map[string]polarindex.Offset, len(m.outTargets)),
	}
	for _, to := range m.outTargets {
		if d, ok := m.deltas.DeltaTo(to); ok {
			st.OutDelta[to] = d
		}
	}
	return st
}

func (m *NodeGeometry) PersistRoot() string { return m.persistRoot }

func (nm *NodeGeometry) persistIndex(off polarindex.Offset) {
	nm.writeIndex(off)
}

func (nm *NodeGeometry) persistTiltVectorAngle() {
	nm.writeIndex(nm.geom.DragIndex)
}

func (nm *NodeGeometry) writeIndex(off polarindex.Offset) {
	if nm.persistRoot == "" {
		return
	}
	err := Files.WriteDragIndex(nm.persistRoot, nm.id, off.Phi, off.Theta, off.R, nm.tilt.TopTiltVectorPhiIdx())
	if err != nil {
		Files.LogPersistErr("index_persist", nm.id, err)
	}
}
