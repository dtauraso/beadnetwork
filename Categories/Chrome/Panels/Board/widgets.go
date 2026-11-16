package Board

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
)

const (
	Px      = 13
	SmallPx = 11
	PadX    = 6
	Gap     = 6
	Radius  = 4
	ChipPx  = 12
	chipPad = 7
)

func RowH() float32 { return 20 }

func W(s string, px float32) float32 { return TextW(s, px, 4) }

func (l *Laid) Box(r Rect, clip uint8, fill, edge Paint) {
	l.Add(Mark{Kind: KBox, R: r, Clip: clip, Fill: fill, Edge: edge, Radius: Radius, Stroke: 1, Alpha: 1})
}

func (l *Laid) Text(x, y, h float32, clip uint8, s string, px float32, ink Paint, bold bool, a Align) float32 {
	wt := uint8(4)
	if bold {
		wt = 7
	}
	return l.text(x, y, h, clip, s, px, ink, bold, wt, a)
}

func (l *Laid) text(x, y, h float32, clip uint8, s string, px float32, ink Paint, bold bool, wt uint8, a Align) float32 {
	w := TextW(s, px, wt)
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
	Dim   bool
	Pri   bool
	Bad   bool
	Tip   string
	Px    float32
}

func (l *Laid) Chip(x, y float32, clip uint8, c Chip) float32 {
	px := c.Px
	if px == 0 {
		px = ChipPx
	}
	w := W(c.Label, px) + 2*chipPad
	r := Rect{X: x, Y: y, W: w, H: RowH()}
	fill, edge, ink := White, Edge, Ink
	switch {
	case c.Pri:
		fill, edge, ink = Accent, Accent, White
	case c.Lit:
		fill, edge, ink = On, Accent, Accent
	}
	if c.Bad {
		ink = Bad
	}
	if c.Dim {
		ink = DimInk
	}
	l.Box(r, clip, fill, edge)
	l.Text(x+w/2, y, r.H, clip, c.Label, px, ink, false, Center)
	if !c.Dim {
		l.Spot(r, clip, Sheet.Cell{Act: c.Act, Tip: c.Tip}, Sheet.PartBody)
	}
	return w
}

type Field struct {
	Label string
	Value string
	Fit   string
	Act   Sheet.Act
	Words bool
	Tip   string
}

func (f Field) weight() uint8 {
	if f.Label != "" {
		return 7
	}
	return 4
}

func FieldW(f Field) float32 {
	fit := f.Fit
	if len(fit) < len(f.Value) {
		fit = f.Value
	}
	lw := float32(0)
	if f.Label != "" {
		lw = W(f.Label+" ", ChipPx)
	}
	return lw + TextW(fit, ChipPx, f.weight()) + 2*PadX
}

func (l *Laid) Field(x, y float32, clip uint8, f Field) float32 {
	wt := f.weight()
	vw := TextW(f.Value, ChipPx, wt)
	w := FieldW(f)
	r := Rect{X: x, Y: y, W: w, H: RowH()}
	l.Add(Mark{Kind: KBox, R: r, Clip: clip, Fill: White, Edge: Edge, Radius: Radius, Stroke: 1, Alpha: 1, Edit: f.Act})
	if f.Label != "" {
		l.Text(x+PadX, y, r.H, clip, f.Label, ChipPx, Ink, false, Left)
	}
	l.Add(Mark{Kind: KText, R: Rect{X: x + w - PadX - vw, Y: y, W: vw, H: r.H}, Clip: clip, Ink: Ink, Text: f.Value, Px: ChipPx, Bold: wt == 7, Align: Right, Alpha: 1, Edit: f.Act})
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
	Faded   bool
	DownTip string
	UpTip   string
}

func (l *Laid) Stepper(x, y float32, clip uint8, s Stepper) float32 {
	return l.Seg(x, y, clip, s)
}
