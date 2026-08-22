package scenerun

import (
	"github.com/dtauraso/wirefold/src/Node/moverreg"
	"github.com/dtauraso/wirefold/src/Node/nodeactor"
	"github.com/dtauraso/wirefold/src/Node/nodeinbox"
	"github.com/dtauraso/wirefold/src/Node/nodemove"
	"github.com/dtauraso/wirefold/src/Node/rulechans"
	SceneB "github.com/dtauraso/wirefold/src/Scene"
	rowtables "github.com/dtauraso/wirefold/src/Scene/rowtables"
	sceneswitch "github.com/dtauraso/wirefold/src/Scene/sceneswitch"
	"github.com/dtauraso/wirefold/src/Scene/viewpersist"
	"github.com/dtauraso/wirefold/src/Scene/viewstate"
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
