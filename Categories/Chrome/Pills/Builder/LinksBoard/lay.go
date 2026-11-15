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
	scopeBar(l, bar, narrow, d)
	if d.Empty != nil {
		empty(l, mid, *d.Empty)
	} else {
		stack := mid
		if d.Insp.Title != "" && d.InspOpen {
			iw := min(inspW, max(inspMinW, mid.W*0.45))
			stack.W -= iw
			l.MostY = inspector(l, Rect{X: mid.X + stack.W, Y: mid.Y, W: iw, H: mid.H}, sy, d.Insp)
		}
		l.MostY = max(l.MostY, columns(l, stack, sy, stack.W < colNarrowW, d))
	}
	strip(l, st, d)
	if d.Toast != "" {
		toast(l, box, st, d)
	}
	if d.Pop != nil {
		popover(l, box, bar, *d.Pop)
	}
	return l
}

func scopeBar(l *Board.Laid, r Rect, narrow bool, d Data) {
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
		x += l.Chip(x, y, clip, Board.Chip{Label: c.Label, Act: c.Act, Px: Board.BtnPx, Tip: "go here"}) + Board.Gap
	}
	if d.S.Act.Op != "" && !narrow {
		sw := Board.W("s", Board.BtnPx) + 2 + Board.W(d.S.Fit, Board.Px) + 2*Board.PadX + 10
		l.Box(Rect{X: x, Y: y - 2, W: sw, H: h + 4}, clip, Board.White, Board.Edge)
		x += 6 + l.Text(x+6, y, h, clip, "s", Board.BtnPx, Board.Ink, false, Board.Left) + 3
		x += l.Field(x, y, clip, d.S) + 4 + Board.Gap
	}
	if d.SNote != "" {
		x += l.Text(x, y, h, clip, d.SNote, Board.SmallPx, Board.Mute, false, Board.Left) + Board.Gap
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
	for i := len(doors) - 1; i >= 0; i-- {
		c := doors[i]
		c.Ghost = true
		if narrow && i < len(d.DoorsShort) {
			c.Label = d.DoorsShort[i]
		}
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

func toast(l *Board.Laid, box, strip Rect, d Data) {
	const bh = 20
	h := float32(bh + 10)
	w := Board.W(d.Toast, Board.Px) + 18 + 10 + Board.W("×", Board.Px)
	if d.Undo.Op != "" {
		w += Board.W("Undo", Board.BtnPx) + 14 + 10
	}
	r := Rect{X: box.X + pad, Y: strip.Y - h - pad, W: min(w, box.W-2*pad), H: h}
	clip := l.Clip(Rect{X: r.X - 4, Y: r.Y - 4, W: r.W + 8, H: r.H + 8})
	l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: r.X, Y: r.Y + 2, W: r.W, H: r.H}, Clip: clip, Fill: Board.Edge, Radius: 6, Alpha: 0.5})
	l.Add(Board.Mark{Kind: Board.KBox, R: r, Clip: clip, Fill: Board.White, Edge: Board.Edge, Radius: 6, Stroke: 1, Alpha: 1})
	x, y := r.X+9, r.Y+5
	x += l.Text(x, y, bh, clip, d.Toast, Board.Px, Board.Ink, false, Board.Left) + 10
	if d.Undo.Op != "" {
		x += l.Btn(x, y, bh, clip, "Undo", d.Undo, "put the pick back", false) + 10
	}
	xw := l.Text(x, y, bh, clip, "×", Board.Px, Board.Mute, false, Board.Left)
	l.Spot(Rect{X: x - 4, Y: y, W: xw + 8, H: bh}, clip, Sheet.Cell{Act: d.Dismiss, Tip: "dismiss"}, Sheet.PartBody)
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
