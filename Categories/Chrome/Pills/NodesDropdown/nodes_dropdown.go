package NodesDropdown

import (
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/Panel"
	NodeBuf "github.com/dtauraso/wirefold/Categories/Node"
)

const Label = "Nodes"

const (
	SwatchSize = 11
	SwatchGap  = 7
	GlyphW     = 8

	DescPadLeft = 18
	DescLineH   = 1.35
)

type Rect = Panel.Rect

type Row struct {
	KindID uint8
	Kind   string

	Fill   string
	Stroke string

	Head   Rect
	Swatch Rect
	Open   bool

	Desc     string
	DescRect Rect
}

type Layout struct {
	Pill    Rect
	Open    bool
	Popover Rect

	Rows []Row
}

type Kind struct {
	KindID uint8
	Name   string
	Open   bool
}

func descLines(desc string, w float32) int {
	if desc == "" {
		return 0
	}
	perLine := int((w - DescPadLeft) / Panel.Advance(Panel.PillFontPx))
	if perLine < 1 {
		perLine = 1
	}
	n := (len(desc) + perLine - 1) / perLine
	if n < 1 {
		n = 1
	}
	return n
}

func Build(st *Panel.PillStack, open bool, kinds []Kind) Layout {
	pill := st.AddPill()
	lay := Layout{Pill: pill, Open: open}
	if !open {
		st.EndGroup()
		return lay
	}

	w := pill.W - 2*Panel.PopoverPad
	rowH := Panel.RowH()
	lineH := Panel.LineHeight(Panel.PillFontPx) * DescLineH

	var contentH float32
	for _, k := range kinds {
		contentH += rowH
		if k.Open {
			a, _ := NodeBuf.AppearanceOf(k.Name)
			contentH += float32(descLines(a.Desc, w)) * lineH
		}
	}

	box, x, y := st.AddPopover(contentH)
	lay.Popover = box

	lay.Rows = make([]Row, len(kinds))
	for i, k := range kinds {
		a, _ := NodeBuf.AppearanceOf(k.Name)
		r := Row{
			KindID: k.KindID,
			Kind:   k.Name,
			Fill:   a.Fill,
			Stroke: a.Stroke,
			Head:   Rect{X: x, Y: y, W: w, H: rowH},
			Swatch: Rect{
				X: x + Panel.RowPadX + GlyphW + SwatchGap,
				Y: y + (rowH-SwatchSize)/2,
				W: SwatchSize, H: SwatchSize,
			},
			Open: k.Open,
			Desc: a.Desc,
		}
		y += rowH
		if k.Open {
			h := float32(descLines(a.Desc, w)) * lineH
			r.DescRect = Rect{X: x + DescPadLeft, Y: y, W: w - DescPadLeft, H: h}
			y += h
		}
		lay.Rows[i] = r
	}
	st.EndGroup()
	return lay
}

type HitKind int

const (
	HitNone HitKind = iota
	HitPill
	HitRow
)

type Hit struct {
	Kind   HitKind
	KindID uint8

	Rect Rect
}

func (l Layout) Hit(x, y float64) Hit {
	if Panel.HitRect(l.Pill, x, y) {
		return Hit{Kind: HitPill, Rect: l.Pill}
	}
	if !l.Open {
		return Hit{}
	}
	for _, r := range l.Rows {
		if Panel.HitRect(r.Head, x, y) {
			return Hit{Kind: HitRow, KindID: r.KindID, Rect: r.Head}
		}
	}
	return Hit{}
}
