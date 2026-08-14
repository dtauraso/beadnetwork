package gesture

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/moverreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

func setSelectionUI(ui *viewstate.UIState, mr *moverreg.MoverRegistry, ctx context.Context, node, edge string) {
	sendMoveFn := func(id string, msg movemsg.Msg) { mr.SendMove(ctx, id, msg) }
	ui.SetSelectionUI(sendMoveFn, node, edge)
}
