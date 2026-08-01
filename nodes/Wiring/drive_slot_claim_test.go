// drive_slot_claim_test.go — asserts item 1 of the DrivenOut structural fix
// (nodes/Wiring/driven_out.go): acquiring a drive-stream slot is possible AT MOST ONCE.
// This is the wiring-time half of the fix — the compile-time half (a plain a.Out(...)
// result can never become a Wiring.DrivenOut at all) is exercised by hand in
// docs/interior-stream-framing.md's "Guard verdict"/"Tests" sections, not by a Go test,
// since a type mismatch is a build failure, not a runtime assertion.
package Wiring

import (
	"context"
	"os"
	"strings"
	"testing"

	T "github.com/dtauraso/wirefold/Trace"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// TestDriveOutSecondSlotClaimFails constructs a BuildArgs directly (same package: its
// fields are unexported, mirroring how a kind's own RegisterBuilder wrapper builds one)
// and calls DriveOut twice with the SAME slot for two different port names — the mistake
// a future kind (or a bad merge) could make. The first call must succeed (a live,
// distinct DrivenOut); the second must be refused: a wiring-time failure REPORTED to
// stderr (never a panic — this happens during single-threaded construction, before any
// node/DriveHeld goroutine exists) and a dead (zero-value) DrivenOut that drives nothing
// rather than silently sharing the first claimant's fd.
func TestDriveOutSecondSlotClaimFails(t *testing.T) {
	a := BuildArgs{
		ctx:             context.Background(),
		name:            "nodeA",
		tr:              T.New(),
		sourceOuts:      &[]*wire.Out{},
		driveSlotClaims: map[int]string{},
	}

	// Capture stderr around the second (colliding) call only, so the assertion is
	// specific to that call's own report.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	origStderr := os.Stderr
	os.Stderr = w

	first := a.DriveOut("Out", 0)
	second := a.DriveOut("OutFanout", 0) // same slot — the collision

	os.Stderr = origStderr
	w.Close()
	var buf strings.Builder
	buf.Grow(4096)
	tmp := make([]byte, 4096)
	for {
		n, rerr := r.Read(tmp)
		if n > 0 {
			buf.Write(tmp[:n])
		}
		if rerr != nil {
			break
		}
	}
	r.Close()
	report := buf.String()

	// first is a dead-end (chan-mode, unwired) Out in this no-loader test build — that's
	// expected (no real PortBindings/stream fds here). What distinguishes it from
	// second is PlaceDrivenAt's outcome: a normal dead-end Out still successfully
	// SENDS (chan mode) rather than failing outright.
	if di := first.PlaceDrivenAt(0, 0); di.Failed() {
		t.Fatalf("first DriveOut call should not itself be refused")
	}
	if second.Paced() || second.Wired() {
		t.Fatalf("second DriveOut call claiming an already-claimed slot must be a dead (zero-value) DrivenOut, got a live one")
	}
	if report == "" {
		t.Fatalf("second DriveOut call claiming an already-claimed slot must report the collision to stderr, got nothing")
	}
	if !strings.Contains(report, "drive-stream collision") || !strings.Contains(report, `"Out"`) || !strings.Contains(report, `"OutFanout"`) {
		t.Fatalf("collision report must name BOTH claimants, got: %s", report)
	}

	// A DriveHeld goroutine handed the refused (dead) DrivenOut must not place anything
	// and must return promptly rather than spinning — PlaceDrivenAt on a nil-out
	// DrivenOut fails immediately (DriveItem.Failed()==true), which is what makes the
	// goroutine exit on its very first cycle. Asserted directly here (not via a live
	// DriveHeld goroutine) since that behavior is gatecommon's, already covered by its
	// own tests; this test only owns the WIRING-time refusal.
	if di := second.PlaceDrivenAt(0, 0); !di.Failed() {
		t.Fatalf("a refused (dead) DrivenOut must fail PlaceDrivenAt, not place a bead")
	}
}
