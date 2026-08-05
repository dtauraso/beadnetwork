// bead_emit.go — the interior-bead emission half of emit_geometry.go, split out as a pure
// move (no logic changes): emitNodeBeads, NoValue, emitHeldBead, emitInputBeads,
// emitRefillSlide. See interior_stream.go for the interiorStream I/O type and
// port_geom_emit.go for the port-geometry helpers; builders.go keeps the
// reflection-driven port-manifest/node-construction half.

package Wiring

import (
	"context"
	"fmt"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

// emitNodeBeads streams node 1's interior 2x2 buffer as a 4-SLOT SNAPSHOT: one
// node-bead event per fixed slot (rows {0,1} × cols {0,1}). The event's x/y/z
// carry the NODE-LOCAL OFFSET (interiorSlotOffset, relative to the node center —
// NOT a world position); TS renders each bead as a child of the node group, so
// the node center is composed by the scene graph and the beads ride the node on
// move (no re-emit needed). backup is the top row (row 0), working is the bottom
// row (row 1); a slot is PRESENT when its row's slice is at least col+1 long,
// ABSENT (popped) otherwise. Absent slots are emitted with present=false (and
// value 0) so TS can clear them — absence can't be rendered, but an explicit empty
// slot can. Discrete positions only (beads snap to slots; no slide yet). Called from
// the node's injected EmitNodeBeads closure whenever the arrays change. Offsets are
// node-local, so no node geometry is needed.
func emitNodeBeads(tr *T.Trace, nodeName string, working, backup []int, stream *interiorStream) {
	const cols = 2
	present := make([]uint8, 0, 4)
	value := make([]int32, 0, 4)
	ox, oy, oz := make([]float32, 0, 4), make([]float32, 0, 4), make([]float32, 0, 4)
	nodeRow := int32(-1)
	if stream != nil {
		nodeRow = stream.nodeRow
	}
	var events []wire.RowEvent
	emitRow := func(row int, slice []int) {
		for col := 0; col < cols; col++ {
			p := interiorSlotOffset(row, col)
			has := col < len(slice)
			v := 0
			if has {
				v = slice[col]
			}
			events = append(events, wire.RowEvent{
				Kind: T.KindNodeBead, NodeRow: nodeRow, Slot: int32(row*cols + col), Value: int32(v),
				PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1,
				X: p.X, Y: p.Y, Z: p.Z,
			})
			present = append(present, boolU8(has))
			value = append(value, int32(v))
			ox, oy, oz = append(ox, float32(p.X)), append(oy, float32(p.Y)), append(oz, float32(p.Z))
		}
	}
	emitRow(0, backup)  // top row = backup
	emitRow(1, working) // bottom row = working
	stream.write(present, value, ox, oy, oz, events)
}

// emitHeldBead streams the Time node's interior as a SINGLE centered
// bead (row 0, col 0) at the node center (offset 0,0,0). The bead is PRESENT when
// NoValue is the sentinel meaning "no value yet" / "no real bead". Real values
// are non-negative indices so NoValue (-1) never collides with a legitimate
// value. Lives here (not gatecommon) because gatecommon imports Wiring —
// gatecommon.NoValue aliases THIS constant, not the reverse, so every package
// that needs the sentinel (including this one, which cannot import gatecommon)
// shares one definition.
const NoValue = -1

// held != NoValue and colored by the held value (0 = white, 1 = black per the
// existing node-bead convention); held == NoValue (no value seen yet) →
// present=false so the interior renders empty. Called from the node's injected
// EmitHeldBead closure only when the held value changes.
func emitHeldBead(tr *T.Trace, nodeName string, held int, stream *interiorStream) {
	has := held != NoValue
	// Only slot (0,0) is meaningful for a Time node; the remaining 3 fixed
	// slots stay absent, matching the fd-3 Interior block's convention for this kind
	// (writeInteriorBlock reads n.interior[slot], and only slot 0 was ever set here).
	v := 0
	if has {
		v = held
	}
	nodeRow := int32(-1)
	if stream != nil {
		nodeRow = stream.nodeRow
	}
	events := []wire.RowEvent{{
		Kind: T.KindNodeBead, NodeRow: nodeRow, Slot: 0, Value: int32(v),
		PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1,
	}}
	// DIAGNOSTIC (task/held-bead-diagnostic): what this node is actually holding as it
	// writes its interior bead, and whether the renderer can paint that value at all
	// (InteriorBeadInstances.tsx draws 0 and 1 only). Rides THIS frame's own EVENTS section,
	// like every other breadcrumb that reaches the .probe logs — a bare tr.Breadcrumb call
	// would go to a sink production does not wire, which is why the first attempt at this
	// logged nothing. Sparse: emitHeldBead runs when a held value CHANGES, not per tick.
	events = append(events, wire.RowEvent{
		Kind: T.KindBreadcrumb, Label: T.BreadcrumbHeldBead, Debug: 1,
		NodeRow: nodeRow, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Value: int32(held),
		Text:  fmt.Sprintf("node=%s held=%d present=%v paintable=%v", nodeName, held, has, held == 0 || held == 1),
	})
	stream.write(
		[]uint8{boolU8(has), 0, 0, 0},
		[]int32{int32(v), 0, 0, 0},
		[]float32{0, 0, 0, 0}, []float32{0, 0, 0, 0}, []float32{0, 0, 0, 0},
		events,
	)
}

// emitInputBeads streams a gate's two held inputs as interior beads: the LEFT
// input on the left of the node (negative x), the RIGHT input on the right
// (positive x), vertically centered. NoValue = not held → present=false. Slot
// keys (0,0)=left, (0,1)=right. Offsets use interiorSlot so they sit inside the
// sphere.
func emitInputBeads(tr *T.Trace, nodeName string, left, right int, stream *interiorStream) {
	s := interiorSlot
	hasL, hasR := left != NoValue, right != NoValue
	vL, vR := 0, 0
	if hasL {
		vL = left
	}
	if hasR {
		vR = right
	}
	nodeRow := int32(-1)
	if stream != nil {
		nodeRow = stream.nodeRow
	}
	events := []wire.RowEvent{
		{Kind: T.KindNodeBead, NodeRow: nodeRow, Slot: 0, Value: int32(vL), PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, X: -s},
		{Kind: T.KindNodeBead, NodeRow: nodeRow, Slot: 1, Value: int32(vR), PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, X: s},
	}
	stream.write(
		[]uint8{boolU8(hasL), boolU8(hasR), 0, 0},
		[]int32{int32(vL), int32(vR), 0, 0},
		[]float32{float32(-s), float32(s), 0, 0}, []float32{0, 0, 0, 0}, []float32{0, 0, 0, 0},
		events,
	)
}

// emitRefillSlide runs the clock-paced animated refill for the Input node's
// interior buffer: the OLD backup row (row 0, top) slides DOWN into the working
// row (row 1, bottom) at human speed (the same wire-bead pulse speed), so a paused
// clock freezes the slide just like every wire. beads is the OLD backup contents
// that are becoming the new working row.
//
// Geometry: each bead animates from its row-0 slot offset to its row-1 slot offset
// — a downward translation of rowPitch = row0.y − row1.y in local y. Duration at
// human speed = rowPitch / PulseSpeedWuPerTick ticks. The loop steps t=0 to t=1
// one cycle per SleepCycle (pause-aware — Tick() freezes under Halt). Each frame:
//   - row 1, every col: present, value = beads[col], offset = lerp(row0,row1,t)
//     (keyed to the DESTINATION bottom slot, sliding down from the top position).
//   - row 0, every col: present=false (the top row is empty during the slide).
//
// At t=1 the bottom beads sit exactly at their row-1 offset.
//
// speedCh is the SAME per-goroutine speed channel the caller's own paced loop
// polls (per-goroutine-clock.md "Delivery"). This loop is a SEPARATE blocking
// loop from the caller's (it is not just one iteration of the caller's flat
// loop — it runs its own SleepCycle cycles until the slide lands), so it must
// poll ApplySpeedNonBlocking itself each cycle; without this a speed change
// sent mid-slide sits unapplied in the channel until the slide finishes and
// the caller's own loop resumes and drains it one cycle later — the in-node
// animation would run at the OLD speed for its entire duration regardless of
// the slider (the bug this fixes).
func emitRefillSlide(ctx context.Context, tr *T.Trace, nodeName string, clk wire.Clock, speedCh <-chan float64, beads []int) {
	if clk == nil || len(beads) == 0 {
		return
	}
	row0Y := interiorSlotOffset(0, 0).Y
	row1Y := interiorSlotOffset(1, 0).Y
	rowPitch := row0Y - row1Y // downward translation distance (local y, positive)
	// Slide runs at the base pulse speed — the same constant speed as the wire
	// beads; the clock is still pause-aware. Duration is a tick count.
	durationTicks := rowPitch / wire.PulseSpeedWuPerTick

	start := clk.Tick()
	emitFrame := func(t float64) {
		for col := 0; col < len(beads); col++ {
			a := interiorSlotOffset(0, col)
			b := interiorSlotOffset(1, col)
			tr.NodeBead(nodeName, 1, col, true, beads[col],
				a.X+(b.X-a.X)*t, a.Y+(b.Y-a.Y)*t, a.Z+(b.Z-a.Z)*t)
		}
		for col := 0; col < len(beads); col++ {
			p := interiorSlotOffset(0, col)
			tr.NodeBead(nodeName, 0, col, false, 0, p.X, p.Y, p.Z)
		}
	}

	emitFrame(0) // initial frame: beads at the top, top row cleared
	for {
		wire.ApplySpeedNonBlocking(clk, speedCh)
		if err := clk.SleepCycle(ctx); err != nil {
			return
		}
		t := float64(clk.Tick()-start) / durationTicks
		if t >= 1 {
			emitFrame(1) // land exactly on the bottom row
			return
		}
		emitFrame(t)
	}
}
