package Dispatch

import (
	"context"
	"fmt"

	NodeKind "github.com/dtauraso/wirefold/src/Node"

	"github.com/dtauraso/wirefold/src/Input/Stdin"
	T "github.com/dtauraso/wirefold/src/Trace"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"

	"github.com/dtauraso/wirefold/src/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/wirefold/src/Node/nodecrud"
	"github.com/dtauraso/wirefold/src/Node/rulenode"
	"github.com/dtauraso/wirefold/src/Scene/sceneswitch"
)

func tiltVectorEdit(ctx context.Context, md *MoveDispatch, speedSinks SliderPanel.Sinks, row int32, attr string) {
	tiltVectorEditFor(ctx, &md.UI, &md.MR, &md.Inboxes, speedSinks, row, attr)
}

func adjustTiltPhi(ctx context.Context, md *MoveDispatch, row int32, up bool) {
	adjustTiltPhiFor(ctx, &md.MR, &md.Inboxes, row, up)
}

func setLatticePoints(md *MoveDispatch, points int32) {
	if points < AngleDropdown.LatticePointsMin || points > AngleDropdown.LatticePointsMax || points%4 != 0 {
		return
	}
	md.UI.LatticePoints = points
	md.Persist.Lattice().Schedule(points)
	md.Inboxes.BroadcastLatticePoints(points)
}

func applyUpdateScene(ctx context.Context, msg Stdin.StdinMsg, md *MoveDispatch, speedSinks SliderPanel.Sinks) {
	if md == nil {
		return
	}
	switch msg.Attr {
	case "selected":
		sceneswitch.SelectScene(&md.Scenes, int(msg.Num))
	case "latticePoints":
		setLatticePoints(md, int32(msg.Num))
	case "viewport":
		md.UI.ViewW = msg.X
		md.UI.ViewH = msg.Y
		md.UI.EmitBreadcrumb(T.RowEvent{
			Label: T.BreadcrumbViewport, NodeRow: -1, PortRow: -1, TargetRow: -1,
			TargetPortRow: -1, EdgeRow: -1, Slot: -1,
			Value: int32(md.UI.FovDeg()),
			Text:  fmt.Sprintf("%.0fx%.0f", msg.X, msg.Y),
		})
	case "create":

		nodecrud.CreateNode(&md.Scenes, &md.UI, &md.MR, uint8(msg.Num), msg.X, msg.Y)
	case "delete":

		nodecrud.DeleteNode(&md.Scenes, &md.UI, &md.RT, msg.Num)
	}
}

func applyUpdateNode(ctx context.Context, msg Stdin.StdinMsg, md *MoveDispatch, _ SliderPanel.Sinks) {
	NodeKind.EditNode(ctx, msg, &md.Rules)
}

func sendRuleEdit(ctx context.Context, md *MoveDispatch, row int, edit rulenode.Edit) {
	NodeKind.SendRuleEdit(ctx, &md.Rules, row, edit)
}
