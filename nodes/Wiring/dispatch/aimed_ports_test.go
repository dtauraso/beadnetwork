package dispatch

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// aimedSrc / aimedSink / aimedPacer are minimal node kinds used as fixtures by other tests
// (the lock cascade). The aimed-port registry itself is gone (edges run node-to-node), and
// position writes route through nodeMover's own goroutine (move_dispatch_construct.go), so these are just
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

// Self-registering, like every production kind — same construction path as the loader.
func init() {
	RegisterBuilder("AimedSrc",
		[]portwiring.PortSpec{{Name: "Out", Dir: portwiring.PortOut}, {Name: "FeedbackIn", Dir: portwiring.PortIn}},
		func(a BuildArgs) (wire.Node, error) {
			n := &aimedSrc{}
			n.Out = a.Out("Out")
			n.FeedbackIn = a.In("FeedbackIn")
			return n, nil
		})
	RegisterBuilder("AimedSink",
		[]portwiring.PortSpec{{Name: "In", Dir: portwiring.PortIn}},
		func(a BuildArgs) (wire.Node, error) {
			n := &aimedSink{}
			n.In = a.In("In")
			return n, nil
		})
	RegisterBuilder("AimedPacer",
		[]portwiring.PortSpec{{Name: "FromSrc", Dir: portwiring.PortIn}, {Name: "Feedback", Dir: portwiring.PortOut}},
		func(a BuildArgs) (wire.Node, error) {
			n := &aimedPacer{}
			n.FromSrc = a.In("FromSrc")
			n.Feedback = a.Out("Feedback")
			return n, nil
		})
}
