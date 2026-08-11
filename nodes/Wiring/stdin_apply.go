// stdin_apply.go — WHAT AN ATTRIBUTE EDIT MEANS.
//
// This file is the APPLICATION-SEMANTICS half of the editor→Go bridge's stdin seam: one
// apply* handler per entity kind, each interpreting the attribute an "edit" record asked
// to set (clock speed, distanceGroup length, tiltVector theta/reset/start, scene
// selection/lattice/create/delete, overlay flag toggles).
//
// The OTHER half — HOW BYTES BECOME A DISPATCH — stays in stdin_reader.go (framing) and stdin_dispatch.go (routing): the wire
// structs, the framed-binary read loop and maxFrameBytes framing, raw-input forwarding and
// the bare save command, and the dispatch TABLES (editOps, updateKindHandlers,
// clockAttrHandlers, overlayAttrHandlers) that route into the handlers below. The tables
// are the ROUTING surface and belong with the reader; what an attribute MEANS belongs
// here. The edit vocabulary is unchanged by this split — the sole op is still update
// (.claude/rules/bridge-surface.md: new capability is a new entity kind or attribute,
// never a new op).

package Wiring

import (
	"strconv"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
)

func applyUpdateClock(msg inputcodec.StdinMsg, md *MoveDispatch, tr *T.Trace, speedSinks []chan float64) {
	if h, ok := clockAttrHandlers[msg.Attr]; ok {
		h(msg, md, speedSinks)
	}
}

// applyUpdateDistanceGroup handles kind=="distanceGroup" attr=="length": one arrow
// click on the "distance home button" toolbar panel. msg.Num is the group index
// (0/1/2, into distanceGroupOrder — time/input/gate); msg.Flag is "up" (×1.1) or
// "down" (÷1.1). Go owns the group definitions and the ×1.1 math (ApplyDistanceGroupTarget,
// distance_groups.go) — the panel sends only which group and which direction.
func applyUpdateDistanceGroup(msg inputcodec.StdinMsg, md *MoveDispatch, tr *T.Trace, speedSinks []chan float64) {
	if md == nil || msg.Attr != "length" {
		return
	}
	dir := -1
	if msg.Flag == "up" {
		dir = 1
	}
	if applyDistanceGroupTarget(md.ctx, &md.ui, &md.mr, &md.lq, msg.Num, dir) {
		md.emitViewFrame(nil)
	}
}

// applyUpdateTiltVector handles kind=="tiltVector" attr=="theta"/"reset"/"start":
// one arrow click on the per-node tilt-vector-angle panel, one click of the RESET button
// (TiltResetButton.tsx), or one click of the START TILT button (TiltVectorButtons.tsx).
// msg.Num is the target node's buffer ROW (never its id/name — no sidecar on this wire,
// .claude/rules/bridge-surface.md); for theta, msg.Flag is "up" (+1 step) or "down"
// (-1 step) — "reset" and "start" carry no direction. ROW ID = NODE ID - 1 by construction
// (persistence-ownership.md), so the row resolves to an id directly — no reverse lookup
// table needed. There is no φ any more — the tilt vector is θ-only end to end
// (task/drop-tilt-vector-phi).
//
// theta/reset each have two routes, decided by whether the target node's OWN kind
// claimed BuildArgs.TiltEditIn at build time (PairNode today — the only kind that owns
// its tilt index independently, per the straightening loop's firing rule): md.sendTiltEdit
// tries that node's dedicated channel first and reports whether one exists. When it does
// NOT (every other kind), this falls back to the old path — md.sendMove onto the node's
// mover (movemsg.KindTiltVectorAngle / movemsg.KindTiltVectorReset) — so the index write +
// persist + re-emit still run on that node's own mover goroutine, unchanged for every kind
// but the pair. A theta click now moves the index by exactly one step and sends/places
// nothing else (task/pair-node-owns-itself split); reset places NO bead either.
//
// start has ONE route only: it is meaningless off the pair's own vector exchange, so it is
// sent to md.sendTiltEdit and simply dropped when that channel does not exist (see the
// "start" branch below) — no mover fallback, unlike theta/reset.
func applyUpdateTiltVector(msg inputcodec.StdinMsg, md *MoveDispatch, tr *T.Trace, speedSinks []chan float64) {
	if md == nil || (msg.Attr != "theta" && msg.Attr != "reset" && msg.Attr != "start") {
		return
	}
	id := strconv.Itoa(msg.Num + 1)
	if _, ok := md.mr.nodeGeoms[id]; !ok {
		return
	}
	if msg.Attr == "reset" {
		// Done setting: the slider's speed governs again. See HumanEditSpeed.
		BroadcastSpeed(speedSinks, md.ui.SliderSpeed())
		if sendTiltEdit(&md.inboxes, md.ctx, id, movemsg.TiltEditMsg{Reset: true}) {
			return
		}
		sendMove(&md.mr, md.ctx, id, movemsg.Msg{Kind: movemsg.KindTiltVectorReset, NodeID: id})
		return
	}
	if msg.Attr == "start" {
		// Done setting — the exchange is about to run, and running it is exactly what the
		// slider's number is about. Sent BEFORE the Start edit so the first cycle of the
		// exchange is already at the intended speed rather than one cycle of human speed.
		BroadcastSpeed(speedSinks, md.ui.SliderSpeed())
		// Start only exists on the pair kind's own dedicated channel (PairNode's
		// VectorOut/outgoingVector) — there is no mover-owned fallback, unlike
		// theta/phi/reset: a kind that never claimed BuildArgs.TiltEditIn has no vector
		// exchange to open, so a Start for it is simply a no-op.
		sendTiltEdit(&md.inboxes, md.ctx, id, movemsg.TiltEditMsg{Start: true})
		return
	}
	// A ▲/▼ ANGLE CLICK — the user is SETTING a tilt. Run every clock at human speed until
	// they start or reset, so the click is answered now rather than a scaled cycle from now
	// (see HumanEditSpeed for why this is not the slider's business). Sent BEFORE the edit,
	// so the very node about to drain this edit is already cycling at that speed.
	BroadcastSpeed(speedSinks, HumanEditSpeed)
	up := msg.Flag == "up"
	if sendTiltEdit(&md.inboxes, md.ctx, id, movemsg.TiltEditMsg{Axis: msg.Attr, Up: up}) {
		return
	}
	sendMove(&md.mr, md.ctx, id, movemsg.Msg{Kind: movemsg.KindTiltVectorAngle, NodeID: id, Axis: msg.Attr, Bool: up})
}

// applyUpdateScene handles kind=="scene" attr=="selected" (one click on the scene tab
// strip) and attr=="latticePoints" (a scene-level pair-lattice point-count edit).
//
// "selected": msg.Num is the tab INDEX into Wiring.SceneTabs — Go owns the tab list, the
// labels it streams on the VIEW frame, and the selection; the strip sends only which tab
// was hit. The switch itself (persist, then end this run so the runner's respawn loads the
// other scene) is SelectScene's — see scene_switch.go for why there is no in-process
// rebuild.
//
// "latticePoints": msg.Num is the new point count. Go owns the valid range (4..64,
// multiples of 4 — nodes/PairNode's newRing panics outside it) and rejects anything else by
// simply ignoring the edit — this is a decoded EXTERNAL message, so an out-of-range value
// must never reach newRing and panic the process. A valid value is persisted to
// view/lattice.json and broadcast to every pair node's own LatticeIn channel
// (BroadcastLatticePoints); it does NOT touch md.ui.speed/clockDivisor — a lattice size
// has no "setting" mode the way a tilt angle does, so there is no HumanEditSpeed-style
// speed override here.
func applyUpdateScene(msg inputcodec.StdinMsg, md *MoveDispatch, tr *T.Trace, speedSinks []chan float64) {
	if md == nil {
		return
	}
	switch msg.Attr {
	case "selected":
		SelectScene(&md.Scenes, int(msg.Num))
	case "latticePoints":
		points := int32(msg.Num)
		if points < 4 || points > 64 || points%4 != 0 {
			return
		}
		md.ui.latticePoints = points
		md.persist.lattice.schedule(points)
		md.BroadcastLatticePoints(points)
	case "create":
		// The palette's drop. Num is the kind id, X/Y the drop's NDC — see
		// scene_structure.go for how that becomes a place, and why this ends the run rather
		// than building anything live.
		md.CreateNode(uint8(msg.Num), msg.X, msg.Y, tr)
	case "delete":
		// The delete key. Num is the target's buffer ROW.
		md.DeleteNode(msg.Num, tr)
	}
}

func applyUpdateOverlays(msg inputcodec.StdinMsg, md *MoveDispatch, tr *T.Trace, speedSinks []chan float64) {
	if md == nil {
		return
	}
	if h, ok := overlayAttrHandlers[msg.Attr]; ok {
		h(msg, md, tr)
	}
	// Persist ON CHANGE (mirrors camera): schedule a debounced write of the new
	// overlay snapshot so toggles survive a reload without an explicit save. No-op until
	// EnableEditPersist arms the writer (nil-receiver / empty-treeRoot guard in schedule).
	// Runs regardless of which (or whether an) attr matched, matching the original
	// switch's post-inner-switch placement.
	md.persist.overlays.schedule(md.ui.ov)
}
