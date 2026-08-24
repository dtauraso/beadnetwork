package Node

import (
	"context"

	beadanimation "github.com/dtauraso/beadnetwork/Categories/Node/BeadAnimation"
)

type NodeBeadAnimation struct {
	id string

	outs   beadanimation.BeadAnimation
	clocks Clocks
}

func NewNodeBeadAnimation(id string, clocks Clocks) *NodeBeadAnimation {
	return &NodeBeadAnimation{id: id, clocks: clocks}
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
