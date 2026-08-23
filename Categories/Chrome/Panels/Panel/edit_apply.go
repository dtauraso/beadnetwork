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
