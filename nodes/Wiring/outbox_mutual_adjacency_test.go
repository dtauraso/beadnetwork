// outbox_mutual_adjacency_test.go — CHECKS BY CODE nm.pending's retain-and-retry send:
// nothing a nodeMover sends is ever dropped, and per-destination delivery order is
// exactly enqueue order (FIFO per destination), INCLUDING when the destination's own
// channel is genuinely full and the retry path must fire.
package Wiring

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestOutboxFIFOPerTargetOrderNoDrop drives nm.pending's retain-and-retry send directly
// (the same mechanism nm.sendMove/flushPending implement): a single sender goroutine
// (mirroring the real shape -- one handler goroutine, its own run loop, doing every send
// for that mover) interleaves many sequenced items across several destination ids, while
// the destinations' inbox channels are deliberately tiny (buffered 1) and drained slowly
// from separate goroutines -- forcing flushPending's retain-on-full-channel path
// explicitly, not merely hoping a big buffer never fills. Asserts every item arrives (no
// drops) and each destination's own subsequence arrives in exactly send order (FIFO per
// target, even across retries).
func TestOutboxFIFOPerTargetOrderNoDrop(t *testing.T) {
	const totalCount = 4000
	dests := []string{"A", "B", "C"}

	chans := map[string]chan moveMsg{}
	for _, d := range dests {
		// Buffered 1 (not "big enough to never fill"): the sender will routinely
		// outrun a deliberately-throttled receiver, forcing flushPending's retain
		// path on every destination, repeatedly.
		chans[d] = make(chan moveMsg, 1)
	}

	nm := &nodeMover{
		resolveDest: func(id string) (chan moveMsg, bool) {
			ch, ok := chans[id]
			return ch, ok
		},
	}

	var mu sync.Mutex
	receivedByTarget := map[string][]int{}
	receivedTotal := 0
	done := make(chan struct{})
	var doneOnce sync.Once

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, d := range dests {
		go func(dest string, ch chan moveMsg) {
			for {
				select {
				case msg := <-ch:
					// Deliberately slow receiver: this is what keeps the tiny
					// buffered-1 channel full most of the time, forcing the
					// sender's retain-and-retry path rather than a single lucky
					// non-blocking send per item.
					time.Sleep(50 * time.Microsecond)
					mu.Lock()
					receivedByTarget[dest] = append(receivedByTarget[dest], msg.AnchorId)
					receivedTotal++
					if receivedTotal == totalCount {
						doneOnce.Do(func() { close(done) })
					}
					mu.Unlock()
				case <-ctx.Done():
					return
				}
			}
		}(d, chans[d])
	}

	// Single sender goroutine (matches production: one handler goroutine's
	// broadcastToEdgesAndPartners/requantizeLocalPolars call sends every outbound message for
	// that mover sequentially, via nm.sendMove == md.enqueueFor(nm)). AnchorId carries
	// the send sequence number so delivery order is directly checkable. This is the
	// same append-then-flush nm.sendMove performs, plus nm.run's own retry-loop call
	// (flushPending on its own, with no new item) standing in for the mover's
	// per-cycle retry.
	go func() {
		for i := 0; i < totalCount; i++ {
			dest := dests[i%len(dests)]
			nm.pending = append(nm.pending, pendingSend{destID: dest, msg: moveMsg{AnchorId: i}})
			nm.flushPending()
		}
		// Keep retrying whatever didn't fit until it all drains (mirrors nm.run's
		// per-cycle flushPending call).
		for {
			mu.Lock()
			left := receivedTotal
			mu.Unlock()
			if left >= totalCount {
				return
			}
			nm.flushPending()
			time.Sleep(time.Millisecond)
		}
	}()

	select {
	case <-done:
	case <-time.After(15 * time.Second):
		mu.Lock()
		got := receivedTotal
		mu.Unlock()
		t.Fatalf("TestOutboxFIFOPerTargetOrderNoDrop: timed out waiting for delivery -- "+
			"got %d/%d (dropped or stuck)", got, totalCount)
	}

	mu.Lock()
	defer mu.Unlock()
	gotTotal := 0
	for dest, seq := range receivedByTarget {
		gotTotal += len(seq)
		for i := 1; i < len(seq); i++ {
			if seq[i] <= seq[i-1] {
				t.Fatalf("target %q: out-of-order delivery at index %d: %v", dest, i, seq)
			}
		}
	}
	if gotTotal != totalCount {
		t.Fatalf("dropped items: delivered %d, sent %d", gotTotal, totalCount)
	}
}
