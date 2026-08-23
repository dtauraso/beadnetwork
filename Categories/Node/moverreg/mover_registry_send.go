package moverreg

import (
	"context"

	"github.com/dtauraso/wirefold/Categories/Node/nodeactor"
	"github.com/dtauraso/wirefold/Categories/Node/nodeactor/owners"
)

func (mr *MoverRegistry) SendMove(ctx context.Context, id string, msg owners.Msg) {
	nm, ok := mr.nodeGeoms[id]
	if !ok {
		return
	}
	nm.SendExternal(ctx, msg)
}

func (mr *MoverRegistry) EnqueueFor(nm *nodeactor.NodeGeometry) func(id string, msg owners.Msg) {
	return nm.EnqueueSend
}
