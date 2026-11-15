package Builder

import (
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const (
	OpFramePop   = "frame-pop"
	OpFrameDraft = "frame-draft"
	OpFrameApply = "frame-apply"
)

func (m *Model) StageFrame(k string, v float64) {
	if m.FrameDraft == nil {
		m.FrameDraft = map[string]float64{}
	}
	m.FrameDraft[k] = v
}

func scopeBar(m Model) Group {
	rows := [][]Cell{{
		Field("s", strconv.Itoa(m.SceneS()), frameAct(FrameS), true, false).Tipped("ticks a twelfth of a turn, one s every node uses — it rises when a start needs more"),
		Field("Δt", strconv.Itoa(m.TickMs), Act{Op: OpTickMs}, true, true).Tipped("ms one human tick lasts"),
		Button("frame ⟳", Act{Op: OpFramePop}, m.FramePop).Tipped("ρ, T_φ, T_θ, c_r — shared by every sequence; Apply reloads the scene"),
	}}
	if m.SNote != "" {
		rows = append(rows, []Cell{Text(m.SNote).Dimmed(true)})
	}
	return Group{Rows: rows}
}

func framePop(m Model) Group {
	f := m.Frame
	field := func(label, k, now, what string) []Cell {
		c := Field(label, now, Act{Op: OpFrameDraft, K: k}, true, false)
		if v, ok := m.FrameDraft[k]; ok {
			c.Value, c.Lit = Num(v), true
		}
		return []Cell{c, Text(what).Dimmed(true)}
	}
	apply := Button("Apply", Act{Op: OpFrameApply}, len(m.FrameDraft) > 0).Dimmed(len(m.FrameDraft) == 0).Tipped("write the frame and reload the scene")
	return Group{Title: "frame — every sequence shares it · Apply reloads the scene", Card: true, Rows: [][]Cell{
		field("ρ", FrameRho, Num(f.Rho), "length of one unit of a link's r"),
		field("T_φ", FrameTphi, strconv.Itoa(f.Tphi), "index steps a turn in φ"),
		field("T_θ", FrameTtheta, strconv.Itoa(f.Ttheta), "index steps a turn in θ"),
		field("c_r", FrameCr, Num(f.Cr), "length of one radial index step"),
		{apply, Button("Cancel", Act{Op: OpFramePop}, false).Tipped("close without changing the frame")},
	}}
}

func scoreView(t *tree, m Model) {
	t.put(pathGroup(m), scopeBar(m))
	if m.FramePop {
		t.put(framePop(m))
	}
	networkRows(t, m)
	startRows(t, m)
	t.put(stepBar(m)...)
	t.put(scoreGroup(m))
	if m.ScoreFocused() {
		t.put(Group{Rows: [][]Cell{{Text(scoreKeysHint).Dimmed(true)}}, Wraps: true})
	}
	t.put(inspectorGroup(m))
}

func (m *Model) Hover(a Act) bool {
	if m.Hot == a {
		return false
	}
	m.Hot = a
	return true
}

func (m Model) Preview() (from, to string, ghosts []Links.Vec3) {
	if t := m.Turning; t.Act.Op != "" {
		return t.Act.Node, t.Act.To, m.turnGhosts(t)
	}
	if m.AtSeqs {
		return "", "", nil
	}
	if h := m.Hot; h.Op == OpStepShow && h.I != m.StepShown() {
		for _, id := range m.Nodes {
			if c, ok := m.StepCol(h.I, id); ok {
				ghosts = append(ghosts, c.Pos)
			}
		}
		return "", "", ghosts
	}
	return m.Hot.Node, m.Hot.To, nil
}

func (m Model) turnGhosts(t Turning) []Links.Vec3 {
	l, ok := m.gridLink(StepEdge{t.Act.I, t.Act.Node, t.Act.To})
	src, ok2 := m.StepCol(t.Act.I, t.Act.Node)
	if !ok || !ok2 || t.presses() == 0 {
		return nil
	}
	g := l
	*g.Delta.At(t.Act.K) = t.From + t.presses()*t.Inc
	ref := Links.RefTurn(src.Ref)
	d := m.Frame.LinkVec(g, ref).Sub(m.Frame.LinkVec(l, ref))
	var out []Links.Vec3
	for _, id := range m.downstream(t.Act.I, l.To) {
		if c, ok := m.StepCol(t.Act.I, id); ok {
			out = append(out, c.Pos.Add(d))
		}
	}
	return out
}

func (m Model) downstream(step int, root string) []string {
	out, seen := []string{root}, map[string]bool{root: true}
	for i := 0; i < len(out); i++ {
		for _, id := range m.Nodes {
			if l, ok := m.gridLink(StepEdge{step, out[i], id}); ok && l.Picked && !seen[id] {
				seen[id] = true
				out = append(out, id)
			}
		}
	}
	return out
}

func (m *Model) SelectEdge(a, b string) bool {
	if m.AtSeqs {
		return false
	}
	step := m.ScoreCell.Step
	if step == 0 {
		step = m.StepShown()
	}
	for _, e := range []StepEdge{{step, a, b}, {step, b, a}} {
		if _, ok := m.gridLink(e); ok {
			m.ScoreCell = e
			return true
		}
	}
	return false
}
