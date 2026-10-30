package Panel

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
	viewH float32
	width float32
	y     float32
	left  bool
}

func NewPillStack(viewW, viewH float32, labels []string) *PillStack {
	return &PillStack{viewW: viewW, viewH: viewH, width: pillWidth(labels), y: PillTop}
}

func NewLeftPillStack(viewW, viewH, top float32, labels []string) *PillStack {
	return &PillStack{viewW: viewW, viewH: viewH, width: pillWidth(labels), y: top, left: true}
}

func pillWidth(labels []string) float32 {
	var w float32
	for _, l := range labels {
		if t := TextWidth(l, PillFontPx); t > w {
			w = t
		}
	}
	return w + 2*PillPadX + CaretW + 2
}

func (s *PillStack) Width() float32 { return s.width }

func (s *PillStack) X() float32 {
	if s.left {
		return OriginX
	}
	return s.viewW - PillRight - s.width
}

func (s *PillStack) AddChip(label string) Rect {
	w := TextWidth(label, PillFontPx) + 2*8 + 2
	h := LineHeight(PillFontPx) + 2*3 + 2
	r := Rect{X: s.viewW - PillRight - w, Y: s.y, W: w, H: h}
	s.y += h + PillGap
	return r
}

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

const PopoverMaxViewFraction = 0.6

func ClipToRoom(full, room, scroll float32) (visible, applied, maxScroll float32) {
	visible = full
	if room > 0 && room < visible {
		visible = room
	}
	maxScroll = full - visible
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scroll > maxScroll {
		scroll = maxScroll
	}
	if scroll < 0 {
		scroll = 0
	}
	return visible, scroll, maxScroll
}

func RoomBelow(viewH, y, pad float32) float32 {
	if viewH <= 0 {
		return 0
	}
	room := viewH - y - pad
	if cap := viewH * PopoverMaxViewFraction; room > cap {
		room = cap
	}
	return room
}

func (s *PillStack) AddScrollingPopover(contentH, scroll float32) (box Rect, contentX, contentY, maxScroll float32) {
	s.y += PopoverGap

	full := contentH + 2*PopoverPad
	visible, scroll, maxScroll := ClipToRoom(full, RoomBelow(s.viewH, s.y, PopoverPad), scroll)

	box = Rect{X: s.X(), Y: s.y, W: s.width, H: visible}
	s.y += box.H
	return box, box.X + PopoverPad, box.Y + PopoverPad - scroll, maxScroll
}

func (s *PillStack) AddWideScrollingPopover(contentW, contentH, scroll, scrollX float32) (box Rect, contentX, contentY, maxScroll, maxScrollX float32) {
	s.y += PopoverGap

	full := contentH + 2*PopoverPad
	visible, scroll, maxScroll := ClipToRoom(full, RoomBelow(s.viewH, s.y, PopoverPad), scroll)

	left := float32(OriginX)
	if s.left {
		left = s.X()
	}
	var roomW float32
	if s.viewW > 0 {
		roomW = s.viewW - PillRight - left
	}
	w, scrollX, maxScrollX := ClipToRoom(max(contentW+2*PopoverPad, s.width), roomW, scrollX)
	x := max(left, s.viewW-PillRight-w)
	if s.left {
		x = left
	}
	box = Rect{X: x, Y: s.y, W: w, H: visible}
	s.y += box.H
	return box, box.X + PopoverPad - scrollX, box.Y + PopoverPad - scroll, maxScroll, maxScrollX
}

func (s *PillStack) EndGroup() { s.y += PillGap }

func RowH() float32 { return LineHeight(PillFontPx) + 2*RowPadY }

func StepperH() float32 {
	return LineHeight(PillFontPx)*2 + RowGap + 2*RowPadY
}

func HeadingH() float32 { return LineHeight(PillHeadingPx) + 2*HeadingPadY }
