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
		if c.Shown {
			l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: k.x, Y: y, W: w, H: 3}, Clip: clip, Fill: Board.Accent, Radius: radius6, Corner: 2, Alpha: 1})
		}
		l.Marks[box].R.H = k.y - y
		y = k.y + colGap
	}
	if d.Add.Op != "" {
		l.Btn(r.X+colGap, y, Board.RowH(), clip, "＋ step", d.Add, "add a step at the end, open", false)
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
	l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: k.x + 1, Y: y + 1, W: k.w - 2, H: cheadH - 1}, Clip: k.clip, Fill: Board.Card, Radius: radius6 - 1, Corner: corner, Alpha: 1})
	glyph, tip := "▾", "close this step"
	if !c.Open {
		glyph, tip = "▸", "open this step"
		l.Spot(Rect{X: k.x, Y: y, W: k.w, H: cheadH}, k.clip, Sheet.Cell{Act: c.Fold, Tip: tip}, Sheet.PartBody)
	}
	l.Text(x+6, y, cheadH, k.clip, glyph, 11, Board.Mute, false, Board.Center)
	l.Spot(Rect{X: k.x, Y: y, W: cpad + 18, H: cheadH}, k.clip, Sheet.Cell{Act: c.Fold, Tip: tip}, Sheet.PartBody)
	x += 12 + 6
	label, shown, show := c.Label, "● in scene", "show in scene"
	if k.narrow {
		label, shown, show = c.Short, "●", "show"
	}
	lw := l.Text(x, y, cheadH, k.clip, label, Board.Px, Board.Ink, true, Board.Left)
	l.Spot(Rect{X: x, Y: y, W: lw, H: cheadH}, k.clip, Sheet.Cell{Act: c.Show, Tip: "show this step in the scene · hover to ghost its places"}, Sheet.PartBody)
	x += lw + 6
	if c.Name.Act.Op != "" {
		l.Editable(x-Board.PadX, y+(cheadH-Board.RowH())/2, 0, k.clip, c.Sub, Board.Px, Board.Mute, c.Name.Act)
	} else if c.Sub != "" && !k.narrow {
		l.Text(x, y, cheadH, k.clip, c.Sub, Board.Px, Board.Mute, false, Board.Left)
	}
	right := k.x + k.w - cpad
	if c.Menu.Op != "" {
		at := dots(l, right, y, cheadH, k.clip, c.MenuOpen, c.Menu, "this step's tools")
		right = at.X - 8
		if c.MenuOpen {
			*k.menus = append(*k.menus, placed{at: at, m: Menu{Items: c.Tools, Drop: c.Drop}})
		}
	}
	if c.Shown {
		l.Weighted(right, y, cheadH, k.clip, shown, 11, Board.Accent, Board.Semi, Board.Right)
	} else if c.Show.Op != "" {
		tw := l.Text(right, y, cheadH, k.clip, show, 11, Board.Mute, false, Board.Right)
		l.Spot(Rect{X: right - tw, Y: y, W: tw, H: cheadH}, k.clip, Sheet.Cell{Act: c.Show, Tip: "show this step in the scene"}, Sheet.PartBody)
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
		var need float32
		for _, lk := range c.Head.Locks {
			need += lockW(lk) + 8
		}
		if x+Board.W(label, 11)+6+need > k.x+k.w-cpad {
			label = "all:"
		}
		x += l.Text(x, y, headH, k.clip, label, 11, Board.Mute, false, Board.Left) + 6
		for _, lk := range c.Head.Locks {
			x += k.lock(x, y+4, lk) + 8
		}
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
	l.Spot(Rect{X: x, Y: y, W: gw, H: h}, k.clip, Sheet.Cell{Act: lk.Cycle, Tip: "lock: click to cycle · off → ↑ → ↓"}, Sheet.PartBody)
	sx := x + gw + lockGap
	ew, mw := pressW(), Board.W(lk.Axis, 12)+10
	l.Box(Rect{X: sx, Y: y, W: 2*ew + mw, H: h}, k.clip, Board.White, Board.Edge)
	l.Line(Rect{X: sx + ew + 0.5, Y: y, W: 0, H: h}, k.clip, Board.Edge, 1, false)
	l.Line(Rect{X: sx + ew + mw - 0.5, Y: y, W: 0, H: h}, k.clip, Board.Edge, 1, false)
	l.Text(sx+ew+mw/2, y, h, k.clip, lk.Axis, 12, Board.Mute, false, Board.Center)
	for _, a := range []struct {
		glyph string
		at    float32
		part  Sheet.Part
		tip   string
	}{{"−", sx, Sheet.PartDown, "press every picked link in this step down on " + lk.Axis}, {"＋", sx + ew + mw, Sheet.PartUp, "press every picked link in this step up on " + lk.Axis}} {
		l.Text(a.at+ew/2, y, h, k.clip, a.glyph, 11, Board.Ink, false, Board.Center)
		l.Spot(Rect{X: a.at, Y: y, W: ew, H: h}, k.clip, Sheet.Cell{Act: lk.Press, Tip: a.tip}, a.part)
	}
	return gw + lockGap + 2*ew + mw
}
