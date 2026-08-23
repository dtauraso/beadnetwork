package viewstate

import (
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/PolarRulesPanel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/TiltPanel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Pills"
	"github.com/dtauraso/wirefold/Categories/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/wirefold/Categories/Chrome/Pills/FitButton"
	"github.com/dtauraso/wirefold/Categories/Chrome/Pills/NodesDropdown"
	"github.com/dtauraso/wirefold/Categories/Chrome/Tabs"
	NodeBuf "github.com/dtauraso/wirefold/Categories/Node"
)

type PanelLayout struct {
	Speed    SliderPanel.Layout
	Tilt     TiltPanel.Layout
	Angle    AngleDropdown.Layout
	Nodes    NodesDropdown.Layout
	Overlays Pills.Layout

	Fit Panel.Rect

	Tabs Tabs.Layout

	Rules PolarRulesPanel.Layout
}

var PillLabels = []string{AngleDropdown.Label, NodesDropdown.Label, Pills.Label}

func (ui *UIState) PanelLayout() PanelLayout {
	st := Panel.New(float32(ui.ViewH))
	pills := Panel.NewPillStack(float32(ui.ViewW), float32(ui.ViewH), PillLabels)

	nodes := make([]AngleDropdown.Node, len(ui.Tilt.Rows))
	for i, row := range ui.Tilt.Rows {
		nodes[i] = AngleDropdown.Node{
			Row:   row,
			Label: ui.Tilt.Labels[i],
			Open:  ui.Angle.GroupOpen[row],
		}
	}

	fit := pills.AddChip(FitButton.FitLabel)

	speed := SliderPanel.Build(st)
	tilt := TiltPanel.Build(st, ui.Tilt.Rows, ui.Tilt.Labels)
	rules := PolarRulesPanel.Build(st, Panel.PanelOpen["nodeRules"](&ui.PN), ui.Rules)

	angle := AngleDropdown.Build(pills, ui.Angle.Open, ui.LatticePoints, nodes)
	nodesPill := NodesDropdown.Build(pills, ui.Nodes.Open && ui.SceneEditable, ui.paletteKinds())
	overlays := Pills.Build(pills, &ui.OV, &ui.PN, ui.OverlaysPill.Scroll)

	return PanelLayout{
		Fit:      fit,
		Tabs:     Tabs.Build(float32(ui.ViewW), ui.TabStrip.Names, ui.TabStrip.Selected),
		Rules:    rules,
		Speed:    speed,
		Tilt:     tilt,
		Angle:    angle,
		Nodes:    nodesPill,
		Overlays: overlays,
	}
}

func (ui *UIState) paletteKinds() []NodesDropdown.Kind {
	if !ui.SceneEditable {
		return nil
	}
	names := NodeBuf.KindNameByID()
	out := make([]NodesDropdown.Kind, 0, len(names))
	for id, name := range names {
		if name == "" || ui.SceneKinds&(1<<uint(id)) == 0 {
			continue
		}
		out = append(out, NodesDropdown.Kind{
			KindID: uint8(id),
			Name:   name,
			Open:   ui.NodesRowOpen[uint8(id)],
		})
	}
	return out
}
