package clock

import "time"

func HumanTick(tickMs int) int64 {
	if tickMs < 1 {
		tickMs = 1
	}
	return int64(time.Since(tickEpoch) / (time.Duration(tickMs) * time.Millisecond))
}
