package LinksBoard

import "github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"

func fieldW(f Board.Field) float32 {
	fit := f.Fit
	if len(fit) < len(f.Value) {
		fit = f.Value
	}
	return Board.W(fit, Board.Px) + 2*Board.PadX
}

func (in *ins) ruler(a Axis) {
	l := in.l
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
	in.y += rulerH + 6
	l.Text(in.x, in.y, Board.RowH(), in.clip, a.Ends, Board.Px, Board.Ink, true, Board.Left)
	in.y += Board.RowH()
	if a.Fixed != "" {
		in.lines(in.x, "🔒 fixed · "+a.Fixed, 11, Board.Warn)
	}
	for _, s := range a.Lines {
		in.lines(in.x, s, 11, Board.Mute)
	}
	in.y += 10
}

func (in *ins) hint(s string) {
	in.lines(in.x, s, 11, Board.Mute)
	in.y += ipadY
}
