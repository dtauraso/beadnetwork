package runtopology

import (
	"os"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	B "github.com/dtauraso/wirefold/Buffer"
	SF "github.com/dtauraso/wirefold/Buffer/streamframe"
	W "github.com/dtauraso/wirefold/nodes/Wiring"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

// wireViewStream makes md the VIEW stream's owner/writer.
//
// The VIEW stream's write side (Step C, memory/feedback_no_single_writer_bridge.md): wire md as the
// stream's owner/writer BEFORE anything that can change camera/overlay/scene-sphere/
// selection/hover reaches it (SeedInitialViewpoint/LoadOverlays/LoadSceneSphere below,
// then the launched movers/stdin reader) — mirrors SetEdgeStreams/SetNodeStreams'
// "wire before it can fire" ordering above. Only when the dedicated fd is actually
// wired (viewStreamWired) — left uncalled otherwise (no WIREFOLD_STREAM_FDS "view"
// entry, e.g. a non-extension launch with no dedicated pipes at all).
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
				events []wire.RowEvent,
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
						DragNodeRow:   dragNodeRow,
						EditRefused:   scene.EditRefused,
						SceneEditable: scene.SceneEditable,
						SceneKinds:    scene.SceneKinds,
						GroupLenTime:  groupLenTime, GroupLenInput: groupLenInput, GroupLenGate: groupLenGate,
						Speed: speed,
					},
					sceneCX, sceneCY, sceneCZ, sceneRadius,
					// The tab strip is CONSTANT for this process's lifetime: the list is
					// Go's own registry and the selection is what this run was loaded
					// with (switching tabs ends the run — scene_switch.go's SelectScene),
					// so it is captured here rather than threaded through MoveDispatch's
					// view-frame signature as if it were live state.
					sceneTabNames, uint16(sceneTabSelected),
					toStreamEvents(events))
			})
	}
}
