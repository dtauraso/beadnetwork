// fixture_kinds_test.go — shared minimal source/sink node kinds used across the Wiring
// package's tests: SrcNode is a one-Out source, SinkNode a one-In sink. Every test that
// uses them wires a SINGLE edge into the sink's one port (not fan-in). Registered once here
// because ~10 test topologies reference the "SrcNode"/"SinkNode" type strings.

package Wiring

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// srcNode is a minimal source kind with one paced Out. Position writes route through
// nodeMover's own goroutine (node_move.go), so no layout plumbing is needed here.
type srcNode struct {
	Out *wire.Out
}

func (n *srcNode) Update(ctx context.Context) {
	<-ctx.Done()
}

// sinkNode is a minimal sink kind with one paced In.
type sinkNode struct {
	In *wire.In
}

func (n *sinkNode) Update(ctx context.Context) {
	<-ctx.Done()
}

// Fixture kinds self-register exactly like production kinds do (RegisterBuilder), so the
// tests exercise the SAME construction path the loader uses rather than a parallel one.
func init() {
	RegisterBuilder("SrcNode",
		[]PortSpec{{Name: "Out", Dir: PortOut}},
		func(a BuildArgs) (wire.Node, error) {
			n := &srcNode{}
			n.Out = a.Out("Out")
			return n, nil
		})
	RegisterBuilder("SinkNode",
		[]PortSpec{{Name: "In", Dir: PortIn}},
		func(a BuildArgs) (wire.Node, error) {
			n := &sinkNode{}
			n.In = a.In("In")
			return n, nil
		})
}
