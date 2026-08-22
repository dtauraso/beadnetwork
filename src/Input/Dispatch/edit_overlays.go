package Dispatch

import (
	"context"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/src/Input/Codec"
	"github.com/dtauraso/wirefold/src/Node/nodeinbox"
	"github.com/dtauraso/wirefold/src/Overlay"
	"github.com/dtauraso/wirefold/src/Scene/scenepersist"
	"github.com/dtauraso/wirefold/src/Scene/viewstate"
)

func applyUpdateOverlays(_ context.Context, msg Codec.StdinMsg, md *MoveDispatch, _ SliderPanel.Sinks) {
	editOverlays(msg, &md.UI, &md.Inboxes, md.Persist.Overlays())
}

func editOverlays(
	msg Codec.StdinMsg,
	ui *viewstate.UIState,
	inboxes *nodeinbox.NodeInboxes,
	persist *scenepersist.Persister[Overlay.OverlayState],
) {
	if msg.Attr != "toggle" {
		return
	}
	toggleOverlayFlag(ui, inboxes, msg.Flag)
	ui.EmitViewFrame(nil)

	persist.Schedule(ui.OV)
}
