// Package selectionstate holds SelectionState: the CURRENTLY-SELECTED (click-select) and
// CURRENTLY-HOVERED (pointer hover) fields, pure UI state parked on the routing directory
// (Wiring.MoveDispatch). It is owned as a field by MoveDispatch (md.ui.sel); there is no
// goroutine — the gesture FSM mutates it directly, serialized by the single-goroutine
// stdin reader.
package selectionstate

// SelectionState carries the current click-selection and pointer-hover state.
type SelectionState struct {
	// Selected is the CURRENTLY-SELECTED node id (click-select), owned by Go. "" = nothing
	// selected. Set by the gesture FSM's click outcome (applySelect) and emitted via
	// KindSelect so the buffer snapshot marks the node's Selected column.
	Selected string
	// SelectedEdge is the CURRENTLY-SELECTED edge label (click-select), owned by Go. "" =
	// no edge selected. Set by the gesture FSM's click outcome (applySelect) and emitted via
	// KindSelect (Edge field) so the buffer snapshot marks the edge's Selected column.
	// Exclusive with `Selected`: selecting an edge clears the node selection and vice versa.
	SelectedEdge string
	// HoverNode / HoverPort / HoverInput are the CURRENTLY-HOVERED entity (pointer hover),
	// owned by Go. The gesture FSM tracks them from the raycast hit on each pointer-move and
	// emits KindHover ONLY when they change (dedupe) so pointer-move doesn't flood the
	// snapshot. HoverPort != "" means a port is hovered (on HoverNode); otherwise HoverNode
	// (possibly "") is the hovered node. "" / "" = nothing hovered.
	HoverNode  string
	HoverPort  string
	HoverInput bool
}
