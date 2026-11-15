package Drag

import (
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

type State struct {
	Index     polarindex.Index
	Constants polarindex.SceneConstants

	OutTargets []string
	OutDelta   map[string]polarindex.Offset
}

type Trim func(delta polarindex.Offset, st State) polarindex.Offset

type Request func(delta polarindex.Offset, st State) map[string]polarindex.Offset

func Apply(trim Trim, delta polarindex.Offset, st State) polarindex.Offset {
	if trim == nil {
		return delta
	}
	return trim(delta, st)
}

func Requested(request Request, delta polarindex.Offset, st State) map[string]polarindex.Offset {
	if request == nil {
		return nil
	}
	return request(delta, st)
}
