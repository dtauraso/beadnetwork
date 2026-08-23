package portwiring

import "context"

type Node interface {
	Update(ctx context.Context)
}
