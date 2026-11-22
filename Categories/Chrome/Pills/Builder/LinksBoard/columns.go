package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
)

const (
	cpad    = 8
	colGap  = 8
	cheadH  = 28
	toolsH  = 26
	headH   = 28
	btnH    = 18
	radius6 = 6
)

type col struct {
	l       *Board.Laid
	clip    uint8
	x, w, y float32
	narrow  bool
	menus   *[]placed
}

func (k *col) rule(y float32, p Board.Paint) {
	k.l.Line(Rect{X: k.x + 1, Y: y + 0.5, W: k.w - 2, H: 0}, k.clip, p, 1, false)
}

func columns(l *Board.Laid, r Rect, sy float32, narrow bool, d Data, menus *[]placed) float32 {
	clip := l.Clip(r)
	l.Add(Board.Mark{Kind: Board.KBox, R: r, Clip: clip, Fill: Board.Back, Alpha: 1})
	w := max(160, r.W-2*colGap)
	y := r.Y + colGap - sy
	if d.New != nil || d.Net != nil {
		return max(0, topCard(l, r, sy, d, menus)+sy-(r.Y+r.H))
	}
	for _, c := range d.Cols {
		k := col{l: l, clip: clip, x: r.X + colGap, w: w, y: y, narrow: narrow, menus: menus}
		box := len(l.Marks)
		edge := Board.Guide
		if c.Shown {
			edge = Board.Accent
		}
		l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: k.x, Y: y, W: w}, Clip: clip, Fill: Board.White, Edge: edge, Radius: radius6, Stroke: 1, Alpha: 1})
		k.header(c)
		if c.Open {
			k.headRow(c)
			for _, n := range c.Nodes {
				k.node(n)
			}
		}
		l.Marks[box].R.H = k.y - y
		y = k.y + colGap
	}
	if d.Add.Op != "" {
		l.Btn(r.X+colGap, y, Board.RowH(), clip, "＋ step", d.Add, false)
		l.Marks[len(l.Marks)-2].Fill = Board.None
		y += Board.RowH() + colGap
	}
	return max(0, y+sy-(r.Y+r.H))
}

func (k *col) header(c Column) {
	l, x, y := k.l, k.x+cpad, k.y
	corner := uint8(2)
	if !c.Open {
		corner = 0
	}
	if !c.Open {
		l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: k.x + 1, Y: y + 1, W: k.w - 2, H: cheadH - 2}, Clip: k.clip, Fill: Board.Card, Radius: radius6 - 1, Corner: corner, Alpha: 1})
	}
	glyph := "▾"
	if !c.Open {
		glyph = "▸"
		l.Spot(Rect{X: k.x, Y: y, W: k.w, H: cheadH}, k.clip, Sheet.Cell{Act: c.Fold}, Sheet.PartBody)
	}
	l.Text(x+4, y, cheadH, k.clip, glyph, 11, Board.Ink, false, Board.Center)
	l.Spot(Rect{X: k.x, Y: y, W: cpad + 14, H: cheadH}, k.clip, Sheet.Cell{Act: c.Fold}, Sheet.PartBody)
	x += 8 + 6
	right := k.x + k.w - cpad
	if c.Menu.Op != "" {
		at := dots(l, right, y, cheadH, k.clip, c.MenuOpen, c.Menu)
		right = at.X - 6
		if c.MenuOpen {
			*k.menus = append(*k.menus, placed{at: at, m: Menu{Items: c.Tools, Drop: c.Drop}})
		}
	}
	label, shown, show := c.Label, "● shown in scene", "show in scene"
	lw := Board.TextW(label, Board.Px, Board.Semi)
	sw := Board.W(c.Sub, 12) + 6
	ew := Board.W(shown, 11)
	if x+lw+6+sw+ew > right {
		shown, show = "●", "show"
		ew = Board.W(shown, 11)
	}
	fitSub := x+lw+6+sw+ew <= right
	if x+lw+ew > right {
		label, lw = c.Short, Board.TextW(c.Short, Board.Px, Board.Semi)
	}
	l.Weighted(x, y, cheadH, k.clip, label, Board.Px, Board.Ink, Board.Semi, Board.Left)
	l.Spot(Rect{X: x, Y: y, W: lw, H: cheadH}, k.clip, Sheet.Cell{Act: c.Show}, Sheet.PartBody)
	x += lw + 6
	if c.Name.Act.Op != "" && fitSub {
		l.Editable(x-Board.PadX, y+(cheadH-Board.RowH())/2, 0, k.clip, c.Sub, 12, Board.Mute, c.Name.Act)
	} else if c.Sub != "" && fitSub {
		l.Text(x, y, cheadH, k.clip, c.Sub, 12, Board.Mute, false, Board.Left)
	}
	if c.Shown {
		l.Add(Board.Mark{Kind: Board.KCircle, R: Rect{X: right - ew, Y: y + cheadH/2 - 4, W: 8, H: 8}, Clip: k.clip, Fill: Board.Accent, Alpha: 1})
		l.Text(right, y, cheadH, k.clip, shown[len("●"):], 11, Board.Accent, false, Board.Right)
	} else if c.Show.Op != "" {
		tw := l.Text(right, y, cheadH, k.clip, show, 11, Board.Mute, false, Board.Right)
		l.Spot(Rect{X: right - tw, Y: y, W: tw, H: cheadH}, k.clip, Sheet.Cell{Act: c.Show}, Sheet.PartBody)
	}
	k.y += cheadH
	if c.Open {
		k.rule(k.y-1, Board.Guide)
	}
}

func (k *col) headRow(c Column) {
	l, x, y := k.l, k.x+cpad, k.y
	switch {
	case c.Progress != "":
		l.Text(x, y, headH, k.clip, c.Progress, 12, Board.Warn, false, Board.Left)
	case c.Head != nil:
		label := "press all " + c.Head.Picked + ":"
		f := flow{x0: x, right: k.x + k.w - cpad, x: x, y: y + 4, h: btnH + 2, gap: 6}
		if c.Head.Play.Cycle.Op != "" {
			px, py := f.at(lockW(c.Head.Play))
			k.lock(px, py, c.Head.Play)
		}
		lx, ly := f.at(Board.W(label, 12))
		l.Text(lx, ly, f.h, k.clip, label, 12, Board.Mute, false, Board.Left)
		for _, lk := range c.Head.Locks {
			lx, ly := f.at(lockW(lk))
			k.lock(lx, ly, lk)
		}
		k.y = f.bottom() + 4
		k.rule(k.y-1, Board.Guide)
		return
	default:
		return
	}
	k.y += headH
	k.rule(k.y-1, Board.Guide)
}

const lockGap = 3

func pressW() float32 { return Board.W("−", 11) + 10 }

func lockW(lk Lock) float32 {
	return Board.W(lk.Label, 12) + 10 + lockGap + 2*pressW() + Board.W(lk.Axis, 12) + 10
}

func (k *col) lock(x, y float32, lk Lock) float32 {
	l, h := k.l, float32(btnH+2)
	gw := Board.W(lk.Label, 12) + 10
	fill, ink := Board.White, Board.Ink
	if lk.On {
		fill, ink = Board.On, Board.Accent
	}
	l.Box(Rect{X: x, Y: y, W: gw, H: h}, k.clip, fill, Board.Edge)
	l.Text(x+gw/2, y, h, k.clip, lk.Label, 12, ink, lk.On, Board.Center)
	l.Spot(Rect{X: x, Y: y, W: gw, H: h}, k.clip, Sheet.Cell{Act: lk.Cycle}, Sheet.PartBody)
	sx := x + gw + lockGap
	ew, mw := pressW(), Board.W(lk.Axis, 12)+10
	l.Box(Rect{X: sx, Y: y, W: 2*ew + mw, H: h}, k.clip, Board.Card, Board.Edge)
	l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: sx + ew, Y: y + 1, W: mw, H: h - 2}, Clip: k.clip, Fill: Board.White, Alpha: 1})
	l.Line(Rect{X: sx + ew + 0.5, Y: y, W: 0, H: h}, k.clip, Board.Edge, 1, false)
	l.Line(Rect{X: sx + ew + mw - 0.5, Y: y, W: 0, H: h}, k.clip, Board.Edge, 1, false)
	if mr := (Rect{X: sx + ew, Y: y, W: mw, H: h}); lk.Value != "" {
		tw := Board.W(lk.Axis, 12)
		l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: mr.X + 1, Y: y + 1, W: mw - 2, H: h - 2}, Clip: k.clip, Stroke: 1, Alpha: 1, Edit: lk.Press})
		l.Add(Board.Mark{Kind: Board.KText, R: Rect{X: mr.X + (mw-tw)/2, Y: y, W: tw, H: h}, Clip: k.clip, Ink: Board.Ink, Text: lk.Axis, Px: 12, Align: Board.Right, Alpha: 1, Edit: lk.Press})
		l.Spot(mr, k.clip, Sheet.Cell{Kind: Sheet.CellField, Value: lk.Value, Act: lk.Press, Typed: true}, Sheet.PartBody)
	} else {
		l.Text(sx+ew+mw/2, y, h, k.clip, lk.Axis, 12, Board.Mute, false, Board.Center)
	}
	for _, a := range []struct {
		glyph string
		at    float32
		part  Sheet.Part
	}{{"−", sx, Sheet.PartDown}, {"＋", sx + ew + mw, Sheet.PartUp}} {
		l.Text(a.at+ew/2, y, h, k.clip, a.glyph, 12, Board.Mute, false, Board.Center)
		l.Spot(Rect{X: a.at, Y: y, W: ew, H: h}, k.clip, Sheet.Cell{Act: lk.Press}, a.part)
	}
	return gw + lockGap + 2*ew + mw
}
