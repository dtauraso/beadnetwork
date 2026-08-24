package Node

import (
	"context"
	"github.com/dtauraso/beadnetwork/Categories/Node/TiltVectors"
	"strconv"
)

type TiltMovers interface {
	HasNode(id string) bool
	SendMove(ctx context.Context, id string, msg Msg)
}

type TiltInbox interface {
	SendTiltEdit(ctx context.Context, id string, m TiltVectors.TiltEditMsg) bool
}

func TiltEdit(ctx context.Context, attr string, row int, dirUp bool, movers TiltMovers, inbox TiltInbox, resume func()) {
	switch attr {
	case "reset", "start":
		ApplyTiltEdit(ctx, int32(row), attr, movers, inbox, resume)
	case "phi":
		AdjustTiltPhi(ctx, int32(row), dirUp, movers, inbox)
	}
}

func ApplyTiltEdit(ctx context.Context, row int32, attr string, movers TiltMovers, inbox TiltInbox, resume func()) {
	id := strconv.Itoa(int(row) + 1)
	if !movers.HasNode(id) {
		return
	}
	resume()
	if attr == "start" {
		inbox.SendTiltEdit(ctx, id, TiltVectors.TiltEditMsg{Start: true})
		return
	}
	if inbox.SendTiltEdit(ctx, id, TiltVectors.TiltEditMsg{Reset: true}) {
		return
	}
	movers.SendMove(ctx, id, Msg{NodeID: id, Body: TiltVectorReset{}})
}

func AdjustTiltPhi(ctx context.Context, row int32, up bool, movers TiltMovers, inbox TiltInbox) {
	id := strconv.Itoa(int(row) + 1)
	if !movers.HasNode(id) {
		return
	}
	if inbox.SendTiltEdit(ctx, id, TiltVectors.TiltEditMsg{Up: up}) {
		return
	}
	movers.SendMove(ctx, id, Msg{NodeID: id, Body: TiltVectorAngle{Up: up}})
}
