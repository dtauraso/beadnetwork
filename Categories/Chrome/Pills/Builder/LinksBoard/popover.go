package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
)

const popW = 320

func popover(l *Board.Laid, box, bar Rect, p Pop) {
	h := Board.RowH()
	rows := float32(len(p.Fields)) + 3
	r := Rect{X: max(box.X+pad, box.X+box.W-popW-pad), Y: bar.Y + bar.H - 2, W: min(popW, box.W-2*pad), H: rows*(h+Board.Gap) + 2*pad}
	clip := l.Clip(r)
	l.Add(Board.Mark{Kind: Board.KBox, R: r, Clip: clip, Fill: Board.White, Edge: Board.Edge, Radius: 6, Stroke: 1, Alpha: 1})
	l.Block(r, clip)
	x, y := r.X+pad, r.Y+pad
	w := l.Text(x, y, h, clip, "Frame", Board.Px, Board.Ink, true, Board.Left)
	tag := Rect{X: x + w + Board.Gap, Y: y + 4, W: Board.W(p.Tag, 11.5) + 16, H: h - 8}
	l.Add(Board.Mark{Kind: Board.KBox, R: tag, Clip: clip, Fill: Board.On, Radius: 99, Alpha: 1})
	l.Weighted(tag.X+8, y, h, clip, p.Tag, 11.5, Board.Accent, Board.Semi, Board.Left)
	y += h + Board.Gap
	for _, f := range p.Fields {
		l.Text(x, y, h, clip, f.Label, Board.Px, Board.Ink, false, Board.Left)
		fx := x + 36
		fx += l.Field(fx, y, clip, f.Field) + Board.Gap
		what := ""
		if f.Was != "" {
			what = "was " + f.Was
		} else if f.Label == "Δt" {
			what = f.What
		}
		l.Text(fx, y, h, clip, what, 11, Board.Mute, false, Board.Left)
		y += h + Board.Gap
	}
	l.Text(x, y, h, clip, "Applying reloads the scene.", 12, Board.Warn, false, Board.Left)
	y += h + Board.Gap
	right := r.X + r.W - pad
	apply := Board.Chip{Label: "Apply & reload", Act: p.Apply, Lit: p.CanApply, Dim: !p.CanApply, Tip: "write the frame and reload the scene"}
	right -= Board.W(apply.Label, Board.Px) + 2*Board.PadX
	l.Chip(right, y, clip, apply)
	cancel := Board.Chip{Label: "Cancel", Act: p.Cancel, Tip: "close without changing the frame"}
	right -= Board.W(cancel.Label, Board.Px) + 2*Board.PadX + Board.Gap
	l.Chip(right, y, clip, cancel)
}
