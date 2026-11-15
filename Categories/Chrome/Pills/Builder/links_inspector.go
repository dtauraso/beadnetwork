package Builder

import (
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const OpScoreCell = "score-cell"

func (m *Model) SelectCell(a Act) { m.ScoreCell = StepEdge{a.I, a.Node, a.To} }

func cellAct(step int, l Links.Link) Act {
	return Act{Op: OpScoreCell, Node: l.From, To: l.To, I: step}
}

func endsAt(l Links.Link, a axis) string {
	sum := l.Sum()
	switch a.k {
	case Links.KPhi:
		return strconv.Itoa(l.Start.Tick.Phi + sum.Phi)
	case Links.KTheta:
		return strconv.Itoa(l.Start.Tick.Theta + sum.Theta)
	}
	return tenths(l.Start.R + float64(sum.R))
}

func inspectorGroup(m Model) Group {
	e := m.ScoreCell
	l, ok := m.gridLink(e)
	if e.Step == 0 || !ok {
		return Group{Rows: [][]Cell{{Text("select a cell in the score to edit it here").Dimmed(true)}}, Card: true}
	}
	pick := Act{Op: OpStepPick, Node: l.From, To: l.To, I: e.Step}
	moves := "off"
	if l.Picked {
		moves = "on"
	}
	rows := [][]Cell{{
		Text(l.From + " → " + l.To).Inked(inkStrong),
		Text("· " + m.StepLabel(e.Step)).Dimmed(true),
		Text("moves in this step").Dimmed(true),
		Button(pickMark(l.Picked)+" "+moves, pick, l.Picked).Tipped("pick or unpick this link for this step"),
	}}
	if !l.Picked {
		rows = append(rows, []Cell{Text("turn it on to give this link a turn in this step").Dimmed(true)})
		return Group{Rows: rows, Card: true}
	}
	head := func(s string) Cell { return Text(s).AboveLastDigit().Inked(inkHead) }
	rows = append(rows, []Cell{Text(""), head("turn"), head("Δ per press"), head("start"), head("ends at")})
	act := func(op, k string) Act { return Act{Op: op, Node: l.From, To: l.To, K: k, I: e.Step} }
	for _, a := range axes {
		dim := a.k == Links.KR && m.FixedR(l)
		rows = append(rows, []Cell{
			Text(a.label).Dimmed(true),
			Field("", signed(a.delta(l)), act(OpLinkPress, a.k), false, true).Fitting(cardFit).Outline(false).Dimmed(dim),
			Field("", strconv.Itoa(a.times(l)), act(OpIncrement, a.k), true, true).Fitting(cardFit).Outline(false),
			Text(a.at(l)).Dimmed(true),
			Text(endsAt(l, a)).Inked(inkStrong),
		})
	}
	return Group{Rows: rows, Card: true}
}
