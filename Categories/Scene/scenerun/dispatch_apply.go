package scenerun

import (
	"context"

	NodeKind "github.com/dtauraso/wirefold/Categories/Node"

	edge "github.com/dtauraso/wirefold/Categories/Node/Edge"

	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/Categories/Scene"

	"github.com/dtauraso/wirefold/Categories/Scene/sceneswitch"
	"github.com/dtauraso/wirefold/Categories/Scene/structuraledit"
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
		sceneswitch.SelectScene(&md.Scenes, int(e.Num))
	case "latticePoints":
		setLatticePoints(md, int32(e.Num))
	case "viewport":
		md.UI.SetViewport(e.X, e.Y)
	case "create":

		structuraledit.CreateNode(&md.Scenes, &md.UI, md.MR.NodeGeoms(), md.nearestNodeTo, uint8(e.Num), e.X, e.Y)
	case "delete":

		structuraledit.DeleteNode(&md.Scenes, &md.UI, &md.RT, e.Num)
	}
}

func applyUpdateNode(ctx context.Context, attr byte, payload []byte, md *MoveDispatch, _ SliderPanel.Sinks) {
	e, ok := NodeKind.DecodeUpdate(payload, attr)
	if !ok {
		return
	}
	NodeKind.EditNode(ctx, e, &md.Rules)
}

func applyUpdateEdge(ctx context.Context, attr byte, payload []byte, md *MoveDispatch, _ SliderPanel.Sinks) {
	if md == nil {
		return
	}
	e, ok := edge.DecodeUpdate(payload, attr)
	if !ok {
		return
	}
	edge.EditEdge(ctx, e, md.Rules.TogglesByEdgeRow)
}

func sendRuleEdit(ctx context.Context, md *MoveDispatch, row int, edit NodeKind.RuleEdit) {
	NodeKind.SendRuleEdit(ctx, &md.Rules, row, edit)
}
