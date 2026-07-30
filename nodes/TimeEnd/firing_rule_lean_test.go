package timeend

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"testing"
	"time"

	T "github.com/dtauraso/wirefold/Trace"
)

// stepWire continuously StepOnceAts pw on a short wall-clock poll until ctx is
// cancelled, matching the production per-cycle StepOnceAt delivery path (no
// blocking delivery loop). clk is this goroutine's OWN clock copy, which
// advances on its own, so a placed bead is carried to delivery once its
// deadline is crossed.
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

// TestHoldFiresAndHoldsOnReceiveLean covers hold/SPEC.md's core contract on
// the one real clock: terminal node, no output. Startup emits the empty
// (noValue) interior bead; on a received value it fires and re-emits the
// held bead with the new value; Held reflects the latest received value.
func TestHoldFiresAndHoldsOnReceiveLean(t *testing.T) {
	const latMs = 20.0
	tr := T.New()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pw := wire.NewPacedWire(int(latMs), 1.0)
	clk := wire.NewRealClock()
	stepWire(ctx, pw, clk.Copy())
	// inSrc is a test-only seeding source on pw: PlaceDrivenAt places a bead
	// (no walker) that the stepWire loop above then drives to delivery,
	// reusing the production placement API to inject the test's input value.
	inSrc := wire.NewPacedOutNoGeom(pw, ctx, "seed", "Out", tr, wire.RuleFireAndForget, 0, "")

	beadCh := make(chan int, 16)
	fires := 0
	node := &TimeEnd{
		Fire:         func() { fires++ },
		Clock:        clk,
		In:           wire.NewInPaced(pw, ctx, "hold", "In", tr, nil, -1),
		EmitHeldBead: func(v int) { beadCh <- v },
	}

	done := make(chan struct{})
	go func() { node.Update(ctx); close(done) }()

	// Startup emits the empty-interior sentinel first.
	select {
	case v := <-beadCh:
		if v != noValue {
			t.Fatalf("startup bead: expected sentinel %d, got %d", noValue, v)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for startup bead")
	}

	if !inSrc.PlaceDrivenAt(7, clk.Tick()).Live() {
		t.Fatal("PlaceDrivenAt returned false")
	}

	// After input arrives (7 != held -1) the changed held bead is emitted.
	select {
	case v := <-beadCh:
		if v != 7 {
			t.Fatalf("held bead after input: expected 7, got %d", v)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for held bead")
	}

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("node.Update did not exit after cancel")
	}

	if node.Held != 7 {
		t.Errorf("Held after fire: expected 7, got %d", node.Held)
	}
	if fires < 1 {
		t.Errorf("expected Fire to be called at least once, got %d", fires)
	}
}
