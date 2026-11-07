package Sheet

import "github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"

const OpDockMax = "dock-max"

func (s *State) dockH(label string) float32 {
	if s.Max[label] {
		return 1 << 20
	}
	return s.Height[label]
}

func (s *State) SetHeight(label string, h float32) {
	if s.Height == nil {
		s.Height = map[string]float32{}
	}
	s.Height[label] = h
	if s.Max[label] {
		s.Max[label] = false
		s.save(label, "max", 0)
	}
	s.save(label, "height", px(h))
}

func (s *State) ToggleMax(label string) {
	if s.Max == nil {
		s.Max = map[string]bool{}
	}
	if s.Max[label] {
		s.SetHeight(label, 1)
		return
	}
	s.Max[label] = true
	s.save(label, "max", 1)
}

func (s *State) StartResize(label string) { s.Resizing = label }

func (s *State) Resize(viewH, y float64) bool {
	if s.Resizing == "" {
		return false
	}
	s.SetHeight(s.Resizing, float32(viewH-y))
	return true
}

func (s *State) EndResize() bool {
	was := s.Resizing != ""
	s.Resizing = ""
	return was
}

func (lay *Layout) addBar(label string, bar Rect, max bool) {
	lay.Pills = append(lay.Pills, PillBox{Label: label, Pill: bar, Open: !max, Bar: true})
}

func barH() float32 { return Panel.PillH() }
