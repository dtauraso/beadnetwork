// port_bindings.go — resolved per-port bindings data + accessors: PortDir/
// PortSpec describe a port's shape, PortBindings holds the resolved paced
// wires (singleBinding/broadcastBinding) keyed by port name, and the
// dead-end fallbacks used when a port name has no paced binding.

package Wiring

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// PortDir describes which direction a port flows.
type PortDir int

const (
	PortIn        PortDir = iota
	PortOut               // single output
	PortBroadcast         // slice output ([]chan<- int)
)

// PortSpec describes one port on a node kind.
type PortSpec struct {
	Name string
	Dir  PortDir
}

// PortBindings holds resolved PacedWires keyed by port name.
// For PortBroadcast ports, use AppendBroadcastWithHandle.
// A port name with no paced binding resolves to a dead-end chan wrapper
// (deadEndIn/deadEndOut/deadEndOutSlice) that neither sends nor receives.
type PortBindings struct {
	// singlePaced holds the resolved paced binding for each single In/Out port.
	// broadcastPaced holds the per-element bindings for each Broadcast fan-out port.
	// Consolidating the formerly-parallel per-edge maps into one struct keeps
	// every field of a binding together and impossible to index-mismatch.
	singlePaced    map[string]singleBinding
	broadcastPaced map[string][]broadcastBinding
	// outSink, when non-nil, collects every paced *Out built for this node keyed
	// by "node.handle" so the loader can index Outs by edge for node-move
	// travel-time updates. Render/run paths leave it nil.
	outSink map[string]*wire.Out
	// clock is the loader's ORIGIN clock, read only by BuildArgs.Clock/Tick
	// (never by a port): it seeds a node's bare `Clock Wiring.Clock` field and the
	// `Tick func() int64` closure at construction. Per per-goroutine-clock.md API
	// demolition, ports/wires no longer hold or hand out a clock at all — a node's
	// own goroutine does exactly one Copy() of its Clock field at its own start.
	// Test builds without a loader leave this nil, and such nodes' Clock/Tick
	// fields simply stay unset (their own zero-value fallback, e.g. gatecommon's
	// defaultTick/defaultSleep).
	clock wire.Clock
	// speedSinks accumulates the SEND end of every speed channel created for
	// this node during construction (one per clock-owning goroutine the node
	// spawns — see injectSpeedChans). It points at the loader's build-wide slice
	// (buildCtx.speedSinks) so every node's channels land in the one list
	// LoadTopology hands back to stdin_reader. nil in test builds with no
	// loader — injectSpeedChans then skips channel creation entirely (a node
	// with no speed channel just never hears a speed change, same as it never
	// had a clock to speed up before this plan).
	speedSinks *[]chan float64
	// md, when non-nil, gives injectClosures's interior-bead Emit* closures access to
	// this node's OWN dedicated interior fd (md.sw.interiorOuts, keyed by node id) and the
	// injected interior-frame builder (md.sw.buildInteriorFrame) — see
	// MoveDispatch.SetNodeStreams / memory/feedback_no_single_writer_bridge.md. Set once
	// per node at construction (loader.go's buildNodes: pb.md = b.md); nil in test builds
	// with no loader, in which case the Emit* closures just skip the dedicated-stream
	// write (tr.NodeBead alone, unchanged).
	md *MoveDispatch
	// vectorOut/vectorIn, when non-nil, map a node id to the SEND/RECEIVE end of its
	// own dedicated tilt-vector channel (tilt_vector_channel.go) — built once, for
	// EVERY node in the whole load, by build.go's allocateVectorChannels phase and
	// shared read-only by every node's BuildArgs.VectorOut/VectorIn call. A node id
	// absent from either map (every kind but Node1/Node2, or an edge whose other
	// endpoint didn't also ask for one) resolves to nil, which the non-blocking
	// send/receive helpers already treat as "nothing wired" — same fallback shape as
	// every other unwired-port case in this file.
	vectorOut map[string]chan TiltVectorMsg
	vectorIn  map[string]chan TiltVectorMsg
}

// singleBinding is the resolved paced binding for one single port. For an INPUT
// port only pw is set (SetSinglePaced); an OUTPUT port also carries its per-edge
// send rule and own geometry (SetSinglePacedRule). The zero value (pw == nil)
// means "no paced binding — fall back to a dead-end chan".
type singleBinding struct {
	pw    *wire.PacedWire
	rule  wire.SendRule
	steps int
	seg   wireSegment
	label string
}

// broadcastBinding is one fan-out element of an Broadcast port: its shared dest wire,
// the concrete source handle (e.g. "ToNext0"), per-edge send rule, and that
// edge's own bead-step count / segment / TS label.
type broadcastBinding struct {
	pw     *wire.PacedWire
	handle string
	rule   wire.SendRule
	steps  int
	seg    wireSegment
	label  string
}

func newPortBindings() PortBindings {
	return PortBindings{
		singlePaced:    map[string]singleBinding{},
		broadcastPaced: map[string][]broadcastBinding{},
	}
}

func (pb *PortBindings) SetSinglePaced(name string, pw *wire.PacedWire) {
	pb.singlePaced[name] = singleBinding{pw: pw}
}

// SetSinglePacedRule binds a single paced output with its per-edge send rule,
// that edge's own bead-step count (docs/bead-lattice.md "The count"), its
// straight-segment endpoints (so the bead's position stream evaluates the exact
// drawn segment), and the TS edge id (label) so the node's EmitGeometry closure
// can stream the segment.
func (pb *PortBindings) SetSinglePacedRule(name string, pw *wire.PacedWire, rule wire.SendRule, steps int, seg wireSegment, label string) {
	pb.singlePaced[name] = singleBinding{pw: pw, rule: rule, steps: steps, seg: seg, label: label}
}

// AppendBroadcastWithHandle is like AppendMultiPaced but records the exact
// source handle (e.g. "ToNext0"), the per-edge send rule, that edge's own
// bead-step count, its straight-segment endpoints, and the TS edge id (label)
// so the node's EmitGeometry closure can stream the segment.
func (pb *PortBindings) AppendBroadcastWithHandle(name, handle string, pw *wire.PacedWire, rule wire.SendRule, steps int, seg wireSegment, label string) {
	pb.broadcastPaced[name] = append(pb.broadcastPaced[name], broadcastBinding{
		pw: pw, handle: handle, rule: rule, steps: steps, seg: seg, label: label,
	})
}

// deadEndIn returns a fresh unbuffered-in-effect receive-only chan for a port
// name with no paced binding. It is never fed a value; it exists only so an
// unwired In field has a non-nil channel to hold.
func (pb *PortBindings) deadEndIn(name string) <-chan int {
	return make(chan int, 1) // chan-name-ok: dead-end placeholder; wire identity is the port `name` (map key)
}

// deadEndOut is deadEndIn's send-only counterpart for an unwired Out field.
func (pb *PortBindings) deadEndOut(name string) chan<- int {
	return make(chan int, 1) // chan-name-ok: dead-end placeholder; wire identity is the port `name` (map key)
}

// deadEndOutSlice is deadEndOut's counterpart for an unwired Broadcast field:
// there is no fan-out recorded for this port name, so it resolves to an empty
// slice of dead-end sends.
func (pb *PortBindings) deadEndOutSlice(name string) []chan<- int {
	return nil
}
