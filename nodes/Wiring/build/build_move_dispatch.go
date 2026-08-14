package build

import (
	"context"
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/nodefiles"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodedrag"
	"github.com/dtauraso/wirefold/nodes/Wiring/scene"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
)

func (b *buildCtx) buildMoveDispatch() error {

	nodeOrder := make([]string, len(b.spec.Nodes))
	for i, n := range b.spec.Nodes {
		nodeOrder[i] = n.ID
	}
	edgeOrder := make([]string, len(b.spec.Edges))
	for i, e := range b.spec.Edges {
		edgeOrder[i] = e.Label
	}
	md, err := dispatch.NewMoveDispatch(b.nodeGeoms, b.edgeEndpoints, b.tr, nodeOrder, edgeOrder, b.clk, &b.speedSinks, b.spec.RowCount)
	if err != nil {
		return fmt.Errorf("buildMoveDispatch: %w", err)
	}

	scenepersist.LoadLatticePoints(&md.UI, b.scenePath)
	if b.hasScene {

		md.UI.SceneSphere = b.sphere
	}

	md.LQ.QuantizedLayout = scene.SceneUsesQuantizedDrag(b.scenePath)

	coplanarEdges := scene.SceneWantsCoplanarEdges(b.scenePath)
	upAxis := scene.SceneWantsUpAxis(b.scenePath)
	if coplanarEdges || upAxis {
		for _, nm := range md.MR.NodeGeoms() {
			nm.SetSceneFlags(coplanarEdges, upAxis)
		}
	}
	for id, off := range b.quantizedOffsets {
		if nm, ok := md.MR.NodeGeoms()[id]; ok {
			nm.SetQuantOffset(off)
		}
	}

	for _, n := range b.spec.Nodes {
		nm, ok := md.MR.NodeGeoms()[n.ID]
		if !ok {
			continue
		}
		nm.SetSelfKind(n.Type)
		nm.SetOrbitRule(n.Orbit)
		nm.SetOrbitActive(nodefiles.LoadOrbitActive(b.scenePath, n.ID))
		if n.TopTiltVectorPhiIdx != nil {
			nm.SetTopTiltVectorPhiIdx(*n.TopTiltVectorPhiIdx)
		}
	}

	kindByID := make(map[string]string, len(b.spec.Nodes))
	for _, n := range b.spec.Nodes {
		kindByID[n.ID] = n.Type
	}
	linkNeighborKind := func(fromID, toID string) {
		nm, ok := md.MR.NodeGeoms()[fromID]
		if !ok {
			return
		}
		nm.AddNeighborKind(toID, kindByID[toID])
	}
	for _, e := range b.spec.Edges {
		linkNeighborKind(e.Source, e.Target)
		linkNeighborKind(e.Target, e.Source)
	}

	for _, e := range b.spec.Edges {
		nm, ok := md.MR.NodeGeoms()[e.Source]
		if !ok {
			continue
		}
		nm.AddOutTarget(e.Target)
	}

	for _, e := range b.spec.Edges {
		d, ok := e.Delta()
		if !ok {
			continue
		}
		if src, ok := md.MR.NodeGeoms()[e.Source]; ok {
			src.SetDeltaTo(e.Target, d)
		}
		if dst, ok := md.MR.NodeGeoms()[e.Target]; ok {
			dst.SetDeltaTo(e.Source, d.Neg())
		}
	}

	for _, nm := range md.MR.NodeGeoms() {
		for to, told := range nodedrag.Requested(nm.SelfKind(), polar.Polar{}, nm) {
			if other, ok := md.MR.NodeGeoms()[to]; ok {
				d := told
				other.SendExternal(context.TODO(), movemsg.Msg{Kind: movemsg.KindDrag, NodeID: to,
					SenderID: nm.ID(), Delta: &d})
			}
		}
	}
	b.md = md
	return nil
}
