// paced_wire_breadcrumb_not_gated_test.go — pins that debug breadcrumbs
// (T.KindBreadcrumb) still emit when edge-bead tracing is OFF (the default,
// WIREFOLD_EDGE_BEAD_TRACE unset). Breadcrumbs are CLAUDE.md's designated
// Go-layer debugging channel and must never go silent depending on an
// unrelated trace-volume knob — see memory/feedback_make_bug_class_unrepresentable.md
// and tools/check-breadcrumb-not-gated.sh, which enforces the same invariant
// statically. This test observes ONLY what this one PacedWire goroutine (the
// test goroutine itself, calling Send synchronously) recorded and drained —
// no second goroutine, no channel delivery between goroutines, no timing (see
// docs/testing-shape.md and CLAUDE.md's "Testing shape" section).
package wire

import (
	"testing"

	T "github.com/dtauraso/wirefold/Trace"
)

// TestBreadcrumbEmitsWithEdgeBeadTraceOff fills a wire's inCh to capacity (so
// the next Send hits the SendBufferFull path, which unconditionally queues a
// KindBreadcrumb RowEvent onto breadcrumbCh — see Send's doc comment), then
// drains breadcrumbCh directly and asserts the breadcrumb row is present. This
// exercises the exact code path check-breadcrumb-not-gated.sh protects: if
// edgeBeadTraceEnabled ever spread to gate this append, the row would be
// missing here even though edgeBeadTraceEnabled is in its default (false/off)
// state — the regression's signature failure, silence instead of an error.
func TestBreadcrumbEmitsWithEdgeBeadTraceOff(t *testing.T) {
	if edgeBeadTraceEnabled {
		t.Fatalf("edgeBeadTraceEnabled = true; this test asserts breadcrumb behavior with " +
			"the DEFAULT (off) state and does not control WIREFOLD_EDGE_BEAD_TRACE — " +
			"run without that env var set")
	}

	pw := NewPacedWire(0, PulseSpeedWuPerMs)

	for i := range wireChanBufferSize {
		select {
		case pw.inCh <- placeRequest{val: i}:
		default:
			t.Fatalf("inCh reported full before reaching wireChanBufferSize (at %d)", i)
		}
	}

	if got := pw.Send(1, beadPlacement{}, 0); got != SendBufferFull {
		t.Fatalf("Send on a full, undrained inCh = %v, want SendBufferFull", got)
	}

	events := pw.drainBreadcrumbEvents()
	if len(events) != 1 {
		t.Fatalf("drainBreadcrumbEvents() after a SendBufferFull = %d events, want 1 "+
			"(breadcrumb must emit regardless of edgeBeadTraceEnabled)", len(events))
	}
	if events[0].Kind != T.KindBreadcrumb {
		t.Fatalf("drained event Kind = %q, want %q", events[0].Kind, T.KindBreadcrumb)
	}
}
