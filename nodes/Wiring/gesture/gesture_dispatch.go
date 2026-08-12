package gesture

import (
	"context"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/gesturefsm"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/layoutquant"
	"github.com/dtauraso/wirefold/nodes/Wiring/moverreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/rowtables"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

type Deps struct {
	MR  *moverreg.MoverRegistry
	UI  *viewstate.UIState
	LQ  *layoutquant.LayoutQuantizer
	RT  *rowtables.RowTables
	Ctx context.Context
}

func HandleRawInput(d Deps, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry, tr *T.Trace) {
	g := &d.UI.Gest
	g.Fov = ev.Fov
	g.Rect = gesturefsm.GestureRect{Left: ev.RectLeft, Top: ev.RectTop, Width: ev.RectWidth, Height: ev.RectHeight}
	if h := rawInputHandlers[ev.Kind]; h != nil {
		h(d, ev, slotReg, tr)
	}
}

var rawInputHandlers = map[string]func(d Deps, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry, tr *T.Trace){
	"pointerdown": func(d Deps, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry, tr *T.Trace) {
		gestPointerDown(d, ev)
	},
	"pointermove": func(d Deps, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry, tr *T.Trace) {
		updateHover(d, ev)
		gestPointerMove(d, ev, tr)
	},
	"pointerup": func(d Deps, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry, tr *T.Trace) {
		gestPointerUp(d, ev)
	},
	"wheel": func(d Deps, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry, tr *T.Trace) {
		gestWheel(d, ev, tr)
	},
	"home": func(d Deps, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry, tr *T.Trace) {
		gestHome(d, ev, tr)
	},
}
