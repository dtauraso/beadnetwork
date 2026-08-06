// pending_bound_test.go — proves nm.pending's declared bound (maxPendingSends,
// mover_registry.go) actually fires when a destination never drains (a wedged peer),
// and confirms the normal path — a destination that drains promptly — never trips it.
// Plain single-goroutine tests only (no synctest changes on this branch).

package Wiring

import (
	"strings"
	"testing"
)

// TestEnqueueForPanicsWhenPendingExceedsBound drives sends to a destination whose
// inbox is NEVER drained (simulating a wedged/dead peer) past maxPendingSends and
// asserts enqueueFor panics, naming the two real causes (wedged peer / outpacing
// sender) rather than a generic "limit exceeded".
func TestEnqueueForPanicsWhenPendingExceedsBound(t *testing.T) {
	mr := &moverRegistry{}
	blockedCh := make(chan moveMsg) // unbuffered, nobody ever receives -> always full
	nm := &nodeGeometry{
		id: "wedged-peer-test",
		resolveDest: func(id string) (chan moveMsg, bool) {
			return blockedCh, true
		},
	}
	send := mr.enqueueFor(nm)

	var panicked bool
	var msg string
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
				msg = r.(string)
			}
		}()
		for i := 0; i <= maxPendingSends; i++ {
			send("peer", moveMsg{})
		}
	}()

	if !panicked {
		t.Fatalf("enqueueFor did not panic after %d retained sends to a wedged destination (want panic once len(pending) > %d)",
			maxPendingSends+1, maxPendingSends)
	}
	t.Logf("verbatim panic: %s", msg)
	if !strings.Contains(msg, "wedged-peer-test") {
		t.Fatalf("panic message does not name the mover id: %q", msg)
	}
	if !strings.Contains(msg, "wedged") || !strings.Contains(msg, "faster") {
		t.Fatalf("panic message must name BOTH causes (wedged/dead peer AND outpacing a live peer), got: %q", msg)
	}
	if strings.Contains(msg, "unresolvable") || strings.Contains(msg, "unknown id") {
		t.Fatalf("panic message must NOT name an unresolvable destination as a cause (flushPending drops those, they can't grow pending): %q", msg)
	}
}

// TestEnqueueForNeverTripsBoundWhenDestinationDrains confirms the normal path — a
// destination whose channel is actively drained — never approaches the bound, however
// many sends are made, because flushPending always removes what it enqueues.
func TestEnqueueForNeverTripsBoundWhenDestinationDrains(t *testing.T) {
	mr := &moverRegistry{}
	liveCh := make(chan moveMsg, moverInboxDepth)
	nm := &nodeGeometry{
		id: "live-peer-test",
		resolveDest: func(id string) (chan moveMsg, bool) {
			return liveCh, true
		},
	}
	send := mr.enqueueFor(nm)

	for range maxPendingSends * 10 {
		send("peer", moveMsg{})
		// Drain immediately, exactly as a live peer's own goroutine would each cycle.
		select {
		case <-liveCh:
		default:
		}
	}

	if len(nm.pending) != 0 {
		t.Fatalf("normal (draining) path left %d items pending; want 0", len(nm.pending))
	}
}
