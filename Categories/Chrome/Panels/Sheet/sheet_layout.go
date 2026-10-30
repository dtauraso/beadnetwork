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
	if c.Boxed {
		w := Panel.TextWidth(c.Key, KeyFontPx) + 2*CellPadX
		if c.Arrows.Op != "" {
			w += lockW()
		}
		return w
	}
	return Panel.TextWidth(c.Key, KeyFontPx) + 2
}

func lockW() float32  { return LockGap + 2*LockArrowW + StepGap }
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

func colsW(g Group) float32 {
	var w float32
	for i, c := range columns(g) {
		if i > 0 {
			w += CellGapX
		}
		w += c
	}
	return w
}

func groupW(g Group) float32 { return max(colsW(g), Panel.TextWidth(g.Title, TitleFontPx)) }

func slideW(g Group, room float32) float32 {
	w := colsW(g)
	if g.Slides && room > 0 {
		w = min(w, room)
	}
	return w
}

func shownW(g Group, room float32) float32 {
	return max(slideW(g, room), Panel.TextWidth(g.Title, TitleFontPx))
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
		room := stack.ContentRoomW()
		var contentW, contentH float32
		for i, g := range p.Groups {
			contentW = max(contentW, shownW(g, room))
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
			dx := float32(0)
			if g.Slides {
				vis := slideW(g, room)
				pb := &lay.Pills[box]
				pb.Slide = Rect{X: x, Y: popover.Y, W: vis, H: popover.H}
				pb.MaxScrollX = colsW(g) - vis
				dx = max(0, min(st.ScrollX[p.Label], pb.MaxScrollX))
			}
			lay.place(box, g, x, y, dx)
			y += groupH(g) + Panel.Gap
		}
	}
	return lay
}

func stepsIn(g Group, i int) bool {
	for _, row := range g.Rows {
		if i < len(row) && row[i].Kind == CellField && row[i].Steps {
			return true
		}
	}
	return false
}

func (lay *Layout) place(box int, g Group, x, y, dx float32) {
	if g.Title != "" {
		lay.Heads = append(lay.Heads, HeadBox{Box: box, Rect: Rect{X: x, Y: y, W: groupW(g), H: titleH()}, Title: g.Title})
		y += titleH() + TitleGap
	}
	cols := columns(g)
	for _, row := range g.Rows {
		cx := x - dx
		for i, c := range row {
			cb := CellBox{Box: box, Cell: c, Rect: Rect{X: cx, Y: y, W: cols[i], H: rowH()}, Slides: g.Slides}
			if c.Arrows.Op != "" {
				right := cx + cols[i]
				tw := cellW(c) - lockW()
				cb.Rect.X, cb.Rect.W = right-lockW()-tw, tw
				ax := right - 2*LockArrowW - StepGap
				cb.Up = Rect{X: ax, Y: y, W: LockArrowW, H: rowH()}
				cb.Down = Rect{X: ax + LockArrowW + StepGap, Y: y, W: LockArrowW, H: rowH()}
			} else if c.Kind == CellField && c.Steps {
				cb.Rect.W = cols[i] - StepGap - StepArrowW
				ax := cx + cb.Rect.W + StepGap
				half := rowH() / 2
				cb.Up = Rect{X: ax, Y: y, W: StepArrowW, H: half}
				cb.Down = Rect{X: ax, Y: y + half, W: StepArrowW, H: half}
			}
			if c.Kind == CellText && c.OverLastDigit && c.Arrows.Op == "" {
				right := cx + cols[i]
				if stepsIn(g, i) {
					right -= StepGap + StepArrowW
				}
				tw := cellW(c)
				cb.Rect.X = min(right-CellPadX-Panel.Advance(ValFontPx)/2-tw/2, right-tw)
				cb.Rect.W = tw
			}
			lay.Cells = append(lay.Cells, cb)
			cx += cols[i] + CellGapX
		}
		y += rowH() + RowGapY
	}
}
