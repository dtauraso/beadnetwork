// drive_stream_wiring_test.go — asserts the WIRING that fixes the framing desync
// documented in docs/interior-stream-framing.md and pinned mechanically by
// interior_stream_concurrent_write_test.go: a node's own interior-stream getter
// (newInteriorStreamGetter, used by its Update-loop closures) and its per-DriveHeld-
// goroutine drive-stream getter (newDriveStreamGetter, used by BuildArgs.DriveOut) must
// NEVER resolve to the same *interiorStream instance — which, before this fix, is exactly
// what happened (both read pb.md.sw.interiorOuts via the SAME lazy-cache-once closure).
//
// This is the test the original bug needed: interior_stream_concurrent_write_test.go
// proves TWO GOROUTINES SHARING ONE WRITER is unsafe in the abstract, but nothing before
// this file asserted that production's OWN wiring never constructs that sharing. A
// regression that made newDriveStreamGetter fall back to pb.md.sw.interiorOuts (a typo, a
// bad merge, "simplifying" DriveOut back to Out) would reintroduce the live-editor bug
// with every other test in this package still green, because
// TestInteriorStreamSustainedFraming and friends only exercise CORRECT wiring — they
// cannot see a wiring mistake that never desyncs a healthy-looking run.
package Wiring

import (
	"io"
	"testing"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// discardWriter is a trivial io.Writer identity distinguishable by pointer, standing in
// for a real dedicated fd (os.NewFile(...)) — this test only cares about WHICH *io.Writer
// instance each getter resolves to, not that bytes actually flow anywhere.
type discardWriter struct{ name string }

func (d *discardWriter) Write(p []byte) (int, error) { return len(p), nil }

// testBuildInteriorFrame is a stand-in for main.go's real buildInteriorFrame closure
// (Buffer.BuildInteriorStreamFrame) — only its non-nilness matters to
// newInteriorStreamGetter/newDriveStreamGetter (a nil builder is treated identically to a
// nil writer: "this stream is absent"), so its actual output is irrelevant here.
func testBuildInteriorFrame(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []wire.RowEvent) []byte {
	return nil
}

// TestDriveStreamNeverSharesNodesInteriorStream is the core wiring assertion: for a node
// with BOTH an interior stream and drive-slot streams wired (setNodeStreams' normal
// production shape), newInteriorStreamGetter and newDriveStreamGetter(slot=0) must
// resolve to DIFFERENT *interiorStream instances, each wrapping a DIFFERENT underlying
// io.Writer — the two emitting goroutines (the node's own Update loop, and its DriveHeld
// goroutine) get genuinely separate fds, not two views onto one.
func TestDriveStreamNeverSharesNodesInteriorStream(t *testing.T) {
	interiorW := &discardWriter{name: "interior"}
	drive0W := &discardWriter{name: "drive-0"}
	drive1W := &discardWriter{name: "drive-1"}

	md := &MoveDispatch{}
	md.sw.interiorOuts = map[string]io.Writer{"nodeA": interiorW}
	md.sw.driveOuts = map[string][driveSlotsPerNode]io.Writer{"nodeA": {drive0W, drive1W}}
	md.sw.buildInteriorFrame = testBuildInteriorFrame

	pb := PortBindings{md: md}

	interiorStreamFn := newInteriorStreamGetter("nodeA", pb)
	drive0Fn := newDriveStreamGetter("nodeA", 0, pb)
	drive1Fn := newDriveStreamGetter("nodeA", 1, pb)

	interior := interiorStreamFn()
	drive0 := drive0Fn()
	drive1 := drive1Fn()

	if interior == nil || drive0 == nil || drive1 == nil {
		t.Fatalf("all three streams must resolve when both interiorOuts and driveOuts are wired: interior=%v drive0=%v drive1=%v", interior, drive0, drive1)
	}
	if interior == drive0 || interior == drive1 || drive0 == drive1 {
		t.Fatalf("a node's interior stream and its drive-slot streams must be THREE DISTINCT *interiorStream instances (three separate goroutines' fds) — got interior=%p drive0=%p drive1=%p", interior, drive0, drive1)
	}
	if interior.out != interiorW {
		t.Fatalf("interior stream's writer must be the node's OWN interiorOuts entry, got a different io.Writer")
	}
	if drive0.out != drive0W {
		t.Fatalf("drive slot 0's writer must be driveOuts[nodeA][0], got a different io.Writer")
	}
	if drive1.out != drive1W {
		t.Fatalf("drive slot 1's writer must be driveOuts[nodeA][1], got a different io.Writer")
	}
	// Cross-check the specific historical bug directly: drive0's writer must not be the
	// SAME writer the node's own interior stream uses — this is the exact "DriveHeld
	// writes the node's shared interiorStream" violation docs/interior-stream-framing.md
	// documents.
	if drive0.out == interior.out {
		t.Fatalf("drive slot 0 shares its writer with the node's own interior stream — this IS the original bug (docs/interior-stream-framing.md)")
	}
}

// TestDriveStreamAbsentWhenUnwired mirrors newInteriorStreamGetter's own nil-safe fallback
// (no WIREFOLD_STREAM_FDS entry -> nil stream, not a crash): a node with no driveOuts entry
// (e.g. a kind that never calls DriveOut, or a slot beyond what a kind uses) gets nil, not
// a stream aliasing something else.
func TestDriveStreamAbsentWhenUnwired(t *testing.T) {
	md := &MoveDispatch{}
	// driveOuts populated (setNodeStreams always does, once "drive" resolves) but with no
	// entry for this particular node id — mirrors a missing nodeMover row, or (more to the
	// point for this test) simply calling DriveOut for a slot/name never actually wired.
	md.sw.driveOuts = map[string][driveSlotsPerNode]io.Writer{}
	pb := PortBindings{md: md}
	if s := newDriveStreamGetter("nodeB", 0, pb)(); s != nil {
		t.Fatalf("newDriveStreamGetter for an unwired node must return nil, got a stream")
	}
}
