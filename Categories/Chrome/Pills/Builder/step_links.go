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

func IsStepEdit(a Act) bool {
	_, ok := editOps[a.Op]
	return a.I > 0 && ok
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
