package gesture

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/edgemover"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/moverreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

func setSelectionUI(ui *viewstate.UIState, mr *moverreg.MoverRegistry, ctx context.Context, node, edge string) {
	sendMoveFn := func(id string, msg movemsg.Msg) { mr.SendMove(ctx, id, msg) }
	sendEdgeSelectFn := func(label string, on bool) { sendEdgeSelect(mr.EdgeMovers(), ctx, label, on) }
	ui.SetSelectionUI(sendMoveFn, sendEdgeSelectFn, node, edge)
}

func sendEdgeSelect(edgeMovers map[string]*edgemover.EdgeMover, ctx context.Context, label string, on bool) {
	em, ok := edgeMovers[label]
	if !ok {
		return
	}
	em.Select(ctx, on)
}
