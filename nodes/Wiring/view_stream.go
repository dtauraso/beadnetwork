// view_stream.go — thin delegators onto md.UI (nodes/Wiring/viewstate), which now owns the
// VIEW stream's own write side (docs/planning/gesture-actor.md's lift). The camera/overlay/
// scene-sphere/selection/hover state these route through is UIState's own
// (viewstate/ui_state.go); the frame-packing logic is viewstate/view_stream.go.
// SetViewStream was deleted (the payoff commit): md.UI is now exported, so its only
// out-of-package caller (runtopology/view_stream.go) calls md.UI.SetViewStream(...)
// directly. EmitBreadcrumb stays a thin delegator — its own in-package callers
// (stdin_dispatch.go) and its out-of-package caller (runtopology/startup_report.go) are
// both left calling md.EmitBreadcrumb(...) unchanged; emitViewFrame stays too, since it is
// unexported and its ~30 in-package call sites (md.emitViewFrame(...)) would otherwise all
// need editing for a purely mechanical rename with no export-blocking to relieve.
package Wiring

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// EmitBreadcrumb writes ev as a structured Breadcrumb event on the VIEW stream. See
// viewstate.UIState.EmitBreadcrumb.
func (md *MoveDispatch) EmitBreadcrumb(ev wire.RowEvent) {
	md.UI.EmitBreadcrumb(ev)
}

// emitViewFrame packs and writes the current camera/overlay/scene-sphere state as this
// goroutine's own VIEW frame. See viewstate.UIState.EmitViewFrame.
func (md *MoveDispatch) emitViewFrame(events []wire.RowEvent) {
	md.UI.EmitViewFrame(events)
}
