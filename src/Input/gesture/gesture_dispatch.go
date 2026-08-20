package gesture

import (
	"context"

	"github.com/dtauraso/wirefold/src/Input/gesturefsm"
	"github.com/dtauraso/wirefold/src/Input/inputcodec"
	"github.com/dtauraso/wirefold/src/Node/Wiring/moverreg"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodemove"
	"github.com/dtauraso/wirefold/src/Node/Wiring/rowtables"
	"github.com/dtauraso/wirefold/src/Node/Wiring/viewstate"
)

type Deps struct {
	MR    *moverreg.MoverRegistry
	UI    *viewstate.UIState
	Mover *nodemove.NodeMover
	RT    *rowtables.RowTables
	Ctx   context.Context
}

func HandleRawInput(d Deps, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry) {
	g := &d.UI.Gest
	g.Rect = gesturefsm.GestureRect{Left: ev.RectLeft, Top: ev.RectTop, Width: ev.RectWidth, Height: ev.RectHeight}
	g.Fov = d.UI.FovDeg()
	if h := rawInputHandlers[ev.Kind]; h != nil {
		h(d, ev, slotReg)
	}
}

var rawInputHandlers = map[string]func(d Deps, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry){
	"pointerdown": func(d Deps, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry) {
		gestPointerDown(d, ev)
	},
	"pointermove": func(d Deps, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry) {
		updateHover(d, ev)
		gestPointerMove(d, ev)
	},
	"pointerup": func(d Deps, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry) {
		gestPointerUp(d, ev)
	},
	"wheel": func(d Deps, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry) {
		gestWheel(d, ev)
	},
	"home": func(d Deps, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry) {
		gestHome(d, ev)
	},
}
