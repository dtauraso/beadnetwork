package Sheet

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
)

func stepsIn(g Group, i int) bool {
	for _, row := range g.Rows {
		if i < len(row) && row[i].Kind == CellField && row[i].Steps {
			return true
		}
	}
	return false
}

func stepPair(right, y float32) (up, down Rect) {
	half := rowH() / 2
	ax := right - StepArrowW
	return Rect{X: ax, Y: y, W: StepArrowW, H: half}, Rect{X: ax, Y: y + half, W: StepArrowW, H: half}
}

func (lay *Layout) place(box int, g Group, x, y, dx float32) {
	if g.Title != "" {
		lay.Heads = append(lay.Heads, HeadBox{Box: box, Rect: Rect{X: x, Y: y, W: groupW(g), H: titleH()}, Title: g.Title, Fold: g.Fold})
		y += titleH() + TitleGap
	}
	cols := columns(g)
	for _, row := range g.Rows {
		cx := x - dx
		for i, c := range row {
			cb := CellBox{Box: box, Cell: c, Rect: Rect{X: cx, Y: y, W: cols[i], H: rowH()}, Slides: g.Slides}
			if c.Arrows.Op != "" {
				right := cx + cols[i]
				tw := textW(c)
				lx := right - lockW()
				cb.Rect.X, cb.Rect.W = lx-tw, tw
				cb.Up, cb.Down = stepPair(right, y)
			} else if c.Kind == CellField && c.Steps {
				cb.Rect.W = cols[i] - StepGap - StepArrowW
				cb.Up, cb.Down = stepPair(cx+cols[i], y)
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
