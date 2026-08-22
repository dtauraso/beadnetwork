package Dispatch

import (
	"context"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/src/Input/Codec"
	"github.com/dtauraso/wirefold/src/Scene/scenepersist"
	"github.com/dtauraso/wirefold/src/Scene/viewstate"
)

func applyUpdatePanels(_ context.Context, msg Codec.StdinMsg, md *MoveDispatch, _ SliderPanel.Sinks) {
	editPanels(msg, &md.UI, md.Persist.Panels())
}

func editPanels(
	msg Codec.StdinMsg,
	ui *viewstate.UIState,
	persist *scenepersist.Persister[Panel.PanelState],
) {
	if msg.Attr != "toggle" {
		return
	}
	if fn, ok := Panel.PanelToggles[msg.Flag]; ok {
		fn(&ui.PN)
	}

	persist.Schedule(ui.PN)
	ui.EmitViewFrame(nil)
}
