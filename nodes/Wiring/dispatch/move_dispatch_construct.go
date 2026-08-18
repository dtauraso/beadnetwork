package dispatch

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/moverreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/polarindex"
	"github.com/dtauraso/wirefold/nodes/clock"
	"github.com/dtauraso/wirefold/tools/topology-vscode/SliderPanel"
	T "github.com/dtauraso/wirefold/tools/topology-vscode/Trace"
)

func NewMoveDispatch(geoms map[string]nodegeom.NodeGeom, edgeEndpoints map[string]inputcodec.EdgeEndpoints, tr *T.Trace, nodeOrder, edgeOrder []string, clk clock.Clock, speedSinks *SliderPanel.Sinks, rowCount int, constants polarindex.SceneConstants) (*MoveDispatch, error) {
	nodeOrder, edgeOrder = resolveSeedOrders(geoms, edgeEndpoints, nodeOrder, edgeOrder)

	md := &MoveDispatch{
		TR: tr,
	}
	md.MR = moverreg.New()
	initMoveDispatchUIDefaults(md)

	if err := md.buildGeomSeeds(geoms, edgeEndpoints, nodeOrder, edgeOrder); err != nil {
		return nil, err
	}
	md.buildNodeMovers(geoms, tr, clk, constants)
	md.wireRuleMesh()
	md.wireMutualPairs(edgeEndpoints)
	md.buildEdgeTable(edgeEndpoints)
	md.wireNodeEdgeIDs()
	md.buildRowTables(rowCount)
	md.wireRuleEditRows()
	md.bindUIClosures()

	return md, nil
}
