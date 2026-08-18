package viewstate

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/angledropdown"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodesdropdown"
	"github.com/dtauraso/wirefold/nodes/Wiring/panelstack"
	"github.com/dtauraso/wirefold/nodes/Wiring/speedpanel"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltpanel"
	B "github.com/dtauraso/wirefold/tools/topology-vscode/Buffer"
)

type PanelLayout struct {
	Speed speedpanel.Layout
	Tilt  tiltpanel.Layout
	Angle angledropdown.Layout
	Nodes nodesdropdown.Layout
}

var PillLabels = []string{angledropdown.Label, "Nodes", "Overlays"}

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

	return PanelLayout{
		Speed: speedpanel.Build(st),
		Tilt:  tiltpanel.Build(st, ui.TiltRows, ui.TiltLabels),
		Angle: angledropdown.Build(pills, ui.AngleOpen, ui.LatticePoints, nodes),
		Nodes: nodesdropdown.Build(pills, ui.NodesOpen && ui.SceneEditable, ui.paletteKinds()),
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
