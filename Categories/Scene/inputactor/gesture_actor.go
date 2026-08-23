package inputactor

import (
	"context"
	"sync"

	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/SliderPanel"

	clock "github.com/dtauraso/wirefold/Categories/Clock"
	"github.com/dtauraso/wirefold/Categories/Input/Drag"
	"github.com/dtauraso/wirefold/Categories/Scene/scenerun"
)

type gestureMsgKind int

const (
	GestureMsgEdit gestureMsgKind = iota
	GestureMsgSave
)

type GestureInboxMsg struct {
	Kind gestureMsgKind

	Op      string
	Entity  byte
	Attr    byte
	Payload []byte
}

const gestureInboxDepth = 64

func StartGestureActor(ctx context.Context, md *scenerun.MoveDispatch, speedSinks SliderPanel.Sinks, clk clock.Clock, inputPath string) (chan GestureInboxMsg, *sync.WaitGroup) {
	inbox := make(chan GestureInboxMsg, gestureInboxDepth)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		reader := NewReader(inputPath)
		mine := clk.Copy()
		wheel := &wheelTotals{}
		for {
			for _, raw := range reader.ReadAll() {
				if ev, ok := Drag.DecodeRawInput(raw); ok {
					wheel.difference(&ev)
					scenerun.HandleRawInputMsg(ctx, ev, md, speedSinks)
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
						scenerun.ApplyEdit(ctx, gm.Op, gm.Entity, gm.Attr, gm.Payload, md, speedSinks)
					case GestureMsgSave:
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

func SendGestureMsgBlocking(ctx context.Context, inbox chan GestureInboxMsg, gm GestureInboxMsg) {
	select {
	case inbox <- gm:
	case <-ctx.Done():
	}
}
