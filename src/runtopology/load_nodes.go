package runtopology

import (
	"fmt"

	beadanimation "github.com/dtauraso/wirefold/src/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/src/Input/dispatch"
	"github.com/dtauraso/wirefold/src/Input/inputcodec"
	"github.com/dtauraso/wirefold/src/Node/Wiring/kindreg"
	"github.com/dtauraso/wirefold/src/Node/Wiring/loadspec"
	"github.com/dtauraso/wirefold/src/Node/Wiring/movemsg"
	"github.com/dtauraso/wirefold/src/Node/Wiring/moverreg"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/src/Node/Wiring/portwiring"
	"github.com/dtauraso/wirefold/src/NodeKinds/nodeapi"
)

func (b *buildCtx) buildNodes() error {

	deps := kindreg.BuildDeps{
		LatticePoints: b.md.UI.LatticePoints,
		ClaimLatticeIn: func(name string) chan int32 {
			sceneToNodeLatticeIn := make(chan int32, moverreg.InboxDepth)
			b.md.Inboxes.ClaimLatticeIn(name, sceneToNodeLatticeIn)
			return sceneToNodeLatticeIn
		},
		ClaimTiltEditIn: func(name string) chan movemsg.TiltEditMsg {
			panelToNodeTiltEditIn := make(chan movemsg.TiltEditMsg, moverreg.InboxDepth)
			b.md.Inboxes.ClaimTiltEditIn(name, panelToNodeTiltEditIn)
			return panelToNodeTiltEditIn
		},
		ClaimSelfDriveGeom: func(name string) *nodeactor.NodeGeometry {
			ng, ok := b.md.MR.NodeGeoms()[name]
			if !ok {
				return nil
			}
			ng.CopyClockSrc()
			return ng
		},
	}
	outSink := map[string]*beadanimation.Sender{}
	nodes := make([]nodeapi.Node, 0, len(b.spec.Nodes))
	for _, n := range b.spec.Nodes {
		bind := kindreg.Registry[n.Type]
		pb := portwiring.NewPortBindings()
		pb.OutSink = outSink
		pb.Clock = b.clk

		pb.SpeedSinks = &b.speedSinks

		pb.RT = b.md.RT
		pb.InteriorEmitters = b.md.Sw.InteriorEmittersPtr()
		pb.BuildInteriorFrame = b.md.Sw.BuildInteriorFramePtr()
		pb.VectorOut = b.vectorOutByNode
		pb.VectorIn = b.vectorInByNode

		for _, port := range bind.Ports {
			switch port.Dir {
			case portwiring.PortIn:
				dk, ok := b.inbound[n.ID][port.Name]
				if ok {
					pb.SetSinglePaced(port.Name, b.destRun[dk])
				}

			case portwiring.PortOut:
				labels := b.outbound[n.ID][port.Name]
				if len(labels) > 0 {

					rule := loadspec.NodeSendRule(n, port.Name)
					lbl := labels[0]
					pb.SetSinglePacedRule(port.Name, b.edgeRun[lbl], rule, lbl)
				}

			case portwiring.PortBroadcast:
				labels := b.outbound[n.ID][port.Name]
				handles := b.outboundHandle[n.ID][port.Name]
				for i, lbl := range labels {
					handle := port.Name
					if i < len(handles) {
						handle = handles[i]
					}

					rule := loadspec.NodeSendRule(n, handle)
					pb.AppendBroadcastWithHandle(port.Name, handle, b.edgeRun[lbl], rule, lbl)
				}

			}
		}

		var tiltPhiIdx int32
		if n.TopTiltVectorPhiIdx != nil {
			tiltPhiIdx = *n.TopTiltVectorPhiIdx
		}
		nd, err := bind.Build(b.ctx, n.ID, n.Data, pb, b.nodeGeoms[n.ID], tiltPhiIdx, deps)
		if err != nil {
			return fmt.Errorf("LoadTopology: build node %q: %w", n.ID, err)
		}
		nodes = append(nodes, nd)
	}
	b.outSink = outSink
	b.nodes = nodes
	return nil
}

func bindDispatch(md *dispatch.MoveDispatch, outSink map[string]*beadanimation.Sender, destRun map[string]*beadanimation.BeadLine) {
	md.MR.Bind(outSink, inputcodec.SlotRegistry(destRun), md.RT.EdgeRowForPair)
}
