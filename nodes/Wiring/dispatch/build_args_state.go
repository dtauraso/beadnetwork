// build_args_state.go — BuildArgs methods for a node's own persisted STATE/DATA (the
// `data.state` seed accessor, and the raw spec data block for fields with no dedicated
// accessor). Split out of build_args.go — see that file's header.

package dispatch

import "github.com/dtauraso/wirefold/nodes/Wiring/loadspec"

// StateSeed returns the persisted `data.state` seed for one field, or def when the spec
// carries none. key is the struct field name with its first letter lowercased — the same
// convention the `wire:"data.state"` tag used (field Held -> key "held").
//
// The seed is OPTIONAL by design: an absent key leaves the kind's own default untouched,
// so "unset" can never collide with a legitimately-held 0.
func (a BuildArgs) StateSeed(key string, def int) int {
	if a.data == nil || a.data.State == nil {
		return def
	}
	if v, ok := a.data.State[key]; ok {
		return v
	}
	return def
}

// Data exposes the raw spec data block for the `wire:"data.<key>"` fields that have no
// dedicated accessor above. Nil when the spec carries no data block.
func (a BuildArgs) Data() *loadspec.NodeData { return a.data }
