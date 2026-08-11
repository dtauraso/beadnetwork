package Wiring

import (
	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/gesturefsm"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
)

// gesture_dispatch.go — the FSM's ENTRY POINT and its raw-input-kind dispatch table. The
// FSM's owned state/types live in gesture.go; the per-event phase handlers this table calls
// into live in gesture_handlers.go.

// HandleRawInput is the FSM entry point: one raw pointer/wheel event → gesture state update
// and (possibly) a camera or topology change. Called by the stdin reader for a
// type=="raw-input" message. slotReg resolves an edge's destination slot; tr emits camera
// events + breadcrumbs. Fire-and-forget: nothing here triggers delivery.
func (md *MoveDispatch) HandleRawInput(ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry, tr *T.Trace) {
	g := &md.UI.Gest
	g.Fov = ev.Fov
	g.Rect = gesturefsm.GestureRect{Left: ev.RectLeft, Top: ev.RectTop, Width: ev.RectWidth, Height: ev.RectHeight}
	if h := rawInputHandlers[ev.Kind]; h != nil {
		h(md, ev, slotReg, tr)
	}
}

// rawInputHandlers is the flat dispatch table for HandleRawInput: raw-input kind →
// handler. An unknown kind is a no-op, matching the switch's absent default.
var rawInputHandlers = map[string]func(md *MoveDispatch, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry, tr *T.Trace){
	"pointerdown": func(md *MoveDispatch, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry, tr *T.Trace) {
		md.gestPointerDown(ev, tr)
	},
	"pointermove": func(md *MoveDispatch, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry, tr *T.Trace) {
		md.updateHover(ev, tr)
		md.gestPointerMove(ev, tr)
	},
	"pointerup": func(md *MoveDispatch, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry, tr *T.Trace) {
		md.gestPointerUp(ev, slotReg, tr)
	},
	"wheel": func(md *MoveDispatch, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry, tr *T.Trace) {
		md.gestWheel(ev, tr)
	},
	"home": func(md *MoveDispatch, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry, tr *T.Trace) {
		md.gestHome(ev, tr)
	},
}
