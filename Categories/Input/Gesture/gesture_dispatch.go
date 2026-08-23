package Gesture

import (
	"context"

	"github.com/dtauraso/wirefold/Categories/Input/Drag"
	"github.com/dtauraso/wirefold/Categories/Node/nodeactor"
	"github.com/dtauraso/wirefold/Categories/Node/nodeactor/owners"
	"github.com/dtauraso/wirefold/Categories/Node/nodemove"
	"github.com/dtauraso/wirefold/Categories/Scene/rowtables"
	"github.com/dtauraso/wirefold/Categories/Scene/viewstate"
)

// The four things a gesture needs from the movers, passed in rather than
// reached for: a gesture picks a node, asks where it is and how big it is, and
// sends it a move.
type Movers struct {
	NodeGeoms  func() map[string]*nodeactor.NodeGeometry
	CenterOf   func(id string) (Vec3, bool)
	BodyRadius func(id string) float64
	SendMove   func(ctx context.Context, id string, msg owners.Msg)
}

type Deps struct {
	MR    Movers
	UI    *viewstate.UIState
	Mover *nodemove.NodeMover
	RT    *rowtables.RowTables
	Ctx   context.Context
}

func HandleRawInput(d Deps, ev Drag.RawInputMsg) {
	g := &d.UI.Gest
	g.Rect = Drag.GestureRect{Left: ev.RectLeft, Top: ev.RectTop, Width: ev.RectWidth, Height: ev.RectHeight}
	g.Fov = d.UI.FovDeg()
	if h := rawInputHandlers[ev.Kind]; h != nil {
		h(d, ev)
	}
}

var rawInputHandlers = map[string]func(d Deps, ev Drag.RawInputMsg){
	"pointerdown": func(d Deps, ev Drag.RawInputMsg) {
		gestPointerDown(d, ev)
	},
	"pointermove": func(d Deps, ev Drag.RawInputMsg) {
		updateHover(d, ev)
		gestPointerMove(d, ev)
	},
	"pointerup": func(d Deps, ev Drag.RawInputMsg) {
		gestPointerUp(d, ev)
	},
	"wheel": func(d Deps, ev Drag.RawInputMsg) {
		gestWheel(d, ev)
	},
	"home": func(d Deps, ev Drag.RawInputMsg) {
		gestHome(d, ev)
	},
}
