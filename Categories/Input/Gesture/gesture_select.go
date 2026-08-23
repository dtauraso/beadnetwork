package Gesture

import (
	"context"

	"github.com/dtauraso/wirefold/Categories/Node/nodeactor/owners"
	"github.com/dtauraso/wirefold/Categories/Scene/viewstate"
)

func setSelectionUI(ui *viewstate.UIState, mr Movers, ctx context.Context, node, edge string) {
	sendMoveFn := func(id string, msg owners.Msg) { mr.SendMove(ctx, id, msg) }
	ui.SetSelectionUI(sendMoveFn, node, edge)
}
