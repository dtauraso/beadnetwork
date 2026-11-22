package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
)

func frameCard(l *Board.Laid, r Rect, sy float32, p Pop) float32 {
	clip := l.Clip(r)
	l.Add(Board.Mark{Kind: Board.KBox, R: r, Clip: clip, Fill: Board.Back, Alpha: 1})
	w := max(160, r.W-2*colGap)
	x, y := r.X+colGap, r.Y+colGap-sy
	c := openFitCard(l, clip, x, y, w, "frame", p.Tag)
	for _, f := range p.Fields {
		what, ink := f.What, Board.Mute
		if f.Was != "" {
			what, ink = "was "+f.Was, Board.Warn
		}
		c.row(f.Label, []fitItem{c.bare(f.Field), c.text(what, ink)})
	}
	c.row("", []fitItem{c.text("Applying reloads the scene.", Board.Warn)})
	c.row("", []fitItem{
		c.chip(Board.Chip{Label: "Apply & reload", Act: p.Apply, Lit: p.CanApply, Dim: !p.CanApply}),
		c.chip(Board.Chip{Label: "Cancel", Act: p.Cancel}),
	})
	y = c.close()
	return max(0, y+sy-(r.Y+r.H))
}
