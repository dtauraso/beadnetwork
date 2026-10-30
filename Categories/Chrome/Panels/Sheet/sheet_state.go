package Sheet

type State struct {
	w *ValueWriter

	Edit    Edit
	Open    map[string]bool
	Scroll  map[string]float32
	ScrollX map[string]float32
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

func (s *State) ScrollXOf(label string) *float32 {
	v := s.ScrollX[label]
	return &v
}

func (s *State) SetScrollX(label string, v float32) {
	if s.ScrollX == nil {
		s.ScrollX = map[string]float32{}
	}
	s.ScrollX[label] = v
}
