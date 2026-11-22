package LinksBoard

import "github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"

func (in *ins) summary(a Axis) {
	l := in.l
	size := min(a.DialSize, in.w())
	a.Dial.Labels = size >= 70
	x := in.left()
	dr := Rect{X: x + (in.sum-size)/2, Y: in.y, W: size, H: size}
	l.Dial(dr, in.clip, a.Dial)
	in.y += size + 10
	l.Text(x, in.y, Board.RowH(), in.clip, a.Ends, Board.Px, Board.Ink, true, Board.Left)
	in.y += Board.RowH()
	for _, s := range a.Lines {
		in.lines(x, s, 11, Board.Mute)
	}
	in.y += 10
}
