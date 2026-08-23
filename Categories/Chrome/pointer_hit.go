package Chrome

import (
	Panels "github.com/dtauraso/wirefold/Categories/Chrome/Panels"
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/PolarRulesPanel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Pills"
	"github.com/dtauraso/wirefold/Categories/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/wirefold/Categories/Chrome/Pills/NodesDropdown"
)

// TargetAt is what the pointer is over, asked of the chrome's own layout. It
// reads no view state: a layout in, a target out.
func TargetAt(pl Layout, x, y float64) Panels.PointerTarget {

	if h := pl.Rules.Hit(x, y); h.Kind != PolarRulesPanel.HitNone {
		return Panels.PointerTarget{Rect: h.Rect, Kind: Panels.PointerInteractive}
	}
	if Panel.HitRect(pl.Fit, x, y) {
		return Panels.PointerTarget{
			Rect: pl.Fit, Kind: Panels.PointerInteractive, Tip: "Frame the whole diagram",
		}
	}
	if i := pl.Tabs.Hit(x, y); i >= 0 && i < len(pl.Tabs.Tabs) {
		return Panels.PointerTarget{Rect: pl.Tabs.Tabs[i].Rect, Kind: Panels.PointerInteractive}
	}
	if h := pl.Nodes.Hit(x, y); h.Kind != NodesDropdown.HitNone {
		return Panels.PointerTarget{Rect: h.Rect, Kind: Panels.PointerInteractive}
	}
	if i := pl.Speed.Hit(x, y); i >= 0 && i < len(pl.Speed.Ticks) {
		return Panels.PointerTarget{Rect: pl.Speed.Ticks[i], Kind: Panels.PointerInteractive}
	}
	if h := pl.Overlays.Hit(x, y); h.Kind != Pills.HitNone || h.Disabled {
		kind := Panels.PointerInteractive
		if h.Disabled {
			kind = Panels.PointerRefusing
		}
		return Panels.PointerTarget{Rect: h.Rect, Kind: kind, Tip: h.Tip}
	}
	if h := pl.Angle.Hit(x, y); h.Kind != AngleDropdown.HitNone {
		return Panels.PointerTarget{Rect: h.Rect, Kind: Panels.PointerInteractive}
	}
	return Panels.PointerTarget{}
}

// TakeWheel scrolls whichever piece is under the pointer, given that piece's
// own scroll to move and the redraw to ask for.
func TakeWheel(pl Layout, overlaysScroll, rulesScroll *float32, x, y, deltaY float64, redraw func()) bool {

	if pl.Overlays.Open && Panel.HitRect(pl.Overlays.Popover, x, y) {
		return scrollBy(overlaysScroll, pl.Overlays.MaxScroll, deltaY, redraw)
	}
	if pl.Rules.Open && Panel.HitRect(pl.Rules.RowsClip, x, y) {
		return scrollBy(rulesScroll, pl.Rules.MaxScroll, deltaY, redraw)
	}
	return false
}

func scrollBy(scroll *float32, max float32, delta float64, redraw func()) bool {
	if max <= 0 {
		return true
	}
	next := clampScroll(*scroll+float32(delta), max)
	if next == *scroll {
		return true
	}
	*scroll = next
	redraw()
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
