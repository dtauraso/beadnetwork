package scenerun

import (
	"github.com/dtauraso/wirefold/Node/moverreg"
	"github.com/dtauraso/wirefold/Node/nodeactor"
	"github.com/dtauraso/wirefold/Node/nodeinbox"
	"github.com/dtauraso/wirefold/Node/nodemove"
	"github.com/dtauraso/wirefold/Node/rulechans"
	SceneB "github.com/dtauraso/wirefold/Scene"
	rowtables "github.com/dtauraso/wirefold/Scene/rowtables"
	sceneswitch "github.com/dtauraso/wirefold/Scene/sceneswitch"
	"github.com/dtauraso/wirefold/Scene/viewpersist"
	"github.com/dtauraso/wirefold/Scene/viewstate"
)

type MoveDispatch struct {
	MR moverreg.MoverRegistry

	GS SceneB.GeomSeeds

	Persist viewpersist.Persisters

	Sw nodeactor.StreamWiring

	UI viewstate.UIState

	Mover nodemove.NodeMover

	Scenes sceneswitch.SceneSwitch

	RT rowtables.RowTables

	Inboxes nodeinbox.NodeInboxes

	Rules rulechans.RuleChannels
}
