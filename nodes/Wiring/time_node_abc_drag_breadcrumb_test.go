package Wiring

// time_node_abc_drag_breadcrumb_test.go — shared test-only plumbing for the "abc-drag"
// breadcrumb neighborSetCRequantize (node_move.go) logs for every direct neighbor that
// receives a moveMsgKindNeighborSetC when a node is dragged: syncBuffer (a race-free debug
// sink), parseBreadcrumbLines/breadcrumbLine (decode), and waitForAbcDrag/abcDragDeltasFor
// (poll on the breadcrumb as the happens-before sync point before reading a recipient's own
// LayoutHolder). Used by neighbor_setc_test.go, rotating_pole_test.go, node_move_test.go,
// drag_persist_e2e_test.go, and subtree_persist_test.go.

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// syncBuffer wraps a bytes.Buffer with a mutex — Trace.Breadcrumb's debugSink write is
// itself mutex-guarded inside Trace (t.mu), but that lock is private to the Trace value;
// this wrapper gives the TEST its own safe read path once the drag has quiesced (we only
// read after pollDragConverged + a settle sleep, never concurrently with a write, but the
// wrapper costs nothing and removes any doubt).
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

// breadcrumbLine is the decoded shape of one {"kind":"breadcrumb",...} JSONL line.
type breadcrumbLine struct {
	Kind  string `json:"kind"`
	Label string `json:"label"`
	Node  string `json:"node"`
	Port  string `json:"port"`
	Value string `json:"value"`
}

// waitForAbcDrag blocks until at least one "abc-drag" breadcrumb naming node has
// appeared in dbg, or fails the test on timeout. neighborSetCRequantize writes the
// recipient's own LocalPolar entry (SetLocalPolar/SetPole) and THEN logs this
// breadcrumb, in that order, in the SAME call on the recipient's OWN goroutine — so
// once the breadcrumb is observed here (through dbg's mutex-guarded buffer), the
// recipient's LayoutHolder is guaranteed to already reflect that write, and reading it
// from the calling (test) goroutine afterward is race-free. See
// TestEveryDragRecipientLogsAbcDragBreadcrumb below for the full argument.
func waitForAbcDrag(t *testing.T, dbg *syncBuffer, node string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		for _, b := range parseBreadcrumbLines(t, dbg.String()) {
			if b.Label == "abc-drag" && b.Node == node {
				return
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for an abc-drag breadcrumb for node %q", node)
		}
		time.Sleep(time.Millisecond)
	}
}

// abcDragDeltaRe matches the "delta=(a,b,c)" substring an "abc-drag" breadcrumb's Value
// carries — the exact same deltaA/deltaB/deltaC ints neighborSetCRequantize passes to
// both Trace.Breadcrumb and Trace.AbcDrag in the same call (node_move.go/quantized_move.go),
// so parsing it here reads the identical production payload the now-removed msgTap used
// to intercept in flight.
var abcDragDeltaRe = regexp.MustCompile(`delta=\((-?\d+),(-?\d+),(-?\d+)\)`)

// abcDragDeltasFor returns every "abc-drag" breadcrumb's (DeltaA,DeltaB,DeltaC) triple
// recorded so far for node, in the order they were written to dbg (dbg is append-only
// and single-writer per node — see waitForAbcDrag's happens-before argument). Callers
// must have already synced past the Nth breadcrumb they care about via
// waitForAbcDragCount before reading, exactly as every other reader of dbg/lh in this
// package does.
func abcDragDeltasFor(t *testing.T, dbg *syncBuffer, node string) [][3]int {
	t.Helper()
	var out [][3]int
	for _, b := range parseBreadcrumbLines(t, dbg.String()) {
		if b.Label != "abc-drag" || b.Node != node {
			continue
		}
		m := abcDragDeltaRe.FindStringSubmatch(b.Value)
		if m == nil {
			t.Fatalf("abc-drag breadcrumb for %s missing delta=(...): %q", node, b.Value)
		}
		var d [3]int
		for i := 0; i < 3; i++ {
			n, err := strconv.Atoi(m[i+1])
			if err != nil {
				t.Fatalf("parse delta component %q: %v", m[i+1], err)
			}
			d[i] = n
		}
		out = append(out, d)
	}
	return out
}

func parseBreadcrumbLines(t *testing.T, raw string) []breadcrumbLine {
	t.Helper()
	var out []breadcrumbLine
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		if line == "" {
			continue
		}
		var b breadcrumbLine
		if err := json.Unmarshal([]byte(line), &b); err != nil {
			t.Fatalf("breadcrumb line is not valid JSON: %v (%q)", err, line)
		}
		out = append(out, b)
	}
	return out
}
