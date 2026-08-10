package interior

import (
	"io"
	"testing"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

// nodeBeadSnapshot captures one InteriorStream.write() call's full 4-slot arrays plus
// its row-resolved RowEvents, so a test can assert on EmitNodeBeads/EmitHeldBead/
// EmitInputBeads' output without a real fd. Slot index == row*2+col (see
// InteriorStream's doc comment).
type nodeBeadSnapshot struct {
	present    []uint8
	value      []int32
	ox, oy, oz []float32
	events     []wire.RowEvent
}

func captureInteriorSnapshot(snap *nodeBeadSnapshot) *InteriorStream {
	return &InteriorStream{
		out: io.Discard,
		buildFrame: func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, evs []wire.RowEvent) []byte {
			snap.present, snap.value = present, value
			snap.ox, snap.oy, snap.oz = ox, oy, oz
			snap.events = evs
			return nil
		},
	}
}

// TestEmitNodeBeadsPositions verifies that EmitNodeBeads streams a 4-SLOT SNAPSHOT
// (all rows {0,1} × cols {0,1}) with positions matching interiorSlotPos for each
// (row,col): top row (0) = backup, bottom row (1) = working. A popped/absent slot is
// present=false (not omitted) so TS can clear it. Always a 4-element snapshot.
func TestEmitNodeBeadsPositions(t *testing.T) {
	// Full state: working=[1,0], backup=[1,0] → 4 present slots.
	tr := T.New()
	var snap nodeBeadSnapshot
	EmitNodeBeads(tr, "in", []int{1, 0}, []int{1, 0}, captureInteriorSnapshot(&snap))
	if len(snap.present) != 4 {
		t.Fatalf("full state: got %d slots, want 4", len(snap.present))
	}
	if len(snap.events) != 4 {
		t.Fatalf("full state: got %d node-bead events, want 4", len(snap.events))
	}

	// Slot index == row*2+col: 0=(0,0) 1=(0,1) 2=(1,0) 3=(1,1). Position asserted from
	// the RowEvent's float64 X/Y/Z (not the ox/oy/oz float32 arrays, which are lossy).
	byslot := map[int32]wire.RowEvent{}
	for _, e := range snap.events {
		byslot[e.Slot] = e
	}
	wantVal := []int32{1, 0, 1, 0}
	for slot := 0; slot < 4; slot++ {
		row, col := slot/2, slot%2
		if snap.present[slot] == 0 {
			t.Errorf("slot (%d,%d): present=false, want true", row, col)
		}
		if snap.value[slot] != wantVal[slot] {
			t.Errorf("slot (%d,%d): value=%d, want %d", row, col, snap.value[slot], wantVal[slot])
		}
		p := InteriorSlotOffset(row, col)
		e, ok := byslot[int32(slot)]
		if !ok {
			t.Errorf("slot (%d,%d): no RowEvent", row, col)
			continue
		}
		if e.X != p.X || e.Y != p.Y || e.Z != p.Z {
			t.Errorf("slot (%d,%d): pos=(%v,%v,%v), want (%v,%v,%v)", row, col, e.X, e.Y, e.Z, p.X, p.Y, p.Z)
		}
	}
	for _, e := range snap.events {
		if e.Kind != T.KindNodeBead {
			t.Fatalf("unexpected event kind: %+v", e)
		}
	}

	// After one pop: working=[1] (end 0 removed). Still a 4-slot snapshot, but the
	// working col-1 slot (row=1,col=1 → slot 3) is now present=false; the other 3
	// are present=true.
	tr2 := T.New()
	var snap2 nodeBeadSnapshot
	EmitNodeBeads(tr2, "in", []int{1}, []int{1, 0}, captureInteriorSnapshot(&snap2))
	if len(snap2.present) != 4 {
		t.Fatalf("after pop: got %d slots, want 4 (snapshot)", len(snap2.present))
	}
	for slot := 0; slot < 4; slot++ {
		emptySlot := slot == 3
		present := snap2.present[slot] != 0
		if emptySlot && present {
			t.Errorf("popped slot (1,1): present=true, want false")
		}
		if !emptySlot && !present {
			row, col := slot/2, slot%2
			t.Errorf("slot (%d,%d): present=false, want true", row, col)
		}
	}
}

// TestInteriorSlotOffsetFormula pins the torus-aware slotOffset(row,col) formula —
// a NODE-LOCAL offset centered at the origin (no node center added):
//
//	slot = InteriorTorusOuterR + interiorBeadGap/2 ; pitch = 2*slot
//	dx = (col-0.5)*pitch ; dy = (0.5-row)*pitch ; dz = 0
//
// Pitch follows bead size, not node radius.
//
// The sphere-fit tests (torus reach must stay inside the node's own sphere radius) live
// in package Wiring's interior_sphere_test.go instead of here: they need Wiring's own
// nodeRadius, and this package must not import Wiring (Wiring imports this package, for
// InteriorStream/EmitNodeBeads/etc — importing back would cycle).
func TestInteriorSlotOffsetFormula(t *testing.T) {
	pitch := 2 * InteriorSlot

	cases := []struct{ row, col int }{{0, 0}, {0, 1}, {1, 0}, {1, 1}}
	for _, tc := range cases {
		got := InteriorSlotOffset(tc.row, tc.col)
		wantX := (float64(tc.col) - 0.5) * pitch
		wantY := (0.5 - float64(tc.row)) * pitch
		if got.X != wantX || got.Y != wantY || got.Z != 0 {
			t.Errorf("slot(%d,%d) = (%v,%v,%v), want (%v,%v,0)", tc.row, tc.col, got.X, got.Y, got.Z, wantX, wantY)
		}
	}
}

// TestInteriorTorusesDoNotOverlap asserts adjacent same-row/col toruses keep a
// non-negative gap: pitch (2*slot) ≥ 2*rt, i.e. torus-to-torus gap ≥ 0.
func TestInteriorTorusesDoNotOverlap(t *testing.T) {
	pitch := 2 * InteriorSlot
	gap := pitch - 2*InteriorTorusOuterR
	if gap < 0 {
		t.Errorf("adjacent toruses overlap: pitch %v < 2*rt %v (gap %v)", pitch, 2*InteriorTorusOuterR, gap)
	}
}
