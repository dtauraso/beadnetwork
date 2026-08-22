package runtopology

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/TiltPanel"
	"github.com/dtauraso/wirefold/src/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/wirefold/src/Scene/scene"

	"github.com/dtauraso/wirefold/src/runtopology/scenerun"
	"github.com/dtauraso/wirefold/src/Node/movemsg"
	"github.com/dtauraso/wirefold/src/Node/nodeactor/nodefiles"
	"github.com/dtauraso/wirefold/src/Node/nodedrag"
	"github.com/dtauraso/wirefold/src/Polar/polarindex"
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
		nm, ok := md.MR.NodeGeoms()[n.ID]
		if !ok {
			continue
		}
		nm.SetSelfKind(n.Type)
		active := nodefiles.LoadDragActive(b.scenePath, n.ID)
		nm.SetDragRule(n.Drag)
		nm.SetDragActive(active)
		rn := nm.RuleNode()
		rn.SetPersistRoot(b.scenePath)
		rn.SeedRule(n.Drag, active)
		rn.SeedKindActive(nodefiles.LoadKindRuleActive(b.scenePath, n.ID))
		selfActive := nodefiles.LoadSelfRuleActive(b.scenePath, n.ID)
		nm.SetSelfRule(n.SelfDrag)
		nm.SetSelfRuleActive(selfActive)
		rn.SeedSelfRule(n.SelfDrag, selfActive)
		if n.TopTiltVectorPhiIdx != nil {
			nm.SetTopTiltVectorPhiIdx(*n.TopTiltVectorPhiIdx)
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

	targetByLabel := make(map[string]string, len(b.spec.Edges))
	sourceByLabel := make(map[string]string, len(b.spec.Edges))
	for _, e := range b.spec.Edges {
		active := nodefiles.LoadEdgeRuleActive(b.scenePath, e.Source, e.Target)
		if src, ok := md.MR.NodeGeoms()[e.Source]; ok {
			src.RuleNode().SeedEdgeActive(e.Target, active)
		}
		if dst, ok := md.MR.NodeGeoms()[e.Target]; ok {
			dst.RuleNode().SeedEdgeActive(e.Source, active)
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
		baseD, ok := e.BaseDeltaIndex()
		if !ok {
			continue
		}
		dragD := e.DragDeltaIndex()
		if src, ok := md.MR.NodeGeoms()[e.Source]; ok {
			src.SetBaseDeltaTo(e.Target, baseD)
			src.SetDragDeltaTo(e.Target, dragD)
		}
		if dst, ok := md.MR.NodeGeoms()[e.Target]; ok {
			dst.SetBaseDeltaTo(e.Source, polarindex.Neg(baseD))
			dst.SetDragDeltaTo(e.Source, polarindex.Neg(dragD))
		}
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
