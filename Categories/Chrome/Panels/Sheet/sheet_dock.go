package Sheet

import "github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"

const (
	OpDockFold = "dock-fold"
	OpDockTab  = "dock-tab"
	DockKey    = "dock"
)

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

func (s *State) StartResize(label string, dy float32) { s.Resizing, s.resizeDY = label, dy }

func (s *State) Resize(viewH, y float64) bool {
	if s.Resizing == "" {
		return false
	}
	if h := float32(viewH-y) + s.resizeDY; h < 2*barH() {
		s.setFolded(s.Resizing, true)
	} else {
		s.SetHeight(s.Resizing, h)
	}
	return true
}

func (s *State) EndResize() bool {
	was := s.Resizing != ""
	s.Resizing = ""
	return was
}

func (s *State) SelectTab(i int) {
	s.Tab = i
	s.setFolded(DockKey, false)
	s.save(DockKey, "tab", int32(i))
}

func (lay *Layout) addBar(labels []string, on int, bar Rect, folded bool) {
	var tabs []DockTab
	x := bar.X + tabInset
	for i, l := range labels {
		w := Panel.TextWidth(l, Panel.PillFontPx) + 2*Panel.PillPadX
		tabs = append(tabs, DockTab{Label: l, Rect: Rect{X: x, Y: bar.Y + tabInset, W: w, H: bar.H - 2*tabInset}, On: i == on})
		x += w + tabInset
	}
	lay.Pills = append(lay.Pills, PillBox{Label: labels[on], Pill: bar, Open: folded, Bar: true, Tabs: tabs})
}

const tabInset = 4

func barH() float32 { return Panel.PillH() + 2*tabInset }

func caretRect(bar Rect) Rect {
	w := float32(Panel.CaretW + 6)
	return Rect{X: bar.X + bar.W - w, Y: bar.Y, W: w, H: bar.H}
}
