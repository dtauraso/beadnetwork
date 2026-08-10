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
// import Wiring — several already do, for BuildArgs itself — while Wiring imports NO kind
// at all. The blank imports that run each kind's init() live in kinds_generated.go at the
// repo root (package main). So a kind may legally receive Wiring types, and BuildArgs can
// name PortBindings/nodegeom.NodeGeom/NodeData. It could NOT live in nodes/wire, which Wiring
// imports and which therefore cannot name any of them.

package Wiring

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/interior"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
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
	geom nodegeom.NodeGeom

	// tiltThetaIdx is this node's PERSISTED tilt-vector-angle index
	// (topo_spec.go's specNode.TopTiltVectorThetaIdx, dereferenced with a 0 default),
	// threaded in from the loaded spec so a kind that owns its own index (TiltVectorAngleSeed)
	// can seed its own struct field from the SAME value the mover used to seed itself with
	// (build.go's old nm.topTiltVectorThetaIdx assignment) — one load-time value, read by
	// whichever goroutine ends up owning it. There is no φ counterpart
	// (task/drop-tilt-vector-phi) — the tilt vector is θ-only end to end.
	tiltThetaIdx int32

	// sourceOuts collects every Out this node resolves. reflectBuild shared one slice
	// between its closure injection and its port wiring; the same slice is threaded here
	// so an Out resolved by a.Out()/a.Broadcast() is still recorded.
	sourceOuts *[]*wire.Out
	// getStream is THIS node's one shared interior-stream getter (lazy-cache-once), so
	// every closure and port that records an event for this node lands on the SAME
	// *interior.InteriorStream instance and shares its cached bead-slot snapshot.
	getStream func() *interior.InteriorStream
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

// The rest of BuildArgs's accessor methods live in their own files, grouped by what they
// reach: build_args_ports.go (In/Out/Broadcast/DriveOut), build_args_beads.go (Fire and the
// Emit* interior-bead-emission closures), build_args_tilt_vector.go (the tilt-vector-angle
// seed/edit channel and the tilt-vector channel ends), build_args_lattice.go (lattice point
// seed/edit channel), build_args_clock.go (Clock/Tick/SpeedCh), build_args_state.go
// (StateSeed/Data), and build_args_selfdrive.go (ClaimSelfDrive) — the same in-package
// method split this branch already did to node_geometry.go and paced_wire.go.

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
		Build: func(ctx context.Context, name string, data *NodeData, pb PortBindings, tr *T.Trace, geom nodegeom.NodeGeom, tiltThetaIdx int32) (wire.Node, error) {
			var sourceOuts []*wire.Out
			return build(BuildArgs{
				ctx: ctx, name: name, data: data, pb: pb, tr: tr,
				geom:            geom,
				sourceOuts:      &sourceOuts,
				getStream:       newInteriorStreamGetter(name, pb),
				driveSlotClaims: map[int]string{},
				tiltThetaIdx:    tiltThetaIdx,
			})
		},
	}
}
