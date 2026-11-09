package Sheet

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
)

type Part int

const (
	PartNone Part = iota
	PartPill
	PartBody
	PartUp
	PartDown
	PartInside
	PartGrab
)

type Hit struct {
	Part  Part
	Label string
	Cell  Cell
	Rect  Rect
	DY    float32
}

func (l Layout) Hit(x, y float64) Hit {
	for _, p := range l.Pills {
		if p.Bar && Panel.HitRect(p.Pill, x, y) {
			if c := caretRect(p.Pill); Panel.HitRect(c, x, y) {
				return Hit{Part: PartBody, Label: p.Label, Cell: Cell{Act: Act{Op: OpDockFold, Node: DockKey}}, Rect: c}
			}
			for i, t := range p.Tabs {
				if Panel.HitRect(t.Rect, x, y) {
					return Hit{Part: PartBody, Label: p.Label, Cell: Cell{Act: Act{Op: OpDockTab, I: i, Node: t.Label}}, Rect: t.Rect}
				}
			}
			return Hit{Part: PartGrab, Label: DockKey, Rect: p.Pill, DY: float32(y) - p.Pill.Y}
		}
		if p.Pill.W > 0 && Panel.HitRect(p.Pill, x, y) {
			return Hit{Part: PartPill, Label: p.Label, Rect: p.Pill}
		}
		if p.Open && p.Grab.W > 0 && Panel.HitRect(p.Grab, x, y) {
			return Hit{Part: PartGrab, Label: DockKey, Rect: p.Grab, DY: float32(y) - (p.Grab.Y + Panel.DockGrabH/2)}
		}
	}
	top := -1
	for i, p := range l.Pills {
		if p.Open && Panel.HitRect(p.Popover, x, y) {
			top = i
		}
	}
	if top < 0 {
		return Hit{}
	}
	for i := len(l.Cells) - 1; i >= 0; i-- {
		cb := l.Cells[i]
		if cb.Box != top || cb.Slides && !Panel.HitRect(l.Pills[top].Slide, x, y) {
			continue
		}
		label := l.Pills[cb.Box].Label
		switch {
		case cb.Up.W > 0 && Panel.HitRect(cb.Up, x, y):
			return Hit{Part: PartUp, Label: label, Cell: cb.Cell, Rect: cb.Up}
		case cb.Down.W > 0 && Panel.HitRect(cb.Down, x, y):
			return Hit{Part: PartDown, Label: label, Cell: cb.Cell, Rect: cb.Down}
		case cb.Cell.Act.Op != "" && Panel.HitRect(cb.Rect, x, y):
			if cb.Cell.Kind == CellField && !cb.Cell.Typed {
				continue
			}
			return Hit{Part: PartBody, Label: label, Cell: cb.Cell, Rect: cb.Rect}
		}
	}
	for _, h := range l.Heads {
		if h.Box == top && h.Fold.Op != "" && Panel.HitRect(h.Rect, x, y) {
			return Hit{Part: PartBody, Label: l.Pills[top].Label, Cell: Cell{Act: h.Fold}, Rect: h.Rect}
		}
	}
	return Hit{Part: PartInside, Label: l.Pills[top].Label}
}

func (l Layout) covering(x, y float64) (PillBox, bool) {
	for i := len(l.Pills) - 1; i >= 0; i-- {
		if p := l.Pills[i]; p.Open && Panel.HitRect(p.Popover, x, y) {
			return p, true
		}
	}
	return PillBox{}, false
}

func (l Layout) Covers(x, y float64) (label, table string, maxScroll, maxScrollX float32, ok bool) {
	p, ok := l.covering(x, y)
	for _, s := range p.Slides {
		if Panel.HitRect(s.Rect, x, y) {
			table, maxScrollX = s.Table, s.Max
		}
	}
	return p.Label, table, p.MaxScroll, maxScrollX, ok
}
