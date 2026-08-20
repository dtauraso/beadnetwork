package owners

import (
	"github.com/dtauraso/wirefold/src/Chrome/Panels/PolarRulesPanel"
)

type Topology struct {
	edgeIDs       []string
	neighborKinds map[string]string
	mutualTargets map[string]bool
	nodeRowFor    func(id string) (int32, bool)

	dragRule   *PolarRulesPanel.DragRule
	dragActive bool

	selfRule   *PolarRulesPanel.DragRule
	selfActive bool
}

func NewTopology() Topology {
	return Topology{dragActive: true, selfActive: true}
}

func (t *Topology) EdgeIDs() []string { return t.edgeIDs }

func (t *Topology) NeighborKinds() map[string]string { return t.neighborKinds }

func (t *Topology) AddMutualTarget(target string) {
	if t.mutualTargets == nil {
		t.mutualTargets = map[string]bool{}
	}
	t.mutualTargets[target] = true
}

func (t *Topology) AddEdgeID(edgeID string) {
	t.edgeIDs = append(t.edgeIDs, edgeID)
}

func (t *Topology) AddNeighborKind(toID, kind string) {
	if t.neighborKinds == nil {
		t.neighborKinds = map[string]string{}
	}
	t.neighborKinds[toID] = kind
}

func (t *Topology) SetDragRule(rule *PolarRulesPanel.DragRule) {
	t.dragRule = rule
}

func (t *Topology) DragRule() *PolarRulesPanel.DragRule { return t.dragRule }

func (t *Topology) SetDragActive(active bool) {
	t.dragActive = active
}

func (t *Topology) DragRuleActive() bool { return t.dragActive }

func (t *Topology) SetSelfRule(rule *PolarRulesPanel.DragRule) {
	t.selfRule = rule
}

func (t *Topology) SelfRule() *PolarRulesPanel.DragRule { return t.selfRule }

func (t *Topology) SetSelfRuleActive(active bool) {
	t.selfActive = active
}

func (t *Topology) SelfRuleActive() bool { return t.selfActive }

func (t *Topology) NeighborKind(toID string) string {
	return t.neighborKinds[toID]
}

func (t *Topology) IsMutualTarget(toID string) bool {
	return t.mutualTargets[toID]
}

func (t *Topology) SetNodeRowFor(fn func(id string) (int32, bool)) {
	t.nodeRowFor = fn
}

func (t *Topology) NodeRowFor(id string) (int32, bool) {
	if t.nodeRowFor == nil {
		return -1, false
	}
	return t.nodeRowFor(id)
}
