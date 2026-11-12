package Sheet

import (
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
)

func cellW(c Cell) float32 {
	switch c.Kind {
	case CellField:
		if c.Bare {
			v := c.Value
			if v == "" {
				v = c.Key
			}
			if len(v) < len(c.Fit) {
				v = c.Fit
			}
			return Panel.TextWidth(v, KeyFontPx) + 2*CellPadX
		}
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
	w := textW(c)
	if c.Fit != "" {
		w = max(w, Panel.TextWidth(c.Fit, KeyFontPx)+2)
	}
	if c.Arrows.Op != "" {
		w += lockW()
	}
	return w
}

func textW(c Cell) float32 {
	if c.Boxed {
		return Panel.TextWidth(c.Key, KeyFontPx) + 2*CellPadX
	}
	return Panel.TextWidth(c.Key, KeyFontPx) + 2
}

func lockW() float32  { return LockGap + StepArrowW }
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

func indentOf(g Group) float32 {
	in := float32(g.Depth) * IndentX
	if g.Indent {
		in += IndentX
	}
	return in
}

func shownW(g Group, room float32) float32 {
	in := indentOf(g)
	return in + max(slideW(g, room-in), Panel.TextWidth(g.Title, TitleFontPx))
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
	var docked []string
	for _, p := range pills {
		if p.Bottom {
			docked = append(docked, p.Label)
		}
	}
	docked = st.ordered(docked)
	on := min(max(st.Tab, 0), len(docked)-1)
	for _, p := range pills {
		if p.Bottom && p.Label != docked[on] {
			continue
		}
		stack := right
		if p.Left {
			stack = left
		}
		box := len(lay.Pills)
		pb := PillBox{Label: p.Label, Left: p.Left, Open: true}
		if !p.Bottom {
			pb.Pill, pb.Open = stack.AddPill(), st.Open[p.Label]
		}
		if !pb.Open {
			stack.EndGroup()
			lay.Pills = append(lay.Pills, pb)
			continue
		}
		room := stack.ContentRoomW()
		if p.Bottom {
			room = stack.DockRoomW()
		}
		groups := make([]Group, len(p.Groups))
		popMax := make([]float32, len(p.Groups))
		for i, g := range p.Groups {
			groups[i], popMax[i] = popWindow(wrapped(g, room), st.PopScroll[p.Label])
		}
		p.Groups = groups
		spots, contentW, contentH := flow(p.Groups, room)
		var popover, bar Rect
		var x, y, maxScroll float32
		if p.Bottom {
			bar, popover, x, y, maxScroll = stack.AddDock(contentH, st.dockH(DockKey), barH(), st.Scroll[p.Label])
			pb.Grab = Rect{X: bar.X, Y: bar.Y - Panel.DockGrabH/2, W: bar.W, H: Panel.DockGrabH}
		} else {
			popover, x, y, maxScroll = stack.AddWideScrollingPopover(contentW, contentH, st.Scroll[p.Label])
			stack.EndGroup()
		}
		pb.Popover, pb.MaxScroll = popover, maxScroll
		lay.Pills = append(lay.Pills, pb)
		top := y
		for gi, g := range p.Groups {
			x, y := x+spots[gi].X, top+spots[gi].Y
			dx := float32(0)
			in := indentOf(g)
			if g.Slides {
				vis := slideW(g, room-in)
				pb := &lay.Pills[box]
				pb.Slide = Rect{X: x, Y: popover.Y, W: max(pb.Slide.W, in+vis), H: popover.H}
				table := g.Table
				if table == "" {
					table = strconv.Itoa(gi)
				}
				most := max(0, colsW(g)-vis)
				pb.Slides = append(pb.Slides, SlideBox{Table: table, Rect: Rect{X: x, Y: y, W: in + vis, H: groupH(g)}, Max: most})
				dx = max(0, min(st.ScrollX[slideKey(p.Label, table)], most))
			}
			if g.Card {
				lay.Cards = append(lay.Cards, CardBox{Box: box, Rect: Rect{X: x + in, Y: y, W: outerW(g, room) - in, H: outerH(g)}, Pop: g.Pop != nil, Cell: len(lay.Cells), Max: popMax[gi]})
				x, y = x+CardPad, y+CardPad
			}
			for d := 0; d < g.Depth; d++ {
				guide := Rect{X: x + float32(d)*IndentX + GuideX, Y: y - Panel.Gap, W: 1, H: outerH(g) + Panel.Gap}
				lay.Cards = append(lay.Cards, CardBox{Box: box, Rect: guide, Cell: len(lay.Cells), Band: BandGuide})
			}
			var right float32
			if g.Row.Op != "" {
				band := Rect{X: x - BandPad, Y: y - RowGapY/2, W: max(room, in+colsW(g)) + 2*BandPad, H: groupH(g)}
				lay.Rows = append(lay.Rows, RowBox{Box: box, Rect: band, Row: g.Row})
				if g.Band == BandSelected {
					lay.Cards = append(lay.Cards, CardBox{Box: box, Rect: band, Cell: len(lay.Cells), Band: BandSelected})
					lay.Nav = true
				}
				right = band.X + band.W - BandPad
			}
			lay.place(box, g, x+in, y, dx, right)
		}
		if p.Bottom {
			lay.addBar(docked, on, bar, st.Folded[DockKey])
		}
	}
	return lay
}
