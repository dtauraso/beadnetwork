package Gesture

import (
	"context"

	"github.com/dtauraso/wirefold/src/Input/Codec"
	"github.com/dtauraso/wirefold/src/Input/Drag"
	"github.com/dtauraso/wirefold/src/Node/moverreg"
	"github.com/dtauraso/wirefold/src/Node/nodemove"
	"github.com/dtauraso/wirefold/src/Scene/rowtables"
	"github.com/dtauraso/wirefold/src/Scene/viewstate"
)

type Deps struct {
	MR    *moverreg.MoverRegistry
	UI    *viewstate.UIState
	Mover *nodemove.NodeMover
	RT    *rowtables.RowTables
	Ctx   context.Context
}

func HandleRawInput(d Deps, ev Codec.RawInputMsg, slotReg Codec.SlotRegistry) {
	g := &d.UI.Gest
	g.Rect = Drag.GestureRect{Left: ev.RectLeft, Top: ev.RectTop, Width: ev.RectWidth, Height: ev.RectHeight}
	g.Fov = d.UI.FovDeg()
	if h := rawInputHandlers[ev.Kind]; h != nil {
		h(d, ev, slotReg)
	}
}

var rawInputHandlers = map[string]func(d Deps, ev Codec.RawInputMsg, slotReg Codec.SlotRegistry){
	"pointerdown": func(d Deps, ev Codec.RawInputMsg, slotReg Codec.SlotRegistry) {
		gestPointerDown(d, ev)
	},
	"pointermove": func(d Deps, ev Codec.RawInputMsg, slotReg Codec.SlotRegistry) {
		updateHover(d, ev)
		gestPointerMove(d, ev)
	},
	"pointerup": func(d Deps, ev Codec.RawInputMsg, slotReg Codec.SlotRegistry) {
		gestPointerUp(d, ev)
	},
	"wheel": func(d Deps, ev Codec.RawInputMsg, slotReg Codec.SlotRegistry) {
		gestWheel(d, ev)
	},
	"home": func(d Deps, ev Codec.RawInputMsg, slotReg Codec.SlotRegistry) {
		gestHome(d, ev)
	},
}
