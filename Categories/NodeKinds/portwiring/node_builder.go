package portwiring

import (
	"context"

	"github.com/dtauraso/wirefold/Categories/Scene/loadspec"
)

type BuildDeps struct {
	LatticePoints int32

	ClaimLatticeIn func(name string) chan int32

	ClaimTiltEditIn func(name string) any

	ClaimSelfDriveGeom func(name string) any
}

type NodeBuilder struct {
	Ports []PortSpec
	Build func(ctx context.Context, name string, data *loadspec.NodeData, pb PortBindings, tiltPhiIdx int32, deps BuildDeps) (Node, error)
}
