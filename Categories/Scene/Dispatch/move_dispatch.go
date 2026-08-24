package Dispatch

import (
	"context"
	"sync"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/AngleDropdown"
	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	"github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Node/ChannelVectors"
	edge "github.com/dtauraso/beadnetwork/Categories/Node/Edge"
	"github.com/dtauraso/beadnetwork/Categories/Node/TiltVectors"
	Flags "github.com/dtauraso/beadnetwork/Categories/Scene/View/Flags"
	"github.com/dtauraso/beadnetwork/Categories/Vector/polarindex"
	SceneB "github.com/dtauraso/beadnetwork/Categories/Scene"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"
	"github.com/dtauraso/beadnetwork/Categories/Scene/View"
)

type MoveDispatch struct {
	MR Movers

	GS SceneB.GeomSeeds

	Sw StreamWiring

	UI View.UIState

	Mover Node.NodeMover

	Scenes Scenes.SceneSwitch

	RT RowTables

	Inboxes TiltVectors.TiltEditInboxes

	ChannelVectorsOn ChannelVectors.OnSwitch

	Rules Node.RuleChannels
}

func (md *MoveDispatch) Start(ctx context.Context) *sync.WaitGroup {
	return md.MR.Start(ctx)
}

func initMoveDispatchUIDefaults(md *MoveDispatch) {
	md.UI.OV = Flags.DefaultOverlayState()
	md.UI.PN = Panel.DefaultPanelState()
	md.UI.Speed = 1
	md.UI.ClockDivisor = 1
	md.UI.LatticePoints = AngleDropdown.DefaultLatticePoints
}

func (md *MoveDispatch) bindUIClosures() {
	md.UI.NodeRowFor = md.RT.NodeRowFor
}

func NewMoveDispatch(geoms map[string]Node.NodeGeom, edgeEndpoints map[string]edge.EdgeEndpoints, nodeOrder, edgeOrder []string, clk clock.Clock, speedSinks *SliderPanel.Sinks, rowCount int, constants polarindex.SceneConstants) (*MoveDispatch, error) {
	nodeOrder, edgeOrder = resolveSeedOrders(geoms, edgeEndpoints, nodeOrder, edgeOrder)

	md := &MoveDispatch{}
	md.MR = NewMovers()
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
