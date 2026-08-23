package scenebuild

import (
	"context"
	"fmt"

	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/TiltPanel"
	clock "github.com/dtauraso/wirefold/Categories/Clock"
	"github.com/dtauraso/wirefold/Categories/Node"
	beadanimation "github.com/dtauraso/wirefold/Categories/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/Categories/Node/TiltVectors"
	"github.com/dtauraso/wirefold/Categories/NodeKinds"
	"github.com/dtauraso/wirefold/Categories/NodeKinds/portwiring"
	"github.com/dtauraso/wirefold/Categories/Scene/loadspec"
	"github.com/dtauraso/wirefold/Categories/Scene/scenerun"
)

func buildNodes(
	ctx context.Context,
	spec loadspec.TopoSpec,
	md *scenerun.MoveDispatch,
	wiring EdgeWiring,
	nodeGeoms map[string]Node.NodeGeom,
	vectorOut, vectorIn map[string]chan TiltPanel.TiltVectorMsg,
	clk clock.Clock,
	speedSinks *SliderPanel.Sinks,
) ([]portwiring.Node, map[string]*beadanimation.Sender, error) {
	deps := portwiring.BuildDeps{
		LatticePoints: md.UI.LatticePoints,
		ClaimLatticeIn: func(name string) chan int32 {
			sceneToNodeLatticeIn := make(chan int32, scenerun.InboxDepth)
			md.Inboxes.ClaimLatticeIn(name, sceneToNodeLatticeIn)
			return sceneToNodeLatticeIn
		},
		ClaimTiltEditIn: func(name string) any {
			panelToNodeTiltEditIn := make(chan TiltVectors.TiltEditMsg, scenerun.InboxDepth)
			md.Inboxes.ClaimTiltEditIn(name, panelToNodeTiltEditIn)
			return panelToNodeTiltEditIn
		},
		ClaimSelfDriveGeom: func(name string) any {
			ng, ok := md.MR.NodeGeoms()[name]
			if !ok {
				return nil
			}
			ng.Clocks().CopyClockSrc()
			return ng
		},
	}

	outSink := map[string]*beadanimation.Sender{}
	nodes := make([]portwiring.Node, 0, len(spec.Nodes))
	for _, n := range spec.Nodes {
		bind, known := NodeKinds.BuilderFor(n.Type)
		if !known {
			return nil, nil, fmt.Errorf("scene names node type %q, which no kind builds — its directory "+
				"under Categories/NodeKinds must declare a Builder, and `go generate ./...` puts it "+
				"in the switch that BuilderFor reads", n.Type)
		}

		pb := NewPortBindings()
		pb.OutSink = outSink
		pb.Clock = clk
		pb.SpeedSinks = speedSinks
		pb.RT = md.RT
		pb.InteriorEmitters = md.Sw.InteriorEmittersPtr()
		pb.VectorOut = vectorOut
		pb.VectorIn = vectorIn
		wiring.BindPorts(&pb, n, bind.Ports)

		var tiltPhiIdx int32
		if n.TopTiltVectorPhiIdx != nil {
			tiltPhiIdx = *n.TopTiltVectorPhiIdx
		}
		nd, err := bind.Build(ctx, n.ID, n.Data, pb, tiltPhiIdx, deps)
		if err != nil {
			return nil, nil, fmt.Errorf("scenebuild: build node %q: %w", n.ID, err)
		}
		nodes = append(nodes, nd)
	}
	return nodes, outSink, nil
}
