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
)

type Hit struct {
	Part  Part
	Label string
	Cell  Cell
	Rect  Rect
}

func (l Layout) Hit(x, y float64) Hit {
	for _, p := range l.Pills {
		if Panel.HitRect(p.Pill, x, y) {
			return Hit{Part: PartPill, Label: p.Label, Rect: p.Pill}
		}
	}
	for _, cb := range l.Cells {
		if !Panel.HitRect(l.Pills[cb.Box].Popover, x, y) {
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
	if p, ok := l.covering(x, y); ok {
		return Hit{Part: PartInside, Label: p.Label}
	}
	return Hit{}
}

func (l Layout) covering(x, y float64) (PillBox, bool) {
	for _, p := range l.Pills {
		if p.Open && Panel.HitRect(p.Popover, x, y) {
			return p, true
		}
	}
	return PillBox{}, false
}

func (l Layout) Covers(x, y float64) (label string, maxScroll float32, ok bool) {
	p, ok := l.covering(x, y)
	return p.Label, p.MaxScroll, ok
}
