// Package portwiring holds the resolved per-port bindings data + accessors split out of
// nodes/Wiring (god-object decomposition): PortDir/PortSpec describe a port's shape,
// PortBindings holds the resolved paced wires (singleBinding/broadcastBinding) keyed by
// port name, and the dead-end fallbacks used when a port name has no paced binding.
//
// PortBindings used to hold a *MoveDispatch back-reference (the `md` field) so a node's
// port-wiring closures could reach the interior-stream fds and row table MoveDispatch
// owns — the exact "leaf reaches UP to the hub for a value" defect this move fixes.
// PortBindings now takes those values directly (RT, InteriorOuts, DriveOuts,
// BuildInteriorFrame — see their own doc comments below for the pointer-indirection each
// needs), set by nodes/Wiring's build_nodes.go at construction; this package needs
// nothing from nodes/Wiring, so nodes/Wiring can import it with no cycle.
package portwiring

import (
	"io"

	"github.com/dtauraso/wirefold/nodes/Wiring/rowtables"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

// DriveSlotsPerNode is a local copy of stream_wiring.go's own driveSlotsPerNode/
// Buffer.DriveSlotsPerNode (2) — kept here rather than importing nodes/Wiring (which
// would cycle) so PortBindings.DriveOuts' field type matches the array size
// streamWiring.driveOuts actually uses, the same "duplicate the small constant"
// precedent BufInteriorSlotsPerNode (port_wiring.go) already follows.
const DriveSlotsPerNode = 2

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
	// OutSink, when non-nil, collects every paced *Out built for this node keyed
	// by "node.handle" so the loader can index Outs by edge for node-move
	// travel-time updates. Render/run paths leave it nil.
	OutSink map[string]*wire.Out
	// Clock is the loader's ORIGIN clock, read only by BuildArgs.Clock/Tick
	// (never by a port): it seeds a node's bare `Clock Wiring.Clock` field and the
	// `Tick func() int64` closure at construction. Per per-goroutine-clock.md API
	// demolition, ports/wires no longer hold or hand out a clock at all — a node's
	// own goroutine does exactly one Copy() of its Clock field at its own start.
	// Test builds without a loader leave this zero, and such nodes' Clock/Tick
	// fields simply stay unset (their own zero-value fallback, e.g. gatecommon's
	// defaultTick/defaultSleep).
	Clock clock.Clock
	// SpeedSinks accumulates the SEND end of every speed channel created for
	// this node during construction (one per clock-owning goroutine the node
	// spawns — see injectSpeedChans). It points at the loader's build-wide slice
	// (buildCtx.speedSinks) so every node's channels land in the one list
	// LoadTopology hands back to stdin_reader. nil in test builds with no
	// loader — injectSpeedChans then skips channel creation entirely (a node
	// with no speed channel just never hears a speed change, same as it never
	// had a clock to speed up before this plan).
	SpeedSinks *[]chan float64
	// RT is a COPY of the loader's row-identity tables (rowtables.RowTables), read only by
	// NodeRowFor to resolve a node id to its buffer row for a stream frame's nodeRow field.
	// Safe to copy by value: RowTables.Build runs once, before buildNodes (buildMoveDispatch
	// precedes buildNodes — build.go), and is never mutated afterward (its own doc comment).
	// Zero value (bare test builds with no loader) makes NodeRowFor always report ok=false.
	RT rowtables.RowTables
	// InteriorOuts/DriveOuts/BuildInteriorFrame give injectClosures's interior-bead Emit*
	// closures (and BuildArgs.DriveOut) access to this node's OWN dedicated interior fd
	// (keyed by node id) and drive-slot fds, plus the injected interior-frame builder — see
	// MoveDispatch.SetNodeStreams / memory/feedback_no_single_writer_bridge.md. These are
	// POINTERS to streamWiring's own fields (build_nodes.go's buildNodes: pb.InteriorOuts =
	// &b.md.sw.interiorOuts, etc.), not copies — SetNodeStreams populates the underlying
	// maps/func LATE, only after LoadTopology returns (main.go), so a plain copy taken here
	// (before that happens) would freeze at nil forever; the pointer indirection is what
	// lets these getters see the values SetNodeStreams installs afterward. nil in test
	// builds with no loader, in which case the Emit* closures just skip the dedicated-stream
	// write (tr.NodeBead alone, unchanged).
	InteriorOuts       *map[string]io.Writer
	DriveOuts          *map[string][DriveSlotsPerNode]io.Writer
	BuildInteriorFrame *func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []wire.RowEvent) []byte
	// VectorOut/VectorIn, when non-nil, map a node id to the SEND/RECEIVE end of its
	// own dedicated tilt-vector channel (tilt_vector_channel.go) — built once, for
	// EVERY node in the whole load, by build.go's allocateVectorChannels phase and
	// shared read-only by every node's BuildArgs.VectorOut/VectorIn call. A node id
	// absent from either map (every kind but PairNode, or an edge whose other
	// endpoint didn't also ask for one) resolves to nil, which the non-blocking
	// send/receive helpers already treat as "nothing wired" — same fallback shape as
	// every other unwired-port case in this file.
	VectorOut map[string]chan tiltvector.TiltVectorMsg
	VectorIn  map[string]chan tiltvector.TiltVectorMsg
}

// singleBinding is the resolved paced binding for one single port. For an INPUT
// port only pw is set (SetSinglePaced); an OUTPUT port also carries its per-edge
// send rule and own geometry (SetSinglePacedRule). The zero value (pw == nil)
// means "no paced binding — fall back to a dead-end chan".
type singleBinding struct {
	pw    *wire.PacedWire
	rule  wire.SendRule
	steps int
	seg   wire.WireSegment
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
	seg    wire.WireSegment
	label  string
}

// NewPortBindings constructs an empty PortBindings.
func NewPortBindings() PortBindings {
	return PortBindings{
		singlePaced:    map[string]singleBinding{},
		broadcastPaced: map[string][]broadcastBinding{},
	}
}

func (pb *PortBindings) SetSinglePaced(name string, pw *wire.PacedWire) {
	pb.singlePaced[name] = singleBinding{pw: pw}
}

// SetSinglePacedRule binds a single paced output with its per-edge send rule,
// that edge's own bead-step count (docs/bead-model/bead-lattice.md "The count"), its
// straight-segment endpoints (so the bead's position stream evaluates the exact
// drawn segment), and the TS edge id (label) so the node's EmitGeometry closure
// can stream the segment.
func (pb *PortBindings) SetSinglePacedRule(name string, pw *wire.PacedWire, rule wire.SendRule, steps int, seg wire.WireSegment, label string) {
	pb.singlePaced[name] = singleBinding{pw: pw, rule: rule, steps: steps, seg: seg, label: label}
}

// AppendBroadcastWithHandle is like AppendMultiPaced but records the exact
// source handle (e.g. "ToNext0"), the per-edge send rule, that edge's own
// bead-step count, its straight-segment endpoints, and the TS edge id (label)
// so the node's EmitGeometry closure can stream the segment.
func (pb *PortBindings) AppendBroadcastWithHandle(name, handle string, pw *wire.PacedWire, rule wire.SendRule, steps int, seg wire.WireSegment, label string) {
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
