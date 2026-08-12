package dispatch

import (
	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/moverreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

func NewMoveDispatch(geoms map[string]nodegeom.NodeGeom, edgeEndpoints map[string]inputcodec.EdgeEndpoints, tr *T.Trace, nodeOrder, edgeOrder []string, clk clock.Clock, speedSinks *[]chan float64, rowCount int) (*MoveDispatch, error) {
	nodeOrder, edgeOrder = resolveSeedOrders(geoms, edgeEndpoints, nodeOrder, edgeOrder)

	md := &MoveDispatch{
		TR: tr,
	}
	md.MR = moverreg.New()
	initMoveDispatchUIDefaults(md)

	if err := md.buildGeomSeeds(geoms, edgeEndpoints, nodeOrder, edgeOrder); err != nil {
		return nil, err
	}
	md.buildNodeMovers(geoms, tr, clk)
	md.wireMutualPairs(edgeEndpoints)
	md.buildEdgeMovers(edgeEndpoints, geoms, tr, clk, speedSinks)
	md.seedPartnerCenters()
	md.wireNodeEdgeIDs()
	md.buildRowTables(rowCount)
	md.bindUIClosures()

	return md, nil
}
