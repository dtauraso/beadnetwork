package Builder

import (
	"math"
	"strconv"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const OpScoreCell = "score-cell"

type Turning struct {
	Act        Act
	X, Y       float64
	Last, Turn float64
	From, Inc  int
}

func twelfths(dx, dy float64) float64 { return math.Atan2(dx, -dy) / (2 * math.Pi) * Links.TurnTicks }

func (m *Model) StartTurn(c Cell, r Sheet.Rect, x, y float64) bool {
	l, ok := m.gridLink(StepEdge{c.Act.I, c.Act.Node, c.Act.To})
	inc := *l.Increment.At(c.Act.K)
	if !ok || c.Act.Op != OpLinkPress || c.Dial.Axis == Sheet.DialNone || inc <= 0 {
		return false
	}
	d := *l.Delta.At(c.Act.K)
	cx, cy := float64(r.X+r.H/2), float64(r.Y+r.H/2)
	m.Turning = Turning{Act: c.Act, X: cx, Y: cy, Last: twelfths(x-cx, y-cy), Turn: float64(d), From: d, Inc: inc}
	return true
}

func (m *Model) TurnTo(x, y float64) bool {
	t := &m.Turning
	if t.Act.Op == "" {
		return false
	}
	u := twelfths(x-t.X, y-t.Y)
	d := math.Mod(u-t.Last+1.5*Links.TurnTicks, Links.TurnTicks) - Links.TurnTicks/2
	t.Turn, t.Last = t.Turn+d, u
	return true
}

func (t Turning) presses() int { return int(math.Round((t.Turn - float64(t.From)) / float64(t.Inc))) }

func (m *Model) EndTurn() (Act, int, bool) {
	t := m.Turning
	m.Turning = Turning{}
	return t.Act, t.presses(), t.Act.Op != ""
}

func (m Model) dialCell(l Links.Link, k string, press Act) Cell {
	d := axisDial(l, k)
	if m.Turning.Act == press {
		d.Ghost = float32(m.Turning.presses() * m.Turning.Inc)
	}
	return Text("").Acting(press).Dialed(d).Tipped("drag around the dial to set the turn · it snaps to Δ per press")
}

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
	rows = append(rows, []Cell{Text(""), head("turn"), Text(""), head("Δ per press"), head("start"), head("ends at")})
	act := func(op, k string) Act { return Act{Op: op, Node: l.From, To: l.To, K: k, I: e.Step} }
	for _, a := range axes {
		dim := a.k == Links.KR && m.FixedR(l)
		dial := Text("")
		if a.k != Links.KR {
			dial = m.dialCell(l, a.k, act(OpLinkPress, a.k))
		}
		rows = append(rows, []Cell{
			Text(a.label).Dimmed(true),
			Field("", signed(a.delta(l)), act(OpLinkPress, a.k), false, true).Fitting(cardFit).Outline(false).Dimmed(dim),
			dial,
			Field("", strconv.Itoa(a.times(l)), act(OpIncrement, a.k), true, true).Fitting(cardFit).Outline(false),
			Text(a.at(l)).Dimmed(true),
			Text(endsAt(l, a)).Inked(inkStrong),
		})
	}
	if note := m.rNote(l); note != "" {
		rows = append(rows, []Cell{Text(note).Dimmed(true)})
	}
	return Group{Rows: rows, Card: true, Wraps: true}
}

func (m Model) rNote(l Links.Link) string {
	for _, id := range m.ManyIn() {
		if id != l.From && id != l.To {
			continue
		}
		var in []string
		for _, f := range m.linksTo(id) {
			in = append(in, f.Name())
		}
		return "🔒 r is fixed: " + id + " has " + strconv.Itoa(len(in)) + " links in (" + strings.Join(in, ", ") + "), so its distance is held"
	}
	if l.Start.R+float64(l.Delta.R) < 1e-9 {
		return "r is at 0, the closest it can go"
	}
	return ""
}
