package runtopology

import (
	"os"

	B "github.com/dtauraso/wirefold/src/Buffer"

	W "github.com/dtauraso/wirefold/src/Input/dispatch"
	SceneB "github.com/dtauraso/wirefold/src/Scene"
)

func wireViewStream(md *W.MoveDispatch, viewFile *os.File, viewStreamWired bool) {
	if viewStreamWired {
		md.UI.SetViewStream(viewFile,
			func(tick uint32, events []B.RowEvent) []byte {
				return SceneB.BuildViewStreamFrame(tick, events)
			})
	}
}
