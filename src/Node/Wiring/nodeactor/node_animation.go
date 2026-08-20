package nodeactor

import (
	"context"
	"io"

	"github.com/dtauraso/wirefold/src/Node/Wiring/nodeactor/owners"
	"github.com/dtauraso/wirefold/src/Node"
)

type NodeBeadAnimation struct {
	id string

	outs   Node.BeadAnimation
	clocks owners.Clocks
}

func (a *NodeBeadAnimation) AddBeadLine(pw *Node.BeadLine, edgeRow int32) {
	a.outs.AddBeadLine(pw, edgeRow)
}

func (a *NodeBeadAnimation) SetBeadStream(w io.Writer, nodeRow int32, buildBeadFrame Node.BeadFrameBuilder) {
	a.outs.SetBeadStream(w, nodeRow, buildBeadFrame)
}

func (a *NodeBeadAnimation) ClearBeadLines() {
	a.outs.ClearBeadLines()
}

func (a *NodeBeadAnimation) SetSleepCh(ch <-chan int64) {
	a.outs.SetSleepCh(ch)
}

func (a *NodeBeadAnimation) StartBeadAnimation(ctx context.Context) {
	go a.outs.RunBeadAnimation(ctx)
}
