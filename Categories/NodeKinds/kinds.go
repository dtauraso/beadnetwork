package NodeKinds

import (
	kindNodePhi "github.com/dtauraso/beadnetwork/Categories/NodeKinds/NodePhi"
	kindNodePhiTheta "github.com/dtauraso/beadnetwork/Categories/NodeKinds/NodePhiTheta"
	kindNodePhiTheta3 "github.com/dtauraso/beadnetwork/Categories/NodeKinds/NodePhiTheta3"
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
	case "NodePhi":
		return kindNodePhi.Builder, true
	case "NodePhiTheta":
		return kindNodePhiTheta.Builder, true
	case "NodePhiTheta3":
		return kindNodePhiTheta3.Builder, true
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
