package scenerun

import (
	"context"
	"sync"

	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Pills/AngleDropdown"
	clock "github.com/dtauraso/wirefold/Categories/Clock"
	"github.com/dtauraso/wirefold/Categories/Node"
	"github.com/dtauraso/wirefold/Categories/Node/ChannelVectors"
	edge "github.com/dtauraso/wirefold/Categories/Node/Edge"
	"github.com/dtauraso/wirefold/Categories/Node/TiltVectors"
	"github.com/dtauraso/wirefold/Categories/Overlay"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
	SceneB "github.com/dtauraso/wirefold/Categories/Scene"
	sceneswitch "github.com/dtauraso/wirefold/Categories/Scene/sceneswitch"
	"github.com/dtauraso/wirefold/Categories/Scene/viewpersist"
	"github.com/dtauraso/wirefold/Categories/Scene/viewstate"
)

type MoveDispatch struct {
	MR Movers

	GS SceneB.GeomSeeds

	Persist viewpersist.Persisters

	Sw StreamWiring

	UI viewstate.UIState

	Mover Node.NodeMover

	Scenes sceneswitch.SceneSwitch

	RT RowTables

	Inboxes TiltVectors.TiltEditInboxes

	ChannelVectorsOn ChannelVectors.OnSwitch

	Rules Node.RuleChannels
}

func (md *MoveDispatch) Start(ctx context.Context) *sync.WaitGroup {
	return md.MR.Start(ctx)
}

func initMoveDispatchUIDefaults(md *MoveDispatch) {
	md.UI.OV = Overlay.DefaultOverlayState()
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
