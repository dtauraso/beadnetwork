package beadchain

import (
	"testing"
	"time"
)

// bead_actor_helpers_test.go — shared fixtures used by more than one of this package's
// bead_actor test files.

// newTestBead builds one bead against a fresh BeadWakeGroup and its own dedicated tick
// channel (a plain, test-controlled channel rather than the real TickBroadcaster, so
// animation ticks can be delivered deterministically without depending on wall time), with
// its observation channel already armed — every read of the bead's state in these tests
// goes through that channel, never through a direct field read from another goroutine
// (`go test -race` enforces this: a direct read of Bead's fields from the test goroutine
// was caught as a genuine data race during this file's own development, exactly the
// cross-goroutine shared-read the ownership model forbids — see bead_actor.go's note next
// to WithObserve).
func newTestBead(offsetR float64) (*Bead, *BeadWakeGroup, chan struct{}, chan struct{}, <-chan BeadSnapshot) {
	g := NewBeadWakeGroup()
	tickCh := make(chan struct{})
	stop := make(chan struct{})
	geom, wake, settle := g.Current()
	b := NewBead(offsetR, geom, wake, settle, tickCh, stop)
	obs := b.WithObserve()
	return b, g, tickCh, stop, obs
}

// waitForSnapshot drains obs (a buffered-1, latest-wins channel) until cond holds on a
// received snapshot, or the timeout elapses. This is a genuine channel RECEIVE — properly
// synchronized with the bead's own pushObserve send — never a poll of the bead's fields.
func waitForSnapshot(t *testing.T, obs <-chan BeadSnapshot, timeout time.Duration, cond func(BeadSnapshot) bool) (BeadSnapshot, bool) {
	t.Helper()
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	var last BeadSnapshot
	for {
		select {
		case snap := <-obs:
			last = snap
			if cond(snap) {
				return snap, true
			}
		case <-deadline.C:
			return last, false
		}
	}
}

func itoa(n int) string {
	if n == 40 {
		return "N=40"
	}
	return "N=1000"
}
