package rowtables

import "github.com/dtauraso/wirefold/src/Input/Codec"

type EdgeEndpoint struct {
	SrcNode, DstNode string
}

type NodeSeed struct {
	ID  string
	Row int
}

type EdgeSeed struct {
	Label            string
	SrcNode, DstNode string
}

type RowTables struct {
	NodeRowTable []string

	EdgeRowTable []string

	EdgeEndpointRowTable []EdgeEndpoint
}

func (rt *RowTables) Build(nodeSeeds []NodeSeed, edgeSeeds []EdgeSeed, rowCount int) {
	if rowCount == 0 {
		rowCount = len(nodeSeeds)
	}
	rt.NodeRowTable = make([]string, rowCount)
	for _, sd := range nodeSeeds {
		if sd.Row < 0 || sd.Row >= rowCount {
			continue
		}
		rt.NodeRowTable[sd.Row] = sd.ID
	}
	rt.EdgeRowTable = make([]string, len(edgeSeeds))
	rt.EdgeEndpointRowTable = make([]EdgeEndpoint, len(edgeSeeds))
	for i, sd := range edgeSeeds {
		rt.EdgeRowTable[i] = sd.Label
		rt.EdgeEndpointRowTable[i] = EdgeEndpoint{SrcNode: sd.SrcNode, DstNode: sd.DstNode}
	}
}

func (rt *RowTables) LookupNodeRow(row int) (nodeID string, ok bool) {
	if row < 0 || row >= len(rt.NodeRowTable) || rt.NodeRowTable[row] == "" {
		return "", false
	}
	return rt.NodeRowTable[row], true
}

func (rt *RowTables) NodeRowFor(nodeID string) (int32, bool) {
	for i, id := range rt.NodeRowTable {
		if id == nodeID {
			return int32(i), true
		}
	}
	return -1, false
}

func (rt *RowTables) LookupEdgeRow(row int) (label string, ok bool) {
	if row < 0 || row >= len(rt.EdgeRowTable) {
		return "", false
	}
	return rt.EdgeRowTable[row], true
}

func (rt *RowTables) EdgeRowForPair(a, b string) (int32, bool) {
	for i, e := range rt.EdgeEndpointRowTable {
		if (e.SrcNode == a && e.DstNode == b) || (e.SrcNode == b && e.DstNode == a) {
			return int32(i), true
		}
	}
	return -1, false
}

func (rt *RowTables) NodeFromHit(h Codec.RawHit) (node string, ok bool) {
	if h.NodeRow >= 0 {
		return rt.LookupNodeRow(h.NodeRow)
	}
	return "", false
}

func (rt *RowTables) EdgeFromHit(h Codec.RawHit) (label string, ok bool) {
	if h.EdgeRow >= 0 {
		return rt.LookupEdgeRow(h.EdgeRow)
	}
	return "", false
}
