package SliderPanel

import (
	clock "github.com/dtauraso/wirefold/Categories/Clock"
	lattice "github.com/dtauraso/wirefold/Categories/Node/BeadAnimation/lattice"
)

const Paused int64 = 0

type Sinks struct {
	Clocks []chan float64
	Anim   []chan int64
}

func SleepMs(num, clockDivisor int64) int64 {
	if clockDivisor < 1 {
		clockDivisor = 1
	}
	if num < 1 {

		return Paused
	}
	return int64(lattice.PulsesPerSlot) * clock.MsPerTick * clock.SpeedNumScale * clockDivisor / num
}

func (s Sinks) SendSpeed(num, clockDivisor int64) { Broadcast(s, num, clockDivisor) }

func Broadcast(sinks Sinks, num, clockDivisor int64) {
	if clockDivisor < 1 {
		clockDivisor = 1
	}
	speed := float64(num) / float64(clock.SpeedNumScale*clockDivisor)
	for _, ch := range sinks.Clocks {
		clock.SendSpeedNonBlocking(ch, speed)
	}

	ms := SleepMs(num, clockDivisor)
	for _, ch := range sinks.Anim {
		clock.SendSleepMsNonBlocking(ch, ms)
	}
}
