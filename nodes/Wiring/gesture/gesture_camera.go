package gesture

import (
	T "github.com/dtauraso/wirefold/Trace"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// CameraViewEvent is the single Camera event every camera-changing action in this package
// hands to EmitViewFrame. Camera decodes entirely from the VIEW frame's own Camera block
// (see buffer-log.ts's decodeEventLine "camera" case) — no row identity to resolve.
// Exported because package dispatch's own viewpoint tests (viewpoint_ops_test.go,
// viewpoint_bridge_test.go) also emit this exact event and dispatch may not duplicate it —
// duplicating the row shape here would be the alias-shim this decomposition forbids.
func CameraViewEvent() []wire.RowEvent {
	return []wire.RowEvent{{Kind: T.KindCamera, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1}}
}
