package time

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"testing"
	stdtime "time"

	T "github.com/dtauraso/wirefold/Trace"
)

// stepWire continuously StepOnceAts pw on a short wall-clock poll until ctx is
// cancelled, matching the production per-cycle StepOnceAt delivery path. clk is
// this goroutine's OWN clock copy; callers must not share it with another goroutine.
func stepWire(ctx context.Context, pw *wire.PacedWire, clk wire.Clock) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			pw.DriveOneCycle(ctx, clk.Tick())
			stdtime.Sleep(stdtime.Millisecond)
		}
	}()
}

// TestFireOnReceiveLean covers time's core fire-on-receive
// contract on the one real clock: on receive it fires and forwards the
// PRIOR held value (starts at Held's zero value) to every ToNext broadcast
// entry, then stores the new value in Held.
func TestFireOnReceiveLean(t *testing.T) {
	const latMs = 40.0
	tr := T.New()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inPw := wire.NewPacedWire(int(latMs), 1.0)
	clk := wire.NewRealClock()
	stepWire(ctx, inPw, clk.Copy())
	// inSrc is a test-only seeding source on inPw: PlaceDrivenAt places a bead
	// (no walker) that the stepWire loop above then drives to delivery,
	// reusing the production placement API to inject the test's input value.
	inSrc := wire.NewPacedOutNoGeom(inPw, ctx, "seed", "Out", tr, wire.RuleFireAndForget, 0, "")

	outPw0 := wire.NewPacedWire(int(latMs), 1.0)
	outPw1 := wire.NewPacedWire(int(latMs), 1.0)
	// Production drives these output wires via each edge's own goroutine
	// (edgeMover.run); this bare-wire unit test has no edgeMover, so it must
	// supply the same per-cycle drive itself, exactly as it already does for
	// the input wire above.
	stepWire(ctx, outPw0, clk.Copy())
	stepWire(ctx, outPw1, clk.Copy())

	node := &Time{
		Fire:  func() {},
		Clock: clk,
		Held:  99, // seed a non-zero prior value to forward
		In:    wire.NewInPaced(inPw, ctx, "in", "In", tr, nil, -1),
		ToNext: wire.Broadcast{
			wire.NewPacedOutNoGeom(outPw0, ctx, "in", "ToNext0", tr,
				wire.RuleFireAndForget, int(latMs), ""),
			wire.NewPacedOutNoGeom(outPw1, ctx, "in", "ToNext1", tr,
				wire.RuleFireAndForget, int(latMs), ""),
		},
	}
	obs0 := wire.NewInPaced(outPw0, ctx, "obs0", "In", tr, nil, -1)
	obs1 := wire.NewInPaced(outPw1, ctx, "obs1", "In", tr, nil, -1)

	done := make(chan struct{})
	go func() { node.Update(ctx); close(done) }()

	if !inSrc.PlaceDrivenAt(7, clk.Tick()).Live() {
		t.Fatal("PlaceDrivenAt returned false")
	}

	waitFor := func(obs *wire.In, want int) {
		t.Helper()
		deadline := stdtime.Now().Add(3 * stdtime.Second)
		for stdtime.Now().Before(deadline) {
			if v, ok := obs.PollRecv(); ok {
				if v != want {
					t.Fatalf("expected %d, got %d", want, v)
				}
				return
			}
			stdtime.Sleep(stdtime.Millisecond)
		}
		t.Fatalf("timeout waiting for value %d", want)
	}

	waitFor(obs0, 99)
	waitFor(obs1, 99)

	cancel()
	select {
	case <-done:
	case <-stdtime.After(2 * stdtime.Second):
		t.Fatal("node.Update did not exit after cancel")
	}

	if node.Held != 7 {
		t.Errorf("Held after fire: expected 7, got %d", node.Held)
	}
}
