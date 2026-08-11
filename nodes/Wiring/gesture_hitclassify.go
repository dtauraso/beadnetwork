package Wiring

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
)

// gesture_hitclassify.go — POINTER-DOWN HIT CLASSIFICATION: what a raycast hit's KIND means
// for the gesture about to start (hitClassifiers, gestPointerDown's dispatch table). This
// is a different concern from rowtables.RowTables' row-index → topology-identity lookups
// (NodeFromHit/EdgeFromHit), which this file's classifiers call into. nodeBodyRadius, the
// node-body sizing gestHome (gesture_handlers.go) uses to frame a fit, moved onto
// moverRegistry (mover_registry.go) as a pure single-owner forward.

// hitClassifiers is gestPointerDown's dispatch table, keyed by the raycast hit kind. The
// switch it replaces was TERMINAL in gestPointerDown (nothing ran after it), so each case's
// `return` becomes a `return` from the handler func here — behavior-equivalent because
// nothing downstream of the switch depended on falling through to it.
var hitClassifiers = map[string]func(md *MoveDispatch, g *gestureState, ev inputcodec.RawInputMsg){
	"handhold": func(md *MoveDispatch, g *gestureState, ev inputcodec.RawInputMsg) {
		// Handhold grab → axis-locked (constrained) orbit. Freeze the sphere rotation frame
		// now (mirrors interaction-handlers.ts: beginSphereRotation on a handhold hit).
		g.handholdDown = true
		beginSphereRotation(&md.ui, &md.mr, &md.lq, ev)
	},
	"node": func(md *MoveDispatch, g *gestureState, ev inputcodec.RawInputMsg) {
		if node, ok := md.RT.NodeFromHit(ev.Hit); ok {
			if c, ok := md.mr.centerOfNode(node); ok {
				g.dragNode = node
				g.dragStartCenter = c
			}
		}
	},
	"empty": func(md *MoveDispatch, g *gestureState, ev inputcodec.RawInputMsg) {
		g.emptyDown = true
		beginSphereRotation(&md.ui, &md.mr, &md.lq, ev)
	},
}
