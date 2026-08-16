package distancegroups

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/moverreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodemove"
	"github.com/dtauraso/wirefold/nodes/Wiring/scene"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
	"github.com/dtauraso/wirefold/nodes/spatial"
)

func ResolveSceneDistanceGroups(ui *viewstate.UIState, scenePath string) {
	s := scene.For(scenePath)
	ui.HasDistanceGroups = s.DistanceGroups

	ui.SceneEditable = s.Editable
	ui.SceneKinds = s.KindMask()
}

func ApplyDistanceGroupTarget(ctx context.Context, ui *viewstate.UIState, mr *moverreg.MoverRegistry, mv *nodemove.NodeMover, groupIdx, dir int) bool {
	rootMove := func(ctx context.Context, target string, newPos spatial.Vec3) bool {
		return mv.RootMove(ctx, mr.NodeGeoms(), target, newPos)
	}
	return ApplyTarget(ctx, ui.HasDistanceGroups, mr.CenterOfNode, rootMove, groupIdx, dir)
}
