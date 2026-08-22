package runtopology

import (
	W "github.com/dtauraso/wirefold/src/Input/dispatch"
	T "github.com/dtauraso/wirefold/src/Trace"
)

func wireViewStream(md *W.MoveDispatch) {
	md.UI.SetViewStream(func(tick uint32, events []T.RowEvent) {
		T.NewLog(T.OwnerView, 0).Append(events)
	})
}
