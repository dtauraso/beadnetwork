package View

import "fmt"

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

	if ui.ViewW <= 0 || ui.ViewH <= 0 {
		ui.ViewBuildFrame(ui.viewTick, events)
		return
	}

	pl := ui.PanelLayout()
	ui.Slider.Write(pl.Speed, ui.Speed)
	ui.Tilt.Write(pl.Tilt)
	ui.Angle.Write(pl.Angle)
	ui.Nodes.Write(pl.Nodes, ui.EditRefused)
	ui.OverlaysPill.Write(pl.Overlays)
	ui.Fit.Write(pl.Fit)
	if len(pl.Tabs.Tabs) == 0 {
		appendTrace(viewTracePath(ui.SceneRoot()), []RowEvent{{
			Kind: KindBreadcrumb, Label: "empty-chrome", Debug: 1,
			NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
			Text: fmt.Sprintf("tabs=0 viewW=%.0f viewH=%.0f names=%d tick=%d",
				ui.ViewW, ui.ViewH, len(ui.TabStrip.Names), ui.viewTick),
		}})
	}
	ui.TabStrip.Write(pl.Tabs)
	ui.Rules.Write(pl.Rules)

	ui.ViewBuildFrame(ui.viewTick, events)
}
