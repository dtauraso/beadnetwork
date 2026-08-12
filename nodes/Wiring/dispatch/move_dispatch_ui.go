// move_dispatch_ui.go — the UI-state phases NewMoveDispatch calls: seed md.UI's startup
// defaults (each overwritten later by its own loader if the loaded scene has one
// persisted), and bind the two closures EmitViewFrame needs but cannot reach directly.

package dispatch

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

// initMoveDispatchUIDefaults sets md.UI's startup defaults, each overwritten later by its
// own loader if the loaded scene has one persisted.
func initMoveDispatchUIDefaults(md *MoveDispatch) {
	md.UI.OV = viewstate.DefaultOverlayState()
	md.UI.Speed = 1                                         // default playback multiplier; LoadSpeed overwrites from view/speed.json if present
	md.UI.ClockDivisor = 1                                  // no scaling until LoadSpeed resolves the loaded scene's own divisor
	md.UI.LatticePoints = scenepersist.DefaultLatticePoints // LoadLatticePoints overwrites from view/lattice.json if present
}

// bindUIClosures binds the two closures EmitViewFrame needs but cannot reach directly
// (md.RT/md.MR are unexported-package-internal from viewstate's point of view — RT is
// exported but UIState cannot hold a *rowtables.RowTables field of its own without
// MoveDispatch handing it one, and DistanceGroupLens needs *moverreg.MoverRegistry, an
// unexported dispatch-owned field). Bound ONCE, here, mirroring ng.msg.sendMove =
// md.MR.EnqueueFor(ng) above — not re-resolved on every emit. Method value md.RT.NodeRowFor
// captures &md.RT (md.RT is addressable through the *MoveDispatch pointer), so it keeps
// seeing whatever RT.Build just populated even though this bind runs after it.
func (md *MoveDispatch) bindUIClosures() {
	md.UI.NodeRowFor = md.RT.NodeRowFor
	mrForLens, uiForLens := &md.MR, &md.UI
	md.UI.DistanceGroupLensFn = func() (float32, float32, float32) {
		return DistanceGroupLens(uiForLens, mrForLens)
	}
}
