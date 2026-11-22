package Builder

import (
	"math"
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/LinksBoard"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const inspectHint = "Click a link row, or an edge in the scene, to inspect it here."

func endsAt(l Links.Link, k string) string {
	sum := l.Sum()
	switch k {
	case Links.KPhi:
		return strconv.Itoa(l.Start.Tick.Phi + sum.Phi)
	case Links.KTheta:
		return strconv.Itoa(l.Start.Tick.Theta + sum.Theta)
	}
	return tenths(l.Start.R + float64(sum.R))
}

func boardInspector(m Model) LinksBoard.Inspector {
	e := m.Touched
	if e.Step == 0 {
		e = m.firstStart()
	}
	if e.Step == startPart {
		return m.startInspector(e)
	}
	l, ok := m.gridLink(e)
	if e.Step <= 0 || !ok {
		return LinksBoard.Inspector{Hint: inspectHint}
	}
	on := map[bool]string{true: "● on", false: "○ off"}[l.Picked]
	in := LinksBoard.Inspector{Title: l.From + " → " + l.To, Sub: "· step " + m.StepLabel(e.Step), Moves: "moves in this step",
		Switch: Board.Chip{Label: on, Act: Act{Op: OpStepPick, Node: l.From, To: l.To, I: e.Step}, Lit: l.Picked}}
	if l.Picked {
		in.Switch, in.Moves = Board.Chip{}, ""
	} else {
		in.Note = "Not picked in this step. Turn it on to give it a turn."
		return in
	}
	act := func(op, k string) Act { return Act{Op: op, Node: l.From, To: l.To, K: k, I: e.Step} }
	for _, a := range axes {
		press := act(OpLinkPress, a.k)
		x := LinksBoard.Axis{Label: a.label + " turn", Short: a.label, IncTag: "Δ per press",
			Turn: stepper(signed(a.delta(l)), press, false), Inc: stepper(itoa(a.times(l)), act(OpIncrement, a.k), true)}
		if a.k == Links.KR {
			x.Label = "r"
			m.rAxis(&x, l)
			in.Axes = append(in.Axes, x)
			continue
		}
		d := boardDial(axisDial(l, a.k))
		d.Inc = float32(a.times(l))
		x.Dial, x.DialSize = d, 72
		x.Ends = a.label + " " + a.at(l) + " → " + endsAt(l, a.k)
		if a.k == Links.KPhi {
			x.DialSize = 120
			end, _ := strconv.Atoi(endsAt(l, a.k))
			s := m.SceneS()
			x.Lines = []string{itoa(end) + " × s" + itoa(s) + " = " + itoa(end*s) + " ticks"}
		}
		in.Axes = append(in.Axes, x)
	}
	return in
}

func (m Model) rAxis(x *LinksBoard.Axis, l Links.Link) {
	start, end := l.Start.R, l.Start.R+float64(l.Sum().R)
	fixed := m.FixedR(l)
	x.Ruler = &Board.Ruler{Start: float32(start), Delta: float32(end - start), Most: float32(max(start, end, 1) * 1.25), Fixed: fixed}
	x.Sub = "tenths of a bead slot · never below 0"
	x.RulerTags = [3]string{"0", "start " + tenths(start), tenths(end)}
	x.Ends = "r " + tenths(start) + " → " + tenths(end)
	atZero := end < 1e-9
	if fixed {
		x.Fixed = m.fixedWhy(l)
	} else if atZero {
		x.Lines = []string{"at 0, the closest it can go"}
	}
}

func (m Model) firstStart() StepEdge {
	for _, id := range m.Nodes {
		if links := m.linksTo(id); len(links) > 0 {
			return StepEdge{startPart, links[0].From, links[0].To}
		}
	}
	return StepEdge{}
}

func (m Model) startInspector(e StepEdge) LinksBoard.Inspector {
	l, ok := m.Link(e.From, e.To)
	if !ok {
		return LinksBoard.Inspector{Hint: inspectHint}
	}
	l.Start = m.Grids[e.To].Base[e.From].Origin
	in := LinksBoard.Inspector{Title: e.From + " → " + e.To, Sub: "· START"}
	act := func(op, k string) Act { return Act{Op: op, Node: l.From, To: l.To, K: k, I: startPart} }
	for _, a := range axes {
		x := LinksBoard.Axis{Label: a.label + " start", Short: a.label, IncTag: "1",
			Turn: stepper(a.at(l), act(OpLinkStart, a.k), true), Inc: stepper(itoa(a.one(l)), act(OpLinkOne, a.k), false)}
		switch a.k {
		case Links.KPhi:
			x.Dial, x.DialSize = Board.Dial{Axis: Board.Phi, Start: float32(l.Start.Tick.Phi), Labels: true}, 92
		case Links.KTheta:
			x.Dial, x.DialSize = Board.Dial{Axis: Board.Theta, Start: float32(l.Start.Tick.Theta)}, 56
		}
		in.Axes = append(in.Axes, x)
	}
	return in
}

func tenths(v float64) string {
	t := math.Round(v * 10)
	s := strconv.FormatInt(int64(math.Abs(t))/10, 10)
	if d := int64(math.Abs(t)) % 10; d != 0 {
		s += "." + strconv.FormatInt(d, 10)
	}
	if t < 0 {
		s = "-" + s
	}
	return s
}

type axis struct {
	label, k          string
	at                func(Links.Link) string
	delta, times, one func(Links.Link) int
}

var axes = []axis{
	{"φ", Links.KPhi, func(l Links.Link) string { return strconv.Itoa(l.Start.Tick.Phi) }, func(l Links.Link) int { return l.Delta.Phi }, func(l Links.Link) int { return l.Increment.Phi }, func(l Links.Link) int { return l.Start.One.Phi }},
	{"θ", Links.KTheta, func(l Links.Link) string { return strconv.Itoa(l.Start.Tick.Theta) }, func(l Links.Link) int { return l.Delta.Theta }, func(l Links.Link) int { return l.Increment.Theta }, func(l Links.Link) int { return l.Start.One.Theta }},
	{"r", Links.KR, func(l Links.Link) string { return tenths(l.Start.R) }, func(l Links.Link) int { return l.Delta.R }, func(l Links.Link) int { return l.Increment.R }, func(l Links.Link) int { return l.Start.One.R }},
}

func linkCount(n int) string {
	switch n {
	case 0:
		return "no links"
	case 1:
		return "1 link"
	}
	return strconv.Itoa(n) + " links"
}
