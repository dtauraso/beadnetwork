package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
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
	scopeBar(l, bar, narrow, d, &menus)
	if len(d.Steps) > 0 && d.Empty == nil {
		sb := Rect{X: mid.X, Y: mid.Y, W: mid.W, H: h + stripPad}
		stepBar(l, sb, narrow, d.Steps)
		mid.Y, mid.H = mid.Y+sb.H, mid.H-sb.H
	}
	if d.Empty != nil {
		empty(l, mid, *d.Empty)
	} else {
		stack := mid
		if d.Insp.Title != "" && d.InspOpen {
			iw := min(inspW, max(inspMinW, mid.W*0.45))
			stack.W -= iw
			l.MostY = inspector(l, Rect{X: mid.X + stack.W, Y: mid.Y, W: iw, H: mid.H}, sy, d.Insp)
		}
		l.MostY = max(l.MostY, columns(l, stack, sy, stack.W < colNarrowW, d, &menus))
	}
	strip(l, st, d)
	if d.Toast != "" {
		refusal(l, box, st, d)
	}
	if d.Pop != nil {
		popover(l, box, bar, *d.Pop)
	}
	for _, p := range menus {
		drawMenu(l, box, p)
	}
	return l
}

func stepBar(l *Board.Laid, r Rect, narrow bool, steps []Board.Chip) {
	clip := l.Clip(r)
	l.Box(r, clip, Board.Card, Board.None)
	l.Line(Rect{X: r.X, Y: r.Y + r.H - 0.5, W: r.W, H: 0}, clip, Board.Guide, 1, false)
	x, y := r.X+pad, r.Y+stripPad/2
	for _, c := range steps {
		if narrow && !c.Lit && c.Label != "＋" {
			continue
		}
		c.Px = Board.BtnPx
		w := Board.W(c.Label, Board.BtnPx) + 2*Board.PadX
		if x+w > r.X+r.W-pad {
			break
		}
		x += l.Chip(x, y, clip, c) + 4
	}
}

func scopeBar(l *Board.Laid, r Rect, narrow bool, d Data, menus *[]placed) {
	clip := l.Clip(r)
	h := Board.RowH()
	l.Line(Rect{X: r.X, Y: r.Y + r.H - 0.5, W: r.W, H: 0}, clip, Board.Guide, 1, false)
	x, y := r.X+pad, r.Y+pad
	for i, c := range d.Crumbs {
		if narrow && c.Short != "" {
			c.Label = c.Short
		}
		if i > 0 && !(narrow && i == 1) {
			x += l.Text(x, y, h, clip, "›", Board.Px, Board.Mute, false, Board.Left) + Board.Gap
		}
		if c.Tag != "" {
			x += l.Text(x, y, h, clip, c.Tag, Board.Px, Board.Mute, false, Board.Left) + Board.Gap
		}
		if c.Edit.Op != "" {
			w := l.Editable(x-Board.PadX, y, 0, clip, c.Label, Board.Px, Board.Ink, c.Edit)
			if c.Act.Op != "" {
				l.Spot(Rect{X: x - Board.PadX, Y: y, W: w, H: h}, clip, Sheet.Cell{Act: c.Act, Tip: "go here"}, Sheet.PartBody)
			}
			x += w - Board.PadX + Board.Gap
			x += l.Chip(x, y, clip, Board.Chip{Label: "✎", Act: c.Rename, Px: Board.BtnPx, Ghost: true, Tip: "rename this sequence"}) + Board.Gap
			continue
		}
		if c.Act.Op == "" {
			x += l.Text(x, y, h, clip, c.Label, Board.Px, Board.Ink, true, Board.Left) + Board.Gap
			continue
		}
		tip := "go here"
		if i == 0 && d.SeqMenu != nil {
			tip = "switch sequence"
		}
		w := l.Chip(x, y, clip, Board.Chip{Label: c.Label, Act: c.Act, Px: Board.BtnPx, Tip: tip, Lit: i == 0 && d.SeqOpen})
		if i == 0 && d.SeqMenu != nil && d.SeqOpen {
			*menus = append(*menus, placed{at: Rect{X: x, Y: y, W: max(w, menuMinW), H: h}, m: *d.SeqMenu})
		}
		x += w + Board.Gap
	}
	if d.S.Act.Op != "" && !narrow {
		x += l.Text(x, y, h, clip, "ticks per 1/12 turn", Board.SmallPx, Board.Mute, false, Board.Left) + 4
		x += l.Field(x, y, clip, d.S) + Board.Gap
	}
	if d.SNote != "" && !narrow {
		nw := Board.W(d.SNote, Board.SmallPx) + 12
		l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: x, Y: y + 3, W: nw, H: h - 6}, Clip: clip, Fill: Board.WarnFill, Radius: 99, Alpha: 1})
		x += l.Text(x+6, y, h, clip, d.SNote, Board.SmallPx, Board.Warn, false, Board.Left) + 12 + Board.Gap
	}
	right := r.X + r.W - pad
	doors := d.Doors
	if d.Insp.Title != "" {
		insp := Board.Chip{Label: "◂ inspect", Act: d.InspTab, Tip: "open the inspector for the selected link"}
		if d.InspOpen {
			insp = Board.Chip{Label: "▸ inspect", Act: d.InspTab, Lit: true, Tip: "hide the inspector · the link stays selected"}
		}
		doors = append(doors[:len(doors):len(doors)], insp)
	}
	if narrow && d.More.Op != "" {
		at := dots(l, right, y, h, clip, d.MoreOpen, d.More, "network, frame, s and the inspector")
		if d.MoreOpen {
			m := Menu{Items: doors, Note: d.SNote}
			if d.S.Act.Op != "" {
				m.Find, m.FindLabel = &d.S, "ticks per 1/12 turn"
			}
			*menus = append(*menus, placed{at: at, m: m})
		}
		return
	}
	for i := len(doors) - 1; i >= 0; i-- {
		c := doors[i]
		c.Ghost = true
		c.Px = Board.BtnPx
		w := Board.W(c.Label, Board.BtnPx) + 2*Board.PadX
		right -= w
		if right < x {
			break
		}
		l.Chip(right, y, clip, c)
		right -= Board.Gap
	}
}

func strip(l *Board.Laid, r Rect, d Data) {
	clip := l.Clip(r)
	l.Box(Rect{X: r.X, Y: r.Y, W: r.W, H: r.H + 4}, clip, Board.Card, Board.None)
	l.Line(Rect{X: r.X, Y: r.Y + 0.5, W: r.W, H: 0}, clip, Board.Guide, 1, false)
	h := Board.RowH()
	x, y := r.X+pad, r.Y+stripPad/2
	x += l.Text(x, y, h, clip, d.Status, 11.5, Board.Mute, false, Board.Left) + 12
	if d.Ghost != "" {
		l.Text(x, y, h, clip, d.Ghost, 11.5, Board.Accent, true, Board.Left)
		return
	}
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
		l.Btn(cx-w/2, y, Board.SegH, clip, e.Door.Label, e.Door.Act, e.Door.Tip, false)
		y += Board.SegH + 10
	}
	l.Text(cx, y, h, clip, e.Sub, 12, Board.Mute, false, Board.Center)
}
