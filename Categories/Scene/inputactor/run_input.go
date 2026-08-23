package inputactor

import (
	"context"
	"os"
	"sync"

	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/SliderPanel"
	clock "github.com/dtauraso/wirefold/Categories/Clock"
	beadanimation "github.com/dtauraso/wirefold/Categories/Node/BeadAnimation"

	"github.com/dtauraso/wirefold/Categories/Scene/scenerun"
)

func StartStdinReader(ctx context.Context, cancel context.CancelFunc, slotReg beadanimation.SlotRegistry, md *scenerun.MoveDispatch, speedSinks SliderPanel.Sinks, clk clock.Clock, inputPath string) (*sync.WaitGroup, *sync.WaitGroup) {
	inbox, gestureWG := StartGestureActor(ctx, slotReg, md, speedSinks, clk, inputPath)

	stdinWG := new(sync.WaitGroup)
	stdinWG.Add(1)
	h := Handlers{
		ApplyEdit: func(op string, entity, attr byte, payload []byte) {
			SendGestureMsgBlocking(ctx, inbox, GestureInboxMsg{
				Kind: GestureMsgEdit,
				Op:   op, Entity: entity, Attr: attr, Payload: payload,
			})
		},
		HandleSave: func() {
			SendGestureMsgBlocking(ctx, inbox, GestureInboxMsg{Kind: GestureMsgSave})
		},
	}
	go func() {
		defer stdinWG.Done()
		RunStdinReader(ctx, os.Stdin, h)
		cancel()
	}()
	return stdinWG, gestureWG
}
