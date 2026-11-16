package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
)

type flow struct {
	x0, right, x, y, h, gap float32
}

func (f *flow) at(w float32) (float32, float32) {
	if f.x > f.x0 && f.x+w > f.right {
		f.x, f.y = f.x0, f.y+f.h+f.gap
	}
	x := f.x
	f.x += w + f.gap
	return x, f.y
}

func (f *flow) bottom() float32 { return f.y + f.h }

func band(l *Board.Laid, r Rect, fill Board.Paint) (uint8, int) {
	clip := l.Clip(r)
	l.Box(r, clip, fill, Board.None)
	return clip, len(l.Marks) - 1
}

func closeBand(l *Board.Laid, r Rect, clip uint8, bg int, f flow) float32 {
	r.H = f.bottom() + pad - r.Y
	l.Clips[clip], l.Marks[bg].R = r, r
	l.Line(Rect{X: r.X, Y: r.Y + r.H - 0.5, W: r.W, H: 0}, clip, Board.Guide, 1, false)
	return r.H
}

func chipW(c Board.Chip) float32 { return Board.W(c.Label, Board.ChipPx) + 14 }

func crumbs(l *Board.Laid, f *flow, clip uint8, cs []Board.Chip) {
	h := Board.RowH()
	sep := Board.W("›", Board.Px)
	for i, c := range cs {
		if i > 0 {
			x, y := f.at(sep)
			l.Text(x, y, h, clip, "›", Board.Px, Board.Mute, false, Board.Left)
		}
		w := min(Board.TextW(c.Label, Board.Px, map[bool]uint8{true: 7, false: 4}[c.Lit]), f.right-f.x0)
		x, y := f.at(w)
		ink := Board.Accent
		if c.Lit {
			ink = Board.Ink
		}
		l.Text(x, y, h, clip, fitText(c.Label, w, c.Lit), Board.Px, ink, c.Lit, Board.Left)
		if !c.Lit && c.Act.Op != "" {
			l.Spot(Rect{X: x, Y: y, W: w, H: h}, clip, Sheet.Cell{Act: c.Act}, Sheet.PartBody)
		}
	}
}

func scopeBar(l *Board.Laid, r Rect, narrow bool, d Data, menus *[]placed) float32 {
	clip, bg := band(l, r, Board.White)
	h := Board.RowH()
	f := flow{x0: r.X + pad, right: r.X + r.W - pad, x: r.X + pad, y: r.Y + pad, h: h, gap: Board.Gap}
	if narrow && d.More.Op != "" {
		f.right -= Board.W("⋯", Board.BtnPx) + 12 + Board.Gap
	}
	if narrow && d.PathMenu.Op != "" {
		c := Board.Chip{Label: "steps ▾", Act: d.PathMenu, Lit: d.PathOpen}
		x, y := f.at(chipW(c))
		l.Chip(x, y, clip, c)
		if d.PathOpen {
			*menus = append(*menus, placed{at: Rect{X: x, Y: y, W: chipW(c), H: h}, pane: true})
		}
	}
	crumbs(l, &f, clip, d.Crumbs)
	if d.S.Act.Op != "" && !narrow {
		s := d.S
		s.Label = "s"
		x, y := f.at(Board.FieldW(s) + 2)
		l.Field(x+2, y, clip, s)
	}
	if d.SNote != "" && !narrow {
		nw := Board.W(d.SNote, Board.SmallPx) + 12
		x, y := f.at(nw)
		l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: x, Y: y + 3, W: nw, H: h - 6}, Clip: clip, Fill: Board.WarnFill, Radius: 99, Alpha: 1})
		l.Text(x+6, y, h, clip, d.SNote, Board.SmallPx, Board.Warn, false, Board.Left)
	}
	doors := d.Doors
	if d.Insp.Title != "" {
		insp := Board.Chip{Label: "◂ inspect", Act: d.InspTab}
		if d.InspOpen {
			insp = Board.Chip{Label: "▸ inspect", Act: d.InspTab, Lit: true}
		}
		doors = append(doors[:len(doors):len(doors)], insp)
	}
	if narrow && d.More.Op != "" {
		at := dots(l, r.X+r.W-pad, r.Y+pad, h, clip, d.MoreOpen, d.More)
		if d.MoreOpen {
			m := Menu{Items: doors, Note: d.SNote}
			if d.S.Act.Op != "" {
				m.Find, m.FindLabel = &d.S, "s · ticks per 1/12 turn"
			}
			*menus = append(*menus, placed{at: at, m: m})
		}
		return closeBand(l, r, clip, bg, f)
	}
	var dw float32
	for _, c := range doors {
		dw += chipW(c) + Board.Gap
	}
	if dw > 0 {
		dw -= Board.Gap
		_, y := f.at(dw)
		x := f.right - dw
		for _, c := range doors {
			x += l.Chip(x, y, clip, c) + Board.Gap
		}
	}
	return closeBand(l, r, clip, bg, f)
}
