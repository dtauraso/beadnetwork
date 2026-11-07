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
	s.Max[label] = !s.Max[label]
	on := int32(0)
	if s.Max[label] {
		on = 1
	}
	s.save(label, "max", on)
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

func maxButton(label string, box Rect, max bool) CellBox {
	glyph := "▲"
	if max {
		glyph = "▼"
	}
	c := Button(glyph, Act{Op: OpDockMax, Node: label}, false)
	w, h := cellW(c), rowH()
	return CellBox{Cell: c, Rect: Rect{X: box.X + Panel.PopoverPad, Y: box.Y + Panel.PopoverPad, W: w, H: h}}
}

func barH() float32 { return rowH() + Panel.Gap }
