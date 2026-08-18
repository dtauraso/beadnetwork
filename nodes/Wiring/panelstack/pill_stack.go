package panelstack

const (
	PillTop   = 44
	PillRight = 12
	PillGap   = 6

	PillFontPx    = 11
	PillGlyphPx   = 9
	PillHeadingPx = 9.5

	PillPadX    = 9
	PillPadY    = 3
	CaretW      = 20
	PopoverPad  = 6
	PopoverGap  = 4
	RowPadX     = 6
	RowPadY     = 4
	RowGap      = 2
	HeadingPadY = 5
)

type PillStack struct {
	viewW float32
	width float32
	y     float32
}

func NewPillStack(viewW float32, labels []string) *PillStack {
	var w float32
	for _, l := range labels {
		if t := TextWidth(l, PillFontPx); t > w {
			w = t
		}
	}
	return &PillStack{viewW: viewW, width: w + 2*PillPadX + CaretW + 2, y: PillTop}
}

func (s *PillStack) Width() float32 { return s.width }

func (s *PillStack) X() float32 { return s.viewW - PillRight - s.width }

func (s *PillStack) AddPill() Rect {
	h := LineHeight(PillFontPx) + 2*PillPadY + 2
	r := Rect{X: s.X(), Y: s.y, W: s.width, H: h}
	s.y += h
	return r
}

func (s *PillStack) AddPopover(contentH float32) (box Rect, contentX, contentY float32) {
	s.y += PopoverGap
	box = Rect{X: s.X(), Y: s.y, W: s.width, H: contentH + 2*PopoverPad}
	s.y += box.H
	return box, box.X + PopoverPad, box.Y + PopoverPad
}

func (s *PillStack) EndGroup() { s.y += PillGap }

func RowH() float32 { return LineHeight(PillFontPx) + 2*RowPadY }

func StepperH() float32 {
	return LineHeight(PillFontPx)*2 + RowGap + 2*RowPadY
}

func HeadingH() float32 { return LineHeight(PillHeadingPx) + 2*HeadingPadY }
