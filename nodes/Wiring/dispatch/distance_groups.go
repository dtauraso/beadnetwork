// distance_groups.go — the one distance-group symbol pinned to package dispatch:
// DistanceGroupLens. ResolveSceneDistanceGroups and ApplyDistanceGroupTarget moved to
// nodes/Wiring/distancegroups (docs/planning/movedispatch-decomposition.md, the remainder
// cluster) — both already read/wrote only already-exported sub-objects. DistanceGroupLens
// stays here because move_dispatch_construct.go's NewMoveDispatch calls it directly to bind
// md.UI.DistanceGroupLensFn; moving it would require dispatch to import distancegroups for
// the delegator while distancegroups would need to import dispatch back for MoveDispatch —
// a real cycle.
package dispatch

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/distancegroups"
	"github.com/dtauraso/wirefold/nodes/Wiring/moverreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

// DistanceGroupLens returns the 3 groups' current max pair lengths, in
// distancegroups.GroupOrder (time, input, gate) — for the VIEW stream's Overlay
// GroupLenTime/GroupLenInput/GroupLenGate columns (read-only reflect; see view_stream.go's
// emitViewFrame). A group whose centers aren't resolvable yet reads 0. Binds
// mr.CenterOfNode once per call and forwards to distancegroups.Lens, which does the actual
// per-group max/scan.
func DistanceGroupLens(ui *viewstate.UIState, mr *moverreg.MoverRegistry) (timeLen, inputLen, gateLen float32) {
	return distancegroups.Lens(ui.HasDistanceGroups, mr.CenterOfNode)
}
