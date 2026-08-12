package nodeapi

import "context"

type Node interface {
	Update(ctx context.Context)
}

func TryEmit(fn func()) {
	if fn != nil {
		fn()
	}
}
