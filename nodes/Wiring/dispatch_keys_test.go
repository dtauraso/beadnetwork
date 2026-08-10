package Wiring

import (
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
)

// TestDispatchTableKeysMatchFingerprint guards the table-driven dispatchers introduced when
// several switches were converted to maps: nothing else checks that a hardcoded string key in
// one of these tables is a real member of the wire vocabulary pinned by
// InputLayoutFingerprint. A typo'd/stale key silently no-ops — worse than the switch it
// replaced, which at least had every case visible next to the enum. Each table's keys must be
// a SUBSET of its corresponding fingerprint-derived enum list; an enum value with no handler is
// legal (forward-compat / not-yet-wired), but a handler key absent from the enum is a bug.
func TestDispatchTableKeysMatchFingerprint(t *testing.T) {
	checks := []struct {
		tableName string
		keys      []string
		enumName  string
		enum      []string
	}{
		{"rawInputHandlers", mapKeys(rawInputHandlers), "InEventKinds", inputcodec.InEventKinds},
		{"hitClassifiers", mapKeys(hitClassifiers), "InHitKinds", inputcodec.InHitKinds},
		{"updateKindHandlers", mapKeys(updateKindHandlers), "InUpdateKinds", inputcodec.InUpdateKinds},
		{"clockAttrHandlers", mapKeys(clockAttrHandlers), "InUpdateAttrs", inputcodec.InUpdateAttrs},
		{"overlayAttrHandlers", mapKeys(overlayAttrHandlers), "InUpdateAttrs", inputcodec.InUpdateAttrs},
		// overlayFlagTraceKind maps each overlay flag NAME → its trace kind. The values are
		// compile-checked (T.Kind* consts), but the string keys can drift from the flag list;
		// guard them here (a key absent from InOverlayFlags is a typo'd/stale flag name).
		{"overlayFlagTraceKind", mapKeys(overlayFlagTraceKind), "InOverlayFlags", inputcodec.InOverlayFlags},
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
// caring about the value type.
func mapKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
