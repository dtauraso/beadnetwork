package runtopology

import (
	"os"

	"github.com/dtauraso/wirefold/nodes/rowevent"

	W "github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	SF "github.com/dtauraso/wirefold/tools/topology-vscode/src/Buffer/streamframe"
)

func wireViewStream(md *W.MoveDispatch, viewFile *os.File, viewStreamWired bool, sceneTabNames []string, sceneTabSelected int) {
	if viewStreamWired {
		md.UI.SetViewStream(viewFile,
			func(tick uint32, events []rowevent.RowEvent) []byte {
				return SF.BuildViewStreamFrame(tick,
					sceneTabNames, uint16(sceneTabSelected),
					events)
			})
	}
}
