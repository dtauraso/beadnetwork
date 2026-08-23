package Dispatch

import (
	"context"

	Panel "github.com/dtauraso/wirefold/Categories/Chrome/Panels/Panel"
	SliderPanel "github.com/dtauraso/wirefold/Categories/Chrome/Panels/SliderPanel"
	NodeKind "github.com/dtauraso/wirefold/Categories/Node"
	edge "github.com/dtauraso/wirefold/Categories/Node/Edge"
	"github.com/dtauraso/wirefold/Categories/Overlay"
	Speed "github.com/dtauraso/wirefold/Categories/Speed"
)

// EDIT_UPDATE_KINDS_START
func (md *MoveDispatch) updateOwners(speedSinks SliderPanel.Sinks) map[string]func(context.Context, byte, []byte) {
	return map[string]func(context.Context, byte, []byte){
		"node": func(ctx context.Context, attr byte, payload []byte) {
			NodeKind.ApplyUpdate(ctx, attr, payload, &md.Rules)
		},
		"edge": func(ctx context.Context, attr byte, payload []byte) {
			edge.ApplyUpdate(ctx, attr, payload, md.Rules.TogglesByEdgeRow)
		},
		"overlays": func(_ context.Context, attr byte, payload []byte) {
			Overlay.ApplyUpdate(attr, payload, &md.UI.OV, &md.ChannelVectorsOn, &md.UI, md.persistOverlays)
		},
		"panels": func(_ context.Context, attr byte, payload []byte) {
			Panel.ApplyUpdate(attr, payload, &md.UI.PN, md.persistPanels, md.redraw)
		},
		"clock": func(_ context.Context, attr byte, payload []byte) {
			Speed.ApplyUpdate(attr, payload, md.speedState(), speedSinks, md.persistSpeed, md.redraw)
		},
		"tiltVector": func(ctx context.Context, attr byte, payload []byte) {
			applyUpdateTiltVector(ctx, attr, payload, md, speedSinks)
		},
		"scene": func(ctx context.Context, attr byte, payload []byte) {
			applyUpdateScene(ctx, attr, payload, md, speedSinks)
		},
	}
}

// EDIT_UPDATE_KINDS_END
