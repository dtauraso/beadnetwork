package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
)

const tabPadX = 10

func tabBand(l *Board.Laid, r Rect, tabs []Board.Chip) float32 {
	clip, bg := band(l, r, Board.Card)
	h := Board.RowH() + 4
	f := flow{x0: r.X + pad, right: r.X + r.W - pad, x: r.X + pad, y: r.Y + 2, h: h, gap: 2}
	for _, t := range tabs {
		w := Board.W(t.Label, Board.Px) + 2*tabPadX
		x, y := f.at(w)
		tr := Rect{X: x, Y: y, W: w, H: h}
		ink := Board.Mute
		if t.Lit {
			ink = Board.Accent
			l.Box(tr, clip, Board.White, Board.None)
			l.Line(Rect{X: x, Y: y + h - 1, W: w, H: 0}, clip, Board.Accent, 2, false)
		}
		l.Text(x+w/2, y, h, clip, t.Label, Board.Px, ink, t.Lit, Board.Center)
		l.Spot(tr, clip, Sheet.Cell{Act: t.Act}, Sheet.PartBody)
	}
	f.y -= pad - 2
	return closeBand(l, r, clip, bg, f)
}
