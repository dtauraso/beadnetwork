package scenerun

import (
	"github.com/dtauraso/wirefold/Categories/Node"
	"github.com/dtauraso/wirefold/Categories/Node/ChannelVectors"
	"github.com/dtauraso/wirefold/Categories/Node/TiltVectors"
	"github.com/dtauraso/wirefold/Categories/Node/rulenode"
	SceneB "github.com/dtauraso/wirefold/Categories/Scene"
	rowtables "github.com/dtauraso/wirefold/Categories/Scene/rowtables"
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

	RT rowtables.RowTables

	Inboxes TiltVectors.TiltEditInboxes

	ChannelVectorsOn ChannelVectors.OnSwitch

	Rules rulenode.RuleChannels
}
