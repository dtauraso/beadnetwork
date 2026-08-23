package viewstate

import (
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/PolarRulesPanel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/TiltPanel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/wirefold/Categories/Chrome/Pills/FitButton"
	"github.com/dtauraso/wirefold/Categories/Chrome/Pills/NodesDropdown"
	"github.com/dtauraso/wirefold/Categories/Chrome/Tabs"
)

type ViewOverlayFlags struct {
	SceneTori, ScenePoles, NodePoles, Handholds, LabelsGlobal, OverlaysVis uint8
	NodeBody, NodeRing, RingPick, SelectionRing, HoverRing                 uint8
	SceneVectors                                                           uint8
	RuleChannels                                                           uint8
	NodePoleSphere                                                         uint8
	AllPoleSpheres                                                         uint8
}

type ViewPanelFlags struct {
	Overlays                                                  uint8
	Node, NodeShape, NodeState, NodePoles, NodeRules          uint8
	Scene, SceneGuides, ScenePoles, SceneVectors, SceneLabels uint8
}

type ViewSceneState struct {
	EditRefused   uint32
	SceneEditable uint8
	SceneKinds    uint32
}

type ViewFrameBuilder func(tick uint32, events []RowEvent)

func (ui *UIState) SetViewStream(buildFrame ViewFrameBuilder) {
	ui.ViewBuildFrame = buildFrame
}

func (ui *UIState) WriteOwnTrace() {
	ui.SetViewStream(func(tick uint32, events []RowEvent) {
		appendTrace(viewTracePath(ui.SceneRoot()), events)
	})
}

func (ui *UIState) EmitBreadcrumb(ev RowEvent) {
	ev.Kind = KindBreadcrumb
	ev.Debug = 1
	ui.EmitViewFrame([]RowEvent{ev})
}

func (ui *UIState) EmitViewFrame(events []RowEvent) {
	if ui.ViewBuildFrame == nil {
		return
	}
	ui.viewTick++

	ui.writeSceneColumns()
	ui.writePointerTargetColumns()
	pl := ui.PanelLayout()
	SliderPanel.WriteValues(ui.sliderPanelValues, pl.Speed, ui.Speed)
	TiltPanel.WriteValues(ui.tiltPanelValues, pl.Tilt)
	AngleDropdown.WriteValues(ui.anglePillValues, pl.Angle)
	NodesDropdown.WriteValues(ui.nodesPillValues, pl.Nodes, ui.EditRefused)
	ui.writeOverlaysPillValues(pl.Overlays)
	FitButton.WriteValues(ui.fitChipValues, pl.Fit)
	Tabs.WriteValues(ui.tabStripValues, pl.Tabs)
	PolarRulesPanel.WriteValues(ui.rulesValues, pl.Rules)

	ui.ViewBuildFrame(ui.viewTick, events)
}
