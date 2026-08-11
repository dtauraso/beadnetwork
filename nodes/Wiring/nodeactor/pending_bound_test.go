// pending_bound_test.go — proves this node's own pending retry queue's declared bound
// (maxPendingSends, consts.go) actually fires when a destination never drains (a wedged
// peer), and confirms the normal path — a destination that drains promptly — never trips
// it. Plain single-goroutine tests only (no synctest changes on this branch).
//
// Moved here from package Wiring in docs/planning/movedispatch-decomposition.md §20: this
// exercises EnqueueSend, now a NodeGeometry method rather than a Wiring-package closure
// (mover_registry.go's old enqueueFor closure body), so the test is a single-goroutine
// subject entirely within this package.

package nodeactor

import (
	"strings"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
)

// trySendMsg wraps a raw movemsg.Msg channel as a non-blocking try-send func value, the
// test-only equivalent of what package Wiring's resolveDest closures return in production
// (NeighborTrySend, node_geometry_accessors.go; edgemover.EdgeMover.TrySendFromSrc/Dst).
func trySendMsg(ch chan movemsg.Msg) func(movemsg.Msg) bool {
	return func(msg movemsg.Msg) bool {
		select {
		case ch <- msg:
			return true
		default:
			return false
		}
	}
}

// TestEnqueueForPanicsWhenPendingExceedsBound drives sends to a destination whose
// inbox is NEVER drained (simulating a wedged/dead peer) past maxPendingSends and
// asserts EnqueueSend panics, naming the two real causes (wedged peer / outpacing
// sender) rather than a generic "limit exceeded".
func TestEnqueueForPanicsWhenPendingExceedsBound(t *testing.T) {
	blockedCh := make(chan movemsg.Msg) // unbuffered, nobody ever receives -> always full
	nm := &NodeGeometry{
		id: "wedged-peer-test",
		msg: nodeMessaging{
			resolveDest: func(id string) (func(movemsg.Msg) bool, bool) {
				return trySendMsg(blockedCh), true
			},
		},
	}

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
			nm.EnqueueSend("peer", movemsg.Msg{})
		}
	}()

	if !panicked {
		t.Fatalf("EnqueueSend did not panic after %d retained sends to a wedged destination (want panic once len(pending) > %d)",
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
	liveCh := make(chan movemsg.Msg, inboxDepth)
	nm := &NodeGeometry{
		id: "live-peer-test",
		msg: nodeMessaging{
			resolveDest: func(id string) (func(movemsg.Msg) bool, bool) {
				return trySendMsg(liveCh), true
			},
		},
	}

	for range maxPendingSends * 10 {
		nm.EnqueueSend("peer", movemsg.Msg{})
		// Drain immediately, exactly as a live peer's own goroutine would each cycle.
		select {
		case <-liveCh:
		default:
		}
	}

	if len(nm.msg.pending) != 0 {
		t.Fatalf("normal (draining) path left %d items pending; want 0", len(nm.msg.pending))
	}
}
