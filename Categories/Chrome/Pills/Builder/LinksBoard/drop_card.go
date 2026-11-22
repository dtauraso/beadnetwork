package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
)

func dropBody(l *Board.Laid, r Rect, sy float32, d Data, menus *[]placed) float32 {
	if d.Pop != nil {
		return frameCard(l, r, sy, *d.Pop)
	}
	if d.New != nil {
		return newNode(l, r, sy, *d.New)
	}
	return network(l, r, sy, *d.Net, menus)
}

func topCard(l *Board.Laid, r Rect, sy float32, d Data, menus *[]placed) float32 {
	if d.New == nil && d.Net == nil && d.Pop == nil {
		return r.Y - sy
	}
	var scratch []placed
	need := dropBody(Board.New(r), Rect{X: r.X, Y: r.Y, W: r.W, H: 1}, 0, d, &scratch) + 1
	dropBody(l, r, sy, d, menus)
	return r.Y - sy + need
}
