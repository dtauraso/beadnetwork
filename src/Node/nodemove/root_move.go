package nodemove

import (
	"context"

	"github.com/dtauraso/wirefold/src/Polar/polar"
	"github.com/dtauraso/wirefold/src/Node/movemsg"
	"github.com/dtauraso/wirefold/src/Node/nodeactor"
	"github.com/dtauraso/wirefold/src/Polar/polarindex"
	"github.com/dtauraso/wirefold/src/spatial"
)

type NodeMover struct{}

func HeldCenters(nodeGeoms map[string]*nodeactor.NodeGeometry, centerOf func(id string) (spatial.Vec3, bool)) map[string]spatial.Vec3 {
	out := make(map[string]spatial.Vec3, len(nodeGeoms))
	for id := range nodeGeoms {
		if c, ok := centerOf(id); ok {
			out[id] = c
		}
	}
	return out
}

func pointerPolar(nm *nodeactor.NodeGeometry, v spatial.Vec3) polar.Polar {
	if rule := nm.SelfRule(); rule != nil && nm.SelfRuleActive() && rule.MaxTheta != nil {
		return polar.Cart2polarAtTheta(v, nm.ScenePolar().Theta)
	}
	return polar.Cart2polar(v)
}

func (mv *NodeMover) RootMove(ctx context.Context, nodeGeoms map[string]*nodeactor.NodeGeometry, nodeID string, target spatial.Vec3) bool {
	nm, ok := nodeGeoms[nodeID]
	if !ok {
		return false
	}

	sc := nm.Constants()
	targetIdx := polarindex.MeasureIndex(pointerPolar(nm, target.Sub(nm.SceneCenter())), sc)

	nm.SendExternal(ctx, movemsg.Msg{Kind: movemsg.KindDrag, NodeID: nodeID, Target: &targetIdx})
	return true
}
