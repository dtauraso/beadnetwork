package stdinreader

import (
	"github.com/dtauraso/wirefold/src/Node/Wiring/angledropdown"
	"github.com/dtauraso/wirefold/src/Node/Wiring/dispatch"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodesdropdown"
	"github.com/dtauraso/wirefold/src/Node/Wiring/overlayspanel"
	"github.com/dtauraso/wirefold/src/Node/Wiring/panelstack"
	"github.com/dtauraso/wirefold/src/Node/Wiring/rulespanel"
	"github.com/dtauraso/wirefold/src/Node/Wiring/viewstate"
)

func panelPointerTarget(md *dispatch.MoveDispatch, x, y float64) viewstate.PointerTarget {
	pl := md.UI.PanelLayout()

	if h := pl.Rules.Hit(x, y); h.Kind != rulespanel.HitNone {
		return viewstate.PointerTarget{Rect: h.Rect, Kind: viewstate.PointerInteractive}
	}
	if panelstack.HitRect(pl.Fit, x, y) {
		return viewstate.PointerTarget{
			Rect: pl.Fit, Kind: viewstate.PointerInteractive, Tip: "Frame the whole diagram",
		}
	}
	if i := pl.Tabs.Hit(x, y); i >= 0 && i < len(pl.Tabs.Tabs) {
		return viewstate.PointerTarget{Rect: pl.Tabs.Tabs[i].Rect, Kind: viewstate.PointerInteractive}
	}
	if h := pl.Nodes.Hit(x, y); h.Kind != nodesdropdown.HitNone {
		return viewstate.PointerTarget{Rect: h.Rect, Kind: viewstate.PointerInteractive}
	}
	if i := pl.Speed.Hit(x, y); i >= 0 && i < len(pl.Speed.Ticks) {
		return viewstate.PointerTarget{Rect: pl.Speed.Ticks[i], Kind: viewstate.PointerInteractive}
	}
	if h := pl.Overlays.Hit(x, y); h.Kind != overlayspanel.HitNone || h.Disabled {
		kind := viewstate.PointerInteractive
		if h.Disabled {
			kind = viewstate.PointerRefusing
		}
		return viewstate.PointerTarget{Rect: h.Rect, Kind: kind, Tip: h.Tip}
	}
	if h := pl.Angle.Hit(x, y); h.Kind != angledropdown.HitNone {
		return viewstate.PointerTarget{Rect: h.Rect, Kind: viewstate.PointerInteractive}
	}
	return viewstate.PointerTarget{}
}
