package Gesture

import (
	"context"

	"github.com/dtauraso/beadnetwork/Categories/Node/Messaging"
	"github.com/dtauraso/beadnetwork/Categories/Scene/View"
)

func setSelectionUI(ui *View.UIState, mr Movers, ctx context.Context, node, edge string) {
	sendMoveFn := func(id string, msg Messaging.Msg) { mr.SendMove(ctx, id, msg) }
	ui.SetSelectionUI(sendMoveFn, node, edge)
}
