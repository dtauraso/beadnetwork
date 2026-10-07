package Topology

import (
	"sort"
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/CardPanel"
)

const CardKind = "NodePhiTheta3"

func CardPanelNodes(spec TopoSpec) (nodes []int, cards map[int]CardPanel.Card) {
	cards = map[int]CardPanel.Card{}
	for _, n := range spec.Nodes {
		if n.Type != CardKind {
			continue
		}
		id, err := strconv.Atoi(n.ID)
		if err != nil || id < 1 || id > CardPanel.NodeCount {
			continue
		}
		var state map[string]int
		if n.Data != nil {
			state = n.Data.State
		}
		nodes = append(nodes, id)
		cards[id] = CardPanel.CardFromState(id, state)
	}
	sort.Ints(nodes)
	return nodes, cards
}
