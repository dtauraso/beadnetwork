package Builder

const (
	OpNodeIdx    = "node-idx"
	OpNodeRef    = "node-ref"
	OpAddNode    = "add-node"
	OpRemoveNode = "remove-node"
	OpConnect    = "connect"
	OpDisconnect = "disconnect"
)

func networkRows(t *tree, m Model) {
	t.add(linkRow{
		id:     Act{Op: OpNetFold},
		open:   openOf(m.NetworkOpen),
		fold:   Act{Op: OpNetFold},
		name:   Text("network"),
		detail: []Cell{Text("what runs now · " + linkCount(len(m.allLinks()))).Dimmed(true)},
		acts:   []Cell{Button("+ node", Act{Op: OpAddNode}, false).Tipped("add a node")},
	})
	if !m.NetworkOpen {
		return
	}
	for _, id := range m.Nodes {
		fold := Act{Op: OpStepFold, Node: id}
		r := linkRow{id: fold, depth: 1, open: openOf(!m.StepFolded(0, id)), fold: fold, name: Text("node " + id), table: "network node " + id}
		if m.StepFolded(0, id) {
			r.detail = []Cell{Text(linkCount(len(m.linksTo(id)))).Dimmed(true)}
		} else {
			r.more = refCells(m, id)
		}
		r.acts = append(connectRow(m, id, "link to"), Button("×", Act{Op: OpRemoveNode, Node: id}, false).Tipped("remove this node"))
		t.add(r)
		if m.StepFolded(0, id) {
			continue
		}
		placeRows(t, m, id)
		for _, l := range m.linksTo(id) {
			e := StepEdge{0, l.From, l.To}
			lr := m.cardRow(linkRowOf(l, l.Picked, Act{}, 2), e)
			lr.acts = []Cell{Button("×", Act{Op: OpDisconnect, Node: l.From, To: l.To}, false).Tipped("remove this link")}
			t.add(lr)
			if m.CardOpen(e) {
				t.put(netCard(l, 3))
			}
		}
	}
}

func connectRow(m Model, from, label string) []Cell {
	row := []Cell{Text(label)}
	for _, to := range m.Nodes {
		if _, has := m.Link(from, to); to != from && !has {
			row = append(row, Button(to, Act{Op: OpConnect, Node: from, To: to}, false))
		}
	}
	if len(row) == 1 {
		return nil
	}
	return row
}
