package NodeKinds

import (
	kindPairNode "github.com/dtauraso/beadnetwork/Categories/NodeKinds/PairNode"
	kindPulseLeft "github.com/dtauraso/beadnetwork/Categories/NodeKinds/PulseLeft"
	kindPulseRight "github.com/dtauraso/beadnetwork/Categories/NodeKinds/PulseRight"
	kindTime "github.com/dtauraso/beadnetwork/Categories/NodeKinds/Time"
	kindTimeEnd "github.com/dtauraso/beadnetwork/Categories/NodeKinds/TimeEnd"
	kindTimeStart "github.com/dtauraso/beadnetwork/Categories/NodeKinds/TimeStart"
	kindinput "github.com/dtauraso/beadnetwork/Categories/NodeKinds/input"
	kindpulse "github.com/dtauraso/beadnetwork/Categories/NodeKinds/pulse"
	kindselectleft "github.com/dtauraso/beadnetwork/Categories/NodeKinds/selectleft"
	kindselectright "github.com/dtauraso/beadnetwork/Categories/NodeKinds/selectright"
)

func BuilderFor(kind string) (Builder, bool) {
	switch kind {
	case "PairNode":
		return kindPairNode.Builder, true
	case "PulseLeft":
		return kindPulseLeft.Builder, true
	case "PulseRight":
		return kindPulseRight.Builder, true
	case "Time":
		return kindTime.Builder, true
	case "TimeEnd":
		return kindTimeEnd.Builder, true
	case "TimeStart":
		return kindTimeStart.Builder, true
	case "Input":
		return kindinput.Builder, true
	case "Pulse":
		return kindpulse.Builder, true
	case "SelectLeft":
		return kindselectleft.Builder, true
	case "SelectRight":
		return kindselectright.Builder, true
	}
	return nil, false
}
