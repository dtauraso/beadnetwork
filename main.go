package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"
	"time"

	"github.com/dtauraso/wirefold/src/Node/clock"
	"github.com/dtauraso/wirefold/runtopology"
)

func Run(topologyPath string) {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	runtopology.RunTopology(ctx, cancel, topologyPath, clock.NewRealClock())
}

func RunTest(dur time.Duration, topologyPath string) {
	ctx, cancel := context.WithTimeout(context.Background(), dur)
	defer cancel()
	runtopology.RunTopology(ctx, cancel, topologyPath, clock.NewRealClock())
}

func main() {
	dur := flag.Duration("duration", 0, "if non-zero, run for this duration then exit (test mode)")
	topologyPath := flag.String("topology", "topology", "path to topology JSON spec")
	flag.Parse()
	if *dur > 0 {
		RunTest(*dur, *topologyPath)
	} else {
		Run(*topologyPath)
	}
}
