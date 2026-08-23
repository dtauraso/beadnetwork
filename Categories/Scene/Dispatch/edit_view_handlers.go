package Dispatch

import (
	"context"

	"github.com/dtauraso/wirefold/Categories/Node"

	Speed "github.com/dtauraso/wirefold/Categories/Speed"

	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/Categories/Overlay"
	"github.com/dtauraso/wirefold/Categories/Scene"
)

func (md *MoveDispatch) persistOverlays(ov Overlay.OverlayState) {
	md.Persist.Overlays().Schedule(ov)
}

func (md *MoveDispatch) persistPanels(pn Panel.PanelState) {
	md.Persist.Panels().Schedule(pn)
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
	Node.TiltEdit(ctx, e.Attr, e.Num, e.Flag == "up", &md.MR, &md.Inboxes, md.resumeSpeed(speedSinks))
}

func (md *MoveDispatch) resumeSpeed(speedSinks SliderPanel.Sinks) func() {
	return func() {
		speedSinks.SendSpeed(Speed.SliderNum(md.UI.Speed), int64(md.UI.ClockDivisor))
	}
}

func tiltVectorEditFor(ctx context.Context, md *MoveDispatch, speedSinks SliderPanel.Sinks, row int32, attr string) {
	Node.ApplyTiltEdit(ctx, row, attr, &md.MR, &md.Inboxes, md.resumeSpeed(speedSinks))
}

func adjustTiltPhiFor(ctx context.Context, md *MoveDispatch, row int32, up bool) {
	Node.AdjustTiltPhi(ctx, row, up, &md.MR, &md.Inboxes)
}
