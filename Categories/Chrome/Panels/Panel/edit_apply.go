package Panel

import (
	"github.com/dtauraso/wirefold/Categories/Input/Stdin"
)

func EditPanels(msg Stdin.StdinMsg, pn *PanelState, persist func(PanelState), redraw func()) {
	if msg.Attr != "toggle" {
		return
	}
	ToggleFlag(pn, msg.Flag)
	persist(*pn)
	redraw()
}

func ToggleFlag(pn *PanelState, flag string) {
	if fn, ok := PanelToggles[flag]; ok {
		fn(pn)
	}
}
