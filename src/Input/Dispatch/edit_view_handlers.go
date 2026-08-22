package Dispatch

import (
	"context"
	"strconv"

	clock "github.com/dtauraso/wirefold/src/Clock"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/src/Input/Stdin"
	"github.com/dtauraso/wirefold/src/Node/movemsg"
	"github.com/dtauraso/wirefold/src/Node/moverreg"
	"github.com/dtauraso/wirefold/src/Node/nodeinbox"
	"github.com/dtauraso/wirefold/src/Overlay"
	"github.com/dtauraso/wirefold/src/Scene/scenepersist"
	"github.com/dtauraso/wirefold/src/Scene/viewstate"
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
	editTiltVector(ctx, msg, &md.UI, &md.MR, &md.Inboxes, speedSinks)
}

func editTiltVector(
	ctx context.Context,
	msg Stdin.StdinMsg,
	ui *viewstate.UIState,
	mr *moverreg.MoverRegistry,
	inboxes *nodeinbox.NodeInboxes,
	speedSinks SliderPanel.Sinks,
) {
	if msg.Attr != "phi" && msg.Attr != "reset" && msg.Attr != "start" {
		return
	}
	id := strconv.Itoa(msg.Num + 1)
	if _, ok := mr.NodeGeoms()[id]; !ok {
		return
	}
	if msg.Attr == "reset" || msg.Attr == "start" {
		tiltVectorEditFor(ctx, ui, mr, inboxes, speedSinks, int32(msg.Num), msg.Attr)
		return
	}
	adjustTiltPhiFor(ctx, mr, inboxes, int32(msg.Num), msg.Flag == "up")
}

func tiltVectorEditFor(
	ctx context.Context,
	ui *viewstate.UIState,
	mr *moverreg.MoverRegistry,
	inboxes *nodeinbox.NodeInboxes,
	speedSinks SliderPanel.Sinks,
	row int32,
	attr string,
) {
	id := strconv.Itoa(int(row) + 1)
	if _, ok := mr.NodeGeoms()[id]; !ok {
		return
	}
	SliderPanel.Broadcast(speedSinks, scenepersist.SliderNum(ui.Speed), int64(ui.ClockDivisor))
	if attr == "start" {
		inboxes.SendTiltEdit(ctx, id, movemsg.TiltEditMsg{Start: true})
		return
	}
	if inboxes.SendTiltEdit(ctx, id, movemsg.TiltEditMsg{Reset: true}) {
		return
	}
	mr.SendMove(ctx, id, movemsg.Msg{Kind: movemsg.KindTiltVectorReset, NodeID: id})
}

func adjustTiltPhiFor(
	ctx context.Context,
	mr *moverreg.MoverRegistry,
	inboxes *nodeinbox.NodeInboxes,
	row int32,
	up bool,
) {
	id := strconv.Itoa(int(row) + 1)
	if _, ok := mr.NodeGeoms()[id]; !ok {
		return
	}
	if inboxes.SendTiltEdit(ctx, id, movemsg.TiltEditMsg{Up: up}) {
		return
	}
	mr.SendMove(ctx, id, movemsg.Msg{Kind: movemsg.KindTiltVectorAngle, NodeID: id, Bool: up})
}
