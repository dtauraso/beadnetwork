package Dispatch

import (
	"context"

	SceneTiltVectors "github.com/dtauraso/wirefold/src/Scene/TiltVectors"

	clock "github.com/dtauraso/wirefold/src/Clock"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/src/Input/Stdin"
	"github.com/dtauraso/wirefold/src/Overlay"
	"github.com/dtauraso/wirefold/src/Scene/scenepersist"
)

func applyUpdateOverlays(_ context.Context, msg Stdin.StdinMsg, md *MoveDispatch, _ SliderPanel.Sinks) {
	Overlay.EditOverlays(msg, &md.UI.OV, &md.Inboxes, &md.UI, md.persistOverlays)
}

func (md *MoveDispatch) persistOverlays(ov Overlay.OverlayState) {
	md.Persist.Overlays().Schedule(ov)
}

func applyUpdatePanels(_ context.Context, msg Stdin.StdinMsg, md *MoveDispatch, _ SliderPanel.Sinks) {
	Panel.EditPanels(msg, &md.UI.PN, md.persistPanels, md.redraw)
}

func (md *MoveDispatch) persistPanels(pn Panel.PanelState) {
	md.Persist.Panels().Schedule(pn)
}

func applyUpdateClock(_ context.Context, msg Stdin.StdinMsg, md *MoveDispatch, speedSinks SliderPanel.Sinks) {
	clock.EditSpeed(msg, md.speedState(), speedSinks, md.persistSpeed, md.redraw)
}

func (md *MoveDispatch) speedState() clock.SpeedState {
	return clock.SpeedState{Speed: &md.UI.Speed, Divisor: md.UI.ClockDivisor}
}

func (md *MoveDispatch) persistSpeed(userSpeed float64) {
	md.Persist.Speed().Schedule(userSpeed)
}

func (md *MoveDispatch) redraw() { md.UI.EmitViewFrame(nil) }

func applyUpdateTiltVector(ctx context.Context, msg Stdin.StdinMsg, md *MoveDispatch, speedSinks SliderPanel.Sinks) {
	SceneTiltVectors.Edit(ctx, msg, &md.MR, &md.Inboxes, md.resumeSpeed(speedSinks))
}

func (md *MoveDispatch) resumeSpeed(speedSinks SliderPanel.Sinks) func() {
	return func() {
		speedSinks.SendSpeed(scenepersist.SliderNum(md.UI.Speed), int64(md.UI.ClockDivisor))
	}
}

func tiltVectorEditFor(ctx context.Context, md *MoveDispatch, speedSinks SliderPanel.Sinks, row int32, attr string) {
	SceneTiltVectors.Apply(ctx, row, attr, &md.MR, &md.Inboxes, md.resumeSpeed(speedSinks))
}

func adjustTiltPhiFor(ctx context.Context, md *MoveDispatch, row int32, up bool) {
	SceneTiltVectors.AdjustPhi(ctx, row, up, &md.MR, &md.Inboxes)
}
