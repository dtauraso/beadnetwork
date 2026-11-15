package LinksBoard

import "github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"

func fieldW(f Board.Field) float32 {
	fit := f.Fit
	if len(fit) < len(f.Value) {
		fit = f.Value
	}
	return Board.W(fit, Board.Px) + 2*Board.PadX
}

func stepW(s Board.Stepper) float32 {
	v := max(26, Board.W(s.Plain, Board.Px)+2*Board.PadX)
	if s.Act.Op != "" {
		v = fieldW(s.Field)
	}
	return 2*(Board.W("＋", Board.BtnPx)+14) + 2*Board.Gap + v
}

func (in *ins) ruler(a Axis) {
	l, h := in.l, Board.RowH()
	x := in.x + l.Weighted(in.x, in.y, headTH, in.clip, a.Label, 11, Board.Mute, Board.Semi, Board.Left) + 6
	in.lines(x, a.Sub, 11, Board.Mute)
	in.y += 4
	rr := Rect{X: in.x + 4, Y: in.y, W: in.w() - 8, H: 20}
	u := *a.Ruler
	u.Big = true
	l.Ruler(rr, in.clip, u)
	most := max(u.Most, u.Start+max(u.Delta, 0), 1)
	at := func(v float32) float32 { return rr.X + rr.W*max(0, v)/most }
	ty := in.y + 20
	l.Text(rr.X, ty, 12, in.clip, a.RulerTags[0], 11, Board.Mute, false, Board.Left)
	sx := min(at(u.Start), rr.X+rr.W-Board.W(a.RulerTags[1], 11))
	sw := l.Text(sx, ty, 12, in.clip, a.RulerTags[1], 11, Board.Mute, false, Board.Left)
	if ex := at(u.Start + u.Delta); u.Delta != 0 && (ex > sx+sw+4 || ex+Board.W(a.RulerTags[2], 11) < sx) {
		l.Text(ex, ty, 12, in.clip, a.RulerTags[2], 11, Board.Mute, false, Board.Left)
	}
	in.y += rulerH + 4
	x = in.x
	right := in.r.X + in.r.W - ipadX
	x += l.Stepper(x, in.y, in.clip, a.Turn) + 6
	if a.Inc.Steps.Op != "" {
		if x+Board.W(a.IncTag, Board.Px)+6+fieldW(a.Inc.Field) > right {
			x, in.y = in.x, in.y+h+4
		}
		x += l.Text(x, in.y, h, in.clip, a.IncTag, Board.Px, Board.Mute, false, Board.Left) + 6
		x += l.Field(x, in.y, in.clip, a.Inc.Field) + 6
	}
	if a.Fixed != "" {
		cw := Board.W("🔒 fixed", 12) + 14
		if x+cw > right {
			x, in.y = in.x, in.y+h+4
		}
		l.Box(Rect{X: x, Y: in.y + 2, W: cw, H: h - 4}, in.clip, Board.White, Board.Edge)
		l.Text(x+cw/2, in.y, h, in.clip, "🔒 fixed", 12, Board.Ink, false, Board.Center)
		x += cw + 6
	}
	if ew := Board.W(a.Ends, Board.Px); right-ew >= x {
		l.Text(right, in.y, h, in.clip, a.Ends, Board.Px, Board.Mute, false, Board.Right)
		in.y += h + 4
	} else {
		in.y += h + 4
		l.Text(in.x, in.y, subH, in.clip, a.Ends, Board.Px, Board.Mute, false, Board.Left)
		in.y += subH + 2
	}
	if a.Fixed != "" {
		in.lines(in.x, a.Fixed, 12, Board.Mute)
	}
	for _, s := range a.Lines {
		in.lines(in.x, s, 11, Board.Mute)
	}
	in.y += 8
}

func (in *ins) hint(s string) {
	parts := wrap(s, in.w()-16, 12)
	bh := float32(len(parts))*subH + 12
	in.l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: in.x, Y: in.y, W: in.w(), H: bh}, Clip: in.clip, Fill: Board.Card, Radius: radius6, Alpha: 1})
	y := in.y + 6
	for _, t := range parts {
		in.l.Text(in.x+8, y, subH, in.clip, t, 12, Board.Mute, false, Board.Left)
		y += subH
	}
	in.y += bh + ipadY
}
