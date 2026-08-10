// build_nodes.go — the node-construction phase of buildFromSpec: builds each
// node from the wire allocation and edge maps computed by earlier phases, then
// binds the resulting Outs and dest wires into the already-built MoveDispatch.

package Wiring

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// buildNodes builds each node from the wire allocation and edge maps computed by
// earlier phases. outSink collects every paced source Out keyed by "node.handle"
// so node-move can update per-edge travel-time on the Out.
func (b *buildCtx) buildNodes() error {
	// currentBuildMD gives BuildArgs methods that need more of MoveDispatch than
	// PortBindings can portably carry (LatticePointsSeed/LatticeIn, TiltEditIn,
	// ClaimSelfDrive — all Wiring-internal state: md.ui, md.inboxes, md.mr) a way to
	// reach it without PortBindings holding a *MoveDispatch back-reference. Set once,
	// here, before any node is built; buildNodes runs single-threaded (LoadTopology's
	// build phase, before any node/mover goroutine exists — see RowTables' own doc
	// comment for the same "built before Start" reasoning), so a package-level var read
	// back by those methods is safe, matching Registry's own existing package-level state.
	currentBuildMD = b.md
	outSink := map[string]*wire.Out{}
	nodes := make([]wire.Node, 0, len(b.spec.Nodes))
	for _, n := range b.spec.Nodes {
		bind := Registry[n.Type]
		pb := newPortBindings()
		pb.outSink = outSink
		pb.clock = b.clk // shared clock for clock-paced interior animation (Input refill slide)
		// &b.speedSinks (not a fresh slice per node): every node's channels append
		// onto the SAME build-wide accumulator, so LoadTopology's one returned
		// list carries every clock-owning goroutine across the whole build.
		pb.speedSinks = &b.speedSinks
		// interiorOuts/driveOuts/buildInteriorFrame give injectClosures's interior-bead
		// Emit* closures access to this node's OWN dedicated interior fd (keyed by node
		// id) + the injected frame builder — the SECOND emitting goroutine per node
		// (memory/feedback_no_single_writer_bridge.md). These are POINTERS into b.md.sw's
		// own fields (see PortBindings' doc comment for why): nil-checked before writing,
		// and stay effectively nil (pointing at an empty/nil map) until SetNodeStreams
		// runs (main.go, after LoadTopology returns).
		pb.rt = b.md.RT
		pb.interiorOuts = &b.md.sw.interiorOuts
		pb.driveOuts = &b.md.sw.driveOuts
		pb.buildInteriorFrame = &b.md.sw.buildInteriorFrame
		pb.vectorOut = b.vectorOutByNode
		pb.vectorIn = b.vectorInByNode

		for _, port := range bind.Ports {
			switch port.Dir {
			case PortIn:
				dk, ok := b.inbound[n.ID][port.Name]
				if ok {
					pb.SetSinglePaced(port.Name, b.destWire[dk])
				}
				// If no inbound edge, a.In() falls back to a dead-end chan.

			case PortOut:
				labels := b.outbound[n.ID][port.Name]
				if len(labels) > 0 {
					// Look up wire by destination of the first outbound edge.
					// For fan-in, the destination port owns the wire.
					// Send rule is node-owned, keyed by this output port name.
					rule := nodeSendRule(n, port.Name)
					lbl := labels[0]
					pb.SetSinglePacedRule(port.Name, b.edgeWire[lbl], rule, b.edgeSteps[lbl], b.edgeSegments[lbl], lbl)
				}
				// If no outbound edge, a.Out() falls back to a dead-end chan.

			case PortBroadcast:
				labels := b.outbound[n.ID][port.Name]
				handles := b.outboundHandle[n.ID][port.Name]
				for i, lbl := range labels {
					handle := port.Name
					if i < len(handles) {
						handle = handles[i]
					}
					// Per-port (per fan-out element): the rule is keyed by the
					// concrete output port name (sourceHandle, e.g. "ToNext0").
					rule := nodeSendRule(n, handle)
					pb.AppendBroadcastWithHandle(port.Name, handle, b.edgeWire[lbl], rule, b.edgeSteps[lbl], b.edgeSegments[lbl], lbl)
				}
				// If no outbound edges, builder falls back to a dead-end slice.
			}
		}

		var tiltThetaIdx int32
		if n.TopTiltVectorThetaIdx != nil {
			tiltThetaIdx = *n.TopTiltVectorThetaIdx
		}
		nd, err := bind.Build(b.ctx, n.ID, n.Data, pb, b.tr, b.nodeGeoms[n.ID], tiltThetaIdx)
		if err != nil {
			return fmt.Errorf("LoadTopology: build node %q: %w", n.ID, err)
		}
		nodes = append(nodes, nd)
	}
	b.outSink = outSink
	b.nodes = nodes
	return nil
}

// bindDispatch binds per-edge source Outs and dest wires into each edgeMover so
// a node-move updates per-edge travel-time.
func (b *buildCtx) bindDispatch() {
	b.md.Bind(b.outSink, inputcodec.SlotRegistry(b.destWire))
}
