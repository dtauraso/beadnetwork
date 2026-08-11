package gesture

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/edgemover"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/moverreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

// setSelectionUI sets the Go-owned selection (node XOR edge, exclusive). Thin delegator to
// ui.SetSelectionUI (nodes/Wiring/viewstate), binding the two closures it needs in place of
// the *moverreg.MoverRegistry/edgeMover-map parameters that type cannot name — same
// bound-func treatment the gesture cluster has always used (moved here from
// dispatch/move_dispatch_api.go in §31, since it is used ONLY by applySelect below).
func setSelectionUI(ui *viewstate.UIState, mr *moverreg.MoverRegistry, ctx context.Context, node, edge string) {
	sendMoveFn := func(id string, msg movemsg.Msg) { mr.SendMove(ctx, id, msg) }
	sendEdgeSelectFn := func(label string, on bool) { sendEdgeSelect(mr.EdgeMovers(), ctx, label, on) }
	ui.SetSelectionUI(sendMoveFn, sendEdgeSelectFn, node, edge)
}

// sendEdgeSelect routes a select/deselect message to one edge's OWN dedicated extIn
// channel (mirrors sendMove's node counterpart) — the edgeMover sets its OWN selected
// field on its own goroutine, no shared map. A blocking send with a ctx-cancel escape
// hatch, same reasoning as sendMove. It needs the edgeMover map, so it is handed to
// UIState.SetSelectionUI as a bound func value (setSelectionUI above), the same shape it
// had in dispatch/move_dispatch_api.go before §31 moved it here.
func sendEdgeSelect(edgeMovers map[string]*edgemover.EdgeMover, ctx context.Context, label string, on bool) {
	em, ok := edgeMovers[label]
	if !ok {
		return
	}
	em.Select(ctx, on)
}
