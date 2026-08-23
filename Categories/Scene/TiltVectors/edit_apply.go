package SceneTiltVectors

import (
	"context"
	"strconv"

	"github.com/dtauraso/wirefold/Categories/Node/movemsg"
)

type Movers interface {
	HasNode(id string) bool
	SendMove(ctx context.Context, id string, msg movemsg.Msg)
}

type TiltInbox interface {
	SendTiltEdit(ctx context.Context, id string, m movemsg.TiltEditMsg) bool
}

func Edit(ctx context.Context, attr string, row int, dirUp bool, movers Movers, inbox TiltInbox, resume func()) {
	switch attr {
	case "reset", "start":
		Apply(ctx, int32(row), attr, movers, inbox, resume)
	case "phi":
		AdjustPhi(ctx, int32(row), dirUp, movers, inbox)
	}
}

func Apply(ctx context.Context, row int32, attr string, movers Movers, inbox TiltInbox, resume func()) {
	id := strconv.Itoa(int(row) + 1)
	if !movers.HasNode(id) {
		return
	}
	resume()
	if attr == "start" {
		inbox.SendTiltEdit(ctx, id, movemsg.TiltEditMsg{Start: true})
		return
	}
	if inbox.SendTiltEdit(ctx, id, movemsg.TiltEditMsg{Reset: true}) {
		return
	}
	movers.SendMove(ctx, id, movemsg.Msg{NodeID: id, Body: movemsg.TiltVectorReset{}})
}

func AdjustPhi(ctx context.Context, row int32, up bool, movers Movers, inbox TiltInbox) {
	id := strconv.Itoa(int(row) + 1)
	if !movers.HasNode(id) {
		return
	}
	if inbox.SendTiltEdit(ctx, id, movemsg.TiltEditMsg{Up: up}) {
		return
	}
	movers.SendMove(ctx, id, movemsg.Msg{NodeID: id, Body: movemsg.TiltVectorAngle{Up: up}})
}
