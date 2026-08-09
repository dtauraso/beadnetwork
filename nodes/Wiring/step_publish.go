package Wiring

// sendStepsNonBlocking delivers steps to ch, latest-wins: if the buffer already holds
// an undrained stale value (the edgeMover's own goroutine hasn't woken to drain it
// since the last publish), that stale value is dropped and replaced — the same
// "producer sends, one consumer owns its copy" shape speedCh already uses
// (per-goroutine-clock.md "Delivery"), applied to edgeMover.stepsIn. A nil ch (this
// edge had no bound edgeMover.stepsIn, or a bare test nodeMover with no outStepsIn)
// makes every case here select `default`, so this is a silent no-op — this node's own
// goroutine never blocks on it.
func sendStepsNonBlocking(ch chan int, steps int) {
	select {
	case ch <- steps:
		return
	default:
	}
	select {
	case <-ch:
	default:
	}
	select {
	case ch <- steps:
	default:
	}
}
