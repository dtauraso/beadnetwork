package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
)

type Rect = Board.Rect

const (
	pad      = 8
	narrowW  = 520
	stripPad = 8
)

func Lay(box Rect, sx, sy float32, d Data) *Board.Laid {
	l := lay(box, sx, sy, d)
	if cx, cy := Board.Clamp(sx, l.MostX), Board.Clamp(sy, l.MostY); cx != sx || cy != sy {
		return lay(box, cx, cy, d)
	}
	return l
}

func lay(box Rect, sx, sy float32, d Data) *Board.Laid {
	l := Board.New(box)
	h := Board.RowH()
	bar := Rect{X: box.X, Y: box.Y, W: box.W, H: h + 2*pad}
	st := Rect{X: box.X, Y: box.Y + box.H - h - stripPad, W: box.W, H: h + stripPad}
	mid := Rect{X: box.X, Y: bar.Y + bar.H, W: box.W, H: st.Y - bar.Y - bar.H}
	narrow := box.W < narrowW
	var menus []placed
	if d.Net != nil {
		netBar(l, bar, *d.Net)
		l.MostY = network(l, mid, sy, *d.Net, &menus)
	} else {
		bar.H = steps(l, box, bar, mid, sy, narrow, d, &menus)
	}
	strip(l, st, d)
	if d.Toast != "" {
		refusal(l, box, st, d)
	}
	if d.Pop != nil {
		popover(l, box, bar, *d.Pop)
	}
	for _, p := range menus {
		if p.pane {
			treePop(l, box, p.at, d)
			continue
		}
		drawMenu(l, box, p)
	}
	return l
}

func steps(l *Board.Laid, box, bar, mid Rect, sy float32, narrow bool, d Data, menus *[]placed) float32 {
	bar.H = scopeBar(l, bar, narrow, d, menus)
	end := mid.Y + mid.H
	if !narrow && (len(d.Tree) > 0 || d.TreeFind != nil) {
		top := bar.Y + bar.H
		l.MostY = treePane(l, Rect{X: mid.X, Y: top, W: treeW, H: end - top}, sy, d, menus)
		mid.X, mid.W = mid.X+treeW, mid.W-treeW
	}
	mid.Y = bar.Y + bar.H
	mid.H = end - mid.Y
	if d.Empty != nil {
		empty(l, mid, *d.Empty)
	} else {
		stack := mid
		if d.Insp.Title != "" && d.InspOpen {
			iw := min(inspW, max(inspMinW, mid.W*0.45))
			stack.W -= iw
			l.MostY = inspector(l, Rect{X: mid.X + stack.W, Y: mid.Y, W: iw, H: mid.H}, sy, d.Insp)
		}
		l.MostY = max(l.MostY, columns(l, stack, sy, stack.W < colNarrowW, d, menus))
	}
	return bar.H
}

func strip(l *Board.Laid, r Rect, d Data) {
	clip := l.Clip(r)
	l.Box(Rect{X: r.X, Y: r.Y, W: r.W, H: r.H + 4}, clip, Board.Card, Board.None)
	l.Line(Rect{X: r.X, Y: r.Y + 0.5, W: r.W, H: 0}, clip, Board.Guide, 1, false)
	h := Board.RowH()
	x, y := r.X+pad, r.Y+stripPad/2
	x += l.Text(x, y, h, clip, d.Status, 11.5, Board.Mute, false, Board.Left) + 12
	hint := "click to open · drag a dial · two-finger scroll"
	if w := Board.W(hint, 11.5); r.X+r.W-pad-w > x+12 {
		l.Text(r.X+r.W-pad, y, h, clip, hint, 11.5, Board.Mute, false, Board.Right)
	}
}

func empty(l *Board.Laid, r Rect, e Empty) {
	clip := l.Clip(r)
	h := Board.RowH()
	cx, y := r.X+r.W/2, r.Y+28
	l.Text(cx, y, h, clip, e.Title, 15, Board.Ink, true, Board.Center)
	y += h + 10
	if e.Door.Act.Op != "" {
		w := Board.W(e.Door.Label, Board.BtnPx) + 14
		l.Btn(cx-w/2, y, Board.SegH, clip, e.Door.Label, e.Door.Act, false)
		y += Board.SegH + 10
	}
	l.Text(cx, y, h, clip, e.Sub, 12, Board.Mute, false, Board.Center)
}
