package Chrome

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/FitButton"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/NodesDropdown"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Tabs"
	Flags "github.com/dtauraso/beadnetwork/Categories/Scene/View/Flags"
)

type Layout struct {
	Speed    SliderPanel.Layout
	Sheet    Sheet.Layout
	Angle    AngleDropdown.Layout
	Nodes    NodesDropdown.Layout
	Overlays Pills.Layout

	Fit Panel.Rect

	Tabs Tabs.Layout
}

type Of struct {
	ViewW, ViewH float64

	SceneEditable bool
	SceneKinds    uint32
	SpeedSlider   bool
	AnglesPill    bool
	PlusButton    bool
	NodePalette   bool
	OverlaysDock  bool
	HiddenFlags   []string
	LatticePoints int32
	Speed         float64

	Overlays *Flags.OverlayState
	Panels   *Panel.PanelState

	Sheet    *Sheet.State
	Builder  Builder.Model
	Angle    AngleDropdown.State
	Nodes    NodesDropdown.State
	Tabs     Tabs.State
	PillsBar Pills.State
}

var PillLabels = []string{AngleDropdown.Label, NodesDropdown.Label, Pills.Label}

func LayoutOf(in Of) Layout {
	pills := Panel.NewPillStack(float32(in.ViewW), float32(in.ViewH), PillLabels)

	tabs := Tabs.Build(float32(in.ViewW), in.Tabs.Names, in.Tabs.Selected, in.Sheet != nil && !in.Sheet.Folded[Sheet.DockKey])
	fit := tabs.Fit
	if fit.W == 0 {
		fit = pills.AddChip(FitButton.FitLabel)
	}
	top := float32(Panel.OriginY)
	if tabs.Strip.H > 0 {
		top = tabs.Strip.Y + tabs.Strip.H + Panel.Gap
	}
	st := Panel.NewAt(top, float32(in.ViewH))

	lay := Layout{Fit: fit, Tabs: tabs}
	if in.SpeedSlider && !in.OverlaysDock {
		lay.Speed = SliderPanel.Build(st)
	}
	if in.AnglesPill && !in.OverlaysDock {
		lay.Angle = AngleDropdown.Build(pills, in.Angle.Open, in.LatticePoints)
	}
	kinds := NodesDropdown.PaletteKinds(in.SceneKinds, in.SceneEditable, in.Nodes.RowOpen)
	switch {
	case !in.NodePalette:
	case in.PlusButton:
		lay.Nodes = NodesDropdown.BuildPlus(float32(in.ViewW), tabs.Strip, in.Nodes.Open && in.SceneEditable, kinds)
	default:
		lay.Nodes = NodesDropdown.Build(pills, in.Nodes.Open && in.SceneEditable, kinds)
	}
	if in.Sheet != nil {
		left := Panel.NewLeftPillStack(float32(in.ViewW), float32(in.ViewH), st.Next(), nil)
		dock := Builder.Pills(in.Builder)
		if in.OverlaysDock && in.SpeedSlider {
			dock = append(dock, SliderPanel.DockPill(in.Speed))
		}
		if in.OverlaysDock && in.AnglesPill {
			dock = append(dock, AngleDropdown.DockPill(in.LatticePoints))
		}
		if in.OverlaysDock {
			dock = append(dock, Pills.DockPill(in.Overlays, in.HiddenFlags))
		}
		lay.Sheet = Sheet.Build(pills, left, dock, in.Sheet)
	}
	if !in.OverlaysDock {
		lay.Overlays = Pills.Build(pills, in.Overlays, in.Panels, in.PillsBar.Scroll)
	}
	return lay
}
