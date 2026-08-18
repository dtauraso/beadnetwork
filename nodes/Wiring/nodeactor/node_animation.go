package nodeactor

import (
	"context"
	"io"

	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/owners"
	"github.com/dtauraso/wirefold/nodes/bead"
)

type NodeAnimation struct {
	id string

	outs   bead.Animation
	clocks owners.Clocks
}

func (a *NodeAnimation) AddBeadRun(pw *bead.BeadRun, edgeRow int32) {
	a.outs.AddBeadRun(pw, edgeRow)
}

func (a *NodeAnimation) SetBeadStream(w io.Writer, nodeRow int32, buildBeadFrame bead.BeadFrameBuilder) {
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
