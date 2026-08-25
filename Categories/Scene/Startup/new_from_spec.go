package Startup

import (
	"context"
	"fmt"

	"github.com/dtauraso/beadnetwork/Categories/Scene/Topology"

	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"

	"github.com/dtauraso/beadnetwork/Categories/Scene/Dispatch"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/AngleDropdown"
	clock "github.com/dtauraso/beadnetwork/Categories/Clock"

	NodeBuf "github.com/dtauraso/beadnetwork/Categories/Node"
	edge "github.com/dtauraso/beadnetwork/Categories/Node/Edge"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

func NewFromSpec(spec Topology.TopoSpec, sphere polar.SceneSphere, hasScene bool, scenePath string, clk clock.Clock, speedSinks *SliderPanel.Sinks, nodeGeoms map[string]NodeBuf.NodeGeom, edgeEndpoints map[string]edge.EdgeEndpoints, baseIndices map[string]polarindex.Index, dragIndices map[string]polarindex.Offset) (*Dispatch.MoveDispatch, error) {

	nodeOrder := make([]string, len(spec.Nodes))
	for i, n := range spec.Nodes {
		nodeOrder[i] = n.ID
	}
	edgeOrder := make([]string, len(spec.Edges))
	for i, e := range spec.Edges {
		edgeOrder[i] = e.Label
	}
	md, err := Dispatch.NewMoveDispatch(nodeGeoms, edgeEndpoints, nodeOrder, edgeOrder, clk, speedSinks, spec.RowCount, spec.Constants)
	if err != nil {
		return nil, fmt.Errorf("Dispatch.NewFromSpec: %w", err)
	}

	md.UI.LatticePoints = AngleDropdown.LatticePointsFor(scenePath)
	if hasScene {

		md.UI.SceneSphere = sphere
	}
	md.UI.Constants = spec.Constants

	s := Scenes.For(scenePath)

	coplanarEdges := s.CoplanarEdges
	upAxis := s.UpAxis
	if coplanarEdges || upAxis {
		for _, nm := range md.MR.NodeGeoms() {
			nm.Flags().SetSceneFlags(coplanarEdges, upAxis)
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
			Topology.SeedNode(nm, n, scenePath)
		}
	}

	md.UI.Tilt.Rows, md.UI.Tilt.Labels = Topology.TiltPanelRows(spec)

	md.UI.Rules.Nodes = Topology.RulePanelNodes(spec, func(id string) bool {
		ng, ok := md.MR.NodeGeoms()[id]
		return ok && ng.HasKindRule()
	})
	md.UI.Rules.SharedRow = -1

	kindByID := make(map[string]string, len(spec.Nodes))
	for _, n := range spec.Nodes {
		kindByID[n.ID] = n.Type
	}

	targetByLabel := make(map[string]string, len(spec.Edges))
	sourceByLabel := make(map[string]string, len(spec.Edges))
	for _, e := range spec.Edges {
		if src, ok := md.MR.NodeGeoms()[e.Source]; ok {
			Topology.SeedEdge(src, e, true, kindByID[e.Target], scenePath)
		}
		if dst, ok := md.MR.NodeGeoms()[e.Target]; ok {
			Topology.SeedEdge(dst, e, false, kindByID[e.Source], scenePath)
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
				other.Msg().SendExternal(context.TODO(), NodeBuf.Msg{NodeID: to,
					Body: NodeBuf.Drag{Delta: &d}})
			}
		}
	}
	return md, nil
}
