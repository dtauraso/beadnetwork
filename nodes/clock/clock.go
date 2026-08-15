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

	SleepPulse(ctx context.Context) error

	WakeOn(wake <-chan struct{})

	Copy() Clock
}
