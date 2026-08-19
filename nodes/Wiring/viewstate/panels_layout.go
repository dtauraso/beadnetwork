package viewstate

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/angledropdown"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodesdropdown"
	"github.com/dtauraso/wirefold/nodes/Wiring/overlayspanel"
	"github.com/dtauraso/wirefold/nodes/Wiring/panelstack"
	"github.com/dtauraso/wirefold/nodes/Wiring/rulespanel"
	"github.com/dtauraso/wirefold/nodes/Wiring/speedpanel"
	"github.com/dtauraso/wirefold/nodes/Wiring/tabstrip"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltpanel"
	B "github.com/dtauraso/wirefold/tools/topology-vscode/Buffer"
	"github.com/dtauraso/wirefold/tools/topology-vscode/OverlaysDropdown"
)

type PanelLayout struct {
	Speed    speedpanel.Layout
	Tilt     tiltpanel.Layout
	Angle    angledropdown.Layout
	Nodes    nodesdropdown.Layout
	Overlays overlayspanel.Layout

	Fit panelstack.Rect

	Tabs tabstrip.Layout

	Rules rulespanel.Layout
}

const FitLabel = "⌂ fit"

var PillLabels = []string{angledropdown.Label, nodesdropdown.Label, overlayspanel.Label}

func (ui *UIState) PanelLayout() PanelLayout {
	st := panelstack.New()
	pills := panelstack.NewPillStack(float32(ui.ViewW), PillLabels)

	nodes := make([]angledropdown.Node, len(ui.TiltRows))
	for i, row := range ui.TiltRows {
		nodes[i] = angledropdown.Node{
			Row:   row,
			Label: ui.TiltLabels[i],
			Open:  ui.AngleGroupOpen[row],
		}
	}

	fit := pills.AddChip(FitLabel)

	speed := speedpanel.Build(st)
	tilt := tiltpanel.Build(st, ui.TiltRows, ui.TiltLabels)
	rules := rulespanel.Build(
		st, OverlaysDropdown.PanelOpen["nodeRules"](&ui.PN),
		ui.RuleNodes, ui.RuleEdit, ui.RuleSharedRow,
	)

	angle := angledropdown.Build(pills, ui.AngleOpen, ui.LatticePoints, nodes)
	nodesPill := nodesdropdown.Build(pills, ui.NodesOpen && ui.SceneEditable, ui.paletteKinds())
	overlays := overlayspanel.Build(pills, &ui.OV, &ui.PN)

	return PanelLayout{
		Fit:      fit,
		Tabs:     tabstrip.Build(float32(ui.ViewW), ui.SceneTabNames, ui.SceneTabSelected),
		Rules:    rules,
		Speed:    speed,
		Tilt:     tilt,
		Angle:    angle,
		Nodes:    nodesPill,
		Overlays: overlays,
	}
}

func (ui *UIState) paletteKinds() []nodesdropdown.Kind {
	if !ui.SceneEditable {
		return nil
	}
	names := B.KindNameByID()
	out := make([]nodesdropdown.Kind, 0, len(names))
	for id, name := range names {
		if name == "" || ui.SceneKinds&(1<<uint(id)) == 0 {
			continue
		}
		out = append(out, nodesdropdown.Kind{
			KindID: uint8(id),
			Name:   name,
			Open:   ui.NodesRowOpen[uint8(id)],
		})
	}
	return out
}
