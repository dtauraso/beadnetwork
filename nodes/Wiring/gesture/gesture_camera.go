package gesture

import (
	"github.com/dtauraso/wirefold/nodes/rowevent"
	T "github.com/dtauraso/wirefold/tools/topology-vscode/Trace"
)

func CameraViewEvent() []rowevent.RowEvent {
	return []rowevent.RowEvent{{Kind: T.KindCamera, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1}}
}
