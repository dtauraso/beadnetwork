package Builder

import (
	"slices"
	"strconv"
)

const PlaceLabel = "Place"

const (
	OpItem   = "item"
	OpMethod = "method"
	OpCard   = "card"
)

func (m Model) Item() string {
	if slices.Contains(m.Nodes, m.Cards.Item) {
		return m.Cards.Item
	}
	if len(m.Nodes) > 0 {
		return m.Nodes[0]
	}
	return ""
}

type run struct {
	method Method
	steps  []CardStep
}

func (m Model) runs() []run {
	item := m.Item()
	var out []run
	for _, me := range Methods {
		if m.Cards.Methods[me.ID] && item != "" {
			out = append(out, run{method: me, steps: me.Run(m, item)})
		}
	}
	return out
}

func (m Model) CardCount() int {
	n := 1
	for _, r := range m.runs() {
		n = max(n, len(r.steps))
	}
	return n
}

func (m Model) CardShown() int { return min(max(1, m.Cards.Card), m.CardCount()) }

func PlacePill(m Model) Pill {
	item := m.Item()
	var nodes []Cell
	for _, id := range m.Nodes {
		nodes = append(nodes, Button(id, Act{Op: OpItem, Node: id}, id == item))
	}
	var methods [][]Cell
	for _, me := range Methods {
		methods = append(methods, []Cell{Button(me.Name, Act{Op: OpMethod, K: me.ID}, m.Cards.Methods[me.ID])})
	}
	card, count := m.CardShown(), m.CardCount()
	stepper := []Cell{Button("◀", Act{Op: OpCard, Sign: -1}, false), Text("step " + strconv.Itoa(card) + " of " + strconv.Itoa(count)), Button("▶", Act{Op: OpCard, Sign: 1}, false)}
	groups := []Group{
		{Title: "node to place", Rows: [][]Cell{nodes}},
		{Title: "methods", Rows: methods},
		{Rows: [][]Cell{stepper}},
	}
	for _, r := range m.runs() {
		var rows [][]Cell
		for i, st := range r.steps {
			title := strconv.Itoa(i+1) + ". " + st.Title
			if !st.OK {
				title += "  ✗"
			}
			rows = append(rows, []Cell{LitText(title, Act{}, i+1 == card)})
			for _, l := range st.Lines {
				rows = append(rows, []Cell{Text("    " + l)})
			}
			rows = append(rows, []Cell{Text("    — " + st.Note)})
		}
		groups = append(groups, Group{Title: r.method.Name, Rows: rows})
	}
	return Pill{Label: PlaceLabel, Left: true, Groups: groups}
}
