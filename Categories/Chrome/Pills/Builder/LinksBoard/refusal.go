package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/LinksBoard/Fill"
)

const noteBH = 20

func pressedAt(l *Board.Laid, at Sheet.Act) (Rect, bool) {
	if at.Op == "" {
		return Rect{}, false
	}
	for i := len(l.Spots) - 1; i >= 0; i-- {
		if c := l.Spots[i].Cell; c.Act == at || c.ArrowAct() == at {
			return l.Spots[i].R, true
		}
	}
	return Rect{}, false
}

func noteW(d Fill.Data) float32 {
	w := Board.W(d.Toast, Board.Px) + 18 + 10 + Board.W("×", Board.Px)
	if d.Undo.Op != "" {
		w += Board.W("Undo", Board.BtnPx) + 14 + 10
	}
	return w
}

func refusal(l *Board.Laid, box, strip Rect, d Fill.Data) {
	h := float32(noteBH + 10)
	w := min(noteW(d), box.W-2*pad)
	r := Rect{X: box.X + pad, Y: strip.Y - h - pad, W: w, H: h}
	beside := false
	if at, ok := pressedAt(l, d.At); ok {
		r.X = min(max(box.X+pad, at.X), box.X+box.W-pad-w)
		r.Y = at.Y + at.H + 4
		if r.Y+h > strip.Y-4 {
			r.Y = at.Y - h - 4
		}
		beside = true
	}
	note(l, r, d, beside)
}

func note(l *Board.Laid, r Rect, d Fill.Data, beside bool) {
	clip := l.Clip(Rect{X: r.X - 4, Y: r.Y - 4, W: r.W + 8, H: r.H + 8})
	edge := Board.Edge
	if beside {
		edge = Board.Warn
	}
	l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: r.X, Y: r.Y + 2, W: r.W, H: r.H}, Clip: clip, Fill: Board.Edge, Radius: 6, Alpha: 0.5})
	l.Add(Board.Mark{Kind: Board.KBox, R: r, Clip: clip, Fill: Board.White, Edge: edge, Radius: 6, Stroke: 1, Alpha: 1})
	l.Block(r, clip)
	x, y := r.X+9, r.Y+5
	x += l.Text(x, y, noteBH, clip, d.Toast, Board.Px, Board.Ink, false, Board.Left) + 10
	if d.Undo.Op != "" {
		x += l.Btn(x, y, noteBH, clip, "Undo", d.Undo, false) + 10
	}
	xw := l.Text(x, y, noteBH, clip, "×", Board.Px, Board.Mute, false, Board.Left)
	l.Spot(Rect{X: x - 4, Y: y, W: xw + 8, H: noteBH}, clip, Sheet.Cell{Act: d.Dismiss}, Sheet.PartBody)
}
