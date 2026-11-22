package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
)

type Rect = Board.Rect

const (
	pad      = 8
	TreeMinW = 120
	colsMinW = 240
	splitW   = 8
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
	bar.H = steps(l, box, bar, mid, sy, narrow, d, &menus)
	strip(l, st, d)
	if d.Toast != "" {
		refusal(l, box, st, d)
	}
	for i := 0; i < len(menus); i++ {
		p := menus[i]
		if p.pane {
			treePop(l, box, p.at, d, &menus)
			continue
		}
		drawMenu(l, box, p)
	}
	return l
}

func steps(l *Board.Laid, box, bar, mid Rect, sy float32, narrow bool, d Data, menus *[]placed) float32 {
	bar.H = scopeBar(l, bar, narrow, d, menus)
	end := mid.Y + mid.H
	var split Rect
	if !narrow && (len(d.Tree) > 0 || d.TreeFind != nil) {
		top := bar.Y + bar.H
		tw := float32(treeW)
		if d.TreeW > 0 {
			tw = d.TreeW
		}
		tw = max(TreeMinW, min(tw, mid.W-colsMinW))
		r := Rect{X: mid.X, Y: top, W: tw, H: end - top}
		split = Rect{X: r.X + tw - splitW/2, Y: top, W: splitW, H: r.H}
		most := treeMost(r, d)
		if a := d.TreeScroll; a.Op != "" {
			a.I = int(most)
			l.Spot(r, l.Clip(r), Sheet.Cell{Act: a}, Sheet.PartInside)
		}
		treePane(l, r, Board.Clamp(d.TreeSY, most), d, menus)
		mid.X, mid.W = mid.X+tw, mid.W-tw
	}
	defer func() {
		if a := d.TreeSplit; a.Op != "" && split.W > 0 {
			a.I = int(box.X)
			l.Spot(split, l.Clip(split), Sheet.Cell{Act: a}, Sheet.PartBody)
		}
	}()
	mid.Y = bar.Y + bar.H
	mid.H = end - mid.Y
	if d.Empty != nil && d.New == nil && d.Net == nil && d.Pop == nil {
		empty(l, mid, *d.Empty)
	} else {
		stack := mid
		if d.InspOpen {
			iw := min(inspW, max(inspMinW, mid.W*0.45))
			if need := cardsW(d); need > 0 {
				iw = max(inspMinW, min(iw, mid.W-need))
			}
			stack.W -= iw
			ir := Rect{X: mid.X + stack.W, Y: mid.Y, W: iw, H: mid.H}
			most := inspector(Board.New(ir), ir, 0, d.Insp)
			if a := d.InspScroll; a.Op != "" {
				a.I = int(most)
				l.Spot(ir, l.Clip(ir), Sheet.Cell{Act: a}, Sheet.PartInside)
			}
			inspector(l, ir, Board.Clamp(d.InspSY, most), d.Insp)
		}
		l.MostY = columns(l, stack, sy, stack.W < colNarrowW, d, menus)
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
