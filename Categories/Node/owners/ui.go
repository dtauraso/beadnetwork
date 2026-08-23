package owners

type UI struct {
	selected, hovered, latchedSel uint8
	hoverPort                     string
	hoverIsInput                  bool
}

func (u *UI) SetSelected(on bool) {
	if on {
		u.selected = 1
	} else {
		u.selected = 0
	}
}

func (u *UI) SetHover(on bool, port string, isInput bool) {
	if on {
		u.hovered = 1
		u.hoverPort = port
		u.hoverIsInput = isInput
	} else {
		u.hovered = 0
		u.hoverPort = ""
		u.hoverIsInput = false
	}
}

func (u *UI) Flags() (selected, hovered, latchedSel uint8) {
	return u.selected, u.hovered, u.latchedSel
}

func (u *UI) SetLatched(on bool) {
	if on {
		u.latchedSel = 1
	} else {
		u.latchedSel = 0
	}
}
