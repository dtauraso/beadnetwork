package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
)

const (
	menuMinW = 190
	menuPad  = 4
)

type Menu struct {
	Note  string
	Empty string
	Items []Board.Chip
	Extra []Board.Chip
	Drop  Board.Chip
}

type placed struct {
	at Rect
	m  Menu
}

func dots(l *Board.Laid, x, y, h float32, clip uint8, open bool, a Act) Rect {
	w := Board.W("⋯", Board.BtnPx) + 12
	r := Rect{X: x - w, Y: y + (h-btnH)/2, W: w, H: btnH}
	fill, edge := Board.None, Board.None
	if open {
		fill, edge = Board.On, Board.Accent
	}
	l.Box(r, clip, fill, edge)
	l.Text(r.X+w/2, r.Y, btnH, clip, "⋯", Board.BtnPx, Board.Ink, false, Board.Center)
	l.Spot(r, clip, Sheet.Cell{Act: a}, Sheet.PartBody)
	return r
}

func menuW(m Menu) float32 {
	w := float32(menuMinW)
	for _, list := range [][]Board.Chip{m.Items, m.Extra, {m.Drop}} {
		for _, c := range list {
			w = max(w, Board.W(c.Label, Board.Px)+Board.W(c.Note, 11)+40)
		}
	}
	return w
}

func drawMenu(l *Board.Laid, box Rect, p placed) {
	m, h := p.m, Board.RowH()
	n, rules := len(m.Items)+len(m.Extra), 0
	if len(m.Extra) > 0 {
		rules++
	}
	if m.Drop.Act.Op != "" {
		n, rules = n+1, rules+1
	}
	if m.Note != "" {
		n++
	}
	if len(m.Items) == 0 && m.Empty != "" {
		n++
	}
	w := min(menuW(m), box.W-2*pad)
	r := Rect{X: p.at.X + p.at.W - w, Y: p.at.Y + p.at.H + 2, W: w, H: float32(n)*h + 2*menuPad + float32(rules)*8}
	r.X = max(box.X+pad, min(r.X, box.X+box.W-pad-w))
	if r.Y+r.H > box.Y+box.H {
		r.Y = max(box.Y, p.at.Y-2-r.H)
	}
	clip := l.Clip(Rect{X: r.X - 4, Y: r.Y - 4, W: r.W + 8, H: r.H + 8})
	l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: r.X, Y: r.Y + 2, W: r.W, H: r.H}, Clip: clip, Fill: Board.Edge, Radius: 6, Alpha: 0.4})
	l.Add(Board.Mark{Kind: Board.KBox, R: r, Clip: clip, Fill: Board.White, Edge: Board.Edge, Radius: 6, Stroke: 1, Alpha: 1})
	l.Block(r, clip)
	x, y := r.X+menuPad, r.Y+menuPad
	iw := r.W - 2*menuPad
	if m.Note != "" {
		l.Text(x+8, y, h, clip, m.Note, Board.SmallPx, Board.Warn, false, Board.Left)
		y += h
	}
	item := func(c Board.Chip, ink Board.Paint) {
		row := Rect{X: x, Y: y, W: iw, H: h}
		if c.Lit {
			l.Add(Board.Mark{Kind: Board.KBox, R: row, Clip: clip, Fill: Board.On, Radius: 4, Alpha: 1})
		}
		if c.Dim {
			ink = Board.Faint
		}
		l.Text(x+8, y, h, clip, c.Label, Board.Px, ink, c.Lit, Board.Left)
		if c.Note != "" {
			l.Text(x+iw-8, y, h, clip, c.Note, 11, Board.Mute, false, Board.Right)
		}
		if !c.Dim {
			l.Spot(row, clip, Sheet.Cell{Act: c.Act}, Sheet.PartBody)
		}
		y += h
	}
	rule := func() {
		y += 4
		l.Line(Rect{X: x, Y: y - 0.5, W: iw, H: 0}, clip, Board.Hair, 1, false)
		y += 4
	}
	for _, c := range m.Items {
		item(c, Board.Ink)
	}
	if len(m.Items) == 0 && m.Empty != "" {
		l.Text(x+8, y, h, clip, m.Empty, 12, Board.Mute, false, Board.Left)
		y += h
	}
	if len(m.Extra) > 0 {
		rule()
		for _, c := range m.Extra {
			item(c, Board.Ink)
		}
	}
	if m.Drop.Act.Op != "" {
		rule()
		item(m.Drop, Board.Warn)
	}
}
