package viewstate

import (
	"encoding/binary"
	"io"

	"github.com/dtauraso/wirefold/nodes/rowevent"

	T "github.com/dtauraso/wirefold/tools/topology-vscode/Trace"
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

type ViewFrameBuilder func(tick uint32, events []rowevent.RowEvent) []byte

func (ui *UIState) SetViewStream(out io.Writer, buildFrame ViewFrameBuilder) {

	ui.viewOut = newViewClaimedStream(&ui.viewClaimed, out)
	ui.ViewBuildFrame = buildFrame
}

func (ui *UIState) EmitBreadcrumb(ev rowevent.RowEvent) {
	ev.Kind = T.KindBreadcrumb
	ev.Debug = 1
	ui.EmitViewFrame([]rowevent.RowEvent{ev})
}

func (ui *UIState) EmitViewFrame(events []rowevent.RowEvent) {
	if ui.ViewBuildFrame == nil {
		return
	}
	ui.viewTick++
	sc := ui.SceneSphere

	dragNodeRow := int32(-1)
	if ui.LastDraggedNode != "" && ui.NodeRowFor != nil {
		if r, ok := ui.NodeRowFor(ui.LastDraggedNode); ok {
			dragNodeRow = r
		}
	}

	ui.writeSceneColumns(sc)
	ui.writePanelColumns()
	ui.writeOverlayColumns(dragNodeRow)
	ui.writeCameraColumns()

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
