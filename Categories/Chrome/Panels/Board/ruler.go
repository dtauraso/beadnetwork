package Board

type Ruler struct {
	Start, Delta, Most float32
	Fixed              bool
	Big                bool
}

func (l *Laid) bar(r Rect, clip uint8, fill Paint, alpha float32) {
	if r.W <= 0 {
		return
	}
	l.Add(Mark{Kind: KBox, R: r, Clip: clip, Fill: fill, Radius: 2, Alpha: alpha})
}

func (l *Laid) Ruler(r Rect, clip uint8, u Ruler) {
	most := max(u.Most, u.Start+max(u.Delta, 0), 1)
	at := func(v float32) float32 { return r.X + r.W*max(0, v)/most }
	end := u.Start + u.Delta
	lo, hi := at(min(u.Start, end)), at(max(u.Start, end))
	knob := Accent
	if u.Fixed {
		knob = Faint
	}
	if u.Big {
		mid := r.Y + r.H/2
		l.bar(Rect{X: r.X, Y: mid - 2, W: r.W, H: 4}, clip, Guide, 1)
		l.bar(Rect{X: r.X, Y: mid - 2, W: at(u.Start) - r.X, H: 4}, clip, StartArc, 1)
		l.bar(Rect{X: lo, Y: mid - 3, W: hi - lo, H: 6}, clip, Accent, 0.9)
		l.Add(Mark{Kind: KBox, R: Rect{X: at(end) - 6, Y: mid - 10, W: 12, H: 20}, Clip: clip, Fill: knob, Radius: 3, Alpha: 1})
		return
	}
	top := r.Y + (r.H-14)/2
	l.bar(Rect{X: r.X, Y: top + 5, W: r.W, H: 3}, clip, Guide, 1)
	l.bar(Rect{X: r.X, Y: top + 5, W: at(u.Start) - r.X, H: 3}, clip, StartArc, 1)
	l.bar(Rect{X: lo, Y: top + 5, W: hi - lo, H: 3}, clip, Accent, 1)
	cy := top + 6.5
	l.Circle(Rect{X: at(end) - 6, Y: cy - 6, W: 12, H: 12}, clip, knob, None)
	l.Add(Mark{Kind: KCircle, R: Rect{X: at(end) - 4, Y: cy - 4, W: 8, H: 8}, Clip: clip, Fill: knob, Edge: White, Stroke: 2, Alpha: 1})
}
