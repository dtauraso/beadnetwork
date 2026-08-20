package runtopology

import (
	"context"
	"sync"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"

	"github.com/dtauraso/wirefold/src/Node/Wiring/dispatch"
	"github.com/dtauraso/wirefold/src/Node/Wiring/inputcodec"
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

func startGestureActor(ctx context.Context, slotReg inputcodec.SlotRegistry, md *dispatch.MoveDispatch, speedSinks SliderPanel.Sinks) (chan gestureInboxMsg, *sync.WaitGroup) {
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
					dispatch.ApplyEdit(ctx, gm.msg, md, speedSinks)
				case gestureMsgRawInput:
					dispatch.HandleRawInputMsg(ctx, gm.msg, slotReg, md, speedSinks)
				case gestureMsgSave:
					dispatch.HandleSaveMsg(md)
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
