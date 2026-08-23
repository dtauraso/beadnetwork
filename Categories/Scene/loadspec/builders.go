package loadspec

import (
	"strconv"

	NodeBuf "github.com/dtauraso/wirefold/Categories/Node"
	"github.com/dtauraso/wirefold/Categories/Node/nodeactor"
	"github.com/dtauraso/wirefold/Categories/Node/nodeactor/nodefiles"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

func KindForID(id uint8) (string, bool) {
	for _, k := range NodeBuf.KnownKinds() {
		if NodeBuf.NodeKindID(k) == id {
			return k, true
		}
	}
	return "", false
}

func NewNodeID(root string) string {
	return strconv.Itoa(LargestNodeID(root) + 1)
}

func SeedNode(m *nodeactor.NodeGeometry, n Node, sceneRoot string) {
	m.SetSelfKind(n.Type)

	rn := m.RuleNode()
	rn.SetPersistRoot(sceneRoot)

	active := nodefiles.LoadDragActive(sceneRoot, n.ID)
	m.Topo().SetDragRule(n.Drag)
	m.Topo().SetDragActive(active)
	rn.SeedRule(n.Drag, active)
	rn.SeedKindActive(nodefiles.LoadKindRuleActive(sceneRoot, n.ID))

	selfActive := nodefiles.LoadSelfRuleActive(sceneRoot, n.ID)
	m.Topo().SetSelfRule(n.SelfDrag)
	m.Topo().SetSelfRuleActive(selfActive)
	rn.SeedSelfRule(n.SelfDrag, selfActive)

	if n.TopTiltVectorPhiIdx != nil {
		m.SetTopTiltVectorPhiIdx(*n.TopTiltVectorPhiIdx)
	}
}

func SeedEdge(m *nodeactor.NodeGeometry, e Edge, src bool, otherKind, sceneRoot string) {
	other := e.Target
	if !src {
		other = e.Source
	}

	m.RuleNode().SeedEdgeActive(other, nodefiles.LoadEdgeRuleActive(sceneRoot, e.Source, e.Target))
	m.Topo().AddNeighborKind(other, otherKind)
	if src {
		m.AddOutTarget(e.Target)
	}

	baseD, ok := e.BaseDeltaIndex()
	if !ok {
		return
	}
	dragD := e.DragDeltaIndex()
	if !src {
		baseD, dragD = polarindex.Neg(baseD), polarindex.Neg(dragD)
	}
	m.SetBaseDeltaTo(other, baseD)
	m.SetDragDeltaTo(other, dragD)
}
