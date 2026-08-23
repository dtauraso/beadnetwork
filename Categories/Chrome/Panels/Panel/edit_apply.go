package Panel

func EditPanels(e Edit, pn *PanelState, persist func(PanelState), redraw func()) {
	ToggleFlag(pn, e.Flag)
	persist(*pn)
	redraw()
}

func ToggleFlag(pn *PanelState, flag string) {
	if fn, ok := PanelToggles[flag]; ok {
		fn(pn)
	}
}

func ApplyUpdate(attr byte, payload []byte, pn *PanelState, persist func(PanelState), redraw func()) {
	e, ok := DecodeUpdate(payload, attr)
	if !ok {
		return
	}
	EditPanels(e, pn, persist, redraw)
}
