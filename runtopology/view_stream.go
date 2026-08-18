package runtopology

import (
	"github.com/dtauraso/wirefold/nodes/bead"
	NodeShape "github.com/dtauraso/wirefold/tools/topology-vscode/Node/Shape"
	"os"

	"github.com/dtauraso/wirefold/nodes/rowevent"

	W "github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
	B "github.com/dtauraso/wirefold/tools/topology-vscode/Buffer"
	SF "github.com/dtauraso/wirefold/tools/topology-vscode/Buffer/streamframe"
)

func wireViewStream(md *W.MoveDispatch, viewFile *os.File, viewStreamWired bool, sceneTabNames []string, sceneTabSelected int) {
	if viewStreamWired {
		md.UI.SetViewStream(viewFile,
			func(tick uint32,
				camPX, camPY, camPZ, camR, camPosPhi, camPosTheta, camUpPhi, camUpTheta float32,
				flags viewstate.ViewOverlayFlags,
				panels viewstate.ViewPanelFlags,
				dragNodeRow int32,
				scene viewstate.ViewSceneState,
				speed float32,
				sceneCX, sceneCY, sceneCZ, sceneRadius float32,
				events []rowevent.RowEvent,
			) []byte {
				return SF.BuildViewStreamFrame(tick,
					camPX, camPY, camPZ, camR, camPosPhi, camPosTheta, camUpPhi, camUpTheta,
					B.OverlayRow{
						SceneTori: flags.SceneTori, ScenePoles: flags.ScenePoles, NodePoles: flags.NodePoles,
						Handholds:    flags.Handholds,
						LabelsGlobal: flags.LabelsGlobal, OverlaysVis: flags.OverlaysVis,
						NodeBody: flags.NodeBody, NodeRing: flags.NodeRing, RingPick: flags.RingPick,
						SelectionRing: flags.SelectionRing, HoverRing: flags.HoverRing,
						SceneVectors:   flags.SceneVectors,
						RuleChannels:   flags.RuleChannels,
						NodePoleSphere: flags.NodePoleSphere,
						AllPoleSpheres: flags.AllPoleSpheres,
						DragNodeRow:    dragNodeRow,
						EditRefused:    scene.EditRefused,
						SceneEditable:  scene.SceneEditable,
						SceneKinds:     scene.SceneKinds,
						Speed:          speed,
					},
					B.PanelRow{
						Overlays: panels.Overlays, Node: panels.Node, NodeShape: panels.NodeShape,
						NodeState: panels.NodeState, NodePoles: panels.NodePoles,
						NodeRules: panels.NodeRules,
						Scene:     panels.Scene, SceneGuides: panels.SceneGuides, ScenePoles: panels.ScenePoles,
						SceneVectors: panels.SceneVectors, SceneLabels: panels.SceneLabels,
					},
					sceneCX, sceneCY, sceneCZ, sceneRadius,
					NodeShape.CanonicalRingSurfacePointsFlat(),
					bead.CanonicalRingSurfacePointsFlat(),

					sceneTabNames, uint16(sceneTabSelected),
					toStreamEvents(events))
			})
	}
}
