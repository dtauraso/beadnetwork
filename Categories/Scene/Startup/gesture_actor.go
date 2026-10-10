package Startup

import (
	"context"
	"os"
	"sync"

	"github.com/dtauraso/beadnetwork/Categories/Scene/Dispatch"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/SliderPanel"

	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Drag"
)

type gestureMsgKind int

const (
	GestureMsgEdit gestureMsgKind = iota
	GestureMsgSave
	GestureMsgCommand
)

type GestureInboxMsg struct {
	Kind gestureMsgKind

	Op      string
	Entity  byte
	Attr    byte
	Payload []byte

	Command Drag.RawInputMsg
}

const gestureInboxDepth = 64

func StartGestureActor(ctx context.Context, md *Dispatch.MoveDispatch, speedSinks SliderPanel.Sinks, clk clock.Clock, inputPath string) (chan GestureInboxMsg, *sync.WaitGroup) {
	inbox := make(chan GestureInboxMsg, gestureInboxDepth)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		reader := Drag.NewInputDirReader(inputPath)
		mine := clk.Copy()
		wheel := &wheelTotals{}
		for {
			for _, raw := range reader.ReadAll() {
				if ev, ok := Drag.DecodeRawInput(raw); ok {
					wheel.difference(&ev)
					Dispatch.HandleRawInputMsg(ctx, ev, md, speedSinks)
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
						Dispatch.ApplyEdit(ctx, gm.Op, gm.Entity, gm.Attr, gm.Payload, md, speedSinks)
					case GestureMsgSave:
						Dispatch.HandleSaveMsg(md)
					case GestureMsgCommand:
						Dispatch.HandleRawInputMsg(ctx, gm.Command, md, speedSinks)
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

func SendGestureMsgBlocking(ctx context.Context, inbox chan<- GestureInboxMsg, gm GestureInboxMsg) {
	select {
	case inbox <- gm:
	case <-ctx.Done():
	}
}

func startStdinReader(ctx context.Context, cancel context.CancelFunc, out chan<- GestureInboxMsg) *sync.WaitGroup {
	stdinWG := new(sync.WaitGroup)
	stdinWG.Add(1)
	h := Drag.Handlers{
		ApplyEdit: func(op string, entity, attr byte, payload []byte) {
			SendGestureMsgBlocking(ctx, out, GestureInboxMsg{
				Kind: GestureMsgEdit,
				Op:   op, Entity: entity, Attr: attr, Payload: payload,
			})
		},
		HandleSave: func() {
			SendGestureMsgBlocking(ctx, out, GestureInboxMsg{Kind: GestureMsgSave})
		},
		HandleCommand: func(ev Drag.RawInputMsg) {
			SendGestureMsgBlocking(ctx, out, GestureInboxMsg{Kind: GestureMsgCommand, Command: ev})
		},
	}
	go func() {
		defer stdinWG.Done()
		Drag.RunStdinReader(ctx, os.Stdin, h)
		cancel()
	}()
	return stdinWG
}
