// scene_capabilities.go — per-scene CAPABILITY queries: what the tree actually being LOADED
// declares about itself (SceneTab's bool/numeric fields, plus its accepted kind mask), keyed
// by the loaded scene's own directory name rather than the anchor. The tab registry lives in
// scene_tabs.go; the switch lives in Wiring's scene_switch.go (a MoveDispatch method cannot
// live in this package); anchor/selection resolution lives in scene_selection.go.
package scene

import (
	"path/filepath"

	B "github.com/dtauraso/wirefold/Buffer"
)

// SceneUsesQuantizedDrag answers, for the tree actually being LOADED, whether the node
// drag snaps to the bead lattice. It takes the loaded scene's own path (not the anchor)
// because the loader knows which tree it is opening but not which tab pointed it there.
// An unknown tree — every test fixture, every one-off run — gets the quantized drag, which
// is what every scene did before scenes were selectable.
func SceneUsesQuantizedDrag(scenePath string) bool {
	base := filepath.Base(filepath.Clean(scenePath))
	for _, t := range SceneTabs {
		if t.Dir == base {
			return t.QuantizedDrag
		}
	}
	return true
}

// SceneWantsCoplanarEdges answers, for the tree being LOADED, whether a node's ring plane
// must contain the edge leaving it (SceneTab.CoplanarEdges). Unknown trees keep the plain
// inward pole, which is what every scene had before this was a choice.
func SceneWantsCoplanarEdges(scenePath string) bool {
	base := filepath.Base(filepath.Clean(scenePath))
	for _, t := range SceneTabs {
		if t.Dir == base {
			return t.CoplanarEdges
		}
	}
	return false
}

// SceneWantsUpAxis answers whether the tree being LOADED aims its node tori and per-node
// vectors straight up (SceneTab.UpAxis). Unknown trees do not — they keep the unrotated
// ring every scene had before ring orientation existed.
func SceneWantsUpAxis(scenePath string) bool {
	base := filepath.Base(filepath.Clean(scenePath))
	for _, t := range SceneTabs {
		if t.Dir == base {
			return t.UpAxis
		}
	}
	return false
}

// SceneClockDivisor answers, for the tree being LOADED, its SceneTab.ClockDivisor. A test
// fixture or one-off tree with no tab entry gets divisor 1 (no scaling) — never 0, which
// would divide the effective speed by zero downstream.
func SceneClockDivisor(scenePath string) float64 {
	base := filepath.Base(filepath.Clean(scenePath))
	for _, t := range SceneTabs {
		if t.Dir == base {
			return t.ClockDivisor
		}
	}
	return 1
}

// SceneHasDistanceGroups answers, for the tree being LOADED, whether the three named
// distance groups apply to it (SceneTab.DistanceGroups). Unknown trees get false — see that
// field's own doc comment for why the ring's node ids must not be read against another
// scene's nodes of the same name.
func SceneHasDistanceGroups(scenePath string) bool {
	base := filepath.Base(filepath.Clean(scenePath))
	for _, t := range SceneTabs {
		if t.Dir == base {
			return t.DistanceGroups
		}
	}
	return false
}

// SceneIsEditable answers, for the tree actually being LOADED, whether it takes structural
// edits (SceneTab.Editable). An UNKNOWN tree — every test fixture, every one-off run — is
// NOT editable: a create writes directories and rewrites counts.json, so the safe answer for
// a tree nobody declared is to leave it alone.
func SceneIsEditable(scenePath string) bool {
	base := filepath.Base(filepath.Clean(scenePath))
	for _, t := range SceneTabs {
		if t.Dir == base {
			return t.Editable
		}
	}
	return false
}

// SceneKindMask is the set of kinds the loaded scene accepts, as a BITMASK over kind ids —
// bit N set means the kind whose Buffer KindId is N may be created here. An empty Kinds list
// (or an unknown tree) yields every bit set: no declared restriction restricts nothing.
//
// A mask rather than a list of names because the wire carries kind IDS, not names, and one
// integer says the whole answer. It rides the Overlay block so the palette can offer exactly
// the kinds the scene will accept, instead of offering all of them and letting Go refuse.
func SceneKindMask(scenePath string) uint32 {
	base := filepath.Base(filepath.Clean(scenePath))
	for _, t := range SceneTabs {
		if t.Dir != base {
			continue
		}
		if len(t.Kinds) == 0 {
			break
		}
		var mask uint32
		for _, k := range t.Kinds {
			if id := B.NodeKindID(k); id != B.KindIDUnknown {
				mask |= 1 << uint(id)
			}
		}
		return mask
	}
	return ^uint32(0)
}
