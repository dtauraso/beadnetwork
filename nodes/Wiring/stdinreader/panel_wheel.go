package stdinreader

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/panelstack"
)

func panelTookWheel(ev inputcodec.RawInputMsg, md *dispatch.MoveDispatch) bool {
	pl := md.UI.PanelLayout()

	if pl.Overlays.Open && panelstack.HitRect(pl.Overlays.Popover, ev.X, ev.Y) {
		if pl.Overlays.MaxScroll <= 0 {
			return true
		}
		md.UI.OverlaysScroll = clampScroll(md.UI.OverlaysScroll+float32(ev.DeltaY), pl.Overlays.MaxScroll)
		md.UI.EmitViewFrame(nil)
		return true
	}
	return false
}

func clampScroll(v, max float32) float32 {
	if v < 0 {
		return 0
	}
	if v > max {
		return max
	}
	return v
}
