package Chrome

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/PolarRulesPanel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/TiltPanel"
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
	Tilt     TiltPanel.Layout
	Sheet    Sheet.Layout
	Angle    AngleDropdown.Layout
	Nodes    NodesDropdown.Layout
	Overlays Pills.Layout

	Fit Panel.Rect

	Tabs Tabs.Layout

	Rules PolarRulesPanel.Layout
}

type Of struct {
	ViewW, ViewH float64

	SceneEditable bool
	SceneKinds    uint32
	PolarRules    bool
	SpeedSlider   bool
	AnglesPill    bool
	PlusButton    bool
	OverlaysDock  bool
	HiddenFlags   []string
	LatticePoints int32

	Overlays *Flags.OverlayState
	Panels   *Panel.PanelState

	Tilt     TiltPanel.State
	Sheet    *Sheet.State
	Builder  Builder.Model
	Angle    AngleDropdown.State
	Nodes    NodesDropdown.State
	Tabs     Tabs.State
	Rules    PolarRulesPanel.State
	PillsBar Pills.State
}

var PillLabels = []string{AngleDropdown.Label, NodesDropdown.Label, Pills.Label}

func LayoutOf(in Of) Layout {
	pills := Panel.NewPillStack(float32(in.ViewW), float32(in.ViewH), PillLabels)

	nodes := make([]AngleDropdown.Node, len(in.Tilt.Rows))
	for i, row := range in.Tilt.Rows {
		nodes[i] = AngleDropdown.Node{
			Row:   row,
			Label: in.Tilt.Labels[i],
			Open:  in.Angle.GroupOpen[row],
		}
	}

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
	if in.SpeedSlider {
		lay.Speed = SliderPanel.Build(st)
	}
	lay.Tilt = TiltPanel.Build(st, in.Tilt.Rows, in.Tilt.Labels)
	if in.PolarRules {
		lay.Rules = PolarRulesPanel.Build(st, Panel.PanelOpen["nodeRules"](in.Panels), in.Rules)
	}
	if in.AnglesPill {
		lay.Angle = AngleDropdown.Build(pills, in.Angle.Open, in.LatticePoints, nodes)
	}
	kinds := NodesDropdown.PaletteKinds(in.SceneKinds, in.SceneEditable, in.Nodes.RowOpen)
	if in.PlusButton {
		lay.Nodes = NodesDropdown.BuildPlus(float32(in.ViewW), tabs.Strip, kinds)
	} else {
		lay.Nodes = NodesDropdown.Build(pills, in.Nodes.Open && in.SceneEditable, kinds)
	}
	if in.Sheet != nil {
		left := Panel.NewLeftPillStack(float32(in.ViewW), float32(in.ViewH), st.Next(), nil)
		dock := Builder.Pills(in.Builder)
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
