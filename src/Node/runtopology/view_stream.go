package runtopology

import (
	"os"

	"github.com/dtauraso/wirefold/src/Node/rowevent"

	W "github.com/dtauraso/wirefold/src/Node/Wiring/dispatch"
	SceneB "github.com/dtauraso/wirefold/src/Scene"
)

func wireViewStream(md *W.MoveDispatch, viewFile *os.File, viewStreamWired bool, sceneTabNames []string, sceneTabSelected int) {
	if viewStreamWired {
		md.UI.SetViewStream(viewFile,
			func(tick uint32, events []rowevent.RowEvent) []byte {
				return SceneB.BuildViewStreamFrame(tick,
					sceneTabNames, uint16(sceneTabSelected),
					events)
			})
	}
}
