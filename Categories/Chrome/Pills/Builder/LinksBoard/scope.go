package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
)

const (
	tinyPx  = 11
	stepGap = 4
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

func scopeBar(l *Board.Laid, r Rect, narrow bool, d Data, menus *[]placed) float32 {
	clip, bg := band(l, r, Board.White)
	h := Board.RowH()
	f := flow{x0: r.X + pad, right: r.X + r.W - pad, x: r.X + pad, y: r.Y + pad, h: h, gap: Board.Gap}
	sepW := Board.W("›", Board.ChipPx) + Board.Gap
	if narrow && d.More.Op != "" {
		f.right -= Board.W("⋯", Board.BtnPx) + 12 + Board.Gap
	}
	for i, c := range d.Crumbs {
		if narrow && c.Short != "" {
			c.Label = c.Short
		}
		lw := Board.TextW(c.Label, Board.Px, 7)
		if c.Act.Op != "" && c.Edit.Op == "" {
			lw = chipW(Board.Chip{Label: c.Label})
		}
		w := lw
		if i > 0 {
			w += sepW
		}
		if c.Tag != "" {
			w += Board.W(c.Tag, Board.Px) + Board.Gap
		}
		x, y := f.at(w)
		if i > 0 {
			x += l.Text(x, y, h, clip, "›", Board.ChipPx, Board.Mute, false, Board.Left) + Board.Gap
		}
		if c.Tag != "" {
			x += l.Text(x, y, h, clip, c.Tag, Board.Px, Board.Mute, false, Board.Left) + Board.Gap
		}
		switch {
		case c.Edit.Op != "":
			ew := l.Editable(x-Board.PadX, y, lw+2*Board.PadX, clip, c.Label, Board.Px, Board.Ink, c.Edit)
			l.Marks[len(l.Marks)-1].Bold = true
			if c.Act.Op != "" {
				l.Spot(Rect{X: x - Board.PadX, Y: y, W: ew, H: h}, clip, Sheet.Cell{Act: c.Act, Tip: "go here"}, Sheet.PartBody)
			}
			rw := chipW(Board.Chip{Label: "✎"})
			rx, ry := f.at(rw)
			l.Chip(rx, ry, clip, Board.Chip{Label: "✎", Act: c.Rename, Ghost: true, Tip: "rename this sequence"})
		case c.Act.Op == "":
			l.Text(x, y, h, clip, c.Label, Board.Px, Board.Ink, true, Board.Left)
		default:
			tip := "go here"
			if i == 0 && d.SeqMenu != nil {
				tip = "switch sequence"
			}
			cw := l.Chip(x, y, clip, Board.Chip{Label: c.Label, Act: c.Act, Tip: tip, Lit: i == 0 && d.SeqOpen})
			if i == 0 && d.SeqMenu != nil && d.SeqOpen {
				*menus = append(*menus, placed{at: Rect{X: x, Y: y, W: max(cw, menuMinW), H: h}, m: *d.SeqMenu})
			}
		}
	}
	if d.Where != "" {
		x, y := f.at(sepW + Board.W(d.Where, Board.Px))
		x += l.Text(x, y, h, clip, "›", Board.ChipPx, Board.Mute, false, Board.Left) + Board.Gap
		l.Text(x, y, h, clip, d.Where, Board.Px, Board.Mute, false, Board.Left)
	}
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
		insp := Board.Chip{Label: "◂ inspect", Act: d.InspTab, Ghost: true, Tip: "open the inspector for the selected link"}
		if d.InspOpen {
			insp = Board.Chip{Label: "▸ inspect", Act: d.InspTab, Lit: true, Tip: "hide the inspector · the link stays selected"}
		}
		doors = append(doors[:len(doors):len(doors)], insp)
	}
	if narrow && d.More.Op != "" {
		at := dots(l, r.X+r.W-pad, r.Y+pad, h, clip, d.MoreOpen, d.More, "network, frame, s and the inspector")
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
			if !c.Lit {
				c.Ghost = true
			}
			x += l.Chip(x, y, clip, c) + Board.Gap
		}
	}
	return closeBand(l, r, clip, bg, f)
}

func stepBar(l *Board.Laid, r Rect, steps []Board.Chip) float32 {
	clip, bg := band(l, r, Board.Card)
	h := Board.RowH()
	f := flow{x0: r.X + pad, right: r.X + r.W - pad, x: r.X + pad, y: r.Y + pad, h: h, gap: stepGap}
	x, y := f.at(Board.W("steps", tinyPx) + 4)
	l.Text(x, y, h, clip, "steps", tinyPx, Board.Mute, false, Board.Left)
	for _, c := range steps {
		x, y := f.at(chipW(c))
		l.Chip(x, y, clip, c)
	}
	hint := "click a step to show it · hover to ghost it"
	hw := Board.W(hint, tinyPx)
	if f.x+hw <= f.right {
		l.Text(f.right, f.y, h, clip, hint, tinyPx, Board.Mute, false, Board.Right)
	} else {
		x, y := f.at(hw)
		l.Text(x, y, h, clip, hint, tinyPx, Board.Mute, false, Board.Left)
	}
	return closeBand(l, r, clip, bg, f)
}
