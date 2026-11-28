package Builder

import (
	"strconv"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/LinksBoard/Fill"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const notTurned = "not turned"

func signed(n int) string {
	if n > 0 {
		return "+" + strconv.Itoa(n)
	}
	return strings.Replace(strconv.Itoa(n), "-", "−", 1)
}

func dialOf(l Links.Link) Sheet.Dial {
	if l.Delta.Phi == 0 && l.Delta.Theta != 0 {
		return axisDial(l, Links.KTheta)
	}
	return axisDial(l, Links.KPhi)
}

func axisDial(l Links.Link, k string) Sheet.Dial {
	if k == Links.KTheta {
		return Sheet.Dial{Axis: Sheet.DialTheta, Turn: float32(l.Delta.Theta)}
	}
	return Sheet.Dial{Axis: Sheet.DialPhi, Turn: float32(l.Delta.Phi)}
}

func boardDial(d Sheet.Dial) Board.Dial {
	p := Board.Phi
	if d.Axis == Sheet.DialTheta {
		p = Board.Theta
	}
	return Board.Dial{Axis: p, Start: d.Start, Turn: d.Turn}
}

func lockOf(m Model, step int, a axis) Fill.Lock {
	st := m.HeadOf(step)
	way, on := st.At(a.k)
	glyph := "·"
	if *on && *way > 0 {
		glyph = "↑"
	} else if *on && *way < 0 {
		glyph = "↓"
	}
	return Fill.Lock{Axis: a.label, Label: a.label + " " + glyph, On: *on, Cycle: Act{Op: OpAllLockCycle, K: a.k, I: step}, Press: Act{Op: OpAllLock, K: a.k, I: step}}
}

func boardCols(m Model) []Fill.Column {
	out := []Fill.Column{m.startCol()}
	for _, n := range m.Numbers() {
		out = append(out, m.stepCol(n))
	}
	return out
}

func (m Model) stepCol(n int) Fill.Column {
	c := Fill.Column{Label: "step " + itoa(n), Short: itoa(n), Sub: m.Names[n], Open: m.StepOpen(n), Shown: n == m.StepShown(),
		Fold: Act{Op: OpStepFold, I: n}}
	if !c.Shown {
		c.Show = Act{Op: OpStepShow, I: n}
	}
	if !c.Open {
		return c
	}
	c.Name = Board.Field{Value: m.Names[n], Fit: "a step name", Act: Act{Op: OpStepName, I: n}, Words: true}
	c.Menu = Act{Op: OpMenu, K: menuStep, I: n}
	c.MenuOpen = m.Menu == c.Menu
	c.Tools = []Board.Chip{
		{Label: "Move earlier", Act: Act{Op: OpStepEarlier, I: n}},
		{Label: "Move later", Act: Act{Op: OpStepLater, I: n}},
		{Label: "Rename", Act: Act{Op: OpRename, K: renameStep, I: n}},
		{Label: "Add step after", Act: Act{Op: OpStepAdd, I: n}},
		{Label: "Duplicate", Act: Act{Op: OpStepCopy, I: n}},
	}
	c.Drop = Board.Chip{Label: "Delete step", Act: Act{Op: OpStepDrop, I: n}}
	hd := &Fill.Head{Picked: itoa(len(m.stepActive(n))) + " picked", Play: playOf(m, n)}
	for _, a := range axes {
		hd.Locks = append(hd.Locks, lockOf(m, n, a))
	}
	c.Head = hd
	if m.StepIsParent(n) {
		return c
	}
	for _, id := range m.Nodes {
		if links := m.linksTo(id); len(links) > 0 {
			c.Nodes = append(c.Nodes, m.nodeOf(n, id, links))
		}
	}
	return c
}

func (m Model) nodeOf(step int, id string, links []Links.Link) Fill.Node {
	var n Fill.Node
	for _, l := range links {
		if step == startPart {
			n.Links = append(n.Links, m.startLink(id, l))
		} else {
			n.Links = append(n.Links, m.linkOf(step, l))
		}
	}
	return n
}

func (m Model) hot(l Links.Link) bool { return m.Hot.Node == l.From && m.Hot.To == l.To }

func (m Model) linkOf(step int, l Links.Link) Fill.Link {
	e := StepEdge{step, l.From, l.To}
	sl, ok := m.stepLink(e)
	lk := Fill.Link{Label: l.From + " → " + l.To, Picked: ok && sl.Picked, Open: m.CardOpen(e), Fold: Act{Op: OpCardFold, Node: l.From, To: l.To, I: step},
		Dial: boardDial(dialOf(sl)), Sum: notTurned, Hot: m.hot(l)}
	if !ok {
		return lk
	}
	lk.Pick = Act{Op: OpStepPick, Node: l.From, To: l.To, I: step}
	if lk.Open {
		lk.Card = m.stepCard(step, sl)
		m.scrollCard(lk.Card, e)
	}
	if !sl.Picked {
		lk.Dial.Turn = 0
		lk.PickOpen = Act{Op: OpLinkPickOpen, Node: l.From, To: l.To, I: step}
		lk.Gap = Act{Op: OpLinkGap, Node: l.From, To: l.To, I: step}
		return lk
	}
	var parts []string
	for _, a := range axes {
		if d := a.delta(sl); d != 0 {
			parts = append(parts, signed(d)+a.label)
		}
	}
	lk.Sum = notTurned
	if len(parts) > 0 {
		lk.Sum = strings.Join(parts, " ")
	}
	return lk
}

func (m Model) startCol() Fill.Column {
	c := Fill.Column{Label: "START", Short: "S", Sub: "where each link begins", Open: m.StartOpen, Fold: Act{Op: OpStartFold}}
	if !c.Open {
		return c
	}
	for _, id := range m.Nodes {
		if links := m.linksTo(id); len(links) > 0 {
			c.Nodes = append(c.Nodes, m.nodeOf(startPart, id, links))
		}
	}
	return c
}

func (m Model) startLink(id string, l Links.Link) Fill.Link {
	e := StepEdge{startPart, l.From, l.To}
	o := l.Start
	lk := Fill.Link{Label: l.From + " → " + l.To, Open: m.CardOpen(e), Fold: Act{Op: OpCardFold, Node: l.From, To: l.To, I: startPart},
		Dial: Board.Dial{Axis: Board.Phi, Start: float32(o.Tick.Phi)}, Sum: "φ" + itoa(o.Tick.Phi) + " θ" + itoa(o.Tick.Theta) + " r" + tenths(o.R),
		Hot: m.hot(l)}
	shown := m.StepShown()
	if sl, ok := m.stepLink(StepEdge{shown, l.From, l.To}); ok {
		lk.Picked, lk.Pick = sl.Picked, Act{Op: OpStepPick, Node: l.From, To: l.To, I: shown}
	}
	if lk.Open {
		lk.Card = m.startBoardCard(id, l)
		m.scrollCard(lk.Card, e)
	}
	return lk
}
