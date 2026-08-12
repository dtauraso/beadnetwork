package dispatch

import (
	"context"
	"sync"
)

func (md *MoveDispatch) Start(ctx context.Context) *sync.WaitGroup {
	return md.MR.Start(ctx)
}
