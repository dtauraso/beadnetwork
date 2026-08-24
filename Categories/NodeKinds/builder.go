package NodeKinds

import (
	"context"

	NodeBuf "github.com/dtauraso/beadnetwork/Categories/Node"
)

type Builder interface {
	Ports() []struct {
		Name string
		Dir  int
	}

	Build(ctx context.Context, name string, data *NodeBuf.NodeData, pb any, tiltPhiIdx int32, deps any) (any, error)
}
