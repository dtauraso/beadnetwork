package LinksBoard

import (
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/LinksBoard/Fill"
)

const (
	nodeRad  = 9
	headLen  = 6
	headTurn = 0.45
	twinGap  = 3
)

type pt struct{ x, y float32 }

func spots(r Rect, ns []Fill.NetNode) ([]pt, float32) {
	minX, maxX, minY, maxY := ns[0].X, ns[0].X, ns[0].Y, ns[0].Y
	for _, nd := range ns {
		minX, maxX, minY, maxY = min(minX, nd.X), max(maxX, nd.X), min(minY, nd.Y), max(maxY, nd.Y)
	}
	m := float32(nodeRad + 6)
	iw, ih := max(1, r.W-2*m), max(1, r.H-2*m)
	s := float32(math.Inf(1))
	if dx := maxX - minX; dx > 1e-6 {
		s = iw / dx
	}
	if dy := maxY - minY; dy > 1e-6 {
		s = min(s, ih/dy)
	}
	if math.IsInf(float64(s), 1) {
		s = 0
	}
	cy, my := (minY+maxY)/2, r.Y+r.H/2
	out := make([]pt, len(ns))
	for i, nd := range ns {
		out[i] = pt{r.X + m + (nd.X-minX)*s, my - (nd.Y-cy)*s}
	}
	return out, (maxX-minX)*s + 2*m
}

func graph(l *Board.Laid, clip uint8, r Rect, n Fill.Net) {
	at, used := spots(r, n.Nodes)
	r.W = min(r.W, used)
	l.Box(r, clip, Board.Back, Board.None)
	for i, row := range n.Cells {
		for j, c := range row {
			if !c.Linked {
				continue
			}
			twin := n.Cells[j][i].Linked
			arrow(l, clip, at[i], at[j], c.Picked, twin)
		}
	}
	for i, nd := range n.Nodes {
		p := at[i]
		box := Rect{X: p.x - nodeRad, Y: p.y - nodeRad, W: 2 * nodeRad, H: 2 * nodeRad}
		l.Circle(box, clip, Board.White, Board.Ink)
		l.Text(p.x, box.Y, box.H, clip, nd.ID, Board.SmallPx, Board.Ink, true, Board.Center)
	}
}

func arrow(l *Board.Laid, clip uint8, p, q pt, picked, twin bool) {
	dx, dy := q.x-p.x, q.y-p.y
	d := float32(math.Hypot(float64(dx), float64(dy)))
	if d < 2*nodeRad+headLen {
		return
	}
	ux, uy := dx/d, dy/d
	if twin {
		p.x, p.y, q.x, q.y = p.x-uy*twinGap, p.y+ux*twinGap, q.x-uy*twinGap, q.y+ux*twinGap
	}
	a := pt{p.x + ux*nodeRad, p.y + uy*nodeRad}
	b := pt{q.x - ux*(nodeRad+1), q.y - uy*(nodeRad+1)}
	ink, stroke := Board.Mute, float32(1.25)
	if picked {
		ink, stroke = Board.Accent, 2
	}
	l.Line(Rect{X: a.x, Y: a.y, W: b.x - a.x, H: b.y - a.y}, clip, ink, stroke, false)
	for _, t := range []float64{headTurn, -headTurn} {
		s, c := math.Sincos(t)
		hx := -(ux*float32(c) - uy*float32(s)) * headLen
		hy := -(ux*float32(s) + uy*float32(c)) * headLen
		l.Line(Rect{X: b.x, Y: b.y, W: hx, H: hy}, clip, ink, stroke, false)
	}
}
