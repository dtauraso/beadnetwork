package moverreg

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/edgegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
)

func (mr *MoverRegistry) drainCenterMirror() {
	if mr.centerMirror == nil {
		mr.centerMirror = map[string]vec3{}
	}
	for id, nm := range mr.nodeGeoms {
		if c, ok := nm.PollCenter(); ok {
			mr.centerMirror[id] = c
		}
	}
}

func (mr *MoverRegistry) CenterOfNode(id string) (vec3, bool) {
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

func (mr *MoverRegistry) NodeQuantOffset(id string) (iTheta, iPhi, iR int, ok bool) {
	nm, exists := mr.nodeGeoms[id]
	if !exists {
		return 0, 0, 0, false
	}
	t, p, r := nm.QuantOffset()
	return t, p, r, true
}

func (mr *MoverRegistry) NearestNodeTo(p vec3) (string, bool) {
	centers := make(map[string]vec3, len(mr.nodeGeoms))
	for id, ng := range mr.nodeGeoms {
		centers[id] = ng.WorldCenter()
	}
	return edgegeom.NearestTo(centers, p)
}
