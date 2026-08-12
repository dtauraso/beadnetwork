package distancegroups

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/layoutquant"
	"github.com/dtauraso/wirefold/nodes/Wiring/moverreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/scene"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

func ResolveSceneDistanceGroups(ui *viewstate.UIState, scenePath string) {
	ui.HasDistanceGroups = scene.SceneHasDistanceGroups(scenePath)

	ui.SceneEditable = scene.SceneIsEditable(scenePath)
	ui.SceneKinds = scene.SceneKindMask(scenePath)
}

func ApplyDistanceGroupTarget(ctx context.Context, ui *viewstate.UIState, mr *moverreg.MoverRegistry, lq *layoutquant.LayoutQuantizer, groupIdx, dir int) bool {
	rootMove := func(ctx context.Context, target string, newPos wire.Vec3) bool {
		return lq.RootMove(ctx, mr.NodeGeoms(), target, newPos)
	}
	return ApplyTarget(ctx, ui.HasDistanceGroups, mr.CenterOfNode, rootMove, groupIdx, dir)
}
