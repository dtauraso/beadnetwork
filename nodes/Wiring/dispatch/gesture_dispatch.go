package dispatch

import (
	"context"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/gesture"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
)

// gesture_dispatch.go — the thin delegator into package gesture (nodes/Wiring/gesture),
// which owns the FSM's entry point, dispatch table, phase handlers, hit classification, and
// leaf actions (moved out of this package in §31,
// docs/planning/movedispatch-decomposition.md). This is now the ONLY gesture-shaped file
// left in package dispatch: it bundles MoveDispatch's already-exported sub-objects (MR/UI/
// LQ/RT) plus an explicit ctx parameter (per Go's own guidance not to store a Context on a
// struct — §35, docs/planning/movedispatch-decomposition.md) into a gesture.Deps and
// forwards.

// HandleRawInput is the FSM entry point: one raw pointer/wheel event → gesture state update
// and (possibly) a camera or topology change. Called by the stdin reader for a
// type=="raw-input" message. ctx comes from the caller (runtopology's gesture actor
// goroutine already has one in scope, matching ApplyEdit's own shape). slotReg resolves an
// edge's destination slot; tr emits camera events + breadcrumbs. Fire-and-forget: nothing
// here triggers delivery.
func (md *MoveDispatch) HandleRawInput(ctx context.Context, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry, tr *T.Trace) {
	gesture.HandleRawInput(gesture.Deps{MR: &md.MR, UI: &md.UI, LQ: &md.LQ, RT: &md.RT, Ctx: ctx}, ev, slotReg, tr)
}
