// paced_wire_breadcrumb_dropped_test.go — pins that a dropped breadcrumb
// (breadcrumbCh's cap-4 non-blocking send losing a row, PacedWire.Send) is
// eventually SURFACED rather than lost for good (bounds-plan.md Step 3,
// "breadcrumbCh"). The cap and the non-blocking send both stay: this test
// asserts the reporting side-channel (droppedBreadcrumbs/
// flushDroppedBreadcrumbs), not a change to breadcrumbCh's capacity or
// blocking behavior.
//
// This exercises only this ONE PacedWire's own goroutine (the test goroutine
// itself, calling Send synchronously) — no second goroutine, no channel
// delivery between goroutines, no timing (docs/testing-shape.md).
package wire

import (
	"testing"

	T "github.com/dtauraso/wirefold/Trace"
)

// TestBreadcrumbDropsAreCountedAndReported fills inCh (so every Send hits the
// SendBufferFull path, which attempts a breadcrumbCh send) enough times that
// breadcrumbCh's cap-4 buffer overflows and starts dropping, WITHOUT ever
// draining breadcrumbCh in between — then drains breadcrumbCh once and
// confirms a "dropped N" breadcrumb reports the exact count of rows lost.
func TestBreadcrumbDropsAreCountedAndReported(t *testing.T) {
	pw := NewPacedWire(0, PulseSpeedWuPerMs)

	// Fill inCh to capacity directly so every subsequent Send call is forced
	// through the SendBufferFull -> breadcrumbCh-send path.
	for i := range wireChanBufferSize {
		select {
		case pw.inCh <- placeRequest{val: i}:
		default:
			t.Fatalf("inCh reported full before reaching wireChanBufferSize (at %d)", i)
		}
	}

	// breadcrumbCh has capacity 4. Call Send enough times (without ever
	// draining breadcrumbCh) that some sends land in the buffer and the rest
	// are dropped. wireChanBufferSize (4096) sends is far more than enough;
	// every call after the 4th SendBufferFull breadcrumb should drop.
	const sends = 10
	for i := range sends {
		if got := pw.Send(i, beadPlacement{}); got != SendBufferFull {
			t.Fatalf("Send(%d) on a full, undrained inCh = %v, want SendBufferFull", i, got)
		}
	}

	wantDropped := sends - cap(pw.breadcrumbCh)
	if pw.droppedBreadcrumbs != wantDropped {
		t.Fatalf("pw.droppedBreadcrumbs = %d, want %d (%d sends - %d channel capacity)",
			pw.droppedBreadcrumbs, wantDropped, sends, cap(pw.breadcrumbCh))
	}

	// Drain breadcrumbCh (the cap-4 buffer of "wire-send-buffer-full" rows
	// from the sends that fit) so the next Send call finds room to report
	// the drop count.
	events := pw.drainBreadcrumbEvents()
	if len(events) != cap(pw.breadcrumbCh) {
		t.Fatalf("drainBreadcrumbEvents() = %d events, want %d (the channel's full capacity)",
			len(events), cap(pw.breadcrumbCh))
	}

	// The NEXT Send call must report the drop before doing anything else
	// (flushDroppedBreadcrumbs, called at Send's top).
	if got := pw.Send(sends, beadPlacement{}); got != SendBufferFull {
		t.Fatalf("Send = %v, want SendBufferFull", got)
	}

	dropReport := pw.drainBreadcrumbEvents()
	if len(dropReport) < 1 {
		t.Fatalf("drainBreadcrumbEvents() after room reappeared = 0 events, want at least 1 "+
			"(a %q report of the %d drops)", T.BreadcrumbLabels[T.BreadcrumbWireBreadcrumbsDropped], wantDropped)
	}
	ev := dropReport[0]
	if ev.Kind != T.KindBreadcrumb || ev.Label != T.BreadcrumbWireBreadcrumbsDropped {
		t.Fatalf("first drained event = %+v, want Kind=%q Label=%d (%q)",
			ev, T.KindBreadcrumb, T.BreadcrumbWireBreadcrumbsDropped,
			T.BreadcrumbLabels[T.BreadcrumbWireBreadcrumbsDropped])
	}
	if int(ev.Value) != wantDropped {
		t.Fatalf("dropped-breadcrumb report Value = %d, want %d", ev.Value, wantDropped)
	}

	if pw.droppedBreadcrumbs != 0 {
		t.Fatalf("pw.droppedBreadcrumbs = %d after a successful report, want 0 (reset)",
			pw.droppedBreadcrumbs)
	}
}
