package dispatch

import (
	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/src/Node/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/src/Node/Wiring/moverreg"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/src/Node/Wiring/polarindex"
	"github.com/dtauraso/wirefold/src/Clock"
)

func NewMoveDispatch(geoms map[string]nodegeom.NodeGeom, edgeEndpoints map[string]inputcodec.EdgeEndpoints, nodeOrder, edgeOrder []string, clk clock.Clock, speedSinks *SliderPanel.Sinks, rowCount int, constants polarindex.SceneConstants) (*MoveDispatch, error) {
	nodeOrder, edgeOrder = resolveSeedOrders(geoms, edgeEndpoints, nodeOrder, edgeOrder)

	md := &MoveDispatch{}
	md.MR = moverreg.New()
	initMoveDispatchUIDefaults(md)

	if err := md.buildGeomSeeds(geoms, edgeEndpoints, nodeOrder, edgeOrder); err != nil {
		return nil, err
	}
	md.buildNodeMovers(geoms, clk, constants)
	md.wireRuleMesh()
	md.wireMutualPairs(edgeEndpoints)
	md.buildEdgeTable(edgeEndpoints)
	md.wireNodeEdgeIDs()
	md.buildRowTables(rowCount)
	md.wireRuleEditRows()
	md.bindUIClosures()

	return md, nil
}
