package nodeactor

import (
	"context"
	"io"

	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/owners"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

type NodeAnimation struct {
	id string

	outs   owners.Outs
	clocks owners.Clocks
}

func (a *NodeAnimation) AddOutWire(pw *wire.PacedWire, edgeRow int32) {
	a.outs.AddOutWire(pw, edgeRow)
}

func (a *NodeAnimation) SetBeadStream(w io.Writer, nodeRow int32, buildBeadFrame owners.BeadFrameBuilder) {
	a.outs.SetBeadStream(w, nodeRow, buildBeadFrame)
}

func (a *NodeAnimation) ClearOutWires() {
	a.outs.ClearOutWires()
}

func (a *NodeAnimation) StartAnimation(ctx context.Context) {
	go a.outs.RunAnimation(ctx)
}
