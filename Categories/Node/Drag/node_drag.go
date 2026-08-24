package Drag

import (
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/PolarRulesPanel"
	"github.com/dtauraso/beadnetwork/Categories/Polar/polarindex"
)

type State struct {
	Index     polarindex.Index
	Constants polarindex.SceneConstants

	Drag   *PolarRulesPanel.DragRule
	DragOn bool

	Self   *PolarRulesPanel.DragRule
	SelfOn bool

	KindOn bool

	OutTargets []string
	OutDelta   map[string]polarindex.Offset

	Inbound map[string]polarindex.Offset
}

type Trim func(delta polarindex.Offset, st State) polarindex.Offset

type Request func(delta polarindex.Offset, st State) map[string]polarindex.Offset

func Apply(trim Trim, delta polarindex.Offset, st State) polarindex.Offset {
	if st.DragOn {
		if trim != nil {
			delta = trim(delta, st)
		} else {
			delta = TrimToDragRule(delta, st)
		}
	}
	return TrimToSelfRule(delta, st)
}

func TrimToSelfRule(delta polarindex.Offset, st State) polarindex.Offset {
	rule := st.Self
	if rule == nil || !st.SelfOn {
		return delta
	}
	sc := st.Constants
	haveIdx := st.Index
	wantIdx := polarindex.Compose(haveIdx, delta, sc)

	if rule.R != nil {
		wantIdx.R = haveIdx.R
	}
	if rule.MaxTheta != nil {
		maxIdx := int(math.Round(*rule.MaxTheta / sc.ConstantTheta()))
		kept := min(max(wantIdx.Theta, -maxIdx), maxIdx)
		if farSide(wantIdx.Theta-kept, sc.MaxIndexTheta) {
			wantIdx.Phi = 0
		}
		wantIdx.Theta = kept
	}
	if rule.Phi != nil {
		wantIdx.Phi = haveIdx.Phi
	}
	return polarindex.Delta(wantIdx, haveIdx)
}

func farSide(thetaGap, turn int) bool {
	if turn <= 0 {
		return false
	}
	half := turn / 2
	gap := ((thetaGap+half)%turn+turn)%turn - half
	if gap < 0 {
		gap = -gap
	}
	return gap > turn/4
}

func Requested(request Request, delta polarindex.Offset, st State) map[string]polarindex.Offset {
	if !st.DragOn || request == nil {
		return nil
	}
	return request(delta, st)
}

func TrimToDragRule(delta polarindex.Offset, st State) polarindex.Offset {
	rule := st.Drag
	if rule == nil || !st.DragOn {
		return delta
	}
	sc := st.Constants
	for _, haveOff := range st.Inbound {
		have := polarindex.OffsetToPolar(haveOff, sc)
		wantOff := polarindex.Sum(haveOff, delta)
		want := polarindex.OffsetToPolar(wantOff, sc)
		trimmed := rule.TrimDelta(have, want)
		trimmedOff := polarindex.MeasureOffset(trimmed, sc)
		delta = polarindex.Sum(trimmedOff, polarindex.Neg(haveOff))
	}
	return delta
}
