package Startup

import (
	"context"
	"fmt"

	"github.com/dtauraso/beadnetwork/Categories/Scene/Topology"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/TiltPanel"
	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	NodeBuf "github.com/dtauraso/beadnetwork/Categories/Node"
	beadanimation "github.com/dtauraso/beadnetwork/Categories/Node/BeadAnimation"
	Ports "github.com/dtauraso/beadnetwork/Categories/Node/Ports"
	"github.com/dtauraso/beadnetwork/Categories/Node/TiltVectors"
	"github.com/dtauraso/beadnetwork/Categories/NodeKinds"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Dispatch"
)

func BuildNodes(
	ctx context.Context,
	spec Topology.TopoSpec,
	md *Dispatch.MoveDispatch,
	lines Ports.EdgeLines,
	nodeGeoms map[string]NodeBuf.NodeGeom,
	vectorOut, vectorIn map[string]chan TiltPanel.TiltVectorMsg,
	clk clock.Clock,
	speedSinks *SliderPanel.Sinks,
) ([]BuiltNode, map[string]*beadanimation.Sender, error) {
	deps := BuildDeps{
		LatticePoints: md.UI.LatticePoints,
		ClaimLatticeIn: func(name string) chan int32 {
			sceneToNodeLatticeIn := make(chan int32, Dispatch.InboxDepth)
			md.Inboxes.ClaimLatticeIn(name, sceneToNodeLatticeIn)
			return sceneToNodeLatticeIn
		},
		ClaimTiltEditIn: func(name string) any {
			panelToNodeTiltEditIn := make(chan TiltVectors.TiltEditMsg, Dispatch.InboxDepth)
			md.Inboxes.ClaimTiltEditIn(name, panelToNodeTiltEditIn)
			return panelToNodeTiltEditIn
		},
		ClaimSelfDriveGeom: func(name string) any {
			ng, ok := md.MR.NodeGeoms()[name]
			if !ok {
				return nil
			}
			return ng
		},
	}

	outSink := map[string]*beadanimation.Sender{}
	nodes := make([]BuiltNode, 0, len(spec.Nodes))
	for _, n := range spec.Nodes {
		bind, known := NodeKinds.BuilderFor(n.Type)
		if !known {
			return nil, nil, fmt.Errorf("scene names node type %q, which no kind builds — its directory "+
				"under Categories/NodeKinds must declare a Builder, and `go generate ./...` puts it "+
				"in the switch that BuilderFor reads", n.Type)
		}

		pb := Ports.NewPortBindings()
		pb.OutSink = outSink
		pb.Clock = clk
		pb.SpeedSinks = speedSinks
		pb.NodeRowOf = md.RT.NodeRowFor
		pb.InteriorEmitters = md.Sw.InteriorEmittersPtr()
		pb.VectorOut = vectorOut
		pb.VectorIn = vectorIn
		bp := bind.Ports()
		declared := make([]Ports.PortSpec, len(bp))
		for i, p := range bp {
			declared[i] = Ports.PortSpec{Name: p.Name, Dir: Ports.PortDir(p.Dir)}
		}
		lines.BindPorts(&pb, n.ID, func(port string) beadanimation.SendRule { return Topology.NodeSendRule(n, port) }, declared)

		var tiltPhiIdx int32
		if n.TopTiltVectorPhiIdx != nil {
			tiltPhiIdx = *n.TopTiltVectorPhiIdx
		}
		nd, err := bind.Build(ctx, n.ID, n.Data, pb, tiltPhiIdx, deps)
		if err != nil {
			return nil, nil, fmt.Errorf("wiring: build node %q: %w", n.ID, err)
		}
		built, ok := nd.(BuiltNode)
		if !ok {
			return nil, nil, fmt.Errorf("wiring: kind %q built something with no Update method for node %q", n.Type, n.ID)
		}
		nodes = append(nodes, built)
	}
	return nodes, outSink, nil
}
