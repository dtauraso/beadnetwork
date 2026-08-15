package dispatch

import (
	geomseeds "github.com/dtauraso/wirefold/nodes/Wiring/geomseeds"
	"github.com/dtauraso/wirefold/nodes/Wiring/layoutquant"
	"github.com/dtauraso/wirefold/nodes/Wiring/moverreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeinbox"
	rowtables "github.com/dtauraso/wirefold/nodes/Wiring/rowtables"
	"github.com/dtauraso/wirefold/nodes/Wiring/rulenode"
	sceneswitch "github.com/dtauraso/wirefold/nodes/Wiring/sceneswitch"
	"github.com/dtauraso/wirefold/nodes/Wiring/streamwire"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"

	T "github.com/dtauraso/wirefold/Trace"
)

type MoveDispatch struct {
	MR moverreg.MoverRegistry

	GS geomseeds.GeomSeeds

	TR *T.Trace

	Persist viewpersist.Persisters

	Sw streamwire.StreamWiring

	UI viewstate.UIState

	LQ layoutquant.LayoutQuantizer

	Scenes sceneswitch.SceneSwitch

	RT rowtables.RowTables

	Inboxes nodeinbox.NodeInboxes

	RuleEdits []chan<- rulenode.Edit

	EdgeRuleToggles []chan<- struct{}
}
