package Gesture

import (
	"context"

	"github.com/dtauraso/wirefold/Input/Drag"
	beadanimation "github.com/dtauraso/wirefold/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/Node/moverreg"
	"github.com/dtauraso/wirefold/Node/nodemove"
	"github.com/dtauraso/wirefold/Scene/rowtables"
	"github.com/dtauraso/wirefold/Scene/viewstate"
)

type Deps struct {
	MR    *moverreg.MoverRegistry
	UI    *viewstate.UIState
	Mover *nodemove.NodeMover
	RT    *rowtables.RowTables
	Ctx   context.Context
}

func HandleRawInput(d Deps, ev Drag.RawInputMsg, slotReg beadanimation.SlotRegistry) {
	g := &d.UI.Gest
	g.Rect = Drag.GestureRect{Left: ev.RectLeft, Top: ev.RectTop, Width: ev.RectWidth, Height: ev.RectHeight}
	g.Fov = d.UI.FovDeg()
	if h := rawInputHandlers[ev.Kind]; h != nil {
		h(d, ev, slotReg)
	}
}

var rawInputHandlers = map[string]func(d Deps, ev Drag.RawInputMsg, slotReg beadanimation.SlotRegistry){
	"pointerdown": func(d Deps, ev Drag.RawInputMsg, slotReg beadanimation.SlotRegistry) {
		gestPointerDown(d, ev)
	},
	"pointermove": func(d Deps, ev Drag.RawInputMsg, slotReg beadanimation.SlotRegistry) {
		updateHover(d, ev)
		gestPointerMove(d, ev)
	},
	"pointerup": func(d Deps, ev Drag.RawInputMsg, slotReg beadanimation.SlotRegistry) {
		gestPointerUp(d, ev)
	},
	"wheel": func(d Deps, ev Drag.RawInputMsg, slotReg beadanimation.SlotRegistry) {
		gestWheel(d, ev)
	},
	"home": func(d Deps, ev Drag.RawInputMsg, slotReg beadanimation.SlotRegistry) {
		gestHome(d, ev)
	},
}
