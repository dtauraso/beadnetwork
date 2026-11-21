package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
)

const dropW = 560

func dropBody(l *Board.Laid, r Rect, sy float32, d Data, menus *[]placed) float32 {
	if d.New != nil {
		return newNode(l, r, sy, *d.New)
	}
	return network(l, r, sy, *d.Net, menus)
}

func dropCard(l *Board.Laid, box, bar, st Rect, sy float32, d Data, menus *[]placed) float32 {
	w := min(dropW, box.W-2*pad)
	top := bar.Y + bar.H - 2
	room := st.Y - pad - top
	var scratch []placed
	need := dropBody(Board.New(box), Rect{X: box.X, Y: top, W: w, H: 1}, 0, d, &scratch) + 1
	r := Rect{X: box.X + box.W - w - pad, Y: top, W: w, H: max(0, min(need, room))}
	l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: r.X, Y: r.Y + 2, W: r.W, H: r.H}, Clip: l.Clip(box), Fill: Board.Edge, Radius: 6, Alpha: 0.4})
	clip := l.Clip(r)
	l.Block(r, clip)
	most := dropBody(l, r, sy, d, menus)
	l.Add(Board.Mark{Kind: Board.KBox, R: r, Clip: clip, Edge: Board.Edge, Radius: 6, Stroke: 1, Alpha: 1})
	return most
}
