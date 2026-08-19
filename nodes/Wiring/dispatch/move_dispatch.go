package dispatch

import (
	geomseeds "github.com/dtauraso/wirefold/nodes/Wiring/geomseeds"
	"github.com/dtauraso/wirefold/nodes/Wiring/moverreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeinbox"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodemove"
	rowtables "github.com/dtauraso/wirefold/nodes/Wiring/rowtables"
	"github.com/dtauraso/wirefold/nodes/Wiring/rulechans"
	sceneswitch "github.com/dtauraso/wirefold/nodes/Wiring/sceneswitch"
	"github.com/dtauraso/wirefold/nodes/Wiring/streamwire"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"

	T "github.com/dtauraso/wirefold/tools/topology-vscode/src/Trace"
)

type MoveDispatch struct {
	MR moverreg.MoverRegistry

	GS geomseeds.GeomSeeds

	TR *T.Trace

	Persist viewpersist.Persisters

	Sw streamwire.StreamWiring

	UI viewstate.UIState

	Mover nodemove.NodeMover

	Scenes sceneswitch.SceneSwitch

	RT rowtables.RowTables

	Inboxes nodeinbox.NodeInboxes

	Rules rulechans.RuleChannels
}
