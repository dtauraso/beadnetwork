package gesture

import (
	T "github.com/dtauraso/wirefold/Trace"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

func CameraViewEvent() []wire.RowEvent {
	return []wire.RowEvent{{Kind: T.KindCamera, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1}}
}
