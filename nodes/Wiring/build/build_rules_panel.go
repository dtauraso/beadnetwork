package build

import (
	"sort"
	"strconv"

	"github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/rulespanel"
)

func buildRulePanelNodes(md *dispatch.MoveDispatch, spec loadspec.TopoSpec) {
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

	nodes := make([]rulespanel.Node, 0, len(spec.Nodes))
	for _, n := range spec.Nodes {
		row, ok := rowOf(n.ID)
		if !ok {
			continue
		}
		rn := rulespanel.Node{Row: row, Label: n.ID, Kind: n.Type}
		for _, e := range spec.Edges {
			if e.Source != n.ID {
				continue
			}
			otherRow, ok := rowOf(e.Target)
			if !ok {
				continue
			}
			rn.Out = append(rn.Out, rulespanel.Edge{
				EdgeRow: edgeRowOf[e.Label], OtherRow: otherRow, OtherLabel: e.Target,
			})
			rn.HasReverse = append(rn.HasReverse, reverse[n.ID][e.Target])
		}
		nodes = append(nodes, rn)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Row < nodes[j].Row })

	md.UI.RuleNodes = nodes
	md.UI.RuleSharedRow = -1
}
