package runtopology

import (
	T "github.com/dtauraso/wirefold/src/Trace"
	"os"


	W "github.com/dtauraso/wirefold/src/Input/dispatch"
	SceneB "github.com/dtauraso/wirefold/src/Scene"
)

func wireViewStream(md *W.MoveDispatch, viewFile *os.File, viewStreamWired bool) {
	if viewStreamWired {
		md.UI.SetViewStream(viewFile,
			func(tick uint32, events []T.RowEvent) []byte {
				T.NewLog(T.OwnerView, 0).Append(events)
				return SceneB.BuildViewStreamFrame(tick)
			})
	}
}
