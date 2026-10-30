package Sheet

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
)

func cellW(c Cell) float32 {
	switch c.Kind {
	case CellField:
		fit := valueField
		if c.Fit != "" {
			fit = c.Fit
		}
		v := c.Value
		if len(v) < len(fit) {
			v = fit
		}
		w := Panel.TextWidth(v, ValFontPx) + 2*CellPadX
		if c.Key != "" {
			w += Panel.TextWidth(c.Key, KeyFontPx) + CellGapX
		}
		if c.Steps {
			w += StepGap + StepArrowW
		}
		return w
	case CellButton:
		return Panel.TextWidth(c.Key, KeyFontPx) + 2*CellPadX
	}
	return Panel.TextWidth(c.Key, KeyFontPx) + 2
}

func rowH() float32   { return Panel.LineHeight(ValFontPx) + 2*CellPadY }
func titleH() float32 { return Panel.LineHeight(TitleFontPx) }

func columns(g Group) []float32 {
	var cols []float32
	for _, row := range g.Rows {
		for i, c := range row {
			if i >= len(cols) {
				cols = append(cols, 0)
			}
			cols[i] = max(cols[i], cellW(c))
		}
	}
	return cols
}

func groupW(g Group) float32 {
	var w float32
	for i, c := range columns(g) {
		if i > 0 {
			w += CellGapX
		}
		w += c
	}
	return max(w, Panel.TextWidth(g.Title, TitleFontPx))
}

func groupH(g Group) float32 {
	h := float32(len(g.Rows)) * (rowH() + RowGapY)
	if g.Title != "" {
		h += titleH() + TitleGap
	}
	return h
}

func Build(right, left *Panel.PillStack, pills []Pill, st *State) Layout {
	var lay Layout
	for _, p := range pills {
		stack := right
		if p.Left {
			stack = left
		}
		box := len(lay.Pills)
		pb := PillBox{Label: p.Label, Left: p.Left, Pill: stack.AddPill(), Open: st.Open[p.Label]}
		if !pb.Open {
			stack.EndGroup()
			lay.Pills = append(lay.Pills, pb)
			continue
		}
		var contentW, contentH float32
		for i, g := range p.Groups {
			contentW = max(contentW, groupW(g))
			if i > 0 {
				contentH += Panel.Gap
			}
			contentH += groupH(g)
		}
		popover, x, y, maxScroll := stack.AddWideScrollingPopover(contentW, contentH, st.Scroll[p.Label])
		stack.EndGroup()
		pb.Popover, pb.MaxScroll = popover, maxScroll
		lay.Pills = append(lay.Pills, pb)
		for _, g := range p.Groups {
			lay.place(box, g, x, y)
			y += groupH(g) + Panel.Gap
		}
	}
	return lay
}

func (lay *Layout) place(box int, g Group, x, y float32) {
	if g.Title != "" {
		lay.Heads = append(lay.Heads, HeadBox{Box: box, Rect: Rect{X: x, Y: y, W: groupW(g), H: titleH()}, Title: g.Title})
		y += titleH() + TitleGap
	}
	cols := columns(g)
	for _, row := range g.Rows {
		cx := x
		for i, c := range row {
			cb := CellBox{Box: box, Cell: c, Rect: Rect{X: cx, Y: y, W: cols[i], H: rowH()}}
			if c.Kind == CellField && c.Steps {
				cb.Rect.W = cols[i] - StepGap - StepArrowW
				ax := cx + cb.Rect.W + StepGap
				half := rowH() / 2
				cb.Up = Rect{X: ax, Y: y, W: StepArrowW, H: half}
				cb.Down = Rect{X: ax, Y: y + half, W: StepArrowW, H: half}
			}
			lay.Cells = append(lay.Cells, cb)
			cx += cols[i] + CellGapX
		}
		y += rowH() + RowGapY
	}
}
