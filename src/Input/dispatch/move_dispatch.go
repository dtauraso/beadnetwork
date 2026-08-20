package dispatch

import (
	geomseeds "github.com/dtauraso/wirefold/src/Node/Wiring/geomseeds"
	"github.com/dtauraso/wirefold/src/Node/Wiring/moverreg"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodeinbox"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodemove"
	rowtables "github.com/dtauraso/wirefold/src/Node/Wiring/rowtables"
	"github.com/dtauraso/wirefold/src/Node/Wiring/rulechans"
	sceneswitch "github.com/dtauraso/wirefold/src/Scene/sceneswitch"
	"github.com/dtauraso/wirefold/src/Node/Wiring/streamwire"
	"github.com/dtauraso/wirefold/src/Node/Wiring/viewpersist"
	"github.com/dtauraso/wirefold/src/Node/Wiring/viewstate"
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
