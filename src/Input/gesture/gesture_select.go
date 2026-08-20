package gesture

import (
	"context"

	"github.com/dtauraso/wirefold/src/Node/movemsg"
	"github.com/dtauraso/wirefold/src/Node/moverreg"
	"github.com/dtauraso/wirefold/src/Scene/viewstate"
)

func setSelectionUI(ui *viewstate.UIState, mr *moverreg.MoverRegistry, ctx context.Context, node, edge string) {
	sendMoveFn := func(id string, msg movemsg.Msg) { mr.SendMove(ctx, id, msg) }
	ui.SetSelectionUI(sendMoveFn, node, edge)
}
