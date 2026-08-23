package scenebuild

import (
	"context"
	"github.com/dtauraso/wirefold/Categories/Scene/Drag"
	"github.com/dtauraso/wirefold/Categories/Scene/scenerun"
	"os"
	"sync"

	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/SliderPanel"
	clock "github.com/dtauraso/wirefold/Categories/Clock"
)

func StartStdinReader(ctx context.Context, cancel context.CancelFunc, md *scenerun.MoveDispatch, speedSinks SliderPanel.Sinks, clk clock.Clock, inputPath string) (*sync.WaitGroup, *sync.WaitGroup) {
	inbox, gestureWG := StartGestureActor(ctx, md, speedSinks, clk, inputPath)

	stdinWG := new(sync.WaitGroup)
	stdinWG.Add(1)
	h := Drag.Handlers{
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
		Drag.RunStdinReader(ctx, os.Stdin, h)
		cancel()
	}()
	return stdinWG, gestureWG
}
