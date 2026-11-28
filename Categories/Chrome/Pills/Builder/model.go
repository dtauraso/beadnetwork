package Builder

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/StepStore"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

type (
	Cell  = Sheet.Cell
	Act   = Sheet.Act
	Pill  = Sheet.Pill
	Group = Sheet.Group

	Step    = StepStore.Step
	StepRow = StepStore.StepRow
	LockOn  = StepStore.LockOn
)

var (
	Text    = Sheet.Text
	LitText = Sheet.LitText
	Button  = Sheet.Button
	Field   = Sheet.Field
)

type Model struct {
	Nodes   []string
	Reports map[string]Links.Node
	Frame   Links.Frame
	S       int
	Refused string

	StepStore.Disk

	Amts      map[string]map[Edge]Amount
	Leaves    [][]int
	Sent      map[Edge]Links.Ints
	Play      *Play
	Folder    *FolderRun
	Rounds    map[string]int
	CardsOpen map[StepEdge]bool
	CardSX    map[StepEdge]float32

	ShownStep   int
	StepClosed  bool
	NetworkOpen bool
	NewOpen     bool
	New         NewNode
	StartOpen   bool
	TreeSY      float32
	TreeW       float32
	splitX0     float32
	Splitting   bool
	StepsOpen   map[StepNode]bool
	FramePop    bool
	FrameDraft  map[string]float64
	Hot         Act
	Menu        Act
	Pressed     Act
	UndoPick    Act
	Scene       string
	OpenTop     int
	Tree        []StepRow
	TreeFind    string

	TickMs  int
	Now     int64
	Timings map[string]NodeTiming
}

func (m Model) On() bool { return len(m.Nodes) > 0 }

func (m Model) ManyIn() []string {
	var out []string
	for id := range m.Reports {
		from := map[string]bool{}
		for _, l := range m.Node(id).Links() {
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
	if m.Reports == nil {
		m.Reports = map[string]Links.Node{}
	}
	m.Reports[r.ID] = r.Node
	if m.Sent == nil {
		m.Sent = map[Edge]Links.Ints{}
	}
	for e := range m.Sent {
		if _, ok := r.Node.InFrom(e.From); e.To == r.ID && !ok {
			delete(m.Sent, e)
		}
	}
	for _, in := range r.Node.In {
		e := Edge{From: in.Link.From, To: r.ID}
		if _, ok := m.Sent[e]; !ok {
			d := in.Link.Delta
			m.Sent[e] = Links.Ints{Phi: d.Phi, Theta: d.Theta, R: d.R}
		}
	}
	if m.Rounds == nil {
		m.Rounds = map[string]int{}
	}
	m.Rounds[r.ID] = r.Round
	if r.Refused != "" {
		m.Refused = r.Refused
	}
}

func (m Model) Node(id string) Links.Node { return m.Reports[id] }

func (m Model) SumOf(l Links.Link) Links.Angles {
	in, _ := m.Node(l.To).InFrom(l.From)
	return in.Sum
}

func (m Model) Link(from, to string) (Links.Link, bool) {
	in, ok := m.Node(to).InFrom(from)
	return in.Link, ok
}

func (m Model) RefOf(id string) Links.Turn { return Links.RefTurn(m.Node(id).Ref) }

func (m Model) SceneS() int { return max(1, m.S) }

func Pills(m Model) []Pill {
	if !m.On() {
		return nil
	}
	return []Pill{LinksPill(m)}
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

type NodeTiming struct {
	H int64
}

func (m *Model) TakeTiming(t Links.Timing) {
	if m.Timings == nil {
		m.Timings = map[string]NodeTiming{}
	}
	m.Timings[t.ID] = NodeTiming{H: t.H}
}

func (m Model) Behind() int {
	b := 0
	for _, id := range m.Nodes {
		if t, ok := m.Timings[id]; !ok || t.H < m.Now-1 {
			b++
		}
	}
	return b
}

func (m Model) TimeKey() string {
	return fmt.Sprintf("%d|%d", m.Now, m.Behind())
}
