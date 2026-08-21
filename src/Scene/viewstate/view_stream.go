package viewstate

import (
	"encoding/binary"
	"io"

	B "github.com/dtauraso/wirefold/src/Buffer"
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

type ViewFrameBuilder func(tick uint32, events []B.RowEvent) []byte

func (ui *UIState) SetViewStream(out io.Writer, buildFrame ViewFrameBuilder) {

	ui.viewOut = newViewClaimedStream(&ui.viewClaimed, out)
	ui.ViewBuildFrame = buildFrame
}

func (ui *UIState) EmitBreadcrumb(ev B.RowEvent) {
	ev.Kind = B.KindBreadcrumb
	ev.Debug = 1
	ui.EmitViewFrame([]B.RowEvent{ev})
}

func (ui *UIState) EmitViewFrame(events []B.RowEvent) {
	if ui.ViewBuildFrame == nil {
		return
	}
	ui.viewTick++

	ui.writeSceneColumns()
	ui.writePointerTargetColumns()
	pl := ui.PanelLayout()
	ui.writeSpeedPanelColumns(pl.Speed)
	ui.writeTiltPanelColumns(pl.Tilt)
	ui.writeAnglePillValues(pl.Angle)
	ui.writeNodesPillValues(pl.Nodes)
	ui.writeOverlaysPillColumns(pl.Overlays)
	ui.writeFitChipValues(pl.Fit)
	ui.writeTabStripValues(pl.Tabs)
	ui.writeRulesPanelValues(pl.Rules)

	frame := ui.ViewBuildFrame(ui.viewTick, events)
	if !ui.viewOut.Ok() {
		return
	}
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], uint32(len(frame)))

	_, _ = ui.viewOut.Write(hdr[:])
	_, _ = ui.viewOut.Write(frame)
}

func boolU8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}
