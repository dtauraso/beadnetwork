package scenebuild

import (
	"context"
	"fmt"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/src/Chrome/Panels/TiltPanel"
	clock "github.com/dtauraso/wirefold/src/Clock"
	beadanimation "github.com/dtauraso/wirefold/src/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/src/Node/movemsg"
	"github.com/dtauraso/wirefold/src/Node/moverreg"
	"github.com/dtauraso/wirefold/src/Node/nodeactor"
	"github.com/dtauraso/wirefold/src/Node/nodegeom"
	"github.com/dtauraso/wirefold/src/NodeKinds/kindreg"
	"github.com/dtauraso/wirefold/src/NodeKinds/nodeapi"
	"github.com/dtauraso/wirefold/src/NodeKinds/portwiring"
	"github.com/dtauraso/wirefold/src/Scene/loadspec"
	"github.com/dtauraso/wirefold/src/Scene/scenerun"
)

func buildNodes(
	ctx context.Context,
	spec loadspec.TopoSpec,
	md *scenerun.MoveDispatch,
	wiring kindreg.EdgeWiring,
	nodeGeoms map[string]nodegeom.NodeGeom,
	vectorOut, vectorIn map[string]chan TiltPanel.TiltVectorMsg,
	clk clock.Clock,
	speedSinks *SliderPanel.Sinks,
) ([]nodeapi.Node, map[string]*beadanimation.Sender, error) {
	deps := kindreg.BuildDeps{
		LatticePoints: md.UI.LatticePoints,
		ClaimLatticeIn: func(name string) chan int32 {
			sceneToNodeLatticeIn := make(chan int32, moverreg.InboxDepth)
			md.Inboxes.ClaimLatticeIn(name, sceneToNodeLatticeIn)
			return sceneToNodeLatticeIn
		},
		ClaimTiltEditIn: func(name string) chan movemsg.TiltEditMsg {
			panelToNodeTiltEditIn := make(chan movemsg.TiltEditMsg, moverreg.InboxDepth)
			md.Inboxes.ClaimTiltEditIn(name, panelToNodeTiltEditIn)
			return panelToNodeTiltEditIn
		},
		ClaimSelfDriveGeom: func(name string) *nodeactor.NodeGeometry {
			ng, ok := md.MR.NodeGeoms()[name]
			if !ok {
				return nil
			}
			ng.CopyClockSrc()
			return ng
		},
	}

	outSink := map[string]*beadanimation.Sender{}
	nodes := make([]nodeapi.Node, 0, len(spec.Nodes))
	for _, n := range spec.Nodes {
		bind := kindreg.Registry[n.Type]

		pb := portwiring.NewPortBindings()
		pb.OutSink = outSink
		pb.Clock = clk
		pb.SpeedSinks = speedSinks
		pb.RT = md.RT
		pb.InteriorEmitters = md.Sw.InteriorEmittersPtr()
		pb.BuildInteriorFrame = md.Sw.BuildInteriorFramePtr()
		pb.VectorOut = vectorOut
		pb.VectorIn = vectorIn
		wiring.BindPorts(&pb, n, bind.Ports)

		var tiltPhiIdx int32
		if n.TopTiltVectorPhiIdx != nil {
			tiltPhiIdx = *n.TopTiltVectorPhiIdx
		}
		nd, err := bind.Build(ctx, n.ID, n.Data, pb, nodeGeoms[n.ID], tiltPhiIdx, deps)
		if err != nil {
			return nil, nil, fmt.Errorf("scenebuild: build node %q: %w", n.ID, err)
		}
		nodes = append(nodes, nd)
	}
	return nodes, outSink, nil
}
