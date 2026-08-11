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

// Deps bundles the sub-objects HandleRawInput and everything it calls need to read/write.
// Every field is an already-exported MoveDispatch sub-object (MR/UI/LQ/RT) plus the one
// context.Context the caller's own goroutine owns — not a new shared type, just the
// explicit-parameter shape §30 used for applyUpdate's table handlers.
type Deps struct {
	MR  *moverreg.MoverRegistry
	UI  *viewstate.UIState
	LQ  *layoutquant.LayoutQuantizer
	RT  *rowtables.RowTables
	Ctx context.Context
}

// HandleRawInput is the FSM entry point: one raw pointer/wheel event → gesture state update
// and (possibly) a camera or topology change. Called (via dispatch.MoveDispatch.HandleRawInput)
// by the stdin reader for a type=="raw-input" message. slotReg resolves an edge's destination
// slot; tr emits camera events + breadcrumbs. Fire-and-forget: nothing here triggers delivery.
func HandleRawInput(d Deps, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry, tr *T.Trace) {
	g := &d.UI.Gest
	g.Fov = ev.Fov
	g.Rect = gesturefsm.GestureRect{Left: ev.RectLeft, Top: ev.RectTop, Width: ev.RectWidth, Height: ev.RectHeight}
	if h := rawInputHandlers[ev.Kind]; h != nil {
		h(d, ev, slotReg, tr)
	}
}

// rawInputHandlers is the flat dispatch table for HandleRawInput: raw-input kind →
// handler. An unknown kind is a no-op, matching the switch's absent default.
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
