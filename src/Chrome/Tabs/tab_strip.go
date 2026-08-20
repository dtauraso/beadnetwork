package Tabs

import "github.com/dtauraso/wirefold/src/Chrome/Panels/Panel"

const (
	Top     = 12
	PadX    = 8
	PadY    = 3
	TabPadX = 8
	TabPadY = 3
	Gap     = 2
	FontPx  = 11
)

type Rect = Panel.Rect

type Tab struct {
	Index int
	Name  string
	Rect  Rect

	Selected bool
}

type Layout struct {
	Strip Rect
	Tabs  []Tab
}

func Build(viewW float32, names []string, selected int) Layout {
	if len(names) == 0 || viewW <= 0 {
		return Layout{}
	}

	tabH := Panel.LineHeight(FontPx) + 2*TabPadY
	widths := make([]float32, len(names))
	var inner float32
	for i, n := range names {
		widths[i] = Panel.TextWidth(n, FontPx) + 2*TabPadX
		inner += widths[i]
		if i > 0 {
			inner += Gap
		}
	}

	strip := Rect{
		X: (viewW - (inner + 2*PadX)) / 2,
		Y: Top,
		W: inner + 2*PadX,
		H: tabH + 2*PadY,
	}

	lay := Layout{Strip: strip, Tabs: make([]Tab, len(names))}
	x := strip.X + PadX
	for i, n := range names {
		lay.Tabs[i] = Tab{
			Index:    i,
			Name:     n,
			Rect:     Rect{X: x, Y: strip.Y + PadY, W: widths[i], H: tabH},
			Selected: i == selected,
		}
		x += widths[i] + Gap
	}
	return lay
}

func (l Layout) Hit(x, y float64) int {
	for _, t := range l.Tabs {
		if Panel.HitRect(t.Rect, x, y) {
			return t.Index
		}
	}
	return -1
}
