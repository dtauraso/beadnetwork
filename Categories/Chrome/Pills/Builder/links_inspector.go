package Builder

import (
	"strconv"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/LinksBoard"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

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
		return &LinksBoard.Card{Note: "Not picked in this step. Click “pick” to give it a turn."}
	}
	c := &LinksBoard.Card{Heads: [3]string{"this step turns", "one press", "ends at"}, StartHead: "starts at", Foot: "“starts at” is shared by every step of this sequence."}
	act := func(op, k string) Act { return Act{Op: op, Node: l.From, To: l.To, K: k, I: step} }
	from := l
	if b, ok := m.Grids[l.To].Base[l.From]; ok {
		from.Start = b.Origin
	}
	for _, a := range axes {
		press := act(OpLinkPress, a.k)
		r := LinksBoard.CardRow{Axis: a.label, Start: stepper(a.at(from), Act{Op: OpLinkStart, Node: l.From, To: l.To, K: a.k, I: startPart}, true), A: stepper(signed(a.delta(l)), press, false), B: stepper(itoa(a.times(l)), act(OpIncrement, a.k), true), Total: itoa(totalOf(l, a.k))}
		r.A.Faded = a.delta(l) == 0
		if a.k == Links.KR {
			m.rRow(c, &r, l)
		} else {
			d := boardDial(axisDial(l, a.k))
			d.Inc = float32(a.times(l))
			r.Dial = &d
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
		r.Why = "r fixed: " + m.fixedShort(l)
	} else if atZero {
		c.Foot = "r is at 0, the closest it can go. " + c.Foot
	}
}

func (m Model) startBoardCard(id string, l Links.Link) *LinksBoard.Card {
	b := m.Grids[id].Base[l.From]
	l.Start = b.Origin
	c := &LinksBoard.Card{Heads: [3]string{"start", "1", l.From + "'s reference"}}
	act := func(op, k string) Act { return Act{Op: op, Node: l.From, To: l.To, K: k, I: startPart} }
	ref := m.Col(l.From).Ref
	turn := func(k string, v int) Board.Stepper {
		return stepper(itoa(v), Act{Op: OpNodeRef, Node: l.From, K: k}, true)
	}
	for _, a := range axes {
		r := LinksBoard.CardRow{Axis: a.label, A: stepper(a.at(l), act(OpLinkStart, a.k), true), B: stepper(itoa(a.one(l)), act(OpLinkOne, a.k), false)}
		switch a.k {
		case Links.KPhi:
			r.Dial = &Board.Dial{Axis: Board.Phi, Start: float32(l.Start.Tick.Phi)}
			r.C = turn(Links.KPhi, ref.Phi)
		case Links.KTheta:
			r.Dial = &Board.Dial{Axis: Board.Theta, Start: float32(l.Start.Tick.Theta)}
			r.C = turn(Links.KTheta, ref.Theta)
		}
		c.Rows = append(c.Rows, r)
	}
	return c
}

func (m Model) fixedShort(l Links.Link) string {
	for _, id := range m.ManyIn() {
		if id == l.From || id == l.To {
			return "node " + id + " has " + strconv.Itoa(len(m.linksTo(id))) + " links in"
		}
	}
	return ""
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
