package viewstate

import (
	"github.com/dtauraso/wirefold/src/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/src/Chrome/Panels/PolarRulesPanel"
	"github.com/dtauraso/wirefold/src/Chrome/Pills"
	"github.com/dtauraso/wirefold/src/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/wirefold/src/Chrome/Pills/NodesDropdown"
)

func (ui *UIState) PointerTargetAt(x, y float64) PointerTarget {
	pl := ui.PanelLayout()

	if h := pl.Rules.Hit(x, y); h.Kind != PolarRulesPanel.HitNone {
		return PointerTarget{Rect: h.Rect, Kind: PointerInteractive}
	}
	if Panel.HitRect(pl.Fit, x, y) {
		return PointerTarget{
			Rect: pl.Fit, Kind: PointerInteractive, Tip: "Frame the whole diagram",
		}
	}
	if i := pl.Tabs.Hit(x, y); i >= 0 && i < len(pl.Tabs.Tabs) {
		return PointerTarget{Rect: pl.Tabs.Tabs[i].Rect, Kind: PointerInteractive}
	}
	if h := pl.Nodes.Hit(x, y); h.Kind != NodesDropdown.HitNone {
		return PointerTarget{Rect: h.Rect, Kind: PointerInteractive}
	}
	if i := pl.Speed.Hit(x, y); i >= 0 && i < len(pl.Speed.Ticks) {
		return PointerTarget{Rect: pl.Speed.Ticks[i], Kind: PointerInteractive}
	}
	if h := pl.Overlays.Hit(x, y); h.Kind != Pills.HitNone || h.Disabled {
		kind := PointerInteractive
		if h.Disabled {
			kind = PointerRefusing
		}
		return PointerTarget{Rect: h.Rect, Kind: kind, Tip: h.Tip}
	}
	if h := pl.Angle.Hit(x, y); h.Kind != AngleDropdown.HitNone {
		return PointerTarget{Rect: h.Rect, Kind: PointerInteractive}
	}
	return PointerTarget{}
}

func (ui *UIState) TakeWheel(x, y, deltaY float64) bool {
	pl := ui.PanelLayout()

	if pl.Overlays.Open && Panel.HitRect(pl.Overlays.Popover, x, y) {
		return ui.scrollBy(&ui.OverlaysScroll, pl.Overlays.MaxScroll, deltaY)
	}
	if pl.Rules.Open && Panel.HitRect(pl.Rules.RowsClip, x, y) {
		return ui.scrollBy(&ui.RulesScroll, pl.Rules.MaxScroll, deltaY)
	}
	return false
}

func (ui *UIState) scrollBy(scroll *float32, max float32, delta float64) bool {
	if max <= 0 {
		return true
	}
	next := clampScroll(*scroll+float32(delta), max)
	if next == *scroll {
		return true
	}
	*scroll = next
	ui.EmitViewFrame(nil)
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
