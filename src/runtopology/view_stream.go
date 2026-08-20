package runtopology

import (
	"os"

	B "github.com/dtauraso/wirefold/src/schema/buffer-layout"

	W "github.com/dtauraso/wirefold/src/Input/dispatch"
	SceneB "github.com/dtauraso/wirefold/src/Scene"
)

func wireViewStream(md *W.MoveDispatch, viewFile *os.File, viewStreamWired bool, sceneTabNames []string, sceneTabSelected int) {
	if viewStreamWired {
		md.UI.SetViewStream(viewFile,
			func(tick uint32, events []B.RowEvent) []byte {
				return SceneB.BuildViewStreamFrame(tick,
					sceneTabNames, uint16(sceneTabSelected),
					events)
			})
	}
}
