package runtopology

import (
	"context"
	"sync"

	"github.com/dtauraso/wirefold/src/Chrome/SliderPanel"

	W "github.com/dtauraso/wirefold/src/Node/Wiring/dispatch"
	"github.com/dtauraso/wirefold/src/Node/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/src/Node/Wiring/stdinreader"
	T "github.com/dtauraso/wirefold/src/Trace"
)

type gestureMsgKind int

const (
	gestureMsgEdit gestureMsgKind = iota
	gestureMsgRawInput
	gestureMsgSave
)

type gestureInboxMsg struct {
	kind gestureMsgKind
	msg  inputcodec.StdinMsg
}

const gestureInboxDepth = 64

func startGestureActor(ctx context.Context, slotReg inputcodec.SlotRegistry, md *W.MoveDispatch, tr *T.Trace, speedSinks SliderPanel.Sinks) (chan gestureInboxMsg, *sync.WaitGroup) {
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
					stdinreader.ApplyEdit(ctx, gm.msg, md, tr, speedSinks)
				case gestureMsgRawInput:
					stdinreader.HandleRawInputMsg(ctx, gm.msg, slotReg, md, tr, speedSinks)
				case gestureMsgSave:
					stdinreader.HandleSaveMsg(md)
				}
			}
		}
	}()
	return inbox, wg
}

func sendGestureMsgBlocking(ctx context.Context, inbox chan gestureInboxMsg, gm gestureInboxMsg) {
	select {
	case inbox <- gm:
	case <-ctx.Done():
	}
}
