package dispatch

import (
	"github.com/dtauraso/wirefold/src/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/src/Node/Wiring/inputcodec"
)

func panelTookWheel(ev inputcodec.RawInputMsg, md *MoveDispatch) bool {
	pl := md.UI.PanelLayout()

	if pl.Overlays.Open && Panel.HitRect(pl.Overlays.Popover, ev.X, ev.Y) {
		return scrollBy(md, &md.UI.OverlaysScroll, pl.Overlays.MaxScroll, ev.DeltaY)
	}
	if pl.Rules.Open && Panel.HitRect(pl.Rules.RowsClip, ev.X, ev.Y) {
		return scrollBy(md, &md.UI.RulesScroll, pl.Rules.MaxScroll, ev.DeltaY)
	}
	return false
}

func scrollBy(md *MoveDispatch, scroll *float32, max float32, delta float64) bool {
	if max <= 0 {
		return true
	}
	next := clampScroll(*scroll+float32(delta), max)
	if next == *scroll {
		return true
	}
	*scroll = next
	md.UI.EmitViewFrame(nil)
	return true
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
