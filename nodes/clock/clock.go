package clock

import (
	"context"
	"time"
)

const MsPerTick = 16

const tickPeriod = MsPerTick * time.Millisecond

const TickPeriod = tickPeriod

type Clock interface {
	Tick() int64

	SleepCycle(ctx context.Context) error

	// SleepPulse waits ONE pulse whatever the speed, where SleepCycle waits ceil(1/speed) of
	// them. A job that is not pacing beads waits this one: speed belongs to bead traversal,
	// and a goroutine that reads input or streams geometry on a speed-scaled sleep makes the
	// bead rate the interaction rate.
	SleepPulse(ctx context.Context) error

	SleepUntilTick(ctx context.Context, target int64) error

	Copy() Clock
}
