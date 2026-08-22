package runtopology

import (
	"context"
	"sync"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"
	beadanimation "github.com/dtauraso/wirefold/src/Node/BeadAnimation"

	clock "github.com/dtauraso/wirefold/src/Clock"
	"github.com/dtauraso/wirefold/src/runtopology/scenerun"
	"github.com/dtauraso/wirefold/src/Input/Drag"
	"github.com/dtauraso/wirefold/src/Input/File"
	"github.com/dtauraso/wirefold/src/Input/Stdin"
)

type gestureMsgKind int

const (
	gestureMsgEdit gestureMsgKind = iota
	gestureMsgSave
)

type gestureInboxMsg struct {
	kind gestureMsgKind
	msg  Stdin.StdinMsg
}

const gestureInboxDepth = 64

func startGestureActor(ctx context.Context, slotReg beadanimation.SlotRegistry, md *scenerun.MoveDispatch, speedSinks SliderPanel.Sinks, clk clock.Clock, inputPath string) (chan gestureInboxMsg, *sync.WaitGroup) {
	inbox := make(chan gestureInboxMsg, gestureInboxDepth)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		reader := File.NewReader(inputPath)
		mine := clk.Copy()
		wheel := &wheelTotals{}
		for {
			for _, raw := range reader.ReadAll() {
				if ev, ok := Drag.DecodeRawInput(raw); ok {
					wheel.difference(&ev)
					scenerun.HandleRawInputMsg(ctx, ev, slotReg, md, speedSinks)
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
						scenerun.ApplyEdit(ctx, gm.msg, md, speedSinks)
					case gestureMsgSave:
						scenerun.HandleSaveMsg(md)
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

type wheelTotals struct {
	x, y float64
	seen bool
}

func (w *wheelTotals) difference(ev *Drag.RawInputMsg) {
	if ev == nil || ev.Kind != "wheel" {
		return
	}
	totalX, totalY := ev.DeltaX, ev.DeltaY
	if !w.seen {
		w.x, w.y, w.seen = totalX, totalY, true
		ev.DeltaX, ev.DeltaY = 0, 0
		return
	}
	ev.DeltaX, ev.DeltaY = totalX-w.x, totalY-w.y
	w.x, w.y = totalX, totalY
}

func sendGestureMsgBlocking(ctx context.Context, inbox chan gestureInboxMsg, gm gestureInboxMsg) {
	select {
	case inbox <- gm:
	case <-ctx.Done():
	}
}
