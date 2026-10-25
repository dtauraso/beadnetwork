package Sheet

import (
	"math"
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

type Model struct {
	Nodes   []string
	Reports map[string]Links.Report
	Frame   Links.Frame
	S       int
	Refused string
}

func (m Model) On() bool { return len(m.Nodes) > 0 }

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
	for _, l := range m.Reports[from].Links {
		if l.To == to {
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

type State struct {
	w *ValueWriter

	Model  Model
	Edit   Edit
	Open   map[string]bool
	Scroll map[string]float32
}

func (s *State) Arm(sceneRoot string) { s.w = NewValueWriter(sceneRoot) }

func (s *State) Toggle(label string) {
	if s.Open == nil {
		s.Open = map[string]bool{}
	}
	s.Open[label] = !s.Open[label]
	if !s.Open[label] && s.Edit.Active {
		s.Edit = Edit{}
	}
}

func (s *State) ScrollOf(label string) *float32 {
	if s.Scroll == nil {
		s.Scroll = map[string]float32{}
	}
	v := s.Scroll[label]
	return &v
}

func (s *State) SetScroll(label string, v float32) {
	if s.Scroll == nil {
		s.Scroll = map[string]float32{}
	}
	s.Scroll[label] = v
}

func (s State) Pills() []Pill {
	if !s.Model.On() {
		return nil
	}
	return []Pill{LinksPill(s.Model)}
}

var RightLabels = []string{LinksLabel}

var LeftLabels = []string{}

func Num(v float64) string {
	if math.Abs(v) < 5e-4 {
		v = 0
	}
	if v == math.Trunc(v) {
		return strconv.Itoa(int(v))
	}
	return strconv.FormatFloat(v, 'f', 2, 64)
}
