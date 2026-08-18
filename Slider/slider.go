package Slider

import (
	lattice "github.com/dtauraso/wirefold/nodes/bead/lattice"
	"github.com/dtauraso/wirefold/nodes/clock"
)

const NumScale = 4

type Sinks struct {
	Clocks []chan float64
	Anim   []chan int64
}

func SleepMs(num, clockDivisor int64) int64 {
	if clockDivisor < 1 {
		clockDivisor = 1
	}
	atOne := int64(lattice.PulsesPerSlot) * clock.MsPerTick * NumScale * clockDivisor
	if num < 1 {

		return atOne * clock.PausedCycleMultiple
	}
	return atOne / num
}

func Broadcast(sinks Sinks, num, clockDivisor int64) {
	if clockDivisor < 1 {
		clockDivisor = 1
	}
	speed := float64(num) / float64(NumScale*clockDivisor)
	for _, ch := range sinks.Clocks {
		clock.SendSpeedNonBlocking(ch, speed)
	}

	ms := SleepMs(num, clockDivisor)
	for _, ch := range sinks.Anim {
		clock.SendSleepMsNonBlocking(ch, ms)
	}
}
