package pulse

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

// TestPulseDrivesHeldValueLean covers pulse's core contract on the one real
// clock: sample-and-hold. It continuously drives its held value to Out
// (starting with the noValue sentinel), and updates the held value (with an
// immediate interior-bead update) whenever a new value arrives on FromInput.
func TestPulseDrivesHeldValueLean(t *testing.T) {
	const latMs = 10.0
	tr := T.New()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inPw := wire.NewPacedWire(latMs*wire.PulseSpeedWuPerMs, wire.PulseSpeedWuPerMs)
	clk := wire.NewRealClock()
	stepWire(ctx, inPw, clk.Copy())
	// inSrc is a test-only seeding source on inPw: PlaceDrivenAt places a bead
	// (no walker) that the stepWire loop above then drives to delivery,
	// reusing the production placement API to inject the test's input value.
	inSrc := wire.NewPacedOutNoGeom(inPw, ctx, "seed", "Out", tr, wire.RuleFireAndForget, 0, 0, "")

	outPw := wire.NewPacedWire(latMs*wire.PulseSpeedWuPerMs, wire.PulseSpeedWuPerMs)
	// Production drives this output wire via its edge's own goroutine
	// (edgeMover.run); this bare-wire unit test has no edgeMover, so it must
	// supply the same per-cycle drive itself.
	stepWire(ctx, outPw, clk.Copy())

	beadCh := make(chan int, 16)
	node := &Node{
		Fire:      func() {},
		Clock:     clk,
		FromInput: wire.NewInPaced(inPw, ctx, "pulse", "FromInput", tr, nil, -1),
		Out: wire.NewPacedOutNoGeom(outPw, ctx, "pulse", "Out", tr,
			wire.RuleFireAndForget, latMs*wire.PulseSpeedWuPerMs, latMs, ""),
		EmitHeldBead: func(v int) { beadCh <- v },
	}
	observer := wire.NewInPaced(outPw, ctx, "obs", "In", tr, nil, -1)

	done := make(chan struct{})
	go func() { node.Update(ctx); close(done) }()

	// Startup emits the empty-interior sentinel first.
	select {
	case v := <-beadCh:
		if v != -1 {
			t.Fatalf("startup bead: expected sentinel -1, got %d", v)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for startup bead")
	}

	if !inSrc.PlaceDrivenAt(5).Live() {
		t.Fatal("PlaceDrivenAt returned false")
	}

	// The interior bead updates the instant input arrives.
	select {
	case v := <-beadCh:
		if v != 5 {
			t.Fatalf("held bead after input: expected 5, got %d", v)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for held bead update")
	}

	// The drive goroutine continuously pulses the held value (5) to Out.
	deadline := time.Now().Add(3 * time.Second)
	got := false
	for time.Now().Before(deadline) {
		if v, ok := observer.PollRecv(); ok && v == 5 {
			got = true
			break
		}
		time.Sleep(time.Millisecond)
	}
	if !got {
		t.Fatal("timeout waiting for driven output value 5")
	}

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("node.Update did not exit after cancel")
	}
}
