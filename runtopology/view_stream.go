package runtopology

import (
	"os"

	"github.com/dtauraso/wirefold/nodes/rowevent"

	B "github.com/dtauraso/wirefold/Buffer"
	SF "github.com/dtauraso/wirefold/Buffer/streamframe"
	W "github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

func wireViewStream(md *W.MoveDispatch, viewFile *os.File, viewStreamWired bool, sceneTabNames []string, sceneTabSelected int) {
	if viewStreamWired {
		md.UI.SetViewStream(viewFile,
			func(tick uint32,
				camPX, camPY, camPZ, camR, camPosTheta, camPosPhi, camUpTheta, camUpPhi float32,
				flags viewstate.ViewOverlayFlags,
				dragNodeRow int32,
				scene viewstate.ViewSceneState,
				groupLenTime, groupLenInput, groupLenGate float32,
				speed float32,
				sceneCX, sceneCY, sceneCZ, sceneRadius float32,
				events []rowevent.RowEvent,
			) []byte {
				return SF.BuildViewStreamFrame(tick,
					camPX, camPY, camPZ, camR, camPosTheta, camPosPhi, camUpTheta, camUpPhi,
					B.OverlayRow{
						SceneTori: flags.SceneTori, ScenePoles: flags.ScenePoles, NodePoles: flags.NodePoles,
						SelSpherePoles: flags.SelSpherePoles, Handholds: flags.Handholds,
						LabelsGlobal: flags.LabelsGlobal, OverlaysVis: flags.OverlaysVis,
						NodeBody: flags.NodeBody, NodeRing: flags.NodeRing, RingPick: flags.RingPick,
						SelectionRing: flags.SelectionRing, HoverRing: flags.HoverRing,
						ReachSphere:   flags.ReachSphere,
						SceneVectors:  flags.SceneVectors,
						DragNodeRow:   dragNodeRow,
						EditRefused:   scene.EditRefused,
						SceneEditable: scene.SceneEditable,
						SceneKinds:    scene.SceneKinds,
						GroupLenTime:  groupLenTime, GroupLenInput: groupLenInput, GroupLenGate: groupLenGate,
						Speed: speed,
					},
					sceneCX, sceneCY, sceneCZ, sceneRadius,

					sceneTabNames, uint16(sceneTabSelected),
					toStreamEvents(events))
			})
	}
}
