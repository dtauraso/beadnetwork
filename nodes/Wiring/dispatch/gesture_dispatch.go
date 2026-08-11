package dispatch

import (
	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/gesture"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
)

// gesture_dispatch.go — the thin delegator into package gesture (nodes/Wiring/gesture),
// which owns the FSM's entry point, dispatch table, phase handlers, hit classification, and
// leaf actions (moved out of this package in §31,
// docs/planning/movedispatch-decomposition.md). This is now the ONLY gesture-shaped file
// left in package dispatch: it bundles MoveDispatch's already-exported sub-objects (MR/UI/
// LQ/RT) plus md.ctx (which stays unexported here — threaded through as an explicit
// gesture.Deps.Ctx value, not exported as a field) into a gesture.Deps and forwards.

// HandleRawInput is the FSM entry point: one raw pointer/wheel event → gesture state update
// and (possibly) a camera or topology change. Called by the stdin reader for a
// type=="raw-input" message. slotReg resolves an edge's destination slot; tr emits camera
// events + breadcrumbs. Fire-and-forget: nothing here triggers delivery.
func (md *MoveDispatch) HandleRawInput(ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry, tr *T.Trace) {
	gesture.HandleRawInput(gesture.Deps{MR: &md.MR, UI: &md.UI, LQ: &md.LQ, RT: &md.RT, Ctx: md.ctx}, ev, slotReg, tr)
}
