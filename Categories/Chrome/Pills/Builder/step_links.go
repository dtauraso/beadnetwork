package Builder

import "github.com/dtauraso/beadnetwork/Categories/Node/Links"

func (m Model) gridLinks(step int, id string) []Links.Link {
	if step == 0 {
		return m.Col(id).Links()
	}
	c, ok := m.StepCol(step, id)
	if !ok {
		return nil
	}
	return c.Links()
}

func (m Model) gridLink(e StepEdge) (Links.Link, bool) {
	c, ok := m.StepCol(e.Step, e.To)
	if !ok {
		return Links.Link{}, false
	}
	in, ok := c.InFrom(e.From)
	return in.Link, ok
}

type Walk struct {
	Steps []int
	I     int
	Round int
	From  []int
	Then  []Links.Envelope
}

func (m Model) Walking() bool { return m.Walk.I < len(m.Walk.Steps) }

func (m *Model) StartWalk() {
	m.Walk = Walk{Steps: []int{m.StepShown()}, From: m.StepFrom, Then: m.Walk.Then}
	m.StepFrom = nil
}

func (m *Model) WalkShow() Links.ShowStep {
	s := Links.ShowStep{Level: m.Level, Step: m.Walk.Steps[m.Walk.I], Steps: m.stepNums()}
	if m.Walk.I == 0 {
		s.From = m.Walk.From
	}
	return s
}

func (m Model) WalkSettled() bool {
	for _, id := range m.Nodes {
		if m.Rounds[id] < m.Walk.Round {
			return false
		}
	}
	return true
}

func (m Model) SNeeded() int {
	need := 1
	for _, g := range m.Grids {
		for _, c := range g.Cols {
			for _, l := range c.Links() {
				need = Links.Lcm(need, l.Start.TicksNeeded())
			}
		}
	}
	return need
}

var editOps = map[string]string{
	OpLinkStart: Links.EditStart, OpLinkOne: Links.EditOne, OpLinkPress: Links.EditPress,
	OpIncrement: Links.EditInc, OpStepPick: Links.EditPick,
}

func IsStartEdit(a Act) bool { return a.I == startPart && (a.Op == OpLinkStart || a.Op == OpLinkOne) }

func IsStepEdit(a Act) bool {
	_, ok := editOps[a.Op]
	return ok && (a.I > 0 || IsStartEdit(a))
}

const (
	startPart   = -1
	OpStartFold = "start-fold"
)

func startRows(t *tree, m Model) {
	fold := Act{Op: OpStartFold}
	t.add(linkRow{id: fold, open: openOf(m.StartOpen), fold: fold, name: Text("start"),
		detail: []Cell{Text("where the sequence begins · each link's START and 1").Dimmed(true)}})
	if !m.StartOpen {
		return
	}
	for _, id := range m.Nodes {
		fold := Act{Op: OpStepFold, Node: id, I: startPart}
		folded := m.StepFolded(startPart, id)
		links := m.linksTo(id)
		r := linkRow{id: fold, depth: 1, open: openOf(!folded), fold: fold, name: Text("node " + id)}
		if len(links) == 0 {
			r.open = leaf
		}
		t.add(r)
		if folded {
			continue
		}
		for _, l := range links {
			e := StepEdge{startPart, l.From, l.To}
			t.add(m.cardRow(linkRowOf(l, false, Act{}, 2), e, false))
			if m.CardOpen(e, false) {
				t.put(startCard(m.Grids[id].Base[l.From], l, 3))
			}
		}
	}
}

func startCard(b Links.Base, l Links.Link, depth int) Group {
	l.Start = b.Origin
	act := func(op, k string) Act { return Act{Op: op, Node: l.From, To: l.To, K: k, I: startPart} }
	head := func(s string) Cell { return Text(s).AboveLastDigit().Inked(inkHead) }
	rows := [][]Cell{{Text(""), head("START"), head("1"), head("Δ so far")}}
	total := map[string]int{Links.KPhi: b.Total.Phi, Links.KTheta: b.Total.Theta, Links.KR: b.Total.R}
	for _, a := range axes {
		rows = append(rows, []Cell{
			Text(a.label).Dimmed(true),
			Field("", a.at(l), act(OpLinkStart, a.k), true, true).Fitting(cardFit).Outline(false),
			Field("", itoa(a.one(l)), act(OpLinkOne, a.k), false, true).Fitting(cardFit).Outline(false),
			Text(itoa(total[a.k])).Dimmed(true),
		})
	}
	return Group{Rows: rows, Card: true, Depth: depth}
}

func StepEditOf(a Act, sign int, typed *float64) Links.StepEdit {
	e := Links.StepEdit{Step: a.I, From: a.Node, Op: editOps[a.Op], K: a.K, Sign: sign}
	if a.Op == OpLinkPress && sign == 0 {
		e.Sign = a.Sign
	}
	if typed != nil {
		e.Typed, e.IsTyped = *typed, true
	}
	return e
}
