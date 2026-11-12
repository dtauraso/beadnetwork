package Builder

import (
	"math"
	"slices"
	"sort"
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

type (
	Cell  = Sheet.Cell
	Act   = Sheet.Act
	Pill  = Sheet.Pill
	Group = Sheet.Group
)

var (
	Text    = Sheet.Text
	LitText = Sheet.LitText
	Button  = Sheet.Button
	Field   = Sheet.Field
)

const (
	inkStrong = Sheet.InkStrong
	inkHead   = Sheet.InkHead
)

type Model struct {
	Nodes   []string
	Grids   map[string]Links.Grid
	Frame   Links.Frame
	S       int
	SNote   string
	Refused string

	Step      Step
	Steps     map[string]*Step
	Root      string
	StepNums  []int
	StepNames map[int]string

	StepFrom  []int
	StepHeads map[int]*Step
	StepFolds map[StepNode]bool

	Seqs        []Seq
	OpenSeq     int
	AtSeqs      bool
	SeqFind     string
	Sel        Act
	ClickRow    Act
	ClickAt     int64
	ShownStep   int
	StepClosed  bool
	NetworkOpen bool
	Touched     StepEdge
	DragStep    int
	Level       string
	Trail       []TreeRow
	Tree        []TreeRow

	Cards    Cards
	CardsDir string

	TickMs  int
	Now     int64
	Timings map[string]NodeTiming
}

func (m Model) On() bool { return len(m.Nodes) > 0 }

func (m Model) ManyIn() []string {
	var out []string
	for id := range m.Grids {
		from := map[string]bool{}
		for _, l := range m.Col(id).Links() {
			from[l.From] = true
		}
		if len(from) > 1 {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

func (m Model) FixedR(l Links.Link) bool {
	fixed := m.ManyIn()
	return slices.Contains(fixed, l.From) || slices.Contains(fixed, l.To)
}

func (m *Model) Take(r Links.Report) {
	if m.Grids == nil {
		m.Grids = map[string]Links.Grid{}
	}
	m.Grids[r.ID] = r.Grid
	if r.Refused != "" {
		m.Refused = r.Refused
	}
}

func (m Model) Col(id string) Links.Col { return m.Grids[id].ShownCol() }

func (m Model) StepCol(step int, id string) (Links.Col, bool) { return m.Grids[id].Col(step) }

func (m *Model) ShowMsg() Links.ShowStep {
	s := Links.ShowStep{Level: m.Level, Step: m.StepShown(), Steps: m.stepNums(), From: m.StepFrom}
	m.StepFrom = nil
	return s
}

func (m Model) SumOf(l Links.Link) Links.Start {
	in, _ := m.Col(l.To).InFrom(l.From)
	return in.Sum
}

func (m Model) Link(from, to string) (Links.Link, bool) {
	in, ok := m.Col(to).InFrom(from)
	return in.Link, ok
}

func (m Model) RefOf(id string) Links.Turn { return Links.RefTurn(m.Col(id).Ref) }

func (m Model) SceneS() int { return max(1, m.S) }

func Pills(m Model) []Pill {
	if !m.On() {
		return nil
	}
	return []Pill{FramePill(m), NodesPill(m), LinksPill(m), GridPill(m), PlacePill(m), TreePill(m)}
}

func Num(v float64) string {
	if math.Abs(v) < 5e-4 {
		v = 0
	}
	if v == math.Trunc(v) {
		return strconv.Itoa(int(v))
	}
	return strconv.FormatFloat(math.Round(v*100)/100, 'f', 2, 64)
}
