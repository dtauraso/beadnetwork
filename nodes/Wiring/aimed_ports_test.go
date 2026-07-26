package Wiring

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// aimedSrc / aimedSink / aimedPacer are minimal node kinds used as fixtures by other tests
// (the lock cascade). The aimed-port registry itself is gone (edges run node-to-node), and
// position writes route through nodeMover's own goroutine (node_move.go), so these are just
// plain kinds now — no layout plumbing to drain.
type aimedSrc struct {
	Out        *wire.Out
	FeedbackIn *wire.In
}

func (n *aimedSrc) Update(ctx context.Context) {
	<-ctx.Done()
}

type aimedSink struct {
	In *wire.In
}

func (n *aimedSink) Update(ctx context.Context) {
	<-ctx.Done()
}

type aimedPacer struct {
	FromSrc  *wire.In
	Feedback *wire.Out
}

func (n *aimedPacer) Update(ctx context.Context) {
	<-ctx.Done()
}

func init() {
	wire.Register("AimedSrc", func() any { return &aimedSrc{} })
	wire.Register("AimedSink", func() any { return &aimedSink{} })
	wire.Register("AimedPacer", func() any { return &aimedPacer{} })
}
