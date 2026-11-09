package Sheet

import "github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"

const popScrollFile = "pop-scroll"

func popStep() float32 { return rowH() + RowGapY }

func (s *State) PopScrollOf(label string) *float32 {
	v := s.PopScroll[label]
	return &v
}

func (s *State) SetPopScroll(label string, v float32) {
	if s.PopScroll == nil {
		s.PopScroll = map[string]float32{}
	}
	s.PopScroll[label] = v
	s.save(label, popScrollFile, px(v))
}

func popWindow(g Group, scroll float32) (Group, float32) {
	hidden := len(g.Rows) - g.MaxRows
	if g.Pop == nil || g.MaxRows <= 0 || hidden <= 0 {
		return g, 0
	}
	most := float32(hidden) * popStep()
	off := int(min(max(scroll, 0), most) / popStep())
	g.Rows = g.Rows[off : off+g.MaxRows]
	return g, most
}

func (l Layout) PopAt(x, y float64) (label string, most float32, ok bool) {
	for _, c := range l.Cards {
		if c.Pop && c.Max > 0 && Panel.HitRect(c.Rect, x, y) {
			return l.Pills[c.Box].Label, c.Max, true
		}
	}
	return "", 0, false
}
