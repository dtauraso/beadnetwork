package stdinreader

import (
	"github.com/dtauraso/wirefold/src/Node/Wiring/dispatch"
	"github.com/dtauraso/wirefold/src/Node/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/src/Node/Wiring/panelstack"
)

func panelTookWheel(ev inputcodec.RawInputMsg, md *dispatch.MoveDispatch) bool {
	pl := md.UI.PanelLayout()

	if pl.Overlays.Open && panelstack.HitRect(pl.Overlays.Popover, ev.X, ev.Y) {
		return scrollBy(md, &md.UI.OverlaysScroll, pl.Overlays.MaxScroll, ev.DeltaY)
	}
	if pl.Rules.Open && panelstack.HitRect(pl.Rules.RowsClip, ev.X, ev.Y) {
		return scrollBy(md, &md.UI.RulesScroll, pl.Rules.MaxScroll, ev.DeltaY)
	}
	return false
}

func scrollBy(md *dispatch.MoveDispatch, scroll *float32, max float32, delta float64) bool {
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
