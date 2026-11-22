package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/LinksBoard/Fill"
)

type flow struct {
	x0, right, x, y, h, gap float32
}

func (f *flow) at(w float32) (float32, float32) {
	if f.x > f.x0 && f.x+w > f.right {
		f.x, f.y = f.x0, f.y+f.h+f.gap
	}
	x := f.x
	f.x += w + f.gap
	return x, f.y
}

func (f *flow) bottom() float32 { return f.y + f.h }

func band(l *Board.Laid, r Rect, fill Board.Paint) (uint8, int) {
	clip := l.Clip(r)
	l.Box(r, clip, fill, Board.None)
	return clip, len(l.Marks) - 1
}

func closeBand(l *Board.Laid, r Rect, clip uint8, bg int, f flow) float32 {
	r.H = f.bottom() + pad - r.Y
	l.Clips[clip], l.Marks[bg].R = r, r
	l.Line(Rect{X: r.X, Y: r.Y + r.H - 0.5, W: r.W, H: 0}, clip, Board.Guide, 1, false)
	return r.H
}

func chipW(c Board.Chip) float32 { return Board.W(c.Label, Board.ChipPx) + 14 }

func crumbs(l *Board.Laid, f *flow, clip uint8, cs []Board.Chip, name Act) {
	h := Board.RowH()
	sep := Board.W("›", Board.ChipPx)
	for i, c := range cs {
		if i > 0 {
			x, y := f.at(sep)
			l.Text(x, y, h, clip, "›", Board.ChipPx, Board.Mute, false, Board.Left)
		}
		if !c.Lit {
			x, y := f.at(chipW(c))
			l.Chip(x, y, clip, c)
			continue
		}
		lw := min(Board.TextW(c.Label, Board.Px, 7), f.right-f.x0-chipW(Board.Chip{Label: "✎"})-Board.Gap-2*Board.PadX)
		x, y := f.at(lw + 2*Board.PadX)
		ew := l.Editable(x, y, lw+2*Board.PadX, clip, fitText(c.Label, lw, true), Board.Px, Board.Ink, name)
		l.Marks[len(l.Marks)-1].Bold = true
		l.Spot(Rect{X: x, Y: y, W: ew, H: h}, clip, Sheet.Cell{Act: c.Act}, Sheet.PartBody)
		pen := Board.Chip{Label: "✎", Act: c.Act}
		rx, ry := f.at(chipW(pen))
		l.Chip(rx, ry, clip, pen)
	}
}

func scopeBar(l *Board.Laid, r Rect, d Fill.Data) float32 {
	clip, bg := band(l, r, Board.White)
	f := flow{x0: r.X + pad, right: r.X + r.W - pad, x: r.X + pad, y: r.Y + pad, h: Board.RowH(), gap: Board.Gap}
	crumbs(l, &f, clip, d.Crumbs, d.CrumbName)
	if c := d.FolderPlay; c != nil {
		x, y := f.at(chipW(*c))
		l.Chip(x, y, clip, *c)
	}
	return closeBand(l, r, clip, bg, f)
}
