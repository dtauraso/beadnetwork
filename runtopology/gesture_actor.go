package runtopology

import (
	"context"
	"sync"

	T "github.com/dtauraso/wirefold/Trace"
	W "github.com/dtauraso/wirefold/nodes/Wiring"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
)

// gesture_actor.go — STEP 1 of docs/planning/gesture-actor.md: cuts the stdin-reader ->
// FSM CYCLE with a channel, giving the gesture FSM its own inbox and its own goroutine.
// This is deliberately additive and behaviour-preserving: the three ops the stdin reader
// used to call inline (ApplyEdit / HandleRawInputMsg / HandleSaveMsg) still run in the
// same order, on the same decoded messages, through the exact same nodes/Wiring entry
// points — only the goroutine they run on changes. VIEW-stream ownership does NOT move
// here (that is step 2); emitViewFrame call sites are untouched by this file.

// gestureMsgKind discriminates the three stdinreader.Handlers operations, now delivered
// as channel sends into the gesture actor's own inbox rather than direct calls.
type gestureMsgKind int

const (
	gestureMsgEdit gestureMsgKind = iota
	gestureMsgRawInput
	gestureMsgSave
)

// gestureInboxMsg is ONE inbox message: raw-input | edit | save, mirroring the
// discriminated union RunStdinReader already decodes (stdin_reader.go's MSG_TYPES_DOC).
// msg is unused for gestureMsgSave — the bare command carries no payload.
type gestureInboxMsg struct {
	kind gestureMsgKind
	msg  inputcodec.StdinMsg
}

// gestureInboxDepth is the declared capacity of the gesture actor's inbox — sized the
// same spirit as moverInboxDepth (mover_registry.go): a queue for a burst between two
// drain passes, not a derived value. The realistic load is raw-input flooding at
// pointer-move/frame rate (~60-120Hz) between two FSM cycles; this is generous headroom
// for that burst.
//
// Unlike a mover's inbox — fed by MULTIPLE peer goroutines, so a bounded non-blocking
// send with a sender-side retry queue is used to keep one slow peer from starving
// another (flushPending) — this inbox has exactly ONE producer (the stdin reader's
// single dispatch goroutine) and ONE consumer (this actor), so there is no fairness
// concern and no cycle: the gesture actor never blocks waiting on the stdin reader, only
// the reverse. A full inbox therefore just means the reader is momentarily ahead of the
// actor; blocking the reader there is plain backpressure, not a deadlock risk. Dropping
// an "edit" or "save" message under load would be a real behaviour change (a discarded
// click or save), which step 1 must not introduce, so a plain buffered channel with a
// ctx.Done()-guarded BLOCKING send is used (see sendGestureMsgBlocking) instead of a
// lossy non-blocking send — matching moverRegistry.sendMove's own external-entry shape,
// not moverRegistry.enqueueFor's peer-to-peer one.
const gestureInboxDepth = 64

// startGestureActor launches the gesture actor's own goroutine: the single receiver of
// gestureInboxMsg values, dispatching each to the exact same nodes/Wiring entry points
// RunStdinReader used to call inline. Returns the inbox channel (for startStdinReader's
// Handlers to send into) and a *sync.WaitGroup covering this one goroutine so joinAll can
// wait on it like every other tracked goroutine. Selects on ctx.Done() at the top of its
// loop — its only blocking call — so cancel-to-return is immediate, the same shape every
// other actor in this package uses (nodeMover.run, edgeMover.run).
func startGestureActor(ctx context.Context, slotReg inputcodec.SlotRegistry, md *W.MoveDispatch, tr *T.Trace, speedSinks []chan float64) (chan gestureInboxMsg, *sync.WaitGroup) {
	inbox := make(chan gestureInboxMsg, gestureInboxDepth)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case gm := <-inbox:
				switch gm.kind {
				case gestureMsgEdit:
					W.ApplyEdit(gm.msg, md, tr, speedSinks)
				case gestureMsgRawInput:
					W.HandleRawInputMsg(gm.msg, slotReg, md, tr)
				case gestureMsgSave:
					W.HandleSaveMsg(md)
				}
			}
		}
	}()
	return inbox, wg
}

// sendGestureMsgBlocking delivers one message to the gesture actor's inbox from the
// stdin reader's own dispatch goroutine. It blocks only under real backpressure (inbox
// full) and always carries a ctx.Done() escape hatch, so a full inbox during shutdown
// can never park the reader forever — the same shape as moverRegistry.sendMove's
// external-entry send.
func sendGestureMsgBlocking(ctx context.Context, inbox chan gestureInboxMsg, gm gestureInboxMsg) {
	select {
	case inbox <- gm:
	case <-ctx.Done():
	}
}
