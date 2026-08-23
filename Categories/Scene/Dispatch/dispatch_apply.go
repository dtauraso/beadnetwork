package Dispatch

import (
	"context"

	NodeKind "github.com/dtauraso/wirefold/Categories/Node"
	"github.com/dtauraso/wirefold/Categories/Overlay"
	Speed "github.com/dtauraso/wirefold/Categories/Speed"

	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/Categories/Scene"

	"github.com/dtauraso/wirefold/Categories/Scene/Drag"
	"github.com/dtauraso/wirefold/Categories/Scene/Gesture"
	"github.com/dtauraso/wirefold/Categories/Scene/Scenes"
)

func tiltVectorEdit(ctx context.Context, md *MoveDispatch, speedSinks SliderPanel.Sinks, row int32, attr string) {
	tiltVectorEditFor(ctx, md, speedSinks, row, attr)
}

func adjustTiltPhi(ctx context.Context, md *MoveDispatch, row int32, up bool) {
	adjustTiltPhiFor(ctx, md, row, up)
}

func setLatticePoints(md *MoveDispatch, points int32) {
	md.UI.SetLatticePoints(points, md.Persist.Lattice().Schedule, md.Inboxes.BroadcastLatticePoints)
}

func applyUpdateScene(ctx context.Context, attr byte, payload []byte, md *MoveDispatch, speedSinks SliderPanel.Sinks) {
	if md == nil {
		return
	}
	e, ok := Scene.DecodeUpdateScene(payload, attr)
	if !ok {
		return
	}
	switch e.Attr {
	case "selected":
		Scenes.SelectScene(&md.Scenes, int(e.Num))
	case "latticePoints":
		setLatticePoints(md, int32(e.Num))
	case "viewport":
		md.UI.SetViewport(e.X, e.Y)
	case "create":

		CreateNode(&md.Scenes, &md.UI, md.MR.NodeGeoms(), md.nearestNodeTo, uint8(e.Num), e.X, e.Y)
	case "delete":

		nodeID, _ := md.RT.LookupNodeRow(e.Num)
		DeleteNode(&md.Scenes, &md.UI, nodeID, e.Num)
	}
}

func sendRuleEdit(ctx context.Context, md *MoveDispatch, row int, edit NodeKind.RuleEdit) {
	NodeKind.SendRuleEdit(ctx, &md.Rules, row, edit)
}

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
	NodeKind.TiltEdit(ctx, e.Attr, e.Num, e.Flag == "up", &md.MR, &md.Inboxes, md.resumeSpeed(speedSinks))
}

func (md *MoveDispatch) resumeSpeed(speedSinks SliderPanel.Sinks) func() {
	return func() {
		speedSinks.SendSpeed(Speed.SliderNum(md.UI.Speed), int64(md.UI.ClockDivisor))
	}
}

func tiltVectorEditFor(ctx context.Context, md *MoveDispatch, speedSinks SliderPanel.Sinks, row int32, attr string) {
	NodeKind.ApplyTiltEdit(ctx, row, attr, &md.MR, &md.Inboxes, md.resumeSpeed(speedSinks))
}

func adjustTiltPhiFor(ctx context.Context, md *MoveDispatch, row int32, up bool) {
	NodeKind.AdjustTiltPhi(ctx, row, up, &md.MR, &md.Inboxes)
}

func (md *MoveDispatch) HandleRawInput(ctx context.Context, ev Drag.RawInputMsg) {
	Gesture.HandleRawInput(Gesture.Deps{MR: md.gestureMovers(), UI: &md.UI, Mover: &md.Mover, RT: &md.RT, Ctx: ctx}, ev)
}
