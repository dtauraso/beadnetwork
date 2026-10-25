package Chrome

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/PolarRulesPanel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/TiltPanel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/AngleDropdown"
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
	LatticePoints int32

	Overlays *Flags.OverlayState
	Panels   *Panel.PanelState

	Tilt     TiltPanel.State
	Sheet    *Sheet.State
	Angle    AngleDropdown.State
	Nodes    NodesDropdown.State
	Tabs     Tabs.State
	Rules    PolarRulesPanel.State
	PillsBar Pills.State
}

var PillLabels = append([]string{AngleDropdown.Label, NodesDropdown.Label, Pills.Label}, Sheet.RightLabels...)

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

	fit := pills.AddChip(FitButton.FitLabel)
	tabs := Tabs.Build(float32(in.ViewW), in.Tabs.Names, in.Tabs.Selected)
	top := float32(Panel.OriginY)
	if tabs.Strip.H > 0 {
		top = tabs.Strip.Y + tabs.Strip.H + Panel.Gap
	}
	st := Panel.NewAt(top, float32(in.ViewH))

	lay := Layout{
		Fit:   fit,
		Tabs:  tabs,
		Speed: SliderPanel.Build(st),
		Tilt:  TiltPanel.Build(st, in.Tilt.Rows, in.Tilt.Labels),
		Rules: PolarRulesPanel.Build(st, Panel.PanelOpen["nodeRules"](in.Panels), in.Rules),
	}
	lay.Angle = AngleDropdown.Build(pills, in.Angle.Open, in.LatticePoints, nodes)
	lay.Nodes = NodesDropdown.Build(pills, in.Nodes.Open && in.SceneEditable, NodesDropdown.PaletteKinds(in.SceneKinds, in.SceneEditable, in.Nodes.RowOpen))
	if in.Sheet != nil {
		left := Panel.NewLeftPillStack(float32(in.ViewW), float32(in.ViewH), st.Next(), Sheet.LeftLabels)
		lay.Sheet = Sheet.Build(pills, left, in.Sheet.Pills(), in.Sheet)
	}
	lay.Overlays = Pills.Build(pills, in.Overlays, in.Panels, in.PillsBar.Scroll)
	return lay
}
