package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
)

const thumbH = 3

func meet(a, b Rect) Rect {
	x0, y0 := max(a.X, b.X), max(a.Y, b.Y)
	x1, y1 := min(a.X+a.W, b.X+b.W), min(a.Y+a.H, b.Y+b.H)
	return Rect{X: x0, Y: y0, W: max(0, x1-x0), H: max(0, y1-y0)}
}

func sideScroll(l *Board.Laid, clip uint8, view Rect, most float32, d Card, cy *float32) (float32, uint8, func()) {
	sx := Board.Clamp(d.SX, most)
	a := d.Scroll
	a.Sign = int(most + 0.5)
	l.Spot(view, clip, Sheet.Cell{Act: a}, Sheet.PartInside)
	gclip := l.Clip(meet(view, l.Clips[clip]))
	return view.X - sx, gclip, func() {
		*cy += gapY
		track := Rect{X: view.X, Y: *cy, W: view.W, H: thumbH}
		l.Add(Board.Mark{Kind: Board.KBox, R: track, Clip: clip, Fill: Board.Hair, Radius: thumbH / 2, Alpha: 1})
		tw := max(24, view.W*view.W/(view.W+most))
		l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: view.X + (view.W-tw)*sx/most, Y: *cy, W: tw, H: thumbH}, Clip: clip, Fill: Board.Edge, Radius: thumbH / 2, Alpha: 1})
		*cy += thumbH
	}
}
