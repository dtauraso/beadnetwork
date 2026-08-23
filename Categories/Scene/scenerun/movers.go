package scenerun

import (
	"context"
	"sync"

	"github.com/dtauraso/wirefold/Categories/Input/Gesture"
	beadanimation "github.com/dtauraso/wirefold/Categories/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/Categories/Node/Edge/edgegeom"
	"github.com/dtauraso/wirefold/Categories/Node/Edge/edgetable"
	"github.com/dtauraso/wirefold/Categories/Node/nodeactor"
	"github.com/dtauraso/wirefold/Categories/Node/nodecrud"
	"github.com/dtauraso/wirefold/Categories/Node/nodegeom"
	"github.com/dtauraso/wirefold/Categories/Node/owners"
)

const InboxDepth = 8

type Movers struct {
	nodeGeoms map[string]*nodeactor.NodeGeometry

	edges map[string]*edgetable.Edge

	edgeOut map[string]*beadanimation.Sender

	centerMirror map[string]Vec3
}

func NewMovers() Movers {
	return Movers{
		nodeGeoms:    map[string]*nodeactor.NodeGeometry{},
		edges:        map[string]*edgetable.Edge{},
		edgeOut:      map[string]*beadanimation.Sender{},
		centerMirror: map[string]Vec3{},
	}
}

func (m *Movers) NodeGeoms() map[string]*nodeactor.NodeGeometry { return m.nodeGeoms }

func (m *Movers) HasNode(id string) bool {
	_, ok := m.nodeGeoms[id]
	return ok
}

func (m *Movers) Edges() map[string]*edgetable.Edge { return m.edges }

func (m *Movers) SeedCenter(id string, c Vec3) { m.centerMirror[id] = c }

func (m *Movers) drainCenterMirror() {
	for id, nm := range m.nodeGeoms {
		if c, ok := nm.Msg().PollCenter(); ok {
			m.centerMirror[id] = Vec3(c)
		}
	}
}

func (m *Movers) CenterOfNode(id string) (Vec3, bool) {
	m.drainCenterMirror()
	c, ok := m.centerMirror[id]
	return c, ok
}

func (m *Movers) NearestNodeTo(p Vec3) (string, bool) {
	m.drainCenterMirror()
	centers := make(map[string]edgegeom.Vec3, len(m.centerMirror))
	for id, c := range m.centerMirror {
		centers[id] = edgegeom.Vec3(c)
	}
	return edgegeom.NearestTo(centers, edgegeom.Vec3(p))
}

func (m *Movers) nodeKind(id string) string {
	if nm, ok := m.nodeGeoms[id]; ok {
		return nm.Kind()
	}
	return ""
}

func (m *Movers) NodeBodyRadius(id string) float64 {
	return nodegeom.NodeRadius(m.nodeKind(id))
}

func (m *Movers) SendMove(ctx context.Context, id string, msg owners.Msg) {
	nm, ok := m.nodeGeoms[id]
	if !ok {
		return
	}
	nm.Msg().SendExternal(ctx, msg)
}

func (m *Movers) EnqueueFor(nm *nodeactor.NodeGeometry) func(id string, msg owners.Msg) {
	return nm.Msg().EnqueueSend
}

func (md *MoveDispatch) nearestNodeTo(p nodecrud.Vec3) (string, bool) {
	return md.MR.NearestNodeTo(Vec3(p))
}

func (md *MoveDispatch) gestureMovers() Gesture.Movers {
	return Gesture.Movers{
		NodeGeoms: md.MR.NodeGeoms,
		CenterOf: func(id string) (Gesture.Vec3, bool) {
			c, ok := md.MR.CenterOfNode(id)
			return Gesture.Vec3(c), ok
		},
		BodyRadius: md.MR.NodeBodyRadius,
		SendMove:   md.MR.SendMove,
	}
}

func (m *Movers) Start(ctx context.Context) *sync.WaitGroup {
	for _, nm := range m.nodeGeoms {
		nm.DeriveOutEdgeGeometryOnce()
	}

	wg := new(sync.WaitGroup)
	wg.Add(len(m.nodeGeoms))
	for _, nm := range m.nodeGeoms {
		go func() {
			defer wg.Done()
			nm.RunGeometry(ctx)
		}()
	}
	return wg
}
