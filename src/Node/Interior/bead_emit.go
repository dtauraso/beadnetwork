package interior

import (
	"context"

	"github.com/dtauraso/wirefold/src/Clock"
	lattice "github.com/dtauraso/wirefold/src/Node/wire/lattice"
)

func EmitNodeBeads(nodeName string, working, backup []int, emitter *Emitter) {
	const cols = 2
	present := make([]uint8, 0, 4)
	value := make([]int32, 0, 4)
	ox, oy, oz := make([]float32, 0, 4), make([]float32, 0, 4), make([]float32, 0, 4)
	emitRow := func(row int, slice []int) {
		for col := 0; col < cols; col++ {
			p := InteriorSlotOffset(row, col)
			has := col < len(slice)
			v := 0
			if has {
				v = slice[col]
			}
			present = append(present, boolU8(has))
			value = append(value, int32(v))
			ox, oy, oz = append(ox, float32(p.X)), append(oy, float32(p.Y)), append(oz, float32(p.Z))
		}
	}
	emitRow(0, backup)
	emitRow(1, working)
	emitter.write(present, value, ox, oy, oz, nil)
}

const NoValue = -1

func EmitHeldBead(nodeName string, held int, emitter *Emitter) {
	has := held != NoValue

	v := 0
	if has {
		v = held
	}
	emitter.write(
		[]uint8{boolU8(has), 0, 0, 0},
		[]int32{int32(v), 0, 0, 0},
		[]float32{0, 0, 0, 0}, []float32{0, 0, 0, 0}, []float32{0, 0, 0, 0},
		nil,
	)
}

func EmitInputBeads(nodeName string, left, right int, emitter *Emitter) {
	s := InteriorSlot
	hasL, hasR := left != NoValue, right != NoValue
	vL, vR := 0, 0
	if hasL {
		vL = left
	}
	if hasR {
		vR = right
	}
	emitter.write(
		[]uint8{boolU8(hasL), boolU8(hasR), 0, 0},
		[]int32{int32(vL), int32(vR), 0, 0},
		[]float32{float32(-s), float32(s), 0, 0}, []float32{0, 0, 0, 0}, []float32{0, 0, 0, 0},
		nil,
	)
}

func EmitRefillSlide(ctx context.Context, nodeName string, clk clock.Clock, speedCh <-chan float64, beads []int) {
	if clk == nil || len(beads) == 0 {
		return
	}
	row0Y := InteriorSlotOffset(0, 0).Y
	row1Y := InteriorSlotOffset(1, 0).Y
	rowPitch := row0Y - row1Y

	durationTicks := rowPitch / lattice.PulseSpeedWuPerTick

	start := clk.Tick()
	emitFrame := func(t float64) {
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
