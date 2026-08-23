package Gesture

import (
	"context"

	"github.com/dtauraso/wirefold/Categories/Scene/Drag"
	"github.com/dtauraso/wirefold/Categories/Node"
	"github.com/dtauraso/wirefold/Categories/Scene/rowtables"
	"github.com/dtauraso/wirefold/Categories/Scene/viewstate"
)

type Movers struct {
	NodeGeoms  func() map[string]*Node.NodeGeometry
	CenterOf   func(id string) (Vec3, bool)
	BodyRadius func(id string) float64
	SendMove   func(ctx context.Context, id string, msg Node.Msg)
}

type Deps struct {
	MR    Movers
	UI    *viewstate.UIState
	Mover *Node.NodeMover
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
