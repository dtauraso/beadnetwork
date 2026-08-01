// build_args.go — the seam that lets a node kind CONSTRUCT ITSELF.
//
// Before this, every kind registered an empty struct and one central function
// (reflectBuild) filled it in by reflection: matching field NAMES ("Fire", "Clock",
// "SpeedCh") and field TYPES to decide what to inject, and struct TAGS to decide what
// data to populate. The knowledge of how a Time node is built lived in Wiring, not in
// nodes/Time.
//
// The cost of that is silence. Nothing checks a field name against what the injector
// looks for, so renaming the `Fire` field to anything else does not fail to compile — it
// simply stays nil and the node quietly never traces a fire. Same for a mistyped tag, or a
// port field whose type drifts.
//
// With BuildArgs a kind writes plain assignments:
//
//	n := &Time{}
//	n.Fire = a.Fire()
//	n.In = a.In("In")
//
// and a rename is a compile error. Nothing here is new BEHAVIOUR — every method below
// returns exactly what reflectBuild's corresponding injection produced; the difference is
// only that the kind asks for it by name instead of being handed it by reflection.
//
// DEPENDENCY DIRECTION (why this type lives in Wiring and not in nodes/wire): the kinds
// import Wiring — several already do, for Wiring.NoValue — while Wiring imports NO kind
// at all. The blank imports that run each kind's init() live in kinds_generated.go at the
// repo root (package main). So a kind may legally receive Wiring types, and BuildArgs can
// name PortBindings/nodeGeom/NodeData. It could NOT live in nodes/wire, which Wiring
// imports and which therefore cannot name any of them.

package Wiring

import (
	"context"
	"fmt"
	"os"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

// BuildArgs carries everything a node kind needs to construct itself. It is passed as ONE
// struct rather than as separate parameters so that adding an input later does not edit
// every kind's build func — with 14 kinds, that churn is the dominant cost of the
// alternative.
//
// The fields are unexported: a kind reaches them through the methods below, which is what
// keeps the construction rules (dead-end fallbacks, row resolution, stream sharing) in one
// place even though the CHOICE of what to build now belongs to the kind.
type BuildArgs struct {
	ctx  context.Context
	name string
	data *NodeData
	pb   PortBindings
	tr   *T.Trace
	geom nodeGeom

	// sourceOuts collects every Out this node resolves. reflectBuild shared one slice
	// between its closure injection and its port wiring; the same slice is threaded here
	// so an Out resolved by a.Out()/a.Broadcast() is still recorded.
	sourceOuts *[]*wire.Out
	// getStream is THIS node's one shared interior-stream getter (lazy-cache-once), so
	// every closure and port that records an event for this node lands on the SAME
	// *interiorStream instance and shares its cached bead-slot snapshot.
	getStream func() *interiorStream
	// driveSlotClaims tracks, for THIS node's ONE build call, which drive slot each
	// DriveOut(portName, slot) call has already claimed (slot -> claiming port name).
	// Allocated once per node in RegisterBuilder's wrapper and never shared across
	// nodes, so this is plain single-threaded bookkeeping during LoadTopology's build
	// phase (before any node/DriveHeld goroutine exists) — no lock needed. A second
	// DriveOut call naming a slot already in this map is the wiring-time failure
	// requirement 1 asks for: it does not construct a second DrivenOut wrapping the
	// SAME underlying drive fd (which would silently reintroduce a two-goroutine-one-fd
	// desync the moment both driven Outs got handed to two DriveHeld goroutines) — see
	// DriveOut below.
	driveSlotClaims map[int]string
}

// Name is this node's spec id.
func (a BuildArgs) Name() string { return a.name }

// Ctx is the build context; it is passed into the paced In/Out wrappers.
func (a BuildArgs) Ctx() context.Context { return a.ctx }

// In resolves an input port by its SPEC name. Paced when the loader bound a wire to it,
// dead-end otherwise (same fallback reflectBuild's wireInPort applied).
func (a BuildArgs) In(portName string) *wire.In {
	return newInPort(portName, a.ctx, a.name, a.pb, a.tr, a.getStream)
}

// Out resolves a single output port by its SPEC name.
func (a BuildArgs) Out(portName string) *wire.Out {
	return newOutPort(portName, a.ctx, a.name, a.pb, a.tr, a.sourceOuts, a.getStream)
}

// Broadcast resolves a fan-out output port by its SPEC name.
func (a BuildArgs) Broadcast(portName string) wire.Broadcast {
	return newBroadcastPort(portName, a.ctx, a.name, a.pb, a.tr, a.sourceOuts, a.getStream)
}

// DriveOut resolves an output port that will be DRIVEN by its own gatecommon.DriveHeld
// goroutine (a SEPARATE goroutine from this node's own Update loop — Pulse/PulseLeft/
// PulseRight/holdflip's shape), instead of Out(). It routes the port's eventSink through
// a DEDICATED per-(node, slot) drive stream (newDriveStreamGetter, Buffer.StreamKindDrive)
// rather than this node's shared getStream — the fix for the framing desync documented in
// docs/interior-stream-framing.md: two goroutines (this node's Update loop and its
// DriveHeld goroutine) must never write the same *interiorStream/fd. slot distinguishes
// multiple DriveHeld outputs on ONE node (Pulse's Out=slot 0, OutFanout=slot 1 — see
// Buffer.DriveSlotsPerNode's doc comment for the current max) and must be a DIFFERENT
// value for each such call on the same node; passing the same slot to two driven outputs
// on one node would make them share a stream, reintroducing this exact bug. A plain
// (non-DriveHeld) Out — only ever written from this node's own Update goroutine — should
// keep using Out(), not DriveOut(): its writes already satisfy the single-writer
// invariant via the shared getStream, and giving it a drive slot would burn an fd for no
// reason.
func (a BuildArgs) DriveOut(portName string, slot int) DrivenOut {
	if a.driveSlotClaims != nil {
		if prior, claimed := a.driveSlotClaims[slot]; claimed {
			// Wiring-time failure, reported not panicked (main.go's own stream-fd-
			// mismatch posture — see its "stream-fd mismatch" Fprintf calls): this runs
			// during LoadTopology's single-threaded build phase, before this node's
			// Update or DriveHeld goroutines exist, so there is no crash-loop risk and
			// nothing to unwind. The SECOND claimant gets a dead DrivenOut (zero value:
			// nil-safe Wired/Paced/Steps/PlaceDrivenAt all degrade to "not driving"),
			// same fallback shape as every other absent-stream case in this file.
			fmt.Fprintf(os.Stderr,
				"drive-stream collision: node %q slot %d already claimed by port %q; port %q "+
					"cannot also claim it — a DriveHeld goroutine for %q would share %q's dedicated "+
					"fd, which is exactly the two-goroutines-one-fd desync docs/interior-stream-"+
					"framing.md documents. %q's driven output stays unwired (drives nothing) instead. "+
					"Give it its own slot (Buffer.DriveSlotsPerNode).\n",
				a.name, slot, prior, portName, portName, prior, portName)
			return DrivenOut{}
		}
		a.driveSlotClaims[slot] = portName
	}
	out := newOutPort(portName, a.ctx, a.name, a.pb, a.tr, a.sourceOuts, newDriveStreamGetter(a.name, slot, a.pb))
	return newDrivenOut(out)
}

// Fire returns this node's fire-trace closure. The node name is captured, so a node
// cannot mis-name itself in the trace. The event lands on this node's OWN interior
// stream; nil-safe when the node has no dedicated interior fd (test builds).
func (a BuildArgs) Fire() func() {
	getStream := a.getStream
	return func() {
		if s := getStream(); s != nil {
			s.WriteEvents([]wire.RowEvent{{
				Kind: T.KindFire, NodeRow: s.nodeRow,
				PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1,
			}})
		}
	}
}

// EmitNodeBeads returns the interior working/backup bead emitter.
func (a BuildArgs) EmitNodeBeads() func(working, backup []int) {
	tr, name, getStream := a.tr, a.name, a.getStream
	return func(working, backup []int) { emitNodeBeads(tr, name, working, backup, getStream()) }
}

// EmitHeldBead returns the single centered held-value bead emitter (held == NoValue
// renders as an empty interior).
func (a BuildArgs) EmitHeldBead() func(held int) {
	tr, name, getStream := a.tr, a.name, a.getStream
	return func(held int) { emitHeldBead(tr, name, held, getStream()) }
}

// EmitInputBeads returns a gate's two-sided held-input bead emitter.
func (a BuildArgs) EmitInputBeads() func(left, right int) {
	tr, name, getStream := a.tr, a.name, a.getStream
	return func(left, right int) { emitInputBeads(tr, name, left, right, getStream()) }
}

// EmitRefillSlide returns the clock-paced refill-slide emitter. The clock and speed
// channel are supplied by the CALLER at invocation time — its own already-Copy()'d clock
// and its own SpeedCh — never captured here, per per-goroutine-clock.md.
func (a BuildArgs) EmitRefillSlide() func(clk wire.Clock, speedCh <-chan float64, beads []int) {
	ctx, tr, name := a.ctx, a.tr, a.name
	return func(clk wire.Clock, speedCh <-chan float64, beads []int) {
		emitRefillSlide(ctx, tr, name, clk, speedCh, beads)
	}
}

// Clock returns the loader's clock ORIGIN, or nil on a test build with no loader. The
// owning goroutine Copy()s it exactly once at its own start — this hands over the origin,
// not a per-goroutine clock.
func (a BuildArgs) Clock() wire.Clock { return a.pb.clock }

// Tick returns a read of the loader clock's current tick, or nil when there is no clock.
func (a BuildArgs) Tick() func() int64 {
	if a.pb.clock == nil {
		return nil
	}
	clk := a.pb.clock
	return func() int64 { return clk.Tick() }
}

// SpeedCh allocates THIS node's buffered-1 speed-delivery channel and registers it with
// the loader's sink, so a speed change reaches this goroutine's own clock copy. Returns
// nil when there is no sink (test builds with no loader). Call it ONCE per node: each
// call allocates and registers another channel, and only the last one a node keeps would
// ever be drained.
func (a BuildArgs) SpeedCh() <-chan float64 {
	if a.pb.speedSinks == nil {
		return nil
	}
	speedCh := make(chan float64, 1)
	*a.pb.speedSinks = append(*a.pb.speedSinks, speedCh)
	return speedCh
}

// StateSeed returns the persisted `data.state` seed for one field, or def when the spec
// carries none. key is the struct field name with its first letter lowercased — the same
// convention the `wire:"data.state"` tag used (field Held -> key "held").
//
// The seed is OPTIONAL by design: an absent key leaves the kind's own default untouched,
// so "unset" can never collide with a legitimately-held 0.
func (a BuildArgs) StateSeed(key string, def int) int {
	if a.data == nil || a.data.State == nil {
		return def
	}
	if v, ok := a.data.State[key]; ok {
		return v
	}
	return def
}

// Data exposes the raw spec data block for the `wire:"data.<key>"` fields that have no
// dedicated accessor above. Nil when the spec carries no data block.
func (a BuildArgs) Data() *NodeData { return a.data }

// RegisterBuilder is how a kind claims ownership of its own construction. Called from the
// kind's init(); the ports it declares are the ports the loader validates against, and
// build is called once per node of that kind.
//
// BuildRegistry skips any kind already present in Registry, so a self-registered kind is
// never overwritten by the reflection fallback — that is what lets the 14 kinds migrate
// ONE AT A TIME while the rest keep working untouched.
func RegisterBuilder(kind string, ports []PortSpec, build func(BuildArgs) (wire.Node, error)) {
	if _, exists := Registry[kind]; exists {
		panic("Wiring.RegisterBuilder: kind already registered: " + kind)
	}
	Registry[kind] = NodeBuilder{
		Ports: ports,
		Build: func(ctx context.Context, name string, data *NodeData, pb PortBindings, tr *T.Trace, geom nodeGeom) (wire.Node, error) {
			var sourceOuts []*wire.Out
			return build(BuildArgs{
				ctx: ctx, name: name, data: data, pb: pb, tr: tr,
				geom:            geom,
				sourceOuts:      &sourceOuts,
				getStream:       newInteriorStreamGetter(name, pb),
				driveSlotClaims: map[int]string{},
			})
		},
	}
}
