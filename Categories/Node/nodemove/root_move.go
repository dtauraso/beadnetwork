package nodemove

import (
	"context"

	"github.com/dtauraso/wirefold/Categories/Node"
	"github.com/dtauraso/wirefold/Categories/Node/nodegeom"
	"github.com/dtauraso/wirefold/Categories/Node/owners"
	"github.com/dtauraso/wirefold/Categories/Polar/polar"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

type NodeMover struct{}

func HeldCenters(nodeGeoms map[string]*Node.NodeGeometry, centerOf func(id string) (Vec3, bool)) map[string]Vec3 {
	out := make(map[string]Vec3, len(nodeGeoms))
	for id := range nodeGeoms {
		if c, ok := centerOf(id); ok {
			out[id] = c
		}
	}
	return out
}

func (mv *NodeMover) RootMove(ctx context.Context, nodeGeoms map[string]*Node.NodeGeometry, nodeID string, target Vec3) bool {
	nm, ok := nodeGeoms[nodeID]
	if !ok {
		return false
	}

	sc := nm.Constants()
	var targetIdx polarindex.Index
	if rule := nm.Topo().SelfRule(); rule != nil && nm.Topo().SelfRuleActive() && rule.MaxTheta != nil {
		targetIdx = nodegeom.IndexAtTheta(nm.SceneCenter(), nodegeom.Vec3(target), nm.ScenePolar().Theta, sc)
	} else {
		targetIdx = polarindex.MeasureIndex(polar.Cart2polar(polar.Vec3(target.Sub(Vec3(nm.SceneCenter())))), sc)
	}

	nm.Msg().SendExternal(ctx, owners.Msg{NodeID: nodeID, Body: owners.Drag{Target: &targetIdx}})
	return true
}
