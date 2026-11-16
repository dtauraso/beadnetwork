package Builder

import (
	"math"
	"strconv"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/LinksBoard"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

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

func stepper(value string, a Act, typed bool) Board.Stepper {
	s := Board.Stepper{Steps: a, Plain: value}
	if typed {
		s.Field = Board.Field{Value: value, Fit: "-00", Act: a}
	}
	return s
}

func totalOf(l Links.Link, k string) int {
	switch k {
	case Links.KPhi:
		return l.Total.Phi
	case Links.KTheta:
		return l.Total.Theta
	}
	return l.Total.R
}

func (m Model) stepCard(step int, l Links.Link) *LinksBoard.Card {
	if !l.Picked {
		return &LinksBoard.Card{Note: "Not picked in this step. Pick it (●) to give it a turn."}
	}
	c := &LinksBoard.Card{Heads: [3]string{"Δ", "nΔ", "total"}, StartHead: "start"}
	act := func(op, k string) Act { return Act{Op: op, Node: l.From, To: l.To, K: k, I: step} }
	from := l
	if b, ok := m.Grids[l.To].Base[l.From]; ok {
		from.Start = b.Origin
	}
	for _, a := range axes {
		press := act(OpLinkPress, a.k)
		r := LinksBoard.CardRow{Axis: a.label, Start: stepper(a.at(from), Act{Op: OpLinkStart, Node: l.From, To: l.To, K: a.k, I: startPart}, true), A: stepper(signed(a.delta(l)), press, false), B: stepper(itoa(a.times(l)), act(OpIncrement, a.k), true), Total: itoa(totalOf(l, a.k))}
		if a.k == Links.KR {
			m.rRow(c, &r, l)
		} else {
			d := boardDial(axisDial(l, a.k))
			if m.Turning.Act == press {
				d.Ghost = float32(m.Turning.presses() * m.Turning.Inc)
			}
			d.Inc = float32(a.times(l))
			r.Dial, r.DialAct = &d, press
		}
		c.Rows = append(c.Rows, r)
	}
	return c
}

func (m Model) rRow(c *LinksBoard.Card, r *LinksBoard.CardRow, l Links.Link) {
	start, end := l.Start.R, l.Start.R+float64(l.Sum().R)
	fixed := m.FixedR(l)
	r.Ruler = &Board.Ruler{Start: float32(start), Delta: float32(end - start), Most: float32(max(start, end) * 1.25), Fixed: fixed}
	atZero := end < 1e-9
	r.A.DownDim, r.A.UpDim = fixed || atZero, fixed
	if fixed {
		c.Fixed = m.fixedWhy(l)
	} else if atZero {
		c.Foot = "r is at 0, the closest it can go"
	}
}

func (m Model) startBoardCard(id string, l Links.Link) *LinksBoard.Card {
	b := m.Grids[id].Base[l.From]
	l.Start = b.Origin
	c := &LinksBoard.Card{Heads: [3]string{"start", "1", ""}}
	act := func(op, k string) Act { return Act{Op: op, Node: l.From, To: l.To, K: k, I: startPart} }
	for _, a := range axes {
		r := LinksBoard.CardRow{Axis: a.label, A: stepper(a.at(l), act(OpLinkStart, a.k), true), B: stepper(itoa(a.one(l)), act(OpLinkOne, a.k), false)}
		switch a.k {
		case Links.KPhi:
			r.Dial = &Board.Dial{Axis: Board.Phi, Start: float32(l.Start.Tick.Phi)}
		case Links.KTheta:
			r.Dial = &Board.Dial{Axis: Board.Theta, Start: float32(l.Start.Tick.Theta)}
		}
		c.Rows = append(c.Rows, r)
	}
	return c
}

func (m Model) fixedWhy(l Links.Link) string {
	for _, id := range m.ManyIn() {
		if id != l.From && id != l.To {
			continue
		}
		var in []string
		for _, f := range m.linksTo(id) {
			in = append(in, f.Name())
		}
		return id + " has " + strconv.Itoa(len(in)) + " links in (" + strings.Join(in, ", ") + "), so its distance is held."
	}
	return ""
}
