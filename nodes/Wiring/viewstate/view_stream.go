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

type ViewFrameBuilder func(tick uint32,
	camPX, camPY, camPZ, camR, camPosPhi, camPosTheta, camUpPhi, camUpTheta float32,
	flags ViewOverlayFlags,
	panels ViewPanelFlags,
	dragNodeRow int32,
	scene ViewSceneState,
	speed float32,
	sceneCX, sceneCY, sceneCZ, sceneRadius, sceneConstantR float32,
	sceneMaxIndexPhi, sceneMaxIndexTheta int32,
	events []rowevent.RowEvent,
) []byte

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
	v := ui.VP.Viewpoint
	sc := ui.SceneSphere

	dragNodeRow := int32(-1)
	if ui.LastDraggedNode != "" && ui.NodeRowFor != nil {
		if r, ok := ui.NodeRowFor(ui.LastDraggedNode); ok {
			dragNodeRow = r
		}
	}

	ui.writeSceneColumns(sc)

	frame := ui.ViewBuildFrame(ui.viewTick,
		float32(v.Pivot.X), float32(v.Pivot.Y), float32(v.Pivot.Z), float32(v.R),
		float32(v.Pos.Phi), float32(v.Pos.Theta), float32(v.Up.Phi), float32(v.Up.Theta),
		ViewOverlayFlags{
			SceneTori:      boolU8(ui.OV.SceneToriVisible),
			ScenePoles:     boolU8(ui.OV.ScenePolesVisible),
			NodePoles:      boolU8(ui.OV.NodePolesVisible),
			Handholds:      boolU8(ui.OV.HandholdsVisible),
			LabelsGlobal:   boolU8(ui.OV.LabelsGlobalVisible),
			OverlaysVis:    boolU8(ui.OV.OverlaysVisible),
			NodeBody:       boolU8(ui.OV.NodeBodyVisible),
			NodeRing:       boolU8(ui.OV.NodeRingVisible),
			RingPick:       boolU8(ui.OV.RingPickVisible),
			SelectionRing:  boolU8(ui.OV.SelectionRingVisible),
			HoverRing:      boolU8(ui.OV.HoverRingVisible),
			SceneVectors:   boolU8(ui.OV.SceneVectorsVisible),
			RuleChannels:   boolU8(ui.OV.RuleChannelsVisible),
			NodePoleSphere: boolU8(ui.OV.NodePoleSphereVisible),
			AllPoleSpheres: boolU8(ui.OV.AllPoleSpheresVisible),
		},
		ViewPanelFlags{
			Overlays:     boolU8(ui.PN.OverlaysOpen),
			Node:         boolU8(ui.PN.NodeOpen),
			NodeShape:    boolU8(ui.PN.NodeShapeOpen),
			NodeState:    boolU8(ui.PN.NodeStateOpen),
			NodePoles:    boolU8(ui.PN.NodePolesOpen),
			NodeRules:    boolU8(ui.PN.NodeRulesOpen),
			Scene:        boolU8(ui.PN.SceneOpen),
			SceneGuides:  boolU8(ui.PN.SceneGuidesOpen),
			ScenePoles:   boolU8(ui.PN.ScenePolesOpen),
			SceneVectors: boolU8(ui.PN.SceneVectorsOpen),
			SceneLabels:  boolU8(ui.PN.SceneLabelsOpen),
		},
		dragNodeRow,
		ViewSceneState{
			EditRefused:   ui.EditRefused,
			SceneEditable: boolU8(ui.SceneEditable),
			SceneKinds:    ui.SceneKinds,
		},
		float32(ui.Speed),
		float32(sc.Center.X), float32(sc.Center.Y), float32(sc.Center.Z), float32(sc.Radius),
		float32(ui.Constants.ConstantR), int32(ui.Constants.MaxIndexPhi), int32(ui.Constants.MaxIndexTheta),
		events,
	)
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
