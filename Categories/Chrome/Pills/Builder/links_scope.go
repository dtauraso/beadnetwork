package Builder

import (
	"strconv"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/LinksBoard"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const (
	OpFramePop   = "frame-pop"
	OpFrameDraft = "frame-draft"
	OpFrameApply = "frame-apply"
	OpToastClose = "toast-close"
)

func (m *Model) StageFrame(k string, v float64) {
	if m.FrameDraft == nil {
		m.FrameDraft = map[string]float64{}
	}
	m.FrameDraft[k] = v
}

func crumbs(m Model) []LinksBoard.Crumb {
	out := []LinksBoard.Crumb{{Label: "Sequences ▾", Short: "Seq ▾", Act: Act{Op: OpMenu, K: menuSeqs}}}
	names := []string{m.seqName()}
	for _, s := range m.Trail {
		names = append(names, s.Label)
	}
	for i, n := range names {
		c := LinksBoard.Crumb{Label: n}
		if i < len(m.Trail) {
			c.Act = Act{Op: OpLevelGo, Node: m.Trail[i].Level, I: m.Trail[i].Step}
		}
		if i == 0 && len(m.Seqs) > 0 {
			c.Tag = m.openSeqTag()
			c.Edit, c.Rename = Act{Op: OpSeqName, I: m.OpenSeq}, Act{Op: OpRename, K: renameSeq, I: m.OpenSeq}
		}
		out = append(out, c)
	}
	return out
}

func framePop(m Model) *LinksBoard.Pop {
	f := m.Frame
	field := func(label, k, now, what string) LinksBoard.PopField {
		p := LinksBoard.PopField{Label: label, What: what, Field: Board.Field{Value: now, Fit: "000.00", Act: Act{Op: OpFrameDraft, K: k}}}
		if v, ok := m.FrameDraft[k]; ok {
			p.Field.Value, p.Was = Num(v), now
		}
		return p
	}
	return &LinksBoard.Pop{Tag: "shared by all " + itoa(max(1, len(m.Seqs))) + " sequences", Apply: Act{Op: OpFrameApply}, Cancel: Act{Op: OpFramePop}, CanApply: len(m.FrameDraft) > 0,
		Fields: []LinksBoard.PopField{
			field("ρ", FrameRho, Num(f.Rho), "length of one unit of r"),
			field("T_φ", FrameTphi, strconv.Itoa(f.Tphi), "index steps a turn in φ"),
			field("T_θ", FrameTtheta, strconv.Itoa(f.Ttheta), "index steps a turn in θ"),
			field("c_r", FrameCr, Num(f.Cr), "one radial index step"),
			{Label: "Δt", What: "ms a human tick lasts · applies now", Field: Board.Field{Value: strconv.Itoa(m.TickMs), Fit: "000.00", Act: Act{Op: OpTickMs}}},
		}}
}

func boardData(m Model) LinksBoard.Data {
	net := Board.Chip{Label: "Network…", Act: Act{Op: OpNetFold}, Tip: "nodes, links and places: what runs now"}
	d := LinksBoard.Data{Crumbs: crumbs(m), SNote: m.SNote,
		S:          Board.Field{Value: strconv.Itoa(m.SceneS()), Fit: "00", Act: frameAct(FrameS), Tip: "ticks a twelfth of a turn, one s every node uses — it rises when a start needs more"},
		Doors:      []Board.Chip{net, {Label: "Frame ⚙", Act: Act{Op: OpFramePop}, Lit: m.FramePop, Tip: "ρ, T_φ, T_θ, c_r, Δt — shared by every sequence; applying reloads the scene"}},
		DoorsShort: []string{"Net…", "⚙"}, Cols: boardCols(m), Insp: boardInspector(m), InspOpen: m.InspOpen, InspTab: Act{Op: OpInspTab},
		Status: "showing step " + itoa(m.StepShown()), Ghost: m.GhostNote(),
		Toast: m.Refused, At: m.Pressed, Dismiss: Act{Op: OpToastClose}}
	if ns := m.stepNums(); len(ns) > 0 {
		d.Add = Act{Op: OpStepAdd, I: ns[len(ns)-1]}
	}
	d.Steps = m.stepStrip(d.Add)
	d.SeqMenu, d.SeqOpen = m.seqMenu(), m.Menu == Act{Op: OpMenu, K: menuSeqs}
	if strings.HasPrefix(m.Refused, "moved the pick") {
		d.Undo = m.UndoPick
	}
	if m.FramePop {
		d.Pop = framePop(m)
	}
	if len(m.allLinks()) == 0 {
		e := &LinksBoard.Empty{Title: "No links yet.", Sub: "add nodes and links, then come back here to animate them", Door: Board.Chip{Label: "Open Network…", Act: net.Act}}
		if len(m.Nodes) == 1 {
			e.Title, e.Sub = "Node "+m.Nodes[0]+" has nothing to turn against.", "A link needs two nodes."
		}
		d.Empty = e
	}
	return d
}

func (m *Model) NotePick(a Act) {
	m.UndoPick = Act{}
	if a.Op != OpStepPick {
		return
	}
	for _, l := range m.linksTo(a.To) {
		if sl, ok := m.gridLink(StepEdge{a.I, l.From, a.To}); ok && sl.Picked && l.From != a.Node {
			m.UndoPick = Act{Op: OpStepPick, Node: l.From, To: a.To, I: a.I}
		}
	}
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

func (m Model) GhostNote() string {
	_, _, ghosts := m.Preview()
	if len(ghosts) == 0 {
		return ""
	}
	n := itoa(len(ghosts)) + " faint node"
	if len(ghosts) > 1 {
		n += "s"
	}
	if t := m.Turning; t.Act.Op != "" {
		return "ghost: the " + n + " in the scene show where " + t.Act.To + " and what follows it land when you let go"
	}
	return "ghost: the " + n + " in the scene show where every node sits at the end of step " + itoa(m.Hot.I)
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
	e, ok := m.edgeIn(m.StepShown(), a, b)
	if !ok {
		return false
	}
	m.setStepOpen(e.Step, true)
	if m.StepFolds == nil {
		m.StepFolds = map[StepNode]bool{}
	}
	m.StepFolds[StepNode{e.Step, e.To}] = false
	m.ToggleCard(Act{Node: e.From, To: e.To, I: e.Step})
	m.Touched = e
	return true
}

func (m Model) edgeIn(step int, a, b string) (StepEdge, bool) {
	for _, e := range []StepEdge{{step, a, b}, {step, b, a}} {
		if _, ok := m.gridLink(e); ok {
			return e, true
		}
	}
	return StepEdge{}, false
}

func (m Model) EdgeHot(a, b string) Act {
	if e, ok := m.edgeIn(m.StepShown(), a, b); ok && !m.AtSeqs {
		return Act{Op: OpLinkRow, Node: e.From, To: e.To, I: e.Step}
	}
	return Act{}
}
