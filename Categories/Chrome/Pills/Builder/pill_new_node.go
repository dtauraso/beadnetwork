package Builder

import (
	"fmt"
	"slices"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/LinksBoard"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const (
	OpNewFold   = "new-fold"
	OpNewCenter = "new-center"
	OpNewAt     = "new-at"
	OpNewAdd    = "new-add"
)

type NewNode struct {
	Center string
	At     Links.Start
	Set    bool
}

func (m Model) NewCenter() string {
	if slices.Contains(m.Nodes, m.New.Center) {
		return m.New.Center
	}
	if len(m.Nodes) == 0 {
		return ""
	}
	return m.Nodes[0]
}

func (m Model) newAt() Links.Start {
	if !m.New.Set {
		return Links.Start{R: 1}
	}
	return m.New.At
}

func (m Model) NewPlace() Links.Place {
	col := m.Col(m.NewCenter())
	v := Links.RefTurn(col.Ref).Rotate(m.Frame.Vec(1, m.newAt()))
	return m.Frame.GridAt(m.Frame.At(col.At).Add(v))
}

func (m *Model) NewEdit(a Act, sign int, typed *float64) {
	at := m.newAt()
	v := map[string]*float64{Links.KR: &at.R, Links.KPhi: &at.Phi, Links.KTheta: &at.Theta}[a.K]
	if v == nil {
		return
	}
	next := *v + float64(sign)
	if typed != nil {
		next = *typed
	}
	if a.K == Links.KR && next < 0 {
		m.Refused = fmt.Sprintf("new node r %v: it must not be below 0, so it stays at %v", next, *v)
		return
	}
	*v = next
	m.New.At, m.New.Set = at, true
}

func newNodeData(m Model) *LinksBoard.NewNode {
	center, at := m.NewCenter(), m.newAt()
	n := &LinksBoard.NewNode{Back: Act{Op: OpNewFold}, Add: Act{Op: OpNewAdd}}
	for _, id := range m.Nodes {
		n.Centers = append(n.Centers, Board.Chip{Label: id, Act: Act{Op: OpNewCenter, Node: id}, Lit: id == center})
	}
	field := func(k string, v float64, label string) LinksBoard.NetField {
		return LinksBoard.NetField{Name: label, Field: Board.Field{Value: Num(v), Fit: "000.00", Act: Act{Op: OpNewAt, K: k}}}
	}
	n.At = []LinksBoard.NetField{field(Links.KPhi, at.Phi, "φ"), field(Links.KTheta, at.Theta, "θ"), field(Links.KR, at.R, "r")}
	p := m.NewPlace()
	c := m.Frame.At(p)
	n.Lands = "lands at (I) r " + Num(p.R) + " φ " + Num(p.Phi) + " θ " + Num(p.Theta) + " · C " + Num(c.X) + " " + Num(c.Y) + " " + Num(c.Z)
	return n
}
