package dispatch

import (
	"context"
	"sync"
)

func (md *MoveDispatch) Start(ctx context.Context) *sync.WaitGroup {
	md.Rules.Start(ctx)
	return md.MR.Start(ctx)
}
