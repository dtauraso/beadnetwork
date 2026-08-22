package nodeactor

import (
	"context"

	beadanimation "github.com/dtauraso/wirefold/src/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/src/Node/nodeactor/owners"
)

type NodeBeadAnimation struct {
	id string

	outs   beadanimation.BeadAnimation
	clocks owners.Clocks
}

func (a *NodeBeadAnimation) AddBeadLine(pw *beadanimation.BeadLine, edgeRow int32) {
	a.outs.AddBeadLine(pw, edgeRow)
}

func (a *NodeBeadAnimation) SetBeadStream(nodeRow int32, buildBeadFrame beadanimation.BeadFrameBuilder, sceneRoot string) {
	a.outs.SetBeadStream(nodeRow, buildBeadFrame, sceneRoot)
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
