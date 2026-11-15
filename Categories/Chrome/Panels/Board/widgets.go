package Board

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
)

const (
	Px      = 13
	SmallPx = 11
	PadX    = 6
	Gap     = 6
	Radius  = 3
)

func RowH() float32 { return Panel.LineHeight(Px) + 6 }

func W(s string, px float32) float32 { return Panel.TextWidth(s, px) }

func (l *Laid) Box(r Rect, clip uint8, fill, edge Paint) {
	l.Add(Mark{Kind: KBox, R: r, Clip: clip, Fill: fill, Edge: edge, Radius: Radius, Stroke: 1, Alpha: 1})
}

func (l *Laid) Text(x, y, h float32, clip uint8, s string, px float32, ink Paint, bold bool, a Align) float32 {
	w := W(s, px)
	r := Rect{X: x, Y: y, W: w, H: h}
	switch a {
	case Center:
		r.X = x - w/2
	case Right:
		r.X = x - w
	}
	l.Add(Mark{Kind: KText, R: r, Clip: clip, Ink: ink, Text: s, Px: px, Bold: bold, Align: Left, Alpha: 1})
	return w
}

func (l *Laid) Line(r Rect, clip uint8, ink Paint, stroke float32, dash bool) {
	l.Add(Mark{Kind: KLine, R: r, Clip: clip, Ink: ink, Stroke: stroke, Dash: dash, Alpha: 1})
}

func (l *Laid) Circle(r Rect, clip uint8, fill, edge Paint) {
	l.Add(Mark{Kind: KCircle, R: r, Clip: clip, Fill: fill, Edge: edge, Stroke: 1.5, Alpha: 1})
}

func (l *Laid) Dial(r Rect, clip uint8, d Dial) {
	l.Add(Mark{Kind: KDial, R: r, Clip: clip, Dial: d, Alpha: 1})
}

type Chip struct {
	Label string
	Act   Sheet.Act
	Lit   bool
	Ghost bool
	Dim   bool
	Tip   string
	Px    float32
}

func (l *Laid) Chip(x, y float32, clip uint8, c Chip) float32 {
	px := c.Px
	if px == 0 {
		px = Px
	}
	w := W(c.Label, px) + 2*PadX
	r := Rect{X: x, Y: y, W: w, H: RowH()}
	fill, edge, ink := White, Edge, Ink
	if c.Ghost {
		fill = None
	}
	if c.Lit {
		fill, edge = On, Accent
	}
	if c.Dim {
		ink = Faint
	}
	l.Box(r, clip, fill, edge)
	l.Text(x+w/2, y, r.H, clip, c.Label, px, ink, false, Center)
	if !c.Dim {
		l.Spot(r, clip, Sheet.Cell{Act: c.Act, Tip: c.Tip}, Sheet.PartBody)
	}
	return w
}

type Field struct {
	Value string
	Fit   string
	Act   Sheet.Act
	Words bool
	Tip   string
}

func (l *Laid) Field(x, y float32, clip uint8, f Field) float32 {
	fit := f.Fit
	if len(fit) < len(f.Value) {
		fit = f.Value
	}
	w := W(fit, Px) + 2*PadX
	r := Rect{X: x, Y: y, W: w, H: RowH()}
	l.Add(Mark{Kind: KBox, R: r, Clip: clip, Fill: White, Edge: Edge, Radius: Radius, Stroke: 1, Alpha: 1, Edit: f.Act})
	l.Add(Mark{Kind: KText, R: Rect{X: x + w - PadX - W(f.Value, Px), Y: y, W: W(f.Value, Px), H: r.H}, Clip: clip, Ink: Ink, Text: f.Value, Px: Px, Align: Right, Alpha: 1, Edit: f.Act})
	l.Spot(r, clip, Sheet.Cell{Kind: Sheet.CellField, Value: f.Value, Act: f.Act, Typed: true, Words: f.Words, Tip: f.Tip}, Sheet.PartBody)
	return w
}

func (l *Laid) Editable(x, y, minW float32, clip uint8, s string, px float32, ink Paint, edit Sheet.Act) float32 {
	h := RowH()
	w := max(minW, W(s, px)+2*PadX)
	l.Add(Mark{Kind: KBox, R: Rect{X: x, Y: y + 2, W: w, H: h - 4}, Clip: clip, Fill: None, Edge: None, Radius: Radius, Stroke: 1, Alpha: 1, Edit: edit})
	l.Add(Mark{Kind: KText, R: Rect{X: x + PadX, Y: y, W: W(s, px), H: h}, Clip: clip, Ink: ink, Text: s, Px: px, Align: Left, Alpha: 1, Edit: edit})
	return w
}

type Stepper struct {
	Field
	Steps   Sheet.Act
	Plain   string
	DownDim bool
	UpDim   bool
	DownTip string
	UpTip   string
}

func (l *Laid) Stepper(x, y float32, clip uint8, s Stepper) float32 {
	h := RowH()
	bw := W("＋", BtnPx) + 14
	down := Rect{X: x, Y: y, W: bw, H: h}
	l.arrow(down, clip, "−", s.Steps, s.DownDim, s.DownTip, Sheet.PartDown)
	x += bw + Gap
	if s.Act.Op != "" {
		x += l.Field(x, y, clip, s.Field) + Gap
	} else {
		w := max(26, W(s.Plain, Px)+2*PadX)
		l.Box(Rect{X: x, Y: y, W: w, H: h}, clip, White, Edge)
		l.Text(x+w-PadX, y, h, clip, s.Plain, Px, Ink, false, Right)
		x += w + Gap
	}
	l.arrow(Rect{X: x, Y: y, W: bw, H: h}, clip, "＋", s.Steps, s.UpDim, s.UpTip, Sheet.PartUp)
	return x + bw - down.X
}

func (l *Laid) arrow(r Rect, clip uint8, glyph string, a Sheet.Act, dim bool, tip string, p Sheet.Part) {
	ink, alpha := Ink, float32(1)
	if dim {
		alpha = 0.4
	}
	l.Box(r, clip, White, Edge)
	l.Text(r.X+r.W/2, r.Y, r.H, clip, glyph, BtnPx, ink, false, Center)
	l.Marks[len(l.Marks)-1].Alpha, l.Marks[len(l.Marks)-2].Alpha = alpha, alpha
	l.Spot(r, clip, Sheet.Cell{Act: a, Dim: dim, Tip: tip}, p)
}

type Ruler struct {
	Start, Delta, Most float32
	Fixed              bool
	Big                bool
}

func (l *Laid) bar(r Rect, clip uint8, fill Paint, alpha float32) {
	if r.W <= 0 {
		return
	}
	l.Add(Mark{Kind: KBox, R: r, Clip: clip, Fill: fill, Radius: 2, Alpha: alpha})
}

func (l *Laid) Ruler(r Rect, clip uint8, u Ruler) {
	most := max(u.Most, u.Start+max(u.Delta, 0), 1)
	at := func(v float32) float32 { return r.X + r.W*max(0, v)/most }
	end := u.Start + u.Delta
	lo, hi := at(min(u.Start, end)), at(max(u.Start, end))
	knob := Accent
	if u.Fixed {
		knob = Faint
	}
	if u.Big {
		mid := r.Y + r.H/2
		l.bar(Rect{X: r.X, Y: mid - 2, W: r.W, H: 4}, clip, Guide, 1)
		l.bar(Rect{X: r.X, Y: mid - 2, W: at(u.Start) - r.X, H: 4}, clip, StartArc, 1)
		l.bar(Rect{X: lo, Y: mid - 3, W: hi - lo, H: 6}, clip, Accent, 0.9)
		l.Add(Mark{Kind: KBox, R: Rect{X: at(end) - 6, Y: mid - 10, W: 12, H: 20}, Clip: clip, Fill: knob, Radius: 3, Alpha: 1})
		return
	}
	top := r.Y + (r.H-14)/2
	l.bar(Rect{X: r.X, Y: top + 5, W: r.W, H: 3}, clip, Guide, 1)
	l.bar(Rect{X: r.X, Y: top + 5, W: at(u.Start) - r.X, H: 3}, clip, StartArc, 1)
	l.bar(Rect{X: lo, Y: top + 5, W: hi - lo, H: 3}, clip, Accent, 1)
	cy := top + 6.5
	l.Circle(Rect{X: at(end) - 6, Y: cy - 6, W: 12, H: 12}, clip, knob, None)
	l.Add(Mark{Kind: KCircle, R: Rect{X: at(end) - 4, Y: cy - 4, W: 8, H: 8}, Clip: clip, Fill: knob, Edge: White, Stroke: 2, Alpha: 1})
}
