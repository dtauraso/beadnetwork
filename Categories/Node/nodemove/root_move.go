package nodemove

import (
	"context"

	"github.com/dtauraso/wirefold/Categories/Node/movemsg"
	"github.com/dtauraso/wirefold/Categories/Node/nodeactor"
	"github.com/dtauraso/wirefold/Categories/Polar/polar"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

type NodeMover struct{}

func HeldCenters(nodeGeoms map[string]*nodeactor.NodeGeometry, centerOf func(id string) (Vec3, bool)) map[string]Vec3 {
	out := make(map[string]Vec3, len(nodeGeoms))
	for id := range nodeGeoms {
		if c, ok := centerOf(id); ok {
			out[id] = c
		}
	}
	return out
}

func pointerPolar(nm *nodeactor.NodeGeometry, v Vec3) polar.Polar {
	if rule := nm.SelfRule(); rule != nil && nm.SelfRuleActive() && rule.MaxTheta != nil {
		return polar.Cart2polarAtTheta(polar.Vec3(v), nm.ScenePolar().Theta)
	}
	return polar.Cart2polar(polar.Vec3(v))
}

func (mv *NodeMover) RootMove(ctx context.Context, nodeGeoms map[string]*nodeactor.NodeGeometry, nodeID string, target Vec3) bool {
	nm, ok := nodeGeoms[nodeID]
	if !ok {
		return false
	}

	sc := nm.Constants()
	targetIdx := polarindex.MeasureIndex(pointerPolar(nm, target.Sub(Vec3(nm.SceneCenter()))), sc)

	nm.SendExternal(ctx, movemsg.Msg{Kind: movemsg.KindDrag, NodeID: nodeID, Target: &targetIdx})
	return true
}
