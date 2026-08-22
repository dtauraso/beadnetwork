package Dispatch

import (
	"context"
	"strconv"

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
	editOverlays(msg, &md.UI, &md.Inboxes, md.Persist.Overlays())
}

func editOverlays(
	msg Stdin.StdinMsg,
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

func applyUpdatePanels(_ context.Context, msg Stdin.StdinMsg, md *MoveDispatch, _ SliderPanel.Sinks) {
	editPanels(msg, &md.UI, md.Persist.Panels())
}

func editPanels(
	msg Stdin.StdinMsg,
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

func applyUpdateClock(_ context.Context, msg Stdin.StdinMsg, md *MoveDispatch, speedSinks SliderPanel.Sinks) {
	editClock(msg, &md.UI, speedSinks, md.Persist.Speed())
}

func editClock(
	msg Stdin.StdinMsg,
	ui *viewstate.UIState,
	speedSinks SliderPanel.Sinks,
	persist *scenepersist.Persister[float64],
) {
	if msg.Attr != "speed" {
		return
	}
	SliderPanel.Broadcast(speedSinks, int64(msg.Num), int64(ui.ClockDivisor))

	userSpeed := float64(msg.Num) / SliderPanel.NumScale
	ui.Speed = userSpeed
	persist.Schedule(userSpeed)
	ui.EmitViewFrame(nil)
}

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
