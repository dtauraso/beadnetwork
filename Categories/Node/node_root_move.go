package Node

import (
	"context"

	"github.com/dtauraso/beadnetwork/Categories/Vector/polar"
	"github.com/dtauraso/beadnetwork/Categories/Vector/polarindex"
)

type NodeMover struct{}

func HeldCenters(nodeGeoms map[string]*NodeGeometry, centerOf func(id string) (Vec3, bool)) map[string]Vec3 {
	out := make(map[string]Vec3, len(nodeGeoms))
	for id := range nodeGeoms {
		if c, ok := centerOf(id); ok {
			out[id] = c
		}
	}
	return out
}

func (mv *NodeMover) RootMove(ctx context.Context, nodeGeoms map[string]*NodeGeometry, nodeID string, target Vec3) bool {
	nm, ok := nodeGeoms[nodeID]
	if !ok {
		return false
	}

	sc := nm.Constants()
	var targetIdx polarindex.Index
	if rule := nm.Topo().SelfRule(); rule != nil && nm.Topo().SelfRuleActive() && rule.MaxTheta != nil {
		targetIdx = IndexAtTheta(nm.SceneCenter(), Vec3(target), nm.ScenePolar().Theta, sc)
	} else {
		targetIdx = polarindex.MeasureIndex(polar.Cart2polar(polar.Vec3(target.Sub(nm.SceneCenter()))), sc)
	}

	nm.Msg().SendExternal(ctx, Msg{NodeID: nodeID, Body: Drag{Target: &targetIdx}})
	return true
}
