package input

import (
	"math"

	NodeDrag "github.com/dtauraso/beadnetwork/Categories/Node/Drag"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

func trimOwnDrag(delta polarindex.Offset, st NodeDrag.State) polarindex.Offset {
	if !st.KindOn {
		return NodeDrag.TrimToDragRule(delta, st)
	}
	delta = snapDeltaTheta(delta, st.Constants)
	delta = NodeDrag.TrimToDragRule(delta, st)
	return holdEqualOutLengths(delta, st)
}

func snapDeltaTheta(delta polarindex.Offset, sc polarindex.SceneConstants) polarindex.Offset {
	if sc.MaxIndexTheta == 0 {
		return delta
	}
	half := sc.MaxIndexTheta / 2
	if half == 0 {
		return delta
	}
	steps := int(math.Round(float64(delta.Theta) / float64(half)))
	delta.Theta = steps * half
	return delta
}

func holdEqualOutLengths(delta polarindex.Offset, st NodeDrag.State) polarindex.Offset {
	longest, shortest, count := 0, 0, 0
	for _, to := range st.OutTargets {
		d, ok := st.OutDelta[to]
		if !ok {
			continue
		}
		if count == 0 || d.R > longest {
			longest = d.R
		}
		if count == 0 || d.R < shortest {
			shortest = d.R
		}
		count++
	}
	if count < 2 || longest == shortest {
		return delta
	}
	delta.R -= longest - shortest
	return delta
}

func equalOutLengths(delta polarindex.Offset, st NodeDrag.State) map[string]polarindex.Offset {
	sc := st.Constants
	paths := map[string]polarindex.Offset{}
	shared := 0
	for _, to := range st.OutTargets {
		p, ok := st.OutDelta[to]
		if !ok {
			continue
		}
		paths[to] = p
		if p.R > shared {
			shared = p.R
		}
	}
	if len(paths) == 0 {
		return nil
	}
	selfWas := st.Index
	selfNow := polarindex.Compose(selfWas, delta, sc)
	out := make(map[string]polarindex.Offset, len(paths))
	for to, p := range paths {
		wants := p
		wants.R = shared
		targetOld := polarindex.Compose(selfWas, p, sc)
		targetNew := polarindex.Compose(selfNow, wants, sc)
		out[to] = polarindex.Delta(targetNew, targetOld)
	}
	return out
}
