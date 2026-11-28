package Builder

import (
	"slices"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/StepStore"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

type Edge = StepStore.Edge

type Amount = StepStore.Amount

func (m Model) stepPath(step int) []int { return append(StepStore.LevelPath(m.Level), step) }

func (m Model) stepDir(p []int) string { return StepStore.PathDir(m.Root, p) }

func (m *Model) ReadAmounts() {
	m.Amts = map[string]map[Edge]Amount{}
	m.Leaves = nil
	if m.Root == "" {
		return
	}
	m.Leaves = StepStore.Leaves(m.Root, nil)
	for _, p := range m.Leaves {
		m.Amts[StepStore.PathKey(p)] = StepStore.ReadAmounts(m.stepDir(p))
	}
}

func under(p, parent []int) bool {
	return len(p) > len(parent) && slices.Equal(p[:len(parent)], parent)
}

func (m Model) IsParent(p []int) bool {
	for _, l := range m.Leaves {
		if under(l, p) {
			return true
		}
	}
	return false
}

func (m Model) StepIsParent(step int) bool { return m.IsParent(m.stepPath(step)) }

func (m Model) amountAt(p []int, e Edge) Amount {
	if a, ok := m.Amts[StepStore.PathKey(p)][e]; ok {
		return a
	}
	return StepStore.NewAmount()
}

func (m Model) AmountOf(step int, e Edge) (Amount, bool) {
	p := m.stepPath(step)
	if m.IsParent(p) {
		return Amount{}, false
	}
	if _, ok := m.Link(e.From, e.To); !ok {
		return Amount{}, false
	}
	return m.amountAt(p, e), true
}

func (m Model) upTo(shown []int, all bool) [][]int {
	var out [][]int
	for _, l := range m.Leaves {
		c := slices.Compare(l, shown)
		if c < 0 || slices.Equal(l, shown) || (all && under(l, shown)) {
			out = append(out, l)
		}
	}
	return out
}

func (m Model) sumOver(leaves [][]int, e Edge) Links.Ints {
	var s Links.Ints
	for _, p := range leaves {
		if a := m.amountAt(p, e); a.Picked {
			s = s.Add(a.D)
		}
	}
	return s
}

func (m Model) pickLeaf(shown []int) []int {
	if !m.IsParent(shown) {
		return shown
	}
	leaves := m.upTo(shown, true)
	return leaves[len(leaves)-1]
}

func (m Model) Through(step int, e Edge) Links.Ints {
	return m.sumOver(m.upTo(m.stepPath(step), true), e)
}

func (m Model) shownPath() []int { return m.stepPath(m.StepShown()) }

func (m Model) target(e Edge) (Links.Ints, bool) {
	shown := m.shownPath()
	return m.sumOver(m.upTo(shown, true), e), m.amountAt(m.pickLeaf(shown), e).Picked
}

func (m *Model) sends(want func(Edge) (Links.Ints, bool)) []Links.Envelope {
	var out []Links.Envelope
	if m.Sent == nil {
		m.Sent = map[Edge]Links.Ints{}
	}
	for _, id := range m.Nodes {
		links := m.linksTo(id)
		var turns []Links.Envelope
		pick, change := "", false
		for _, l := range links {
			e := Edge{From: l.From, To: l.To}
			t, on := want(e)
			if d := t.Sub(m.Sent[e]); !d.Zero() {
				turns = append(turns, Links.Envelope{To: id, Msg: Links.PickReq{From: l.From}},
					Links.Envelope{To: id, Msg: Links.Amount{From: l.From, D: d}},
					Links.Envelope{To: id, Msg: Links.Unpick{From: l.From}})
				m.Sent[e] = t
			}
			if on {
				pick = l.From
			}
			change = change || on != l.Picked
		}
		if len(turns) == 0 && !change {
			continue
		}
		for _, l := range links {
			out = append(out, Links.Envelope{To: id, Msg: Links.Unpick{From: l.From}})
		}
		out = append(out, turns...)
		if pick != "" {
			out = append(out, Links.Envelope{To: id, Msg: Links.PickReq{From: pick}})
		}
	}
	return out
}

func (m *Model) Sync() []Links.Envelope { return m.sends(m.target) }

func (m *Model) setAmount(p []int, e Edge, a Amount) error {
	key := StepStore.PathKey(p)
	if m.Amts == nil {
		m.Amts = map[string]map[Edge]Amount{}
	}
	if m.Amts[key] == nil {
		m.Amts[key] = map[Edge]Amount{}
	}
	m.Amts[key][e] = a
	if !slices.ContainsFunc(m.Leaves, func(l []int) bool { return slices.Equal(l, p) }) {
		m.Leaves = append(m.Leaves, slices.Clone(p))
		slices.SortFunc(m.Leaves, slices.Compare)
	}
	if m.Root == "" {
		return nil
	}
	return StepStore.WriteAmount(m.stepDir(p), e, a)
}

func (m *Model) RemoveEdges(drop func(Edge) bool) error {
	for _, p := range m.Leaves {
		for e := range m.Amts[StepStore.PathKey(p)] {
			if !drop(e) {
				continue
			}
			delete(m.Amts[StepStore.PathKey(p)], e)
			if err := StepStore.RemoveAmount(m.stepDir(p), e); err != nil {
				return err
			}
		}
	}
	return nil
}
