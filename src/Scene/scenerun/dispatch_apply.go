package scenerun

import (
	"context"

	NodeKind "github.com/dtauraso/wirefold/src/Node"

	"github.com/dtauraso/wirefold/src/Input/Stdin"
	edge "github.com/dtauraso/wirefold/src/Node/Edge"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"

	"github.com/dtauraso/wirefold/src/Node/nodecrud"
	"github.com/dtauraso/wirefold/src/Node/rulenode"
	"github.com/dtauraso/wirefold/src/Scene/sceneswitch"
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
		md.UI.SetViewport(msg.X, msg.Y)
	case "create":

		nodecrud.CreateNode(&md.Scenes, &md.UI, &md.MR, uint8(msg.Num), msg.X, msg.Y)
	case "delete":

		nodecrud.DeleteNode(&md.Scenes, &md.UI, &md.RT, msg.Num)
	}
}

func applyUpdateNode(ctx context.Context, msg Stdin.StdinMsg, md *MoveDispatch, _ SliderPanel.Sinks) {
	NodeKind.EditNode(ctx, msg, &md.Rules)
}

func applyUpdateEdge(ctx context.Context, msg Stdin.StdinMsg, md *MoveDispatch, _ SliderPanel.Sinks) {
	if md == nil {
		return
	}
	edge.EditEdge(ctx, msg, md.Rules.TogglesByEdgeRow)
}

func sendRuleEdit(ctx context.Context, md *MoveDispatch, row int, edit rulenode.Edit) {
	NodeKind.SendRuleEdit(ctx, &md.Rules, row, edit)
}
