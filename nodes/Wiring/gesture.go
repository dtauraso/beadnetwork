package Wiring

import (
	"math"

	T "github.com/dtauraso/wirefold/Trace"
)

// gesture.go — the GESTURE STATE MACHINE. It consumes RAW pointer/wheel input (forwarded
// fire-and-forget from TS behind USE_RAW_INPUT) plus the stateless raycast hit, owns the
// in-progress gesture bookkeeping (origin, button, phase, frozen rotation frame), and
// decides what the raw input MEANS — orbit / zoom / pan / drag / wire. This is the one place
// gesture state lives (the spec's "gesture state machine lives in Go, in one place"); TS
// holds none of it. The leaf ACTIONS invoked by the phase handlers below (orbit/drag/hover/
// select) live in gesture_actions.go, and hit-resolution helpers (row index → topology
// identity) live in gesture_hit.go — this file keeps only the FSM types, state, dispatch,
// and the phase handlers that own the transitions.
//
// The camera OUTCOMES are produced through the already-tested polar viewpoint ops
// (OrbitViewpoint / ZoomViewpoint / PanViewpoint → spherical.go), fed by the renderer-edge
// camera math in gesture_camera.go (ported formula-for-formula from the TS handlers). This
// file adds no new orbit/rotation math — it only sequences gestures and calls the ported
// helpers.
//
// States:
//   idle      — nothing in progress.
//   pending   — pointer is down; not yet past the move slop. Resolves to a drag/rotate on
//               the first move past MOVE_SLOP_PX, or to a click/wire on pointer-up.
//   rotating  — empty-space great-circle orbit about a frozen region-focus pivot.
//   dragging  — node body drag (world target on a camera-facing plane → RootMove).
//   wiring    — an unconnected port is being dragged toward another port to wire an edge.
//   portMove  — a CONNECTED port is being dragged along its node's ring (ring-anchor snap).
//   handhold  — a handhold grab-sphere is dragged for axis-locked (constrained) orbit.
//
// Phase 7 closed the interaction gaps: click-select is Go-owned (md.ui.sel.selected +
// KindSelect trace → buffer Selected column); handhold-constrained orbit and
// connected-port ring-move are ported here formula-faithfully from
// interaction-handlers.ts. Wire-drop no longer creates an edge — the create/delete edit
// ops were removed end-to-end (no live sender ever emitted them; the only trigger was
// this drop path, which unconditionally tore down live in-flight beads via
// PacedWire.Restore()). A wiring drag now simply resets on pointer-up.

type gesturePhase int

const (
	gestIdle gesturePhase = iota
	gestPending
	gestRotating
	gestDragging
	gestWiring
	gestPortMove
	gestHandhold
)

// gestureState is the FSM's owned bookkeeping. Zero value = idle.
type gestureState struct {
	phase gesturePhase

	// pointer-down snapshot + running previous position (client pixels)
	downX, downY float64
	prevX, prevY float64
	button       int

	// smoothX/smoothY are the AVERAGING ("fat") cursor driving rotation: an exponential
	// moving average of the raw pointer position, so the rotation follows a continuously
	// blurred cursor (never holds, never freezes — just lags-free-smooths jitter). Seeded to
	// the raw position when a rotation drag (gestRotating/gestHandhold) begins.
	smoothX, smoothY float64
	// secondary is true when the pointer-down was a SECONDARY (button 2) press — a
	// two-finger trackpad tap. Such a press is always a tap-select and NEVER converts to a
	// drag/rotate, so it stays `gestPending` through any finger drift and resolves to a
	// select on pointer-up.
	secondary bool

	// empty-space rotation gate + the entity grabbed at pointer-down
	emptyDown bool

	// node-drag target
	dragNode        string
	dragStartCenter vec3

	// wiring source port (unconnected port grabbed at pointer-down)
	wireNode  string
	wirePort  string
	wireInput bool

	// connected-port ring-move (portMove): the grabbed port + its node's center at grab.
	portMoveNode   string
	portMovePort   string
	portMoveInput  bool
	portMoveCenter vec3

	// handhold-constrained orbit gate (set at pointer-down on a handhold hit).
	handholdDown bool

	// rotation frame, FROZEN at gesture start (mirrors beginSphereRotation): the pivot,
	// its screen-pixel center, and pixels-per-radian for screenToPolar.
	rotPivot     vec3
	rotCx, rotCy float64
	rotPxPerRad  float64

	// per-gesture render params captured from the raw events
	fov  float64
	rect gestureRect
}

type gestureRect struct{ left, top, width, height float64 }

func (r gestureRect) aspect() float64 {
	if r.height == 0 {
		return 1
	}
	return r.width / r.height
}

// HandleRawInput is the FSM entry point: one raw pointer/wheel event → gesture state update
// and (possibly) a camera or topology change. Called by the stdin reader for a
// type=="raw-input" message. slotReg resolves an edge's destination slot; tr emits camera
// events + breadcrumbs. Fire-and-forget: nothing here triggers delivery.
func (md *MoveDispatch) HandleRawInput(ev rawInputMsg, slotReg SlotRegistry, tr *T.Trace) {
	g := &md.ui.gest
	g.fov = ev.Fov
	g.rect = gestureRect{left: ev.RectLeft, top: ev.RectTop, width: ev.RectWidth, height: ev.RectHeight}
	if h := rawInputHandlers[ev.Kind]; h != nil {
		h(md, ev, slotReg, tr)
	}
}

// rawInputHandlers is the flat dispatch table for HandleRawInput: raw-input kind →
// handler. An unknown kind is a no-op, matching the switch's absent default.
var rawInputHandlers = map[string]func(md *MoveDispatch, ev rawInputMsg, slotReg SlotRegistry, tr *T.Trace){
	"pointerdown": func(md *MoveDispatch, ev rawInputMsg, slotReg SlotRegistry, tr *T.Trace) {
		md.gestPointerDown(ev, tr)
	},
	"pointermove": func(md *MoveDispatch, ev rawInputMsg, slotReg SlotRegistry, tr *T.Trace) {
		md.updateHover(ev, tr)
		md.gestPointerMove(ev, tr)
	},
	"pointerup": func(md *MoveDispatch, ev rawInputMsg, slotReg SlotRegistry, tr *T.Trace) {
		md.gestPointerUp(ev, slotReg, tr)
	},
	"wheel": func(md *MoveDispatch, ev rawInputMsg, slotReg SlotRegistry, tr *T.Trace) {
		md.gestWheel(ev, tr)
	},
	"home": func(md *MoveDispatch, ev rawInputMsg, slotReg SlotRegistry, tr *T.Trace) {
		md.gestHome(ev, tr)
	},
}

// gestHome handles a "home" (fit-to-content) command: Go frames ALL nodes from its OWN held
// geometry with the SAME fit math the TS HomeButton used (homeFitPose), then installs the
// result via SetViewpoint + EmitViewpoint — the exact path a gesture uses. The FSM's own
// viewpoint IS the framed pose; EmitViewpoint streams it out on this goroutine's own
// per-owner VIEW frame (the buffer VIEW stream) and it persists on the polar save path. TS
// sent no pose, only render context (fov + aspect). Because the FSM's own viewpoint now IS
// the framed pose, the next orbit/pan/zoom builds on it (no snap-back). Does nothing when
// there are no nodes, mirroring HomeButton's early return.
func (md *MoveDispatch) gestHome(ev rawInputMsg, tr *T.Trace) {
	centers := md.heldCenters()
	radius := make(map[string]float64, len(centers))
	for id := range centers {
		radius[id] = md.nodeBodyRadius(id)
	}
	pivot, r, pos, up, ok := homeFitPose(centers, radius, ev.Fov, md.ui.gest.rect.aspect())
	if !ok {
		return
	}
	md.SetViewpoint(pivot, r, pos, up)
	md.EmitViewpoint(tr)
}

// nodeBodyRadius is the node's body sphere radius used to size the home fit. It reuses the
// SAME nodeRadius the pre-branch HomeButton framed with (geometry-helpers.ts nodeRadius ←
// getNodeGeometry(id).radius, the streamed radius the buffer also renders), i.e. the shared
// port_geometry.go nodeRadius(kind) = min(width,height)/CurveParamNodeRadiusDivisor with the
// (110,60) default for an unknown kind. Framing an unknown-kind node as a zero-size POINT
// (the earlier behavior) tightened the fit vs the pre-branch, which framed it at radius 15.
func (md *MoveDispatch) nodeBodyRadius(id string) float64 {
	return nodeRadius(md.NodeKind(id))
}

// pixelToNDC mirrors geometry-helpers.ts pixelToNDC.
func (g *gestureState) pixelToNDC(x, y float64) (nx, ny float64) {
	nx = ((x-g.rect.left)/g.rect.width)*2 - 1
	ny = -((y-g.rect.top)/g.rect.height)*2 + 1
	return nx, ny
}

func (md *MoveDispatch) gestPointerDown(ev rawInputMsg, tr *T.Trace) {
	g := &md.ui.gest
	g.downX, g.downY = ev.X, ev.Y
	g.prevX, g.prevY = ev.X, ev.Y
	g.button = ev.Button
	g.secondary = ev.Button == 2 // two-finger trackpad tap → always a tap-select
	g.phase = gestPending
	g.emptyDown = false
	g.dragNode = ""
	g.wireNode = ""
	g.portMoveNode = ""
	g.handholdDown = false

	if h, ok := hitClassifiers[ev.Hit.Kind]; ok {
		h(md, g, ev)
	}
}

// hitClassifiers is gestPointerDown's dispatch table, keyed by the raycast hit kind. The
// switch it replaces was TERMINAL in gestPointerDown (nothing ran after it), so each case's
// `return` becomes a `return` from the handler func here — behavior-equivalent because
// nothing downstream of the switch depended on falling through to it.
var hitClassifiers = map[string]func(md *MoveDispatch, g *gestureState, ev rawInputMsg){
	"port": func(md *MoveDispatch, g *gestureState, ev rawInputMsg) {
		node, port, isInput, ok := md.portFromHit(ev.Hit)
		if !ok {
			return
		}
		if md.portConnected(node, port, isInput) {
			// Connected port → ring-move along the node's ring. Freeze the node center
			// (the ring plane is z = center.z) at grab, mirroring portMoveRef.nodeCenter.
			// (A plain click without crossing the drag slop still resolves via the
			// gestPending fallthrough on pointer-up, so this doesn't block select-mode
			// `port ∈ torus` authoring — only an actual drag reaches gestPortMove.)
			if c, ok := md.centerOfNode(node); ok {
				g.portMoveNode, g.portMovePort, g.portMoveInput = node, port, isInput
				g.portMoveCenter = c
			}
			return
		}
		g.wireNode, g.wirePort, g.wireInput = node, port, isInput
	},
	"handhold": func(md *MoveDispatch, g *gestureState, ev rawInputMsg) {
		// Handhold grab → axis-locked (constrained) orbit. Freeze the sphere rotation frame
		// now (mirrors interaction-handlers.ts: beginSphereRotation on a handhold hit).
		g.handholdDown = true
		md.beginSphereRotation(ev)
	},
	"node": func(md *MoveDispatch, g *gestureState, ev rawInputMsg) {
		if node, ok := md.nodeFromHit(ev.Hit); ok {
			if c, ok := md.centerOfNode(node); ok {
				g.dragNode = node
				g.dragStartCenter = c
			}
		}
	},
	"empty": func(md *MoveDispatch, g *gestureState, ev rawInputMsg) {
		g.emptyDown = true
		md.beginSphereRotation(ev)
	},
}

func (md *MoveDispatch) gestPointerMove(ev rawInputMsg, tr *T.Trace) {
	g := &md.ui.gest
	if g.phase == gestIdle {
		return
	}
	dx := ev.X - g.downX
	dy := ev.Y - g.downY
	dist := math.Hypot(dx, dy)

	// A secondary (two-finger) press never becomes a drag/rotate — it is a tap-select, so
	// it stays gestPending through any finger drift and resolves on pointer-up.
	if g.phase == gestPending && dist > gestureMoveSlopPx && !g.secondary {
		for _, edge := range commitEdges {
			if edge.guard(g) {
				edge.action(md, g, ev, tr)
				g.phase = edge.to
				break
			}
		}
	}

	if apply, ok := applyAction[g.phase]; ok {
		apply(md, g, ev, tr)
	}
}

func (md *MoveDispatch) gestPointerUp(ev rawInputMsg, slotReg SlotRegistry, tr *T.Trace) {
	g := &md.ui.gest
	switch {
	case g.phase == gestPortMove:
		md.applyPortMove(ev) // final ring-anchor flush
	case g.phase == gestDragging:
		md.applyNodeDragTarget(ev) // final target flush
	case g.phase == gestHandhold, g.phase == gestRotating:
		// Rotation completed (free or handhold-constrained): nothing to flush.
	case g.phase == gestPending:
		// Click → Go-owned selection. A node hit selects it; empty space clears the
		// selection. md.ui.sel.selected is the authoritative selection; Select() emits it so the
		// buffer snapshot marks the node's Selected column.
		md.applySelect(ev, tr)
	}
	wasDragging := g.phase == gestDragging
	g.reset(&md.ui.vp.viewpoint)
	if wasDragging {
		// The drag just ended: g.dragNode is now "" (cleared by reset above), so the
		// Overlay block's DragNodeRow column must go back to -1 promptly rather than
		// waiting for the next unrelated view-frame emit. Mirrors commitDragStart's own
		// emitViewFrame call at drag START.
		md.emitViewFrame(nil)
	}
}

// reset clears the gesture FSM back to idle at the end of every gesture (pointer-up).
// It also clears vp.lockedAxis (the handhold-constrained-orbit rotation axis frozen at
// gesture start — see viewpoint.lockedAxis's doc comment) so that field's own "nil
// between gestures" doc is actually true: lockedAxis is gesture-scoped state, exactly
// like dragNode/wireNode/portMoveNode above, it just happens to live on viewpoint
// instead of gestureState (frozen once per handhold gesture in orbit's lazy-init path).
// Today it is always overwritten before use anyway (every new gesture reseeds it via
// SetViewpoint/seedOrbitPivot before orbit ever reads it), so this had no live bug —
// but reset() is the obvious single home for "gesture-scoped state ends here", so it
// belongs here rather than living only as an unenforced comment.
func (g *gestureState) reset(vp *viewpoint) {
	g.phase = gestIdle
	g.emptyDown = false
	g.dragNode = ""
	g.wireNode = ""
	g.portMoveNode = ""
	g.handholdDown = false
	g.secondary = false
	vp.lockedAxis = nil
}

// gestWheel mirrors interaction-handlers.ts handleWheelNative: ctrl+wheel = zoom-to-cursor
// dolly (expressed as a PAN in the polar model — a pivot translation, not a radius change),
// plain wheel = screen-space pan. Both first seed the viewpoint to region-focus, then pan.
func (md *MoveDispatch) gestWheel(ev rawInputMsg, tr *T.Trace) {
	vp := md.ui.vp.viewpoint
	eye := eyeOf(vp)
	pivot := regionFocus(vp, md.heldCenters())

	if ev.Ctrl {
		// Zoom-to-cursor: move the camera TOWARD the node under the cursor along the cursor→node
		// line, KEEPING the look direction — so that node stays fixed under the mouse. It does NOT
		// re-aim: re-aiming (snapping the camera to look straight at the node) is what recentered
		// the view and threw the cursor off. PanViewpoint translates the whole camera (pivot+eye
		// ride together); pos/up are unchanged, so the node keeps projecting to the same pixel.
		// The cursor→node pick is a screen-space selection at the input boundary (projectNDC).
		mouseNdcX, mouseNdcY := md.ui.gest.pixelToNDC(ev.X, ev.Y)
		basis := basisFromViewpoint(vp.pos, vp.up)
		aspect := md.ui.gest.rect.aspect()
		target := pivot
		best := math.Inf(1)
		for _, c := range md.heldCenters() {
			nx, ny, inFront := projectNDC(c, eye, basis, ev.Fov, aspect)
			if !inFront {
				continue
			}
			if d := math.Hypot(nx-mouseNdcX, ny-mouseNdcY); d < best {
				best = d
				target = c
			}
		}
		toTarget := target.Sub(eye)
		distP := toTarget.Length()
		rayDir := anglesToWorldOffset(1, vp.pos.Theta, vp.pos.Phi).Scale(-1) // forward, if AT the node
		if distP > 1e-9 {
			rayDir = toTarget.Scale(1 / distP)
		}
		// Move the eye ALONG the cursor→node ray. amt>0 = toward the node (zoom in). The step is a
		// fraction of the remaining distance (fast approach when far), FLOORED at a scene-scaled
		// minimum so you can push THROUGH the node instead of asymptotically creeping to it — a
		// pilot camera flies past nodes. No stop-short clamp.
		amt := 1 - math.Pow(gestureZoomBase, ev.DeltaY)
		step := distP * amt
		if minStep := vp.r * (gestureZoomBase - 1); math.Abs(step) < minStep {
			step = math.Copysign(minStep, amt)
		}
		md.PanViewpoint(rayDir.Scale(step), tr)
		return
	}

	// Plain wheel = LATERAL pan = STRAFE THE CAMERA (free-camera model): the camera body slides
	// sideways through the fixed scene. Pan SPEED is scaled by the camera's OWN focal distance
	// (vp.r), NOT by eye-to-nearest-content — the latter collapses when zoom dollies the eye up
	// to a node, which is exactly what made pan crawl after zooming in (and coupled pan to zoom).
	// vp.r is a stable scene-scale property (set by home/framing, unchanged by the dolly), so pan
	// stays a usable pilot speed at any zoom. The displacement is built in polar; PanViewpoint
	// translates pivot+eye together with the look direction unchanged. The scene does not move.
	fovRad := ev.Fov * math.Pi / 180
	worldPerPixel := (2 * vp.r * math.Tan(fovRad/2)) / md.ui.gest.rect.height
	disp := panDisplacementPolar(vp.pos, vp.up, ev.DeltaX, ev.DeltaY, worldPerPixel)
	md.PanViewpoint(disp, tr)
}
