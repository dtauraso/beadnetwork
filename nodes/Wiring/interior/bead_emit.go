package interior

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/rowevent"
	"github.com/dtauraso/wirefold/nodes/clock"
	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"

	T "github.com/dtauraso/wirefold/Trace"
)

func EmitNodeBeads(tr *T.Trace, nodeName string, working, backup []int, stream *InteriorStream) {
	const cols = 2
	present := make([]uint8, 0, 4)
	value := make([]int32, 0, 4)
	ox, oy, oz := make([]float32, 0, 4), make([]float32, 0, 4), make([]float32, 0, 4)
	nodeRow := int32(-1)
	if stream != nil {
		nodeRow = stream.nodeRow
	}
	var events []rowevent.RowEvent
	emitRow := func(row int, slice []int) {
		for col := 0; col < cols; col++ {
			p := InteriorSlotOffset(row, col)
			has := col < len(slice)
			v := 0
			if has {
				v = slice[col]
			}
			events = append(events, rowevent.RowEvent{
				Kind: T.KindNodeBead, NodeRow: nodeRow, Slot: int32(row*cols + col), Value: int32(v),
				PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1,
				X: p.X, Y: p.Y, Z: p.Z,
			})
			present = append(present, boolU8(has))
			value = append(value, int32(v))
			ox, oy, oz = append(ox, float32(p.X)), append(oy, float32(p.Y)), append(oz, float32(p.Z))
		}
	}
	emitRow(0, backup)
	emitRow(1, working)
	stream.write(present, value, ox, oy, oz, events)
}

const NoValue = -1

func EmitHeldBead(tr *T.Trace, nodeName string, held int, stream *InteriorStream) {
	has := held != NoValue

	v := 0
	if has {
		v = held
	}
	nodeRow := int32(-1)
	if stream != nil {
		nodeRow = stream.nodeRow
	}
	events := []rowevent.RowEvent{{
		Kind: T.KindNodeBead, NodeRow: nodeRow, Slot: 0, Value: int32(v),
		PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1,
	}}
	stream.write(
		[]uint8{boolU8(has), 0, 0, 0},
		[]int32{int32(v), 0, 0, 0},
		[]float32{0, 0, 0, 0}, []float32{0, 0, 0, 0}, []float32{0, 0, 0, 0},
		events,
	)
}

func EmitInputBeads(tr *T.Trace, nodeName string, left, right int, stream *InteriorStream) {
	s := InteriorSlot
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
	events := []rowevent.RowEvent{
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

func EmitRefillSlide(ctx context.Context, tr *T.Trace, nodeName string, clk clock.Clock, speedCh <-chan float64, beads []int) {
	if clk == nil || len(beads) == 0 {
		return
	}
	row0Y := InteriorSlotOffset(0, 0).Y
	row1Y := InteriorSlotOffset(1, 0).Y
	rowPitch := row0Y - row1Y

	durationTicks := rowPitch / lattice.PulseSpeedWuPerTick

	start := clk.Tick()
	emitFrame := func(t float64) {
		for col := 0; col < len(beads); col++ {
			a := InteriorSlotOffset(0, col)
			b := InteriorSlotOffset(1, col)
			tr.NodeBead(nodeName, 1, col, true, beads[col],
				a.X+(b.X-a.X)*t, a.Y+(b.Y-a.Y)*t, a.Z+(b.Z-a.Z)*t)
		}
		for col := 0; col < len(beads); col++ {
			p := InteriorSlotOffset(0, col)
			tr.NodeBead(nodeName, 0, col, false, 0, p.X, p.Y, p.Z)
		}
	}

	emitFrame(0)
	for {
		clock.ApplySpeedNonBlocking(clk, speedCh)
		if err := clk.SleepCycle(ctx); err != nil {
			return
		}
		t := float64(clk.Tick()-start) / durationTicks
		if t >= 1 {
			emitFrame(1)
			return
		}
		emitFrame(t)
	}
}
