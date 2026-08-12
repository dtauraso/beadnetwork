package stdinreader

import (
	"context"
	"strconv"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	"github.com/dtauraso/wirefold/nodes/Wiring/distancegroups"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenestructure"
	"github.com/dtauraso/wirefold/nodes/Wiring/sceneswitch"
)

func applyUpdateClock(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, tr *T.Trace, speedSinks []chan float64) {
	if h, ok := clockAttrHandlers[msg.Attr]; ok {
		h(msg, md, speedSinks)
	}
}

func applyUpdateDistanceGroup(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, tr *T.Trace, speedSinks []chan float64) {
	if md == nil || msg.Attr != "length" {
		return
	}
	dir := -1
	if msg.Flag == "up" {
		dir = 1
	}
	if distancegroups.ApplyDistanceGroupTarget(ctx, &md.UI, &md.MR, &md.LQ, msg.Num, dir) {
		md.UI.EmitViewFrame(nil)
	}
}

func applyUpdateTiltVector(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, tr *T.Trace, speedSinks []chan float64) {
	if md == nil || (msg.Attr != "theta" && msg.Attr != "reset" && msg.Attr != "start") {
		return
	}
	id := strconv.Itoa(msg.Num + 1)
	if _, ok := md.MR.NodeGeoms()[id]; !ok {
		return
	}
	if msg.Attr == "reset" {

		scenepersist.BroadcastSpeed(speedSinks, scenepersist.SliderSpeed(&md.UI))
		if md.Inboxes.SendTiltEdit(ctx, id, movemsg.TiltEditMsg{Reset: true}) {
			return
		}
		md.MR.SendMove(ctx, id, movemsg.Msg{Kind: movemsg.KindTiltVectorReset, NodeID: id})
		return
	}
	if msg.Attr == "start" {

		scenepersist.BroadcastSpeed(speedSinks, scenepersist.SliderSpeed(&md.UI))

		md.Inboxes.SendTiltEdit(ctx, id, movemsg.TiltEditMsg{Start: true})
		return
	}

	scenepersist.BroadcastSpeed(speedSinks, scenepersist.HumanEditSpeed)
	up := msg.Flag == "up"
	if md.Inboxes.SendTiltEdit(ctx, id, movemsg.TiltEditMsg{Axis: msg.Attr, Up: up}) {
		return
	}
	md.MR.SendMove(ctx, id, movemsg.Msg{Kind: movemsg.KindTiltVectorAngle, NodeID: id, Axis: msg.Attr, Bool: up})
}

func applyUpdateScene(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, tr *T.Trace, speedSinks []chan float64) {
	if md == nil {
		return
	}
	switch msg.Attr {
	case "selected":
		sceneswitch.SelectScene(&md.Scenes, int(msg.Num))
	case "latticePoints":
		points := int32(msg.Num)
		if points < 4 || points > 64 || points%4 != 0 {
			return
		}
		md.UI.LatticePoints = points
		md.Persist.Lattice().Schedule(points)
		md.Inboxes.BroadcastLatticePoints(points)
	case "create":

		scenestructure.CreateNode(&md.Scenes, &md.UI, &md.MR, uint8(msg.Num), msg.X, msg.Y, tr)
	case "delete":

		scenestructure.DeleteNode(&md.Scenes, &md.UI, &md.RT, msg.Num, tr)
	}
}

func applyUpdateOverlays(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, tr *T.Trace, speedSinks []chan float64) {
	if md == nil {
		return
	}
	if h, ok := overlayAttrHandlers[msg.Attr]; ok {
		h(msg, md, tr)
	}

	md.Persist.Overlays().Schedule(md.UI.OV)
}
