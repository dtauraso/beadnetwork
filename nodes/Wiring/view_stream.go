// view_stream.go — thin delegators onto md.UI (nodes/Wiring/viewstate), which now owns the
// VIEW stream's own write side (docs/planning/gesture-actor.md's lift). The camera/overlay/
// scene-sphere/selection/hover state these route through is UIState's own
// (viewstate/ui_state.go); the frame-packing logic is viewstate/view_stream.go. Kept as
// thin MoveDispatch delegators — not deleted — so the many in-package call sites
// (md.emitViewFrame(...), md.EmitBreadcrumb(...)) and main.go's md.SetViewStream(...) are
// unchanged text; SetViewStream is one of the export-blocked methods a follow-up commit may
// delete once md.UI's own export makes the delegator redundant for its external caller
// (runtopology/view_stream.go).
package Wiring

import (
	"io"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

// SetViewStream installs the VIEW stream's write side. See viewstate.UIState.SetViewStream.
func (md *MoveDispatch) SetViewStream(out io.Writer, buildFrame viewstate.ViewFrameBuilder) {
	md.UI.SetViewStream(out, buildFrame)
}

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
