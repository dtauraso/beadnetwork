package scenerun

import (
	"context"
	"sync"

	"github.com/dtauraso/wirefold/Chrome/Panels/SliderPanel"
	beadanimation "github.com/dtauraso/wirefold/Node/BeadAnimation"

	clock "github.com/dtauraso/wirefold/Clock"
	"github.com/dtauraso/wirefold/Input/Drag"
	"github.com/dtauraso/wirefold/Input/File"
	"github.com/dtauraso/wirefold/Input/Stdin"
)

type gestureMsgKind int

const (
	GestureMsgEdit gestureMsgKind = iota
	GestureMsgSave
)

type GestureInboxMsg struct {
	Kind gestureMsgKind
	Msg  Stdin.StdinMsg
}

const gestureInboxDepth = 64

func StartGestureActor(ctx context.Context, slotReg beadanimation.SlotRegistry, md *MoveDispatch, speedSinks SliderPanel.Sinks, clk clock.Clock, inputPath string) (chan GestureInboxMsg, *sync.WaitGroup) {
	inbox := make(chan GestureInboxMsg, gestureInboxDepth)
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
					HandleRawInputMsg(ctx, ev, slotReg, md, speedSinks)
				}
			}

		drain:
			for {
				select {
				case <-ctx.Done():
					return
				case gm := <-inbox:
					switch gm.Kind {
					case GestureMsgEdit:
						ApplyEdit(ctx, gm.Msg, md, speedSinks)
					case GestureMsgSave:
						HandleSaveMsg(md)
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

func SendGestureMsgBlocking(ctx context.Context, inbox chan GestureInboxMsg, gm GestureInboxMsg) {
	select {
	case inbox <- gm:
	case <-ctx.Done():
	}
}
