package beadchain

import (
	"bytes"
	"runtime"
	"runtime/pprof"
	"strings"
	"testing"
	"time"
)

// bead_idle_test.go — an idle bead goroutine costs nothing: with no drag, no tick, and no
// geometry event, it is parked in Go's "[select]" state, never running/runnable.

// --- Idle costs nothing ---------------------------------------------------------------

// TestIdleBeadIsBlockedNotRunnable: with no drag, no tick, and no geometry event, every
// bead goroutine's own frame (Bead.run, chan receive) shows up in a goroutine dump as
// blocked in Go's "[select]" state — never running/runnable — which is what "parked at
// zero CPU" means at the runtime level. This is the direct behavioural evidence for the
// claim tools/network/beads/check-no-select-default.sh backs structurally: default: would make the frame
// appear as "[running]" in a tight loop instead of "[select]" here.
func TestIdleBeadIsBlockedNotRunnable(t *testing.T) {
	const n = 50
	stops := make([]chan struct{}, n)
	for i := 0; i < n; i++ {
		b, _, _, stop, _ := newTestBead(float64(i))
		stops[i] = stop
		b.Start()
	}
	defer func() {
		for _, s := range stops {
			close(s)
		}
	}()

	// Give the goroutines a moment to actually park (get scheduled and reach their select).
	runtime.Gosched()
	time.Sleep(50 * time.Millisecond)

	var buf bytes.Buffer
	if err := pprof.Lookup("goroutine").WriteTo(&buf, 2); err != nil {
		t.Fatalf("goroutine profile: %v", err)
	}
	dump := buf.String()

	sections := strings.Split(dump, "\n\n")
	found := 0
	for _, sec := range sections {
		if !strings.Contains(sec, "beadchain.(*Bead).run") {
			continue
		}
		found++
		if !strings.Contains(sec, "[select]") {
			t.Fatalf("an idle Bead.run goroutine is not parked in [select] (possible spin):\n%s", sec)
		}
	}
	if found < n {
		t.Fatalf("expected %d idle Bead.run stacks, found %d", n, found)
	}
}
