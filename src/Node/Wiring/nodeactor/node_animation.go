package nodeactor

import (
	"context"
	"io"

	"github.com/dtauraso/wirefold/src/Node/Wiring/nodeactor/owners"
	"github.com/dtauraso/wirefold/src/Node"
)

type NodeAnimation struct {
	id string

	outs   Node.Animation
	clocks owners.Clocks
}

func (a *NodeAnimation) AddBeadRun(pw *Node.BeadRun, edgeRow int32) {
	a.outs.AddBeadRun(pw, edgeRow)
}

func (a *NodeAnimation) SetBeadStream(w io.Writer, nodeRow int32, buildBeadFrame Node.BeadFrameBuilder) {
	a.outs.SetBeadStream(w, nodeRow, buildBeadFrame)
}

func (a *NodeAnimation) ClearBeadRuns() {
	a.outs.ClearBeadRuns()
}

func (a *NodeAnimation) SetSleepCh(ch <-chan int64) {
	a.outs.SetSleepCh(ch)
}

func (a *NodeAnimation) StartAnimation(ctx context.Context) {
	go a.outs.RunAnimation(ctx)
}
