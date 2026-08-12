package nodeactor

func (u *nodeUI) SetSelected(on bool) {
	if on {
		u.selected = 1
	} else {
		u.selected = 0
	}
}

func (u *nodeUI) SetHover(on bool, port string, isInput bool) {
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

func (u *nodeUI) SetLatched(on bool) {
	if on {
		u.latchedSel = 1
	} else {
		u.latchedSel = 0
	}
}
