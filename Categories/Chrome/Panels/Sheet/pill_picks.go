package Sheet

import "github.com/dtauraso/beadnetwork/Categories/Node/Links"

const (
	OpPick      = "pick"
	OpUnpick    = "unpick"
	OpIncrement = "increment"
	OpLocking   = "locking"
)

func (m Model) allows(l Links.Link) bool {
	for _, x := range m.allLinks() {
		if x.Picked && x.To == l.To && x.From != l.From {
			return false
		}
	}
	return true
}

func (m Model) Picked() []Links.Link {
	var out []Links.Link
	for _, l := range m.allLinks() {
		if l.Picked {
			out = append(out, l)
		}
	}
	return out
}

func pickGroup(m Model) Group {
	var row []Cell
	for _, l := range m.allLinks() {
		if !l.Picked && m.allows(l) {
			row = append(row, Button(l.Name(), Act{Op: OpPick, Node: l.From, To: l.To}, false))
		}
	}
	if len(row) == 0 {
		row = []Cell{Text("no edge left to pick")}
	}
	return Group{Title: "pick an edge — at most one picked edge into each node", Rows: [][]Cell{row}}
}

func lockSign(l Links.Link, k string) int {
	if k == Links.KTheta {
		return l.Lock.Theta
	}
	return l.Lock.Phi
}

func signButtons(lockOf func(k string) int, act func(k string, sign int) Act, k string) []Cell {
	return []Cell{
		Button("+", act(k, 1), lockOf(k) == 1),
		Button("−", act(k, -1), lockOf(k) == -1),
	}
}

func pickedGroup(m Model) Group {
	var rows [][]Cell
	for _, l := range m.Picked() {
		act := func(op, k string) Act { return Act{Op: op, Node: l.From, To: l.To, K: k} }
		press := func(k string, sign int) Act { return Act{Op: OpLinkPress, Node: l.From, To: l.To, K: k, Sign: sign} }
		lockOf := func(k string) int { return lockSign(l, k) }
		mark := "○"
		if l.Locking {
			mark = "●"
		}
		row := []Cell{Text(l.Name()), Field("Δφ by", Num(l.Increment.Phi), act(OpIncrement, Links.KPhi), true, false)}
		row = append(row, signButtons(lockOf, press, Links.KPhi)...)
		row = append(row, Field("Δθ by", Num(l.Increment.Theta), act(OpIncrement, Links.KTheta), true, false))
		row = append(row, signButtons(lockOf, press, Links.KTheta)...)
		row = append(row, Button("unpick", act(OpUnpick, ""), false), Button(mark, act(OpLocking, ""), l.Locking))
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return Group{Title: "picked", Rows: [][]Cell{{Text("no edges picked")}}}
	}
	return Group{Title: "picked — + / − steps Δ; ● then + or − locks that way, a lit one unlocks", Rows: append([][]Cell{allRow(m)}, rows...)}
}

const (
	OpAllStep    = "all-step"
	OpAllBoth    = "all-both"
	OpAllLocking = "all-locking"
)

func allRow(m Model) []Cell {
	st := m.Step
	lockOf := func(k string) int {
		if k == Links.KTheta {
			return st.LockAll.Theta
		}
		return st.LockAll.Phi
	}
	press := func(k string, sign int) Act { return Act{Op: OpAllStep, K: k, Sign: sign} }
	locked := st.LockAll != (Links.Lock{})
	for _, l := range m.Picked() {
		locked = locked || l.Lock != (Links.Lock{})
	}
	up, down := "+", "−"
	if locked {
		up, down = "▲", "▼"
	}
	mark := "○"
	if st.Locking {
		mark = "●"
	}
	row := []Cell{Text("all"), Text("Δφ")}
	row = append(row, signButtons(lockOf, press, Links.KPhi)...)
	row = append(row, Text("Δθ"))
	row = append(row, signButtons(lockOf, press, Links.KTheta)...)
	row = append(row, Button(up+" both", Act{Op: OpAllBoth, Sign: 1}, false), Button(down+" both", Act{Op: OpAllBoth, Sign: -1}, false))
	if !m.RowsLocking() {
		row = append(row, Button(mark, Act{Op: OpAllLocking}, st.Locking))
	}
	return row
}
