package Dispatch

import (
	"sort"
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Node"
	edge "github.com/dtauraso/beadnetwork/Categories/Node/Edge"
	"github.com/dtauraso/beadnetwork/Categories/Node/Edge/edgegeom"
)

func edgeEnds(geoms map[string]Node.NodeGeom) map[string]edgegeom.NodeEnd {
	out := make(map[string]edgegeom.NodeEnd, len(geoms))
	for id, g := range geoms {
		out[id] = edgegeom.NodeEnd{
			Center: edgegeom.Vec3(Node.NodeWorldPos(g)),
			OuterR: Node.NodeTorusOuterR(g.Kind),
		}
	}
	return out
}

func resolveSeedOrders(geoms map[string]Node.NodeGeom, edgeEndpoints map[string]edge.EdgeEndpoints, nodeOrder, edgeOrder []string) ([]string, []string) {
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

func (md *MoveDispatch) buildGeomSeeds(geoms map[string]Node.NodeGeom, edgeEndpoints map[string]edge.EdgeEndpoints, nodeOrder, edgeOrder []string) error {
	md.GS.NodeSeeds = make([]Node.Seed, 0, len(nodeOrder))
	for i, id := range nodeOrder {
		g, ok := geoms[id]
		if !ok {
			continue
		}

		row := i
		if n, err := strconv.Atoi(id); err == nil {
			row = n - 1
		}
		md.GS.NodeSeeds = append(md.GS.NodeSeeds, Node.NewSeed(id, g, row))
	}
	md.GS.EdgeSeeds = make([]edgegeom.Seed, 0, len(edgeOrder))
	for _, label := range edgeOrder {
		ep, ok := edgeEndpoints[label]
		if !ok {
			continue
		}

		seed, err := edgegeom.NewSeed(label, ep.Source, ep.Target, edgeEnds(geoms))
		if err != nil {
			return err
		}
		md.GS.EdgeSeeds = append(md.GS.EdgeSeeds, seed)
	}
	return nil
}

func (md *MoveDispatch) buildRowTables(rowCount int) {
	rtNodeSeeds := make([]NodeSeed, len(md.GS.NodeSeeds))
	for i, sd := range md.GS.NodeSeeds {
		rtNodeSeeds[i] = NodeSeed{ID: sd.ID, Row: sd.Row}
	}
	rtEdgeSeeds := make([]EdgeSeed, len(md.GS.EdgeSeeds))
	for i, sd := range md.GS.EdgeSeeds {
		rtEdgeSeeds[i] = EdgeSeed{Label: sd.Label, SrcNode: sd.SrcNode, DstNode: sd.DstNode}
	}
	md.RT.Build(rtNodeSeeds, rtEdgeSeeds, rowCount)
}
