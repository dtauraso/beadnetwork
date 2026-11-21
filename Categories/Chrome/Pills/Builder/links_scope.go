package Builder

import (
	"strconv"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/LinksBoard"
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

func framePop(m Model) *LinksBoard.Pop {
	f := m.Frame
	field := func(label, k, now, what string) LinksBoard.PopField {
		p := LinksBoard.PopField{Label: label, What: what, Field: Board.Field{Value: now, Fit: "000.00", Act: Act{Op: OpFrameDraft, K: k}}}
		if v, ok := m.FrameDraft[k]; ok {
			p.Field.Value, p.Was = Num(v), now
		}
		return p
	}
	return &LinksBoard.Pop{Tag: "shared by every step", Apply: Act{Op: OpFrameApply}, Cancel: Act{Op: OpFramePop}, CanApply: len(m.FrameDraft) > 0,
		Fields: []LinksBoard.PopField{
			field("ρ", FrameRho, Num(f.Rho), "length of one unit of r"),
			field("T_φ", FrameTphi, strconv.Itoa(f.Tphi), "index steps a turn in φ"),
			field("T_θ", FrameTtheta, strconv.Itoa(f.Ttheta), "index steps a turn in θ"),
			field("c_r", FrameCr, Num(f.Cr), "one radial index step"),
			{Label: "Δt", What: "ms a human tick lasts · applies now", Field: Board.Field{Value: strconv.Itoa(m.TickMs), Fit: "000.00", Act: Act{Op: OpTickMs}}},
		}}
}

func boardData(m Model) LinksBoard.Data {
	net := Board.Chip{Label: "Network ▾", Act: Act{Op: OpNetFold}, Lit: m.NetworkOpen && !m.NewOpen}
	d := LinksBoard.Data{
		S:     Board.Field{Value: strconv.Itoa(m.SceneS()), Fit: "00", Act: frameAct(FrameS)},
		Doors: []Board.Chip{net, {Label: "Nodes ▾", Act: Act{Op: OpNewFold}, Lit: m.NewOpen}, {Label: "Frame ⚙", Act: Act{Op: OpFramePop}, Lit: m.FramePop}},
		More:  Act{Op: OpMenu, K: menuMore}, MoreOpen: m.Menu == Act{Op: OpMenu, K: menuMore}, Cols: boardCols(m), Insp: boardInspector(m), InspOpen: m.InspOpen, InspTab: Act{Op: OpInspTab},
		Status: "showing step " + itoa(m.StepShown()),
		Toast:  m.Refused, At: m.Pressed, Dismiss: Act{Op: OpToastClose}}
	if ns := m.stepNums(); len(ns) > 0 {
		d.Add = Act{Op: OpStepAdd, I: ns[len(ns)-1]}
	}
	m.treeData(&d)
	d.Crumbs, d.CrumbName = m.Crumbs()
	if d.TreeFind != nil {
		d.PathMenu, d.PathOpen = Act{Op: OpTreePop}, m.TreePop
		d.TreeScroll, d.TreeSY = Act{Op: OpTreeScroll}, m.TreeSY
	}
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

func (m Model) Preview() (from, to string) {
	if t := m.Turning; t.Act.Op != "" {
		return t.Act.Node, t.Act.To
	}
	return m.Hot.Node, m.Hot.To
}

func (m *Model) SelectEdge(a, b string) bool {
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
	if e, ok := m.edgeIn(m.StepShown(), a, b); ok {
		return Act{Op: OpLinkRow, Node: e.From, To: e.To, I: e.Step}
	}
	return Act{}
}
