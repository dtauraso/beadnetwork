package loadspec

import (
	"sort"
	"strconv"

	"github.com/dtauraso/wirefold/Chrome/Panels/PolarRulesPanel"
	"github.com/dtauraso/wirefold/Node/nodedrag"
)

func (spec TopoSpec) RulePanelNodes() []PolarRulesPanel.Node {
	rowOf := func(id string) (int32, bool) {
		n, err := strconv.Atoi(id)
		if err != nil || n < 1 {
			return -1, false
		}
		return int32(n - 1), true
	}

	edgeRowOf := make(map[string]int32, len(spec.Edges))
	for i, e := range spec.Edges {
		edgeRowOf[e.Label] = int32(i)
	}

	reverse := make(map[string]map[string]bool, len(spec.Nodes))
	for _, e := range spec.Edges {
		if reverse[e.Target] == nil {
			reverse[e.Target] = map[string]bool{}
		}
		reverse[e.Target][e.Source] = true
	}

	nodes := make([]PolarRulesPanel.Node, 0, len(spec.Nodes))
	for _, n := range spec.Nodes {
		row, ok := rowOf(n.ID)
		if !ok {
			continue
		}
		rn := PolarRulesPanel.Node{Row: row, Label: n.ID, Kind: n.Type, HasKindRule: nodedrag.HasKindRule(n.Type)}
		for _, e := range spec.Edges {
			if e.Source != n.ID {
				continue
			}
			otherRow, ok := rowOf(e.Target)
			if !ok {
				continue
			}
			rn.Out = append(rn.Out, PolarRulesPanel.Edge{
				EdgeRow: edgeRowOf[e.Label], OtherRow: otherRow, OtherLabel: e.Target,
			})
			rn.HasReverse = append(rn.HasReverse, reverse[n.ID][e.Target])
		}
		nodes = append(nodes, rn)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Row < nodes[j].Row })

	return nodes
}
