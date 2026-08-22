package scenerun

import (
	"github.com/dtauraso/wirefold/Categories/Node/moverreg"
	"github.com/dtauraso/wirefold/Categories/Node/nodeactor"
	"github.com/dtauraso/wirefold/Categories/Node/nodeinbox"
	"github.com/dtauraso/wirefold/Categories/Node/nodemove"
	"github.com/dtauraso/wirefold/Categories/Node/rulechans"
	SceneB "github.com/dtauraso/wirefold/Categories/Scene"
	rowtables "github.com/dtauraso/wirefold/Categories/Scene/rowtables"
	sceneswitch "github.com/dtauraso/wirefold/Categories/Scene/sceneswitch"
	"github.com/dtauraso/wirefold/Categories/Scene/viewpersist"
	"github.com/dtauraso/wirefold/Categories/Scene/viewstate"
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
