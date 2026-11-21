package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
)

type NewNode struct {
	Back, Add Act
	Centers   []Board.Chip
	At        []NetField
}

func newBar(l *Board.Laid, r Rect, n NewNode) {
	clip := l.Clip(r)
	h := Board.RowH()
	l.Line(Rect{X: r.X, Y: r.Y + r.H - 0.5, W: r.W, H: 0}, clip, Board.Guide, 1, false)
	x, y := r.X+pad, r.Y+pad
	x += l.Chip(x, y, clip, Board.Chip{Label: "← steps", Act: n.Back, Px: Board.BtnPx}) + Board.Gap
	x += l.Text(x, y, h, clip, "›", Board.Px, Board.Mute, false, Board.Left) + Board.Gap
	l.Text(x, y, h, clip, "New node", Board.Px, Board.Ink, true, Board.Left)
}

func newNode(l *Board.Laid, r Rect, sy float32, n NewNode) float32 {
	clip := l.Clip(r)
	l.Add(Board.Mark{Kind: Board.KBox, R: r, Clip: clip, Fill: Board.Back, Alpha: 1})
	h := Board.RowH()
	w := max(160, r.W-2*colGap)
	x, y := r.X+colGap, r.Y+colGap-sy
	ch := cheadH + 3*(h+Board.Gap) + cpad
	l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: x, Y: y, W: w, H: ch}, Clip: clip, Fill: Board.White, Edge: Board.Guide, Radius: radius6, Stroke: 1, Alpha: 1})
	l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: x + 1, Y: y + 1, W: w - 2, H: cheadH - 1}, Clip: clip, Fill: Board.Card, Radius: radius6 - 1, Corner: 2, Alpha: 1})
	tx := x + cpad
	tx += l.Text(tx, y+(cheadH-h)/2, h, clip, "new node", Board.Px, Board.Ink, true, Board.Left) + Board.Gap
	l.Text(tx, y+(cheadH-h)/2, h, clip, "placed from a center node, as a link from it would be", Board.SmallPx, Board.Mute, false, Board.Left)
	cx, cy := x+cpad, y+cheadH+Board.Gap
	fx := cx + l.Text(cx, cy, h, clip, "center", Board.SmallPx, Board.Mute, false, Board.Left)
	fx = max(fx+Board.Gap, cx+netHeadW)
	for _, c := range n.Centers {
		c.Px = Board.BtnPx
		fx += l.Chip(fx, cy, clip, c) + Board.Gap
	}
	cy += h + Board.Gap
	fx = cx + netHeadW
	l.Text(cx, cy, h, clip, "angles and r", Board.SmallPx, Board.Mute, false, Board.Left)
	for _, f := range n.At {
		fx += l.Text(fx, cy, h, clip, f.Name, Board.SmallPx, Board.Mute, false, Board.Left) + 3
		fx += l.Field(fx, cy, clip, f.Field) + Board.Gap
	}
	cy += h + Board.Gap
	l.Chip(cx, cy, clip, Board.Chip{Label: "＋ add node", Act: n.Add, Lit: true, Px: Board.BtnPx})
	y += ch + colGap
	return max(0, y+sy-(r.Y+r.H))
}
