// builders.go — reflection-driven port-manifest and node construction.
//
// Adding a kind: register one entry in kindRegistry. The struct fields
// determine the port manifest automatically:
//   - *Wiring.In       → PortIn
//   - *Wiring.Out      → PortOut
//   - Wiring.Broadcast  → PortBroadcast
//   - all other field types are ignored
//
// Non-channel fields can be populated from data.* JSON values via struct tags:
//   - wire:"data.<key>"  reads NodeData.<Key> where <Key> is <key> with its
//                        first letter uppercased. Any exported field on NodeData
//                        is reachable this way (e.g. data.init → NodeData.Init).
//                        Slice fields are copied, not aliased.
//                        Mismatched or absent fields are silently skipped.
//   - wire:"data.state"  reads NodeData.State[lowerFirst(fieldName)] (int).
//                        The map key is the struct field name with its first
//                        letter lowercased (e.g. field Held → key "held").
//
// The resolved per-port bindings data (PortDir/PortSpec/PortBindings and their
// dead-end fallbacks) live in port_bindings.go. Wiring each port field to its
// binding and this node's interior stream lives in port_wiring.go. The
// loader-facing kind registry (NodeBuilder/Registry/BuildRegistry) lives in
// node_registry.go.

package Wiring

import (
	"context"
	"fmt"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"reflect"

	T "github.com/dtauraso/wirefold/Trace"
)

var (
	tInPtr              = reflect.TypeFor[*wire.In]()
	tOutPtr             = reflect.TypeFor[*wire.Out]()
	tBroadcast          = reflect.TypeFor[wire.Broadcast]()
	tFireFunc           = reflect.TypeFor[func()]()
	tEmitBeadsFunc      = reflect.TypeFor[func(working, backup []int)]()
	tEmitHeldFunc       = reflect.TypeFor[func(held int)]()
	tEmitInputBeadsFunc = reflect.TypeFor[func(left, right int)]()
	tRefillSlideFunc    = reflect.TypeFor[func(clk wire.Clock, speedCh <-chan float64, beads []int)]()
	tTickFunc           = reflect.TypeFor[func() int64]()
)

// reflectPorts walks the exported fields of the struct pointed to by sample
// and returns a PortSpec for each channel field that carries int.
// Chan-of-chan fields and non-channel fields are silently skipped.
// Anonymous (embedded) struct fields are recursed so port fields promoted
// from an embedded struct (e.g. gatecommon.GateNode) are discovered.
func reflectPorts(sample any) []PortSpec {
	t := reflect.TypeOf(sample).Elem()
	return collectPorts(t)
}

func collectPorts(t reflect.Type) []PortSpec {
	var ports []PortSpec
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Anonymous {
			ft := f.Type
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				ports = append(ports, collectPorts(ft)...)
			}
			continue
		}
		switch f.Type {
		case tInPtr:
			ports = append(ports, PortSpec{Name: f.Name, Dir: PortIn})
		case tOutPtr:
			ports = append(ports, PortSpec{Name: f.Name, Dir: PortOut})
		case tBroadcast:
			ports = append(ports, PortSpec{Name: f.Name, Dir: PortBroadcast})
		}
	}
	return ports
}

// injectFunc sets the named func-typed field on v to fn, but only when the field
// exists, is settable, and has exactly the expected type `want`. This is the one
// shape every closure injection in reflectBuild shares (Fire, EmitGeometry, the
// Emit* bead closures, Tick); structs lacking the field are left
// untouched. Returns whether the field was set.
func injectFunc(v reflect.Value, name string, want reflect.Type, fn any) bool {
	f := v.FieldByName(name)
	if !f.IsValid() || !f.CanSet() || f.Type() != want {
		return false
	}
	f.Set(reflect.ValueOf(fn))
	return true
}

// reflectBuild wires pb into the struct pointed to by nodePtr via reflection,
// then returns it cast to Node. ctx is required when pb contains PacedWire
// bindings (paced mode); it is passed into the In/Out wrappers.
//
// The three concerns are split into named helpers, each called in the same
// order the original monolithic function performed them (behavior unchanged):
//   - injectClosures: Fire/EmitGeometry/EmitNodeBeads/EmitHeldBead/EmitInputBeads/
//     EmitRefillSlide/Tick closure injection.
//   - wirePorts: tag-driven (struct-shape-driven) port wiring — In/Out/Broadcast
//     fields set from pb's resolved bindings.
//   - populateData: wire:"data.<key>" / wire:"data.state" tag-driven data
//     population.
func reflectBuild(ctx context.Context, name string, data *NodeData, pb PortBindings, newNode func() any, tr *T.Trace, geom nodeGeom, partnerCenter partnerCenterFn) (wire.Node, error) {
	nodePtr := newNode()
	v := reflect.ValueOf(nodePtr).Elem()

	// getStream is THIS node's one shared interior-stream getter (lazy-cache-once — see
	// its doc comment): every closure/port that records a Fire/Recv/Send/NodeBead event
	// for this node calls the SAME func, so they all land on the SAME *interiorStream
	// instance (and share its cached bead-slot snapshot).
	getStream := newInteriorStreamGetter(name, pb)

	var sourceOuts []*wire.Out
	injectClosures(ctx, v, name, pb, tr, geom, &sourceOuts, partnerCenter, getStream)
	wirePorts(ctx, v, nodePtr, name, pb, tr, &sourceOuts, getStream)
	populateData(v, nodePtr, data)

	node, ok := nodePtr.(wire.Node)
	if !ok {
		return nil, fmt.Errorf("reflectBuild: %T does not implement Node", nodePtr)
	}
	return node, nil
}

// injectClosures injects every func-typed closure field reflectBuild supports
// (Fire, EmitGeometry, the Emit* interior-bead closures, and — when a shared
// clock is present — EmitRefillSlide/Tick). Each injection is a no-op
// when the struct lacks the matching field (injectFunc's contract). Returns the
// sourceOuts slice that EmitGeometry's closure reads for per-edge segments;
// wirePorts appends to it as it resolves each Out/Broadcast binding, and the
// closure (which fires later, at node startup) sees the completed slice.
// sourceOuts is owned by the caller (reflectBuild) and shared with wirePorts,
// which appends to it as it resolves each Out/Broadcast binding; the EmitGeometry
// closure reads through the same pointer so it sees the completed slice.
func injectClosures(ctx context.Context, v reflect.Value, name string, pb PortBindings, tr *T.Trace, geom nodeGeom, sourceOuts *[]*wire.Out, partnerCenter partnerCenterFn, getStream func() *interiorStream) {
	// Inject Fire closure if the struct has a `Fire func()` field. The closure
	// captures the node name so the node calls n.Fire() with no arguments and
	// cannot mis-name itself in the trace. The RowEvent flush below lands this Fire
	// on THIS node's OWN interior-stream frame (KindFire is fully decentralized — it
	// never rides the VIEW stream's fallback bucket) — this node's own Update goroutine
	// is the sole owner of when it fires, so it resolves its own NodeRow at the call
	// site (owner_events.go) via the shared interiorStream (getStream), never a shared
	// accumulator. writeEvents is nil-safe (no-op) when this node has no dedicated
	// interior fd (test builds without a loader).
	injectFunc(v, "Fire", tFireFunc, func() {
		if s := getStream(); s != nil {
			s.WriteEvents([]wire.RowEvent{{
				Kind: T.KindFire, NodeRow: s.nodeRow,
				PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1,
			}})
		}
	})

	// EmitGeometry is deliberately left UNINJECTED (the `EmitGeometry func()` field on
	// node structs stays nil, and Wiring.TryEmit(n.EmitGeometry) no-ops at node startup —
	// see node.go's TryEmit). It used to be the node's own Update-loop startup emit of
	// its node-geometry event AND each outgoing edge's segment (tr.NodeGeometry/
	// tr.Geometry), duplicating the identical values nodeMover/edgeMover's own
	// goroutine-start emit now produces (node_mover.go: nodeMover.run/edgeMover.run each
	// call their own emitGeometry/recomputeGeometry once before their loop). This node
	// struct field and sourceOuts (still populated by wirePorts, now otherwise unread by
	// this function) are kept only because deleting the struct fields themselves would
	// be a wider, unrelated churn across every node kind package; the field being
	// present-but-nil is equivalent to it not existing for every live code path
	// (geom/partnerCenter/sourceOuts are otherwise still referenced below/by callers,
	// so their params stay; nothing left in THIS function reads them for geometry).

	// Inject EmitNodeBeads closure if the struct has an `EmitNodeBeads
	// func(working, backup []int)` field (node 1's interior buffer). Emits one
	// node-bead event per present interior bead. The node's Update calls it with the
	// LIVE working/backup contents whenever the arrays change.
	injectFunc(v, "EmitNodeBeads", tEmitBeadsFunc, func(working, backup []int) {
		emitNodeBeads(tr, name, working, backup, getStream())
	})

	// Inject EmitHeldBead closure if the struct has an `EmitHeldBead func(held int)`
	// field (Time's interior held-value bead): a SINGLE centered node-bead
	// (slot 0,0 at offset 0,0,0) colored by the held value; held == -1 →
	// present=false (empty interior).
	injectFunc(v, "EmitHeldBead", tEmitHeldFunc, func(held int) {
		emitHeldBead(tr, name, held, getStream())
	})

	// Inject EmitInputBeads closure if the struct has an `EmitInputBeads
	// func(left, right int)` field (a gate's two-sided held-input beads): LEFT input
	// on the left of the node, RIGHT on the right; -1 = not held → present=false.
	injectFunc(v, "EmitInputBeads", tEmitInputBeadsFunc, func(left, right int) {
		emitInputBeads(tr, name, left, right, getStream())
	})

	// EmitRefillSlide func(clk Clock, speedCh <-chan float64, beads []int): the
	// clock-paced refill slide (the OLD backup beads slide DOWN from row 0 into row
	// 1 at wire-bead speed; a paused clock freezes it). The clock AND speed channel
	// are parameters the CALLER supplies at invocation time (its own
	// already-Copy()'d clock and its own SpeedCh — see input.Node.Update, which
	// calls n.EmitRefillSlide(clk, n.SpeedCh, beads) with the same copy/channel its
	// own loop paces on) rather than values captured here from pb.clock: capturing
	// the loader's origin in this closure would hand every future call a read into
	// a clock this goroutine never Copy()'d for itself (per-goroutine-clock.md
	// flagged this as a residual — this closure no longer needs pb.clock at all, so
	// it is unconditional). The speed channel must be threaded through too: this
	// slide runs its OWN blocking SleepCycle loop separate from the caller's main
	// loop, so it must poll ApplySpeedNonBlocking itself each cycle or a speed
	// change sent mid-slide sits unapplied until the slide finishes.
	injectFunc(v, "EmitRefillSlide", tRefillSlideFunc, func(clk wire.Clock, speedCh <-chan float64, beads []int) {
		emitRefillSlide(ctx, tr, name, clk, speedCh, beads)
	})

	// The remaining injections seed a node's OWN clock storage from the loader's
	// origin, once, at construction — a test build without a loader leaves pb.clock
	// nil and these fields stay unset (each node falls back to its own wall-clock/
	// no-loader behavior, e.g. gatecommon's defaultTick/defaultSleep).
	if pb.clock != nil {
		clk := pb.clock
		// Tick func() int64: current tick (pause-aware) off the origin clock. Used
		// only as a chan-mode/no-Out-yet fallback for "now" by gatecommon.GateNode;
		// the paced path takes its own Copy() of the Clock field below instead.
		injectFunc(v, "Tick", tTickFunc, func() int64 { return clk.Tick() })
		// Clock Wiring.Clock: the node's OWN clock storage, seeded from the loader's
		// origin so the node's goroutine can Copy() it exactly once at its own
		// start — this field is
		// never read repeatedly by anything outside the node's own goroutine, and it
		// is never reached through a port. Only fields typed exactly Wiring.Clock
		// (e.g. input.Node.Clock, gatecommon.GateNode.Clock) receive this; other
		// nodes are unaffected.
		tClockType := reflect.TypeFor[wire.Clock]()
		injectFunc(v, "Clock", tClockType, clk)
	}

	injectSpeedChans(v, pb)
}

// populateData performs tag-driven data population: wire:"data.<key>" or
// wire:"data.state" struct tags on nodePtr's fields, read from data (a nil
// data leaves every tagged field untouched, matching the original guard).
func populateData(v reflect.Value, nodePtr any, data *NodeData) {
	if data == nil {
		return
	}
	t := reflect.TypeOf(nodePtr).Elem()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("wire")
		if tag == "" {
			continue
		}
		fv := v.Field(i)
		if !fv.CanSet() {
			continue
		}
		const dataPrefix = "data."
		const stateTag = "data.state"
		if tag == stateTag {
			// key is field name with first letter lowercased. The seed is
			// OPTIONAL: an absent key leaves the constructor default untouched
			// (the empty sentinel for held-bearing kinds), so "unset" can never
			// collide with a legitimately-held 0. Only a present key — a real
			// authored starting value — overrides the default.
			key := lowerFirst(f.Name)
			if val, ok := data.State[key]; ok {
				fv.Set(reflect.ValueOf(val))
			}
		} else if len(tag) > len(dataPrefix) && tag[:len(dataPrefix)] == dataPrefix {
			key := tag[len(dataPrefix):]
			if len(key) == 0 {
				continue
			}
			exportedKey := exportedFieldName(key)
			src := reflect.ValueOf(data).Elem().FieldByName(exportedKey)
			if !src.IsValid() || src.Type() != fv.Type() {
				continue
			}
			if src.Kind() == reflect.Slice {
				if src.IsNil() {
					continue
				}
				cp := reflect.MakeSlice(src.Type(), src.Len(), src.Len())
				reflect.Copy(cp, src)
				fv.Set(cp)
			} else {
				fv.Set(src)
			}
		}
	}
}

// verticalRingNormal and flatRingNormal are the two great-circle ring normals
// streamed on every node-geometry event so TS never hardcodes ring orientation.
// vertical: ring stands upright (normal points along +Z world axis).
// flat: ring lies flat (normal points along +Y world axis, Three y-up convention).
const (
	verticalRingNormalX, verticalRingNormalY, verticalRingNormalZ = 0.0, 0.0, 1.0
	flatRingNormalX, flatRingNormalY, flatRingNormalZ             = 0.0, 1.0, 0.0
)
