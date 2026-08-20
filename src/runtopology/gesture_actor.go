package runtopology

import (
	"context"
	"sync"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"

	clock "github.com/dtauraso/wirefold/src/Clock"
	"github.com/dtauraso/wirefold/src/Input/dispatch"
	"github.com/dtauraso/wirefold/src/Input/inputcodec"
	"github.com/dtauraso/wirefold/src/Input/inputfile"
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

func startGestureActor(ctx context.Context, slotReg inputcodec.SlotRegistry, md *dispatch.MoveDispatch, speedSinks SliderPanel.Sinks, clk clock.Clock, inputPath string) (chan gestureInboxMsg, *sync.WaitGroup) {
	inbox := make(chan gestureInboxMsg, gestureInboxDepth)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		reader := inputfile.NewReader(inputPath)
		mine := clk.Copy()
		for {
			if raw, fresh := reader.Read(); fresh {
				if msg, ok := inputcodec.DecodeInputRecord(raw); ok && msg.Type == "raw-input" {
					dispatch.HandleRawInputMsg(ctx, msg, slotReg, md, speedSinks)
				}
			}

		drain:
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
				default:
					break drain
				}
			}

			if err := mine.SleepCycle(ctx); err != nil {
				return
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
