package dispatch

import (
	"github.com/dtauraso/wirefold/src/Node/moverreg"
	"github.com/dtauraso/wirefold/src/Node/nodeinbox"
	"github.com/dtauraso/wirefold/src/Node/nodemove"
	"github.com/dtauraso/wirefold/src/Node/rulechans"
	rowtables "github.com/dtauraso/wirefold/src/Scene/rowtables"
	sceneswitch "github.com/dtauraso/wirefold/src/Scene/sceneswitch"
	"github.com/dtauraso/wirefold/src/Scene/viewpersist"
	"github.com/dtauraso/wirefold/src/Scene/viewstate"
	geomseeds "github.com/dtauraso/wirefold/src/runtopology/geomseeds"
	"github.com/dtauraso/wirefold/src/runtopology/streamwire"
)

type MoveDispatch struct {
	MR moverreg.MoverRegistry

	GS geomseeds.GeomSeeds

	Persist viewpersist.Persisters

	Sw streamwire.StreamWiring

	UI viewstate.UIState

	Mover nodemove.NodeMover

	Scenes sceneswitch.SceneSwitch

	RT rowtables.RowTables

	Inboxes nodeinbox.NodeInboxes

	Rules rulechans.RuleChannels
}
