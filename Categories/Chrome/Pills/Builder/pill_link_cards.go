package Builder

import (
	"math"
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/LinksBoard/Fill"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const OpCardScroll = "card-scroll"

func (m Model) scrollCard(c *Fill.Card, e StepEdge) {
	if c == nil {
		return
	}
	c.Scroll, c.SX = Act{Op: OpCardScroll, Node: e.From, To: e.To, I: e.Step}, m.CardSX[e]
}

func (m *Model) CardWheel(pl Sheet.Layout, x, y, dx, dy float64) bool {
	if math.Abs(dx) <= math.Abs(dy) {
		return false
	}
	for _, p := range pl.Pills {
		l, ok := p.Board.(*Board.Laid)
		if !ok {
			continue
		}
		for _, s := range l.Spots {
			a := s.Cell.Act
			if a.Op != OpCardScroll || !l.In(s, x, y) {
				continue
			}
			e := StepEdge{a.I, a.Node, a.To}
			if m.CardSX == nil {
				m.CardSX = map[StepEdge]float32{}
			}
			m.CardSX[e] = Board.Clamp(m.CardSX[e]+float32(dx), float32(a.Sign))
			return true
		}
	}
	return false
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
