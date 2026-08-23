package NodeKinds

import (
	"context"

	"github.com/dtauraso/wirefold/Categories/Scene/loadspec"
)

type Builder interface {
	Ports() []struct {
		Name string
		Dir  int
	}

	Build(ctx context.Context, name string, data *loadspec.NodeData, pb any, tiltPhiIdx int32, deps any) (any, error)
}
