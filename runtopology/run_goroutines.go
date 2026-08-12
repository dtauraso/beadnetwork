package runtopology

import (
	"context"
	"os"
	"sync"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
	W "github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/stdinreader"
)

func startStdinReader(ctx context.Context, cancel context.CancelFunc, slotReg inputcodec.SlotRegistry, md *W.MoveDispatch, tr *T.Trace, speedSinks []chan float64) (*sync.WaitGroup, *sync.WaitGroup) {
	inbox, gestureWG := startGestureActor(ctx, slotReg, md, tr, speedSinks)

	stdinWG := new(sync.WaitGroup)
	stdinWG.Add(1)
	h := stdinreader.Handlers{
		ApplyEdit: func(msg inputcodec.StdinMsg) {
			sendGestureMsgBlocking(ctx, inbox, gestureInboxMsg{kind: gestureMsgEdit, msg: msg})
		},
		HandleRawInput: func(msg inputcodec.StdinMsg) {
			sendGestureMsgBlocking(ctx, inbox, gestureInboxMsg{kind: gestureMsgRawInput, msg: msg})
		},
		HandleSave: func() {
			sendGestureMsgBlocking(ctx, inbox, gestureInboxMsg{kind: gestureMsgSave})
		},
	}
	go func() {
		defer stdinWG.Done()
		stdinreader.RunStdinReader(ctx, os.Stdin, h)
		cancel()
	}()
	return stdinWG, gestureWG
}

func launchNodeGoroutines(ctx context.Context, nodes []wire.Node) *sync.WaitGroup {
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
