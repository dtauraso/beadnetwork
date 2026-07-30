package pacer

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"testing"
	"time"

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
			time.Sleep(time.Millisecond)
		}
	}()
}

// TestPacerChangeStepFeedbackLean covers pacer's core contract on the one
// real clock: on each received value it fires and emits a change-step
// feedback bead on FeedbackOut — 1 the first time / when the value changes,
// 0 when it repeats.
func TestPacerChangeStepFeedbackLean(t *testing.T) {
	const latMs = 10.0
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

	outPw := wire.NewPacedWire(int(latMs), 1.0)
	// Production drives this output wire via its edge's own goroutine
	// (edgeMover.run); this bare-wire unit test has no edgeMover, so it must
	// supply the same per-cycle drive itself.
	stepWire(ctx, outPw, clk.Copy())

	node := &Node{
		Fire:      func() {},
		Clock:     clk,
		FromInput: wire.NewInPaced(inPw, ctx, "pacer", "FromInput", tr, nil, -1),
		FeedbackOut: wire.NewPacedOutNoGeom(outPw, ctx, "pacer", "FeedbackOut", tr,
			wire.RuleFireAndForget, int(latMs), ""),
	}
	observer := wire.NewInPaced(outPw, ctx, "obs", "In", tr, nil, -1)

	done := make(chan struct{})
	go func() { node.Update(ctx); close(done) }()

	waitFor := func(want int) {
		t.Helper()
		deadline := time.Now().Add(3 * time.Second)
		for time.Now().Before(deadline) {
			if v, ok := observer.PollRecv(); ok {
				if v != want {
					t.Fatalf("expected feedback step %d, got %d", want, v)
				}
				return
			}
			time.Sleep(time.Millisecond)
		}
		t.Fatalf("timeout waiting for feedback step %d", want)
	}

	// First value ever seen -> step=1 (change from noValue).
	if !inSrc.PlaceDrivenAt(5, clk.Tick()).Live() {
		t.Fatal("PlaceDrivenAt returned false")
	}
	waitFor(1)

	// Same value again -> step=0.
	if !inSrc.PlaceDrivenAt(5, clk.Tick()).Live() {
		t.Fatal("PlaceDrivenAt returned false")
	}
	waitFor(0)

	// Different value -> step=1.
	if !inSrc.PlaceDrivenAt(6, clk.Tick()).Live() {
		t.Fatal("PlaceDrivenAt returned false")
	}
	waitFor(1)

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("node.Update did not exit after cancel")
	}

	if node.Held != 6 {
		t.Errorf("Held after fires: expected 6, got %d", node.Held)
	}
}
