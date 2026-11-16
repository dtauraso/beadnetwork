package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
)

func (in *ins) summary(a Axis) {
	l := in.l
	size := min(a.DialSize, in.w())
	a.Dial.Labels = size >= 70
	dr := Rect{X: in.x + (in.w()-size)/2, Y: in.y, W: size, H: size}
	l.Dial(dr, in.clip, a.Dial)
	if a.DialAct.Op != "" {
		axis := Sheet.DialPhi
		if a.Dial.Axis == Board.Theta {
			axis = Sheet.DialTheta
		}
		l.Spot(dr, in.clip, Sheet.Cell{Act: a.DialAct, Dial: Sheet.Dial{Axis: axis}}, Sheet.PartBody)
	}
	in.y += size + 10
	l.Text(in.x, in.y, Board.RowH(), in.clip, a.Ends, Board.Px, Board.Ink, true, Board.Left)
	in.y += Board.RowH()
	for _, s := range a.Lines {
		in.lines(in.x, s, 11, Board.Mute)
	}
	in.y += 10
}
