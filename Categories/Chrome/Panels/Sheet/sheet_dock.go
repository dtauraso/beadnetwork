package Sheet

import "github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"

const OpDockFold = "dock-fold"

func (s *State) dockH(label string) float32 {
	if s.Folded[label] {
		return 1
	}
	return s.Height[label]
}

func (s *State) SetHeight(label string, h float32) {
	if s.Height == nil {
		s.Height = map[string]float32{}
	}
	s.Height[label] = h
	s.setFolded(label, false)
	s.save(label, "height", px(h))
}

func (s *State) ToggleFold(label string) { s.setFolded(label, !s.Folded[label]) }

func (s *State) setFolded(label string, on bool) {
	if s.Folded[label] == on {
		return
	}
	if s.Folded == nil {
		s.Folded = map[string]bool{}
	}
	s.Folded[label] = on
	v := int32(0)
	if on {
		v = 1
	}
	s.save(label, "folded", v)
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

func (lay *Layout) addBar(label string, bar Rect, folded bool) {
	lay.Pills = append(lay.Pills, PillBox{Label: label, Pill: bar, Open: folded, Bar: true})
}

func barH() float32 { return Panel.PillH() }
