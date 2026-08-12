package scene

import (
	"path/filepath"

	B "github.com/dtauraso/wirefold/Buffer"
)

func SceneUsesQuantizedDrag(scenePath string) bool {
	base := filepath.Base(filepath.Clean(scenePath))
	for _, t := range SceneTabs {
		if t.Dir == base {
			return t.QuantizedDrag
		}
	}
	return true
}

func SceneWantsCoplanarEdges(scenePath string) bool {
	base := filepath.Base(filepath.Clean(scenePath))
	for _, t := range SceneTabs {
		if t.Dir == base {
			return t.CoplanarEdges
		}
	}
	return false
}

func SceneWantsUpAxis(scenePath string) bool {
	base := filepath.Base(filepath.Clean(scenePath))
	for _, t := range SceneTabs {
		if t.Dir == base {
			return t.UpAxis
		}
	}
	return false
}

func SceneClockDivisor(scenePath string) float64 {
	base := filepath.Base(filepath.Clean(scenePath))
	for _, t := range SceneTabs {
		if t.Dir == base {
			return t.ClockDivisor
		}
	}
	return 1
}

func SceneHasDistanceGroups(scenePath string) bool {
	base := filepath.Base(filepath.Clean(scenePath))
	for _, t := range SceneTabs {
		if t.Dir == base {
			return t.DistanceGroups
		}
	}
	return false
}

func SceneIsEditable(scenePath string) bool {
	base := filepath.Base(filepath.Clean(scenePath))
	for _, t := range SceneTabs {
		if t.Dir == base {
			return t.Editable
		}
	}
	return false
}

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
