package scenebuild

import (
	"context"
	"fmt"

	"github.com/dtauraso/wirefold/Categories/Scene/scenerun"

	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Pills/AngleDropdown"
	clock "github.com/dtauraso/wirefold/Categories/Clock"
	"github.com/dtauraso/wirefold/Categories/Scene/loadspec"
	"github.com/dtauraso/wirefold/Categories/Scene/scene"

	edge "github.com/dtauraso/wirefold/Categories/Node/Edge"
	"github.com/dtauraso/wirefold/Categories/Node/nodeactor/owners"
	"github.com/dtauraso/wirefold/Categories/Node/nodegeom"
	"github.com/dtauraso/wirefold/Categories/Polar/polar"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

func NewFromSpec(spec loadspec.TopoSpec, sphere polar.SceneSphere, hasScene bool, scenePath string, clk clock.Clock, speedSinks *SliderPanel.Sinks, nodeGeoms map[string]nodegeom.NodeGeom, edgeEndpoints map[string]edge.EdgeEndpoints, baseIndices map[string]polarindex.Index, dragIndices map[string]polarindex.Offset) (*scenerun.MoveDispatch, error) {

	nodeOrder := make([]string, len(spec.Nodes))
	for i, n := range spec.Nodes {
		nodeOrder[i] = n.ID
	}
	edgeOrder := make([]string, len(spec.Edges))
	for i, e := range spec.Edges {
		edgeOrder[i] = e.Label
	}
	md, err := scenerun.NewMoveDispatch(nodeGeoms, edgeEndpoints, nodeOrder, edgeOrder, clk, speedSinks, spec.RowCount, spec.Constants)
	if err != nil {
		return nil, fmt.Errorf("scenerun.NewFromSpec: %w", err)
	}

	md.UI.LatticePoints = AngleDropdown.LatticePointsFor(scenePath)
	if hasScene {

		md.UI.SceneSphere = sphere
	}
	md.UI.Constants = spec.Constants

	s := scene.For(scenePath)

	coplanarEdges := s.CoplanarEdges
	upAxis := s.UpAxis
	if coplanarEdges || upAxis {
		for _, nm := range md.MR.NodeGeoms() {
			nm.SetSceneFlags(coplanarEdges, upAxis)
		}
	}
	for id, off := range baseIndices {
		if nm, ok := md.MR.NodeGeoms()[id]; ok {
			nm.SetBaseIndex(off)
		}
	}
	for id, off := range dragIndices {
		if nm, ok := md.MR.NodeGeoms()[id]; ok {
			nm.SetDragIndex(off)
		}
	}

	for _, n := range spec.Nodes {
		if nm, ok := md.MR.NodeGeoms()[n.ID]; ok {
			loadspec.SeedNode(nm, n, scenePath)
		}
	}

	md.UI.TiltRows, md.UI.TiltLabels = spec.TiltPanelRows()

	md.UI.RuleNodes = spec.RulePanelNodes(func(id string) bool {
		ng, ok := md.MR.NodeGeoms()[id]
		return ok && ng.HasKindRule()
	})
	md.UI.RuleSharedRow = -1

	kindByID := make(map[string]string, len(spec.Nodes))
	for _, n := range spec.Nodes {
		kindByID[n.ID] = n.Type
	}

	targetByLabel := make(map[string]string, len(spec.Edges))
	sourceByLabel := make(map[string]string, len(spec.Edges))
	for _, e := range spec.Edges {
		if src, ok := md.MR.NodeGeoms()[e.Source]; ok {
			loadspec.SeedEdge(src, e, true, kindByID[e.Target], scenePath)
		}
		if dst, ok := md.MR.NodeGeoms()[e.Target]; ok {
			loadspec.SeedEdge(dst, e, false, kindByID[e.Source], scenePath)
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
		for to, told := range nm.RequestedDrag(polarindex.Offset{}) {
			if other, ok := md.MR.NodeGeoms()[to]; ok {
				d := told
				other.SendExternal(context.TODO(), owners.Msg{NodeID: to,
					Body: owners.Drag{Delta: &d}})
			}
		}
	}
	return md, nil
}
