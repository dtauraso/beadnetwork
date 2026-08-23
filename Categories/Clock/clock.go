package clock

import "context"

type Clock interface {
	Tick() int64

	SleepCycle(ctx context.Context) error

	SleepPulse(ctx context.Context) error

	WakeOn(wake <-chan struct{})

	Copy() Clock
}
