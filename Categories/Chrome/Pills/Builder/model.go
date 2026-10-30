package Builder

import (
	"math"
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
	StepDir string

	Unfold    map[string]bool
	UnfoldDir string

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

func (m *Model) ToggleUnfold(k string) {
	if m.Unfold == nil {
		m.Unfold = map[string]bool{}
	}
	m.Unfold[k] = !m.Unfold[k]
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
