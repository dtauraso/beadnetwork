package runtopology

import (
	"github.com/dtauraso/wirefold/nodes/bead"
	NodeShape "github.com/dtauraso/wirefold/tools/topology-vscode/Node/Shape"
	"os"

	"github.com/dtauraso/wirefold/nodes/rowevent"

	W "github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
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
				events []rowevent.RowEvent,
			) []byte {
				return SF.BuildViewStreamFrame(tick,
					camPX, camPY, camPZ, camR, camPosPhi, camPosTheta, camUpPhi, camUpTheta,
					NodeShape.CanonicalRingSurfacePointsFlat(),
					bead.CanonicalRingSurfacePointsFlat(),

					sceneTabNames, uint16(sceneTabSelected),
					toStreamEvents(events))
			})
	}
}
