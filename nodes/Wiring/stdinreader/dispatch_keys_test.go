package stdinreader

import (
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

// TestDispatchTableKeysMatchFingerprint guards this package's table-driven dispatchers: a
// hardcoded string key here must be a real member of the wire vocabulary pinned by
// InputLayoutFingerprint. Moved here from nodes/Wiring/dispatch (§30,
// docs/planning/movedispatch-decomposition.md) alongside updateKindHandlers/
// clockAttrHandlers/overlayAttrHandlers themselves; the gesture cluster's own
// rawInputHandlers/hitClassifiers checks stayed in
// nodes/Wiring/dispatch/dispatch_keys_test.go.
func TestDispatchTableKeysMatchFingerprint(t *testing.T) {
	checks := []struct {
		tableName string
		keys      []string
		enumName  string
		enum      []string
	}{
		{"updateKindHandlers", mapKeys(updateKindHandlers), "InUpdateKinds", inputcodec.InUpdateKinds},
		{"clockAttrHandlers", mapKeys(clockAttrHandlers), "InUpdateAttrs", inputcodec.InUpdateAttrs},
		{"overlayAttrHandlers", mapKeys(overlayAttrHandlers), "InUpdateAttrs", inputcodec.InUpdateAttrs},
		// viewstate.OverlayFlagTraceKind maps each overlay flag NAME → its trace kind. The
		// values are compile-checked (T.Kind* consts), but the string keys can drift from the
		// flag list; guard them here (a key absent from InOverlayFlags is a typo'd/stale flag
		// name).
		{"OverlayFlagTraceKind", mapKeys(viewstate.OverlayFlagTraceKind), "InOverlayFlags", inputcodec.InOverlayFlags},
	}

	for _, c := range checks {
		valid := make(map[string]bool, len(c.enum))
		for _, v := range c.enum {
			valid[v] = true
		}
		for _, k := range c.keys {
			if !valid[k] {
				t.Errorf("%s has key %q which is not a member of %s (valid: %v)", c.tableName, k, c.enumName, c.enum)
			}
		}
	}
}

// mapKeys extracts the string keys of a map[string]V for the membership check above, without
// caring about the value type. Note: this table-key membership check is a different
// package's own copy of the same one-liner nodes/Wiring/dispatch/dispatch_keys_test.go
// keeps for its own tables (its editOps table lives there, this one's tables live here) —
// same "own trivial copy per package" precedent bool_u8.go's header already documents.
func mapKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
