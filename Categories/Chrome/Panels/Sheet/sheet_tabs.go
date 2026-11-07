package Sheet

import "slices"

func (s *State) ordered(labels []string) []string {
	var out []string
	for _, l := range s.Order {
		if slices.Contains(labels, l) {
			out = append(out, l)
		}
	}
	for _, l := range labels {
		if !slices.Contains(out, l) {
			out = append(out, l)
		}
	}
	return out
}

func (s *State) StartTabDrag(label string) { s.Dragging = label }

func (s *State) EndTabDrag() bool {
	was := s.Dragging != ""
	s.Dragging = ""
	return was
}

func (s *State) DragTab(lay Layout, x float64) bool {
	if s.Dragging == "" {
		return false
	}
	for _, p := range lay.Pills {
		if !p.Bar {
			continue
		}
		var labels []string
		to := 0
		for _, t := range p.Tabs {
			labels = append(labels, t.Label)
			if t.Label != s.Dragging && float64(t.Rect.X+t.Rect.W/2) < x {
				to++
			}
		}
		from := slices.Index(labels, s.Dragging)
		if from < 0 || from == to {
			return true
		}
		labels = slices.Insert(slices.Delete(labels, from, from+1), to, s.Dragging)
		s.Order = labels
		for i, l := range labels {
			s.save(l, "tab-pos", int32(i))
		}
		s.Tab = to
		s.save(DockKey, "tab", int32(to))
	}
	return true
}
