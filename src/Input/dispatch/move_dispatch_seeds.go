package dispatch

import (
	"sort"
	"strconv"

	geomseeds "github.com/dtauraso/wirefold/src/runtopology/geomseeds"
	"github.com/dtauraso/wirefold/src/Input/inputcodec"
	"github.com/dtauraso/wirefold/src/Node/nodegeom"
	rowtables "github.com/dtauraso/wirefold/src/Scene/rowtables"
)

func resolveSeedOrders(geoms map[string]nodegeom.NodeGeom, edgeEndpoints map[string]inputcodec.EdgeEndpoints, nodeOrder, edgeOrder []string) ([]string, []string) {
	if nodeOrder == nil {
		nodeOrder = make([]string, 0, len(geoms))
		for id := range geoms {
			nodeOrder = append(nodeOrder, id)
		}
		sort.Strings(nodeOrder)
	}
	if edgeOrder == nil {
		edgeOrder = make([]string, 0, len(edgeEndpoints))
		for label := range edgeEndpoints {
			edgeOrder = append(edgeOrder, label)
		}
		sort.Strings(edgeOrder)
	}
	return nodeOrder, edgeOrder
}

func (md *MoveDispatch) buildGeomSeeds(geoms map[string]nodegeom.NodeGeom, edgeEndpoints map[string]inputcodec.EdgeEndpoints, nodeOrder, edgeOrder []string) error {
	md.GS.NodeSeeds = make([]geomseeds.NodeGeomSeed, 0, len(nodeOrder))
	for i, id := range nodeOrder {
		g, ok := geoms[id]
		if !ok {
			continue
		}

		row := i
		if n, err := strconv.Atoi(id); err == nil {
			row = n - 1
		}
		md.GS.NodeSeeds = append(md.GS.NodeSeeds, geomseeds.BuildNodeSeed(id, i, g, row))
	}
	md.GS.EdgeSeeds = make([]geomseeds.EdgeGeomSeed, 0, len(edgeOrder))
	for _, label := range edgeOrder {
		ep, ok := edgeEndpoints[label]
		if !ok {
			continue
		}

		seed, err := geomseeds.BuildEdgeSeed(label, ep, geoms)
		if err != nil {
			return err
		}
		md.GS.EdgeSeeds = append(md.GS.EdgeSeeds, seed)
	}
	return nil
}

func (md *MoveDispatch) buildRowTables(rowCount int) {
	rtNodeSeeds := make([]rowtables.NodeSeed, len(md.GS.NodeSeeds))
	for i, sd := range md.GS.NodeSeeds {
		rtNodeSeeds[i] = rowtables.NodeSeed{ID: sd.ID, Row: sd.Row}
	}
	rtEdgeSeeds := make([]rowtables.EdgeSeed, len(md.GS.EdgeSeeds))
	for i, sd := range md.GS.EdgeSeeds {
		rtEdgeSeeds[i] = rowtables.EdgeSeed{Label: sd.Label, SrcNode: sd.SrcNode, DstNode: sd.DstNode}
	}
	md.RT.Build(rtNodeSeeds, rtEdgeSeeds, rowCount)
}
