package runtopology

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/TiltPanel"
	"github.com/dtauraso/wirefold/src/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/wirefold/src/Scene/scene"

	"github.com/dtauraso/wirefold/src/Node/movemsg"
	"github.com/dtauraso/wirefold/src/Node/nodedrag"
	"github.com/dtauraso/wirefold/src/Polar/polarindex"
	"github.com/dtauraso/wirefold/src/runtopology/scenerun"
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
	md, err := scenerun.NewMoveDispatch(b.nodeGeoms, b.edgeEndpoints, nodeOrder, edgeOrder, b.clk, &b.speedSinks, b.spec.RowCount, b.spec.Constants)
	if err != nil {
		return fmt.Errorf("buildMoveDispatch: %w", err)
	}

	md.UI.LatticePoints = AngleDropdown.LatticePointsFor(b.scenePath)
	if b.hasScene {

		md.UI.SceneSphere = b.sphere
	}
	md.UI.Constants = b.spec.Constants

	s := scene.For(b.scenePath)

	coplanarEdges := s.CoplanarEdges
	upAxis := s.UpAxis
	if coplanarEdges || upAxis {
		for _, nm := range md.MR.NodeGeoms() {
			nm.SetSceneFlags(coplanarEdges, upAxis)
		}
	}
	for id, off := range b.baseIndices {
		if nm, ok := md.MR.NodeGeoms()[id]; ok {
			nm.SetBaseIndex(off)
		}
	}
	for id, off := range b.dragIndices {
		if nm, ok := md.MR.NodeGeoms()[id]; ok {
			nm.SetDragIndex(off)
		}
	}

	for _, n := range b.spec.Nodes {
		if nm, ok := md.MR.NodeGeoms()[n.ID]; ok {
			nm.SeedFromSpec(n, b.scenePath)
		}
	}

	for row := 0; row < b.spec.RowCount; row++ {
		for _, n := range b.spec.Nodes {
			id, err := strconv.Atoi(n.ID)
			if err != nil || id-1 != row || !TiltPanel.KindWantsVectorChannel(n.Type) {
				continue
			}
			md.UI.TiltRows = append(md.UI.TiltRows, int32(row))
			md.UI.TiltLabels = append(md.UI.TiltLabels, n.ID)
		}
	}

	buildRulePanelNodes(md, b.spec)

	kindByID := make(map[string]string, len(b.spec.Nodes))
	for _, n := range b.spec.Nodes {
		kindByID[n.ID] = n.Type
	}

	targetByLabel := make(map[string]string, len(b.spec.Edges))
	sourceByLabel := make(map[string]string, len(b.spec.Edges))
	for _, e := range b.spec.Edges {
		if src, ok := md.MR.NodeGeoms()[e.Source]; ok {
			src.SeedEdge(e, true, kindByID[e.Target], b.scenePath)
		}
		if dst, ok := md.MR.NodeGeoms()[e.Target]; ok {
			dst.SeedEdge(e, false, kindByID[e.Source], b.scenePath)
		}
		targetByLabel[e.Label] = e.Target
		sourceByLabel[e.Label] = e.Source
	}

	md.Rules.TogglesByEdgeRow = make([]chan<- struct{}, len(md.RT.EdgeRowTable))
	for row, label := range md.RT.EdgeRowTable {
		src, okS := md.MR.NodeGeoms()[sourceByLabel[label]]
		target, okT := targetByLabel[label], true
		if !okS || !okT {
			continue
		}
		md.Rules.TogglesByEdgeRow[row] = src.RuleNode().EdgeToggleChannel(target)
	}

	for _, nm := range md.MR.NodeGeoms() {
		nm.RuleNode().BroadcastSelf()
	}

	for _, nm := range md.MR.NodeGeoms() {
		for to, told := range nodedrag.Requested(nm.SelfKind(), polarindex.Offset{}, nm) {
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
