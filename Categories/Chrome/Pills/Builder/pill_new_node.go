package Builder

import (
	"fmt"
	"math"
	"slices"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/LinksBoard/Fill"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const (
	OpNewFold   = "new-fold"
	OpNewCenter = "new-center"
	OpNewAt     = "new-at"
	OpNewAdd    = "new-add"
	OpNodeDrop  = "node-drop"
)

type NewNode struct {
	Center string
	At     Links.Angles
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

func (m Model) newAt() Links.Angles {
	if !m.New.Set {
		return Links.Angles{R: 1}
	}
	return m.New.At
}

func (m Model) NewPlace() Links.Place {
	col := m.Node(m.NewCenter())
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
	if a.K != Links.KR && next != math.Trunc(next) {
		m.Refused = fmt.Sprintf("new node %s %v: not a whole number of ticks, so it stays at %v", map[string]string{Links.KPhi: "φ", Links.KTheta: "θ"}[a.K], next, *v)
		return
	}
	if a.K == Links.KR && next < 0 {
		m.Refused = fmt.Sprintf("new node r %v: it must not be below 0, so it stays at %v", next, *v)
		return
	}
	*v = next
	m.New.At, m.New.Set = at, true
}

func newNodeData(m Model) *Fill.NewNode {
	center, at := m.NewCenter(), m.newAt()
	n := &Fill.NewNode{Add: Act{Op: OpNewAdd}}
	for _, id := range m.Nodes {
		n.Centers = append(n.Centers, Board.Chip{Label: id, Act: Act{Op: OpNewCenter, Node: id}, Lit: id == center})
		n.Drops = append(n.Drops, Board.Chip{Label: "✕ " + id, Act: Act{Op: OpNodeDrop, Node: id}})
	}
	field := func(k string, v float64, label string) Fill.NetField {
		return Fill.NetField{Name: label, Field: Board.Field{Value: Num(v), Fit: "000.00", Act: Act{Op: OpNewAt, K: k}}}
	}
	n.At = []Fill.NetField{field(Links.KPhi, at.Phi, "φ"), field(Links.KTheta, at.Theta, "θ"), field(Links.KR, at.R, "r")}
	return n
}
