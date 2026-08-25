package Dispatch

import (
	"context"
	"sync"

	"github.com/dtauraso/beadnetwork/Categories/Scene/View"

	"github.com/dtauraso/beadnetwork/Categories/Node"
	beadanimation "github.com/dtauraso/beadnetwork/Categories/Node/BeadAnimation"
	"github.com/dtauraso/beadnetwork/Categories/Node/Edge/edgegeom"
	"github.com/dtauraso/beadnetwork/Categories/Node/Edge/edgetable"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Gesture"
)

const InboxDepth = 8

type Movers struct {
	nodeGeoms map[string]*Node.NodeGeometry

	edges map[string]*edgetable.Edge

	edgeOut map[string]*beadanimation.Sender

	centerMirror map[string]Vec3
}

func NewMovers() Movers {
	return Movers{
		nodeGeoms:    map[string]*Node.NodeGeometry{},
		edges:        map[string]*edgetable.Edge{},
		edgeOut:      map[string]*beadanimation.Sender{},
		centerMirror: map[string]Vec3{},
	}
}

func (m *Movers) NodeGeoms() map[string]*Node.NodeGeometry { return m.nodeGeoms }

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
	return Node.NodeRadius(m.nodeKind(id))
}

func (m *Movers) SendMove(ctx context.Context, id string, msg Node.Msg) {
	nm, ok := m.nodeGeoms[id]
	if !ok {
		return
	}
	nm.Msg().SendExternal(ctx, msg)
}

func (m *Movers) EnqueueFor(nm *Node.NodeGeometry) func(id string, msg Node.Msg) {
	return nm.Msg().EnqueueSend
}

func (md *MoveDispatch) nearestNodeTo(p View.Vec3) (string, bool) {
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
		nm.DeriveOutEdgeGeometry()
	}

	wg := new(sync.WaitGroup)
	wg.Add(len(m.nodeGeoms))
	for _, nm := range m.nodeGeoms {
		go func() {
			defer wg.Done()
			nm.RunGeometry(ctx)
		}()
		if nm.HasTopVector() {
			wg.Add(1)
			go func() {
				defer wg.Done()
				nm.RunTopVector(ctx)
			}()
		}
	}
	return wg
}
