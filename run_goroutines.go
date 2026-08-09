package main

import (
	"context"
	"os"
	"sync"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
	W "github.com/dtauraso/wirefold/nodes/Wiring"
)

// startStdinReader launches the editor→Go bridge dispatch loop and returns the WaitGroup
// that tracks ONLY that loop.
//
// Read the editor→Go bridge: "edit" JSON lines (op = create/update/delete)
// from stdin. When stdin reaches EOF (extension host disconnect), cancel the context.
//
// stdinWG tracks ONLY this dispatch-loop goroutine, not RunStdinReader's internal
// frame-reader goroutine. That inner goroutine blocks in io.ReadFull(os.Stdin),
// which does NOT select on ctx — it is unblocked only by closing the fd (which
// RunStdinReader itself arranges when r is an io.Closer and ctx is done). On a
// non-pollable fd that close could still leave the read parked, so waiting on it
// here would turn a leak into a hang. RunStdinReader's dispatch loop, in contrast,
// selects on ctx.Done() and returns immediately on cancel regardless of the frame
// reader's state — that promptness is what stdinWG actually certifies. The frame
// reader goroutine is deliberately left un-waited (detached); in production it
// outlives the process only as long as it takes the OS to tear down the closed fd,
// which is bounded by process exit, not by this WaitGroup.
func startStdinReader(ctx context.Context, cancel context.CancelFunc, slotReg W.SlotRegistry, md *W.MoveDispatch, tr *T.Trace, speedSinks []chan float64) *sync.WaitGroup {
	stdinWG := new(sync.WaitGroup)
	stdinWG.Add(1)
	go func() {
		defer stdinWG.Done()
		W.RunStdinReader(ctx, os.Stdin, slotReg, md, tr, speedSinks)
		cancel()
	}()
	return stdinWG
}

// launchNodeGoroutines starts every node's own Update loop and returns the WaitGroup
// covering them.
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

// joinAll blocks until every tracked goroutine has exited.
//
// Wait for every tracked goroutine to exit — node Update loops, nodeMover/
// edgeMover goroutines, and the stdin dispatch loop — before closing the trace.
// No grace timeout: every one of these goroutines' only blocking call is
// SleepCycle, which selects on ctx.Done(), so cancel-to-return is bounded by one
// clock tick (~16ms), not by an arbitrary grace window. If a goroutine ever fails
// to exit, wg.Wait() below hangs visibly instead of silently proceeding past a
// still-running goroutine — a hang names the bug; a grace timeout hides it.
func joinAll(wg, moverWG, stdinWG *sync.WaitGroup) {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		moverWG.Wait()
		stdinWG.Wait()
		close(done)
	}()
	<-done
}
