package scenerun

import (
	"context"

	"github.com/dtauraso/wirefold/Categories/Node/TiltVectors"

	Speed "github.com/dtauraso/wirefold/Categories/Speed"

	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/Categories/Overlay"
	"github.com/dtauraso/wirefold/Categories/Scene"
	"github.com/dtauraso/wirefold/Categories/Scene/scenepersist"
)

func applyUpdateOverlays(_ context.Context, attr byte, payload []byte, md *MoveDispatch, _ SliderPanel.Sinks) {
	e, ok := Overlay.DecodeUpdate(payload, attr)
	if !ok {
		return
	}
	Overlay.EditOverlays(e, &md.UI.OV, &md.ChannelVectorsOn, &md.UI, md.persistOverlays)
}

func (md *MoveDispatch) persistOverlays(ov Overlay.OverlayState) {
	md.Persist.Overlays().Schedule(ov)
}

func applyUpdatePanels(_ context.Context, attr byte, payload []byte, md *MoveDispatch, _ SliderPanel.Sinks) {
	e, ok := Panel.DecodeUpdate(payload, attr)
	if !ok {
		return
	}
	Panel.EditPanels(e, &md.UI.PN, md.persistPanels, md.redraw)
}

func (md *MoveDispatch) persistPanels(pn Panel.PanelState) {
	md.Persist.Panels().Schedule(pn)
}

func applyUpdateClock(_ context.Context, attr byte, payload []byte, md *MoveDispatch, speedSinks SliderPanel.Sinks) {
	e, ok := Speed.DecodeUpdate(payload, attr)
	if !ok {
		return
	}
	Speed.EditSpeed(e, md.speedState(), speedSinks, md.persistSpeed, md.redraw)
}

func (md *MoveDispatch) speedState() Speed.SpeedState {
	return Speed.SpeedState{Speed: &md.UI.Speed, Divisor: md.UI.ClockDivisor}
}

func (md *MoveDispatch) persistSpeed(userSpeed float64) {
	md.Persist.Speed().Schedule(userSpeed)
}

func (md *MoveDispatch) redraw() { md.UI.EmitViewFrame(nil) }

func applyUpdateTiltVector(ctx context.Context, attr byte, payload []byte, md *MoveDispatch, speedSinks SliderPanel.Sinks) {
	e, ok := Scene.DecodeUpdateTiltVector(payload, attr)
	if !ok {
		return
	}
	TiltVectors.Edit(ctx, e.Attr, e.Num, e.Flag == "up", &md.MR, &md.Inboxes, md.resumeSpeed(speedSinks))
}

func (md *MoveDispatch) resumeSpeed(speedSinks SliderPanel.Sinks) func() {
	return func() {
		speedSinks.SendSpeed(scenepersist.SliderNum(md.UI.Speed), int64(md.UI.ClockDivisor))
	}
}

func tiltVectorEditFor(ctx context.Context, md *MoveDispatch, speedSinks SliderPanel.Sinks, row int32, attr string) {
	TiltVectors.Apply(ctx, row, attr, &md.MR, &md.Inboxes, md.resumeSpeed(speedSinks))
}

func adjustTiltPhiFor(ctx context.Context, md *MoveDispatch, row int32, up bool) {
	TiltVectors.AdjustPhi(ctx, row, up, &md.MR, &md.Inboxes)
}
