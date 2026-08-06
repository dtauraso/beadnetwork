// pair_node_self_clock_speed_test.go — regression coverage for the bug where a
// self-driven pair node's own RENDER clock (nodeGeometry.clk — read by writeStreamFrame's
// frame tick and chainBeads' animation, per_node_self.go's own doc comment) never applied a
// delivered speed change at all: ClaimSelfDrive copied clockSrc into geom.clk exactly ONCE,
// at build time, and nothing ever polled ApplySpeedNonBlocking on it afterward — unlike a
// ring node, whose nodeMover.run polls its own m.speedCh every cycle (mover_registry.go's
// finalizeActors). The kind's own SEPARATE clock (n.Clock, polled via its own SpeedCh in the
// kind's Update loop) DID receive speed changes and correctly paced wire delivery, which is
// why the defect was invisible to anything that only checked bead delivery/dot-turn timing:
// the visible thing (rendered bead motion) ran at the un-scaled rate regardless.
//
// This test asserts what ONE goroutine's own clock state does (docs/testing-shape.md): no
// second goroutine is launched, no cross-goroutine delivery is exercised — Step is called
// directly, synchronously, on the test goroutine, exactly the same shape
// speed_delivery_test.go already uses for ApplySpeedNonBlocking itself.
package Wiring

import (
	"context"
	"testing"
	"time"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// TestSelfDrivenGeometryClockAppliesDeliveredSpeed: a self-driven node's geometry clock
// (the one writeStreamFrame/chainBeads read) must speed up when a speed change is delivered
// on its own PairNodeSelf.speedCh, polled once per Step cycle — the same "faster than the
// old rate" wall-clock assertion TestApplySpeedNonBlockingAppliesOnWake already uses.
func TestSelfDrivenGeometryClockAppliesDeliveredSpeed(t *testing.T) {
	root := writeTree(t) // generic two-node/one-edge fixture (writeTree, scene_edit_persist_test.go)
	md := loadTreeMD(t, root)

	geom := md.mr.nodeGeoms["1"]
	if geom == nil {
		t.Fatalf("no nodeGeometry for node 1")
	}
	// Reproduce exactly what ClaimSelfDrive does at build time: copy clockSrc into clk once,
	// then wire this node's own dedicated speed channel (the fix under test).
	if geom.clockSrc != nil {
		geom.clk = geom.clockSrc.Copy()
	}
	speedCh := make(chan float64, 1)
	self := &PairNodeSelf{geom: geom, speedCh: speedCh}
	ctx := context.Background()

	// Baseline: default speed 1, ~2 ticks over 2 periods.
	before := geom.clk.Tick()
	time.Sleep(2 * wire.TickPeriod)
	self.Step(ctx, geom.clk.Tick())
	afterDefault := geom.clk.Tick()
	if advanced := afterDefault - before; advanced > 5 {
		t.Fatalf("baseline advance too fast before any speed change: advanced=%d ticks", advanced)
	}

	// Deliver a much faster speed the same way LoadSpeed/clockAttrHandlers broadcast it —
	// SendSpeedNonBlocking onto this clock's own channel — then let Step poll+apply it
	// BEFORE the measurement window starts (SetSpeed only affects the rate going forward
	// from the moment it is applied, same ordering TestApplySpeedNonBlockingAppliesOnWake
	// uses: send, apply-on-wake, THEN measure the next window).
	wire.SendSpeedNonBlocking(speedCh, 8)
	self.Step(ctx, geom.clk.Tick())

	before2 := geom.clk.Tick()
	time.Sleep(2 * wire.TickPeriod)
	self.Step(ctx, geom.clk.Tick())
	after2 := geom.clk.Tick()
	if advanced := after2 - before2; advanced < 5 {
		t.Fatalf("self-driven node geometry clock did not apply delivered speed: advanced=%d ticks over 2 periods at speed 8 (want >=5, i.e. clearly faster than the old speed-1 rate) — geom.clk is stuck at its build-time copy", advanced)
	}
}
