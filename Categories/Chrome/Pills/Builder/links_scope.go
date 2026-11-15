package Builder

import "strconv"

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
	t.put(pathGroup(m), viewGroup(m), scopeBar(m))
	if m.FramePop {
		t.put(framePop(m))
	}
	t.put(scoreGroup(m))
	if m.ScoreFocused() {
		t.put(Group{Rows: [][]Cell{{Text(scoreKeysHint).Dimmed(true)}}, Wraps: true})
	}
	t.put(inspectorGroup(m))
}
