package Board

import "github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"

const (
	Semi    = 6
	SegH    = 20
	segB    = 17
	segMinV = 28
	BtnPx   = 12
)

func (l *Laid) Weighted(x, y, h float32, clip uint8, s string, px float32, ink Paint, weight uint8, a Align) float32 {
	w := l.text(x, y, h, clip, s, px, ink, false, weight, a)
	l.Marks[len(l.Marks)-1].Weight = weight
	return w
}

func (l *Laid) Seg(x, y float32, clip uint8, s Stepper) float32 {
	v := s.Plain
	if s.Act.Op != "" {
		v = s.Value
	}
	vw := max(segMinV, W(v, BtnPx)+10)
	w := 2*segB + vw
	r := Rect{X: x, Y: y, W: w, H: SegH}
	l.Box(r, clip, Card, Edge)
	l.Add(Mark{Kind: KBox, R: Rect{X: x + segB, Y: y + 1, W: vw, H: SegH - 2}, Clip: clip, Fill: White, Alpha: 1})
	l.Line(Rect{X: x + segB + 0.5, Y: y, W: 0, H: SegH}, clip, Edge, 1, false)
	l.Line(Rect{X: x + segB + vw - 0.5, Y: y, W: 0, H: SegH}, clip, Edge, 1, false)
	l.segEnd(Rect{X: x, Y: y, W: segB, H: SegH}, clip, "−", s.Steps, s.DownDim, Sheet.PartDown)
	vr := Rect{X: x + segB, Y: y, W: vw, H: SegH}
	ink := Ink
	if s.Faded {
		ink = Mute
	}
	if s.Act.Op != "" {
		l.Add(Mark{Kind: KBox, R: Rect{X: vr.X + 1, Y: vr.Y + 1, W: vr.W - 2, H: vr.H - 2}, Clip: clip, Radius: 0, Stroke: 1, Alpha: 1, Edit: s.Act})
		tw := W(v, BtnPx)
		l.Add(Mark{Kind: KText, R: Rect{X: vr.X + (vw-tw)/2, Y: y, W: tw, H: SegH}, Clip: clip, Ink: ink, Text: v, Px: BtnPx, Align: Right, Alpha: 1, Edit: s.Act})
		l.Spot(vr, clip, Sheet.Cell{Kind: Sheet.CellField, Value: v, Act: s.Act, Typed: true}, Sheet.PartBody)
	} else {
		l.Text(vr.X+vw/2, y, SegH, clip, v, BtnPx, ink, false, Center)
	}
	l.segEnd(Rect{X: x + segB + vw, Y: y, W: segB, H: SegH}, clip, "＋", s.Steps, s.UpDim, Sheet.PartUp)
	return w
}

func (l *Laid) segEnd(r Rect, clip uint8, glyph string, a Sheet.Act, dim bool, p Sheet.Part) {
	ink := Mute
	if dim {
		ink = DimInk
	}
	l.Text(r.X+r.W/2, r.Y, r.H, clip, glyph, 12, ink, false, Center)
	l.Spot(r, clip, Sheet.Cell{Act: a, Dim: dim}, p)
}

func (l *Laid) Btn(x, y, h float32, clip uint8, label string, a Sheet.Act, dim bool) float32 {
	w := W(label, BtnPx) + 14
	l.Box(Rect{X: x, Y: y, W: w, H: h}, clip, White, Edge)
	ink := Ink
	if dim {
		ink = Faint
	}
	l.Text(x+w/2, y, h, clip, label, BtnPx, ink, false, Center)
	if !dim {
		l.Spot(Rect{X: x, Y: y, W: w, H: h}, clip, Sheet.Cell{Act: a}, Sheet.PartBody)
	}
	return w
}
