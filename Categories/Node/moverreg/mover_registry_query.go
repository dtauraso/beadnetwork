package moverreg

import (
	"context"

	"github.com/dtauraso/wirefold/Categories/Node/Edge/edgegeom"
	"github.com/dtauraso/wirefold/Categories/Node/movemsg"
	"github.com/dtauraso/wirefold/Categories/Node/nodeactor"
	"github.com/dtauraso/wirefold/Categories/Node/nodegeom"
)

func (mr *MoverRegistry) drainCenterMirror() {
	if mr.centerMirror == nil {
		mr.centerMirror = map[string]Vec3{}
	}
	for id, nm := range mr.nodeGeoms {
		if c, ok := nm.PollCenter(); ok {
			mr.centerMirror[id] = Vec3(c)
		}
	}
}

func (mr *MoverRegistry) CenterOfNode(id string) (Vec3, bool) {
	mr.drainCenterMirror()
	c, ok := mr.centerMirror[id]
	return c, ok
}

func (mr *MoverRegistry) SendMove(ctx context.Context, id string, msg movemsg.Msg) {
	nm, ok := mr.nodeGeoms[id]
	if !ok {
		return
	}
	nm.SendExternal(ctx, msg)
}

func (mr *MoverRegistry) EnqueueFor(nm *nodeactor.NodeGeometry) func(id string, msg movemsg.Msg) {
	return nm.EnqueueSend
}

func (mr *MoverRegistry) nodeKind(nodeID string) string {
	if nm, ok := mr.nodeGeoms[nodeID]; ok {
		return nm.Kind()
	}
	return ""
}

func (mr *MoverRegistry) NodeBodyRadius(id string) float64 {
	return nodegeom.NodeRadius(mr.nodeKind(id))
}

func (mr *MoverRegistry) NearestNodeTo(p Vec3) (string, bool) {
	centers := make(map[string]edgegeom.Vec3, len(mr.nodeGeoms))
	for id, ng := range mr.nodeGeoms {
		centers[id] = edgegeom.Vec3(ng.WorldCenter())
	}
	return edgegeom.NearestTo(centers, edgegeom.Vec3(p))
}
