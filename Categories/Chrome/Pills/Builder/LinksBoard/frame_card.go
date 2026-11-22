package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
)

func frameCard(l *Board.Laid, r Rect, sy float32, p Pop) float32 {
	clip := l.Clip(r)
	l.Add(Board.Mark{Kind: Board.KBox, R: r, Clip: clip, Fill: Board.Back, Alpha: 1})
	h := Board.RowH()
	w := max(160, r.W-2*colGap)
	x, y := r.X+colGap, r.Y+colGap-sy
	ch := cheadH + float32(len(p.Fields)+2)*(h+Board.Gap) + cpad
	l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: x, Y: y, W: w, H: ch}, Clip: clip, Fill: Board.White, Edge: Board.Guide, Radius: radius6, Stroke: 1, Alpha: 1})
	l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: x + 1, Y: y + 1, W: w - 2, H: cheadH - 1}, Clip: clip, Fill: Board.Card, Radius: radius6 - 1, Corner: 2, Alpha: 1})
	tx := x + cpad
	tx += l.Text(tx, y+(cheadH-h)/2, h, clip, "frame", Board.Px, Board.Ink, true, Board.Left) + Board.Gap
	l.Text(tx, y+(cheadH-h)/2, h, clip, p.Tag, Board.SmallPx, Board.Mute, false, Board.Left)
	cx, cy := x+cpad, y+cheadH+Board.Gap
	for _, f := range p.Fields {
		l.Text(cx, cy, h, clip, f.Label, Board.SmallPx, Board.Mute, false, Board.Left)
		fx := cx + netHeadW
		fx += l.Field(fx, cy, clip, f.Field) + Board.Gap
		what, ink := f.What, Board.Mute
		if f.Was != "" {
			what, ink = "was "+f.Was, Board.Warn
		}
		l.Text(fx, cy, h, clip, what, Board.SmallPx, ink, false, Board.Left)
		cy += h + Board.Gap
	}
	l.Text(cx, cy, h, clip, "Applying reloads the scene.", Board.SmallPx, Board.Warn, false, Board.Left)
	cy += h + Board.Gap
	fx := cx + l.Chip(cx, cy, clip, Board.Chip{Label: "Apply & reload", Act: p.Apply, Lit: p.CanApply, Dim: !p.CanApply, Px: Board.BtnPx}) + Board.Gap
	l.Chip(fx, cy, clip, Board.Chip{Label: "Cancel", Act: p.Cancel, Px: Board.BtnPx})
	y += ch + colGap
	return max(0, y+sy-(r.Y+r.H))
}
