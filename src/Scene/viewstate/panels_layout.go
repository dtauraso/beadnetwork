package viewstate

import (
	"github.com/dtauraso/wirefold/src/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/src/Chrome/Panels/PolarRulesPanel"
	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/src/Chrome/Panels/TiltPanel"
	"github.com/dtauraso/wirefold/src/Chrome/Pills"
	"github.com/dtauraso/wirefold/src/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/wirefold/src/Chrome/Pills/NodesDropdown"
	"github.com/dtauraso/wirefold/src/Chrome/Tabs"
	NodeBuf "github.com/dtauraso/wirefold/src/Node"
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

const FitLabel = "⌂ fit"

var PillLabels = []string{AngleDropdown.Label, NodesDropdown.Label, Pills.Label}

func (ui *UIState) PanelLayout() PanelLayout {
	st := Panel.New(float32(ui.ViewH))
	pills := Panel.NewPillStack(float32(ui.ViewW), float32(ui.ViewH), PillLabels)

	nodes := make([]AngleDropdown.Node, len(ui.TiltRows))
	for i, row := range ui.TiltRows {
		nodes[i] = AngleDropdown.Node{
			Row:   row,
			Label: ui.TiltLabels[i],
			Open:  ui.AngleGroupOpen[row],
		}
	}

	fit := pills.AddChip(FitLabel)

	speed := SliderPanel.Build(st)
	tilt := TiltPanel.Build(st, ui.TiltRows, ui.TiltLabels)
	rules := PolarRulesPanel.Build(
		st, Panel.PanelOpen["nodeRules"](&ui.PN),
		ui.RuleNodes, ui.RuleEdit, ui.RuleSharedRow, ui.RulesScroll,
	)

	angle := AngleDropdown.Build(pills, ui.AngleOpen, ui.LatticePoints, nodes)
	nodesPill := NodesDropdown.Build(pills, ui.NodesOpen && ui.SceneEditable, ui.paletteKinds())
	overlays := Pills.Build(pills, &ui.OV, &ui.PN, ui.OverlaysScroll)

	return PanelLayout{
		Fit:      fit,
		Tabs:     Tabs.Build(float32(ui.ViewW), ui.SceneTabNames, ui.SceneTabSelected),
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
