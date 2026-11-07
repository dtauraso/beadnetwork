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

type Model struct {
	Nodes   []string
	Reports map[string]Links.Report
	Frame   Links.Frame
	S       int
	Refused string

	Step    Step
	Steps   map[string]*Step
	StepDir string

	StepPicks map[StepEdge]bool
	StepLinks map[StepEdge]*Links.Link

	Seqs    []Seq
	OpenSeq int

	Cards    Cards
	CardsDir string

	TickMs  int
	Now     int64
	Timings map[string]NodeTiming
}

func (m Model) RowsLocking() bool {
	for _, l := range m.Picked() {
		if l.Locking {
			return true
		}
	}
	return false
}

func (m Model) On() bool { return len(m.Nodes) > 0 }

func (m Model) ManyIn() []string {
	var out []string
	for id, r := range m.Reports {
		from := map[string]bool{}
		for _, l := range r.Links {
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
		m.Reports = map[string]Links.Report{}
	}
	m.Reports[r.ID] = r
	if r.Refused != "" {
		m.Refused = r.Refused
	}
}

func (m Model) Link(from, to string) (Links.Link, bool) {
	for _, l := range m.Reports[to].Links {
		if l.From == from {
			return l, true
		}
	}
	return Links.Link{}, false
}

func (m Model) ScaleOf(id string) int {
	if s := m.Reports[id].S; s > 0 {
		return s
	}
	return max(1, m.S)
}

func Pills(m Model) []Pill {
	if !m.On() {
		return nil
	}
	pills := []Pill{FramePill(m), NodesPill(m), LinksPill(m)}
	if len(m.Seqs) > 0 {
		pills = append(pills, SequencesPill(m))
	}
	return append(pills, PlacePill(m), TreePill(m))
}

var RightLabels = []string{FrameLabel, NodesLabel, LinksLabel}

var LeftLabels = []string{SeqLabel, PlaceLabel, TreeLabel}

func Num(v float64) string {
	if math.Abs(v) < 5e-4 {
		v = 0
	}
	if v == math.Trunc(v) {
		return strconv.Itoa(int(v))
	}
	return strconv.FormatFloat(math.Round(v*100)/100, 'f', 2, 64)
}
