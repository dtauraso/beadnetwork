package runtopology

import (
	"context"
	"os"
	"sync"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"
	clock "github.com/dtauraso/wirefold/src/Clock"
	beadanimation "github.com/dtauraso/wirefold/src/Node/BeadAnimation"

	"github.com/dtauraso/wirefold/src/NodeKinds/nodeapi"

	W "github.com/dtauraso/wirefold/src/Input/Dispatch"
	"github.com/dtauraso/wirefold/src/Input/Stdin"
)

func startStdinReader(ctx context.Context, cancel context.CancelFunc, slotReg beadanimation.SlotRegistry, md *W.MoveDispatch, speedSinks SliderPanel.Sinks, clk clock.Clock, inputPath string) (*sync.WaitGroup, *sync.WaitGroup) {
	inbox, gestureWG := startGestureActor(ctx, slotReg, md, speedSinks, clk, inputPath)

	stdinWG := new(sync.WaitGroup)
	stdinWG.Add(1)
	h := Stdin.Handlers{
		ApplyEdit: func(msg Stdin.StdinMsg) {
			sendGestureMsgBlocking(ctx, inbox, gestureInboxMsg{kind: gestureMsgEdit, msg: msg})
		},
		HandleSave: func() {
			sendGestureMsgBlocking(ctx, inbox, gestureInboxMsg{kind: gestureMsgSave})
		},
	}
	go func() {
		defer stdinWG.Done()
		Stdin.RunStdinReader(ctx, os.Stdin, h)
		cancel()
	}()
	return stdinWG, gestureWG
}

func launchNodeGoroutines(ctx context.Context, nodes []nodeapi.Node) *sync.WaitGroup {
	wg := new(sync.WaitGroup)
	wg.Add(len(nodes))
	for _, node := range nodes {
		go func() {
			defer wg.Done()
			node.Update(ctx)
		}()
	}
	return wg
}

func joinAll(wg, moverWG, stdinWG, gestureWG *sync.WaitGroup) {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		moverWG.Wait()
		stdinWG.Wait()
		gestureWG.Wait()
		close(done)
	}()
	<-done
}
