// dispatch_edit.go — ROUTING for a decoded editor→Go message.
//
// One job: given a inputcodec.StdinMsg, decide what runs. The read loop (stdin_reader.go) hands this
// file a message and never looks at it again; the handlers that say what an attribute MEANS
// live in dispatch_apply.go. What is here is the tables in between — op → update, entity kind →
// per-kind handler, attribute → its effect — plus the two top-level types that need no
// table (raw-input and save).
//
// Every level is deliberately FORWARD-COMPAT: an unknown op/kind/attr is ignored rather
// than an error, because a newer webview may send a message this binary predates. That is
// also why the tables are sentinel-fenced and guarded — a silently-ignored edit looks
// exactly like a working one from the outside (tools/bridge/check-input-attr-dispatched.sh,
// tools/bridge/check-edit-op-parity.sh).
//
// This file (and dispatch_apply.go) moved here from nodes/Wiring/dispatch (§30,
// docs/planning/movedispatch-decomposition.md) — the bulk-move pass that landed the stdin
// cluster once md.ctx stopped being the last field these handlers needed to read directly
// (ApplyEdit now takes ctx as an explicit parameter, matching what its sole production
// caller — runtopology's gesture actor goroutine — already has in scope). Every handler here
// takes *dispatch.MoveDispatch (an exported type) and reaches it only through exported
// fields/methods, so this package now DOES import nodes/Wiring/dispatch for that type —
// stdin_reader.go's own framing/back-pressure/shutdown machinery below does not and still
// takes its three dispatch operations as plain function values (Handlers), so the framing
// half of this package still never needs to know MoveDispatch exists.
package stdinreader

import (
	"context"

	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

// HandleRawInputMsg hands a raw pointer/wheel event + stateless raycast hit to the
// gesture state machine, which owns gesture bookkeeping and produces camera/topology
// changes. ctx comes from the caller (runtopology's gesture actor goroutine already has
// one in scope, matching ApplyEdit's own shape). Fire-and-forget — nothing on this seam
// triggers delivery.
func HandleRawInputMsg(ctx context.Context, msg inputcodec.StdinMsg, slotReg inputcodec.SlotRegistry, md *dispatch.MoveDispatch, tr *T.Trace) {
	if md != nil && msg.Event != nil {
		md.HandleRawInput(ctx, *msg.Event, slotReg, tr)
	}
}

// HandleSaveMsg persists Go's OWN authoritative scene state (overlay visibility and the
// scene sphere) in response to the bare "save" command. The camera pose is already
// continuously flushed elsewhere (scene_camera_persist.go).
func HandleSaveMsg(md *dispatch.MoveDispatch) {
	if md == nil {
		return
	}
	md.Persist.Overlays().Schedule(md.UI.OV)
	// Persist the scene sphere immediately (not debounced) so save reliably activates
	// the polar-load path (scene_sphere_persist.go LoadSceneSphere) — until the sphere
	// is in sphere.json, reload stays on cartesian x/y/z.
	md.Persist.Sphere().Schedule(md.UI.SceneSphere)
}

// viewstate.OverlayToggles (the FLAG name → OverlayState flip-method table) and
// viewstate.OverlayFlagTraceKind (the FLAG name → Trace.Kind* string, so applyUpdate's toggle
// case can hand emitViewFrame the ONE event that flag's toggle logged) are both GENERATED
// into nodes/Wiring/viewstate/overlay_state.go from the SAME OVERLAY_FLAG_NAMES source, so
// they cannot drift apart by flag-name set — a flag missing its Trace kind constant fails
// the Go build rather than silently omitting the emit. See
// tools/gen-node-defs/overlay_gen.go's writeOverlayGen.

// applyEdit dispatches one geometry-CRUD edit by its op. The sole op is update
// (matched by value so it stays invisible to the message-kind-parity guard, which
// fences only top-level msg.Type kinds).
//
//   - update: set an ATTRIBUTE on a typed entity (msg.Kind). Live entities are
//     overlays (attr "toggle": flip the named flag) and clock (attr "speed").
//     (Camera / node-move / port-anchor edits are produced in-process by the
//     gesture FSM from raw-input, so they never cross this seam.)
//
// The create/delete edge ops were removed end-to-end: no TS sender ever emitted them,
// and the create path's only live trigger — a port-drop gesture — tore down a live
// wire's in-flight beads via PacedWire.Restore. The destination-keyed inputcodec.SlotRegistry
// stays live for delivery/movers (md.MR.Bind), but the reader no longer indexes it here.
//
// Unknown ops/kinds/attrs are ignored (forward-compat).
// EDIT_OPS_START
var editOps = map[string]func(context.Context, inputcodec.StdinMsg, *dispatch.MoveDispatch, *T.Trace, []chan float64){
	"update": applyUpdate,
}

// EDIT_OPS_END

// ApplyEdit takes ctx as an explicit parameter rather than reading it off md — md.ctx
// stays unexported by design (§30, docs/planning/movedispatch-decomposition.md); the
// caller (runtopology's gesture actor) already has ctx in scope from its own goroutine.
func ApplyEdit(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, tr *T.Trace, speedSinks []chan float64) {
	if h, ok := editOps[msg.Op]; ok {
		h(ctx, msg, md, tr, speedSinks)
	}
}

// applyUpdate routes an op=="update" edit to the entity named by msg.Kind, setting the
// requested attribute. Live entities: overlays (toggle one flag) and clock (set the
// playback-speed multiplier — Go-owned state, the slider just signals the value).
// Unknown kinds/attrs are ignored (forward-compat). Dispatch is a NESTED table: the
// top-level kind→handler table below routes to a per-kind handler that owns its own
// attr-level table (applyUpdateClock / applyUpdateOverlays), so each kind can keep
// kind-specific behavior that runs regardless of which attr matched (e.g. overlays'
// unconditional persist-on-change) without distorting the attr dispatch itself.
// EDIT_UPDATE_KINDS_START
var updateKindHandlers = map[string]func(context.Context, inputcodec.StdinMsg, *dispatch.MoveDispatch, *T.Trace, []chan float64){
	"clock":         applyUpdateClock,
	"overlays":      applyUpdateOverlays,
	"distanceGroup": applyUpdateDistanceGroup,
	"scene":         applyUpdateScene,
	"tiltVector":    applyUpdateTiltVector,
}

// EDIT_UPDATE_KINDS_END

func applyUpdate(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, tr *T.Trace, speedSinks []chan float64) {
	if h, ok := updateKindHandlers[msg.Kind]; ok {
		h(ctx, msg, md, tr, speedSinks)
	}
}

// clockAttrHandlers is the attr-level table for kind=="clock".
var clockAttrHandlers = map[string]func(msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, speedSinks []chan float64){
	"speed": func(msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, speedSinks []chan float64) {
		// msg.Num carries the playback multiplier in QUARTER-UNITS (an integer 0..8:
		// the SpeedSlider's six-value table 0, 0.25, 0.5, 0.75, 1, 2 sent as 0, 1, 2, 3,
		// 4, 8) — input-layout.ts's encodeClockSpeed and input_codec.go's decode agree on
		// this exact integer form so a fractional multiplier survives msg.Num's int type
		// with no truncation. Divide back to the real multiplier here, the one place that
		// interprets it.
		userSpeed := float64(msg.Num) / 4.0
		// divisor is this scene's own ClockDivisor, resolved once at load into md.clockDivisor
		// (LoadSpeed) — GO-OWNED and never crosses the bridge. userSpeed stays the number the
		// slider shows and speed.json persists; only the EFFECTIVE rate reaching the clocks
		// (EffectiveClockSpeed, scene_speed_persist.go) is scaled, so a live edit and the
		// load-time seed can never disagree.
		divisor := 1.0
		if md != nil {
			divisor = md.UI.ClockDivisor
		}
		effective := scenepersist.EffectiveClockSpeed(userSpeed, divisor)
		// SetSpeed left the Clock INTERFACE in the per-goroutine-clock demolition (item 4):
		// nothing outside a goroutine's own copy may mutate it anymore, since a copy is
		// owned by exactly one goroutine.
		// Delivery (per-goroutine-clock.md "Delivery"): broadcast the new speed to
		// EVERY clock-owning goroutine's own channel (collected once, at load,
		// before any goroutine spawned — see LoadTopology's speedSinks return
		// value). This RunStdinReader goroutine is the sole writer of any of these
		// channels; SendSpeedNonBlocking never blocks on a
		// receiver that is asleep or never reads (latest-wins coalescing).
		for _, ch := range speedSinks {
			clock.SendSpeedNonBlocking(ch, effective)
		}
		if md == nil {
			return
		}
		// Mirror the USER's speed on this goroutine so the VIEW frame's Speed column
		// reflects it (the webview slider reads this back — no local default state,
		// memory/feedback_reflect_dont_create_store.md), and persist it UNSCALED (scene-level,
		// this view-owner goroutine's own file — .claude/rules/persistence-ownership.md). The
		// divisor never crosses the bridge and never reaches disk.
		md.UI.Speed = userSpeed
		md.Persist.Speed().Schedule(userSpeed)
		md.UI.EmitViewFrame(nil)
	},
}

// overlayAttrHandlers is the attr-level table for kind=="overlays".
var overlayAttrHandlers = map[string]func(msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, tr *T.Trace){
	"toggle": func(msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, tr *T.Trace) {
		// Flip the named flag — Go owns the state; TS just signals the flip.
		if fn, ok := viewstate.OverlayToggles[msg.Flag]; ok {
			fn(&md.UI.OV, tr)
			// Structured buffer counterpart of the "pole-toggle-go" debug breadcrumb
			// OverlayState.Toggle* logs for scene/node poles (viewstate.OverlayFlagBreadcrumbScope,
			// viewstate/overlay_state.go): only this goroutine (RunStdinReader's dispatch loop, the
			// VIEW stream's owner) ever calls md.UI.EmitBreadcrumb, so it is safe here with
			// no lock. Value=visible(0/1); the scope ("scene"/"nodes") rides the
			// sanctioned free-form Text column since it names which pole-flag fired, not
			// a typed row ref.
			if scope, ok := viewstate.OverlayFlagBreadcrumbScope[msg.Flag]; ok {
				md.UI.EmitBreadcrumb(wire.RowEvent{Label: T.BreadcrumbPoleToggleGo, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1, Value: int32(boolU8(viewstate.OverlayFlagValue[msg.Flag](&md.UI.OV))), Text: scope})
			}
			// Decentralized (Step C, memory/feedback_no_single_writer_bridge.md): this goroutine (the sole
			// caller of every overlay Toggle*) also writes its own VIEW frame directly,
			// carrying the one flag that just changed — matches the ONE tr.X(bool) event
			// the toggle already logged.
			if kind, ok := viewstate.OverlayFlagTraceKind[msg.Flag]; ok {
				md.UI.EmitViewFrame([]wire.RowEvent{{Kind: kind, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1}})
			}
		}
	},
}
