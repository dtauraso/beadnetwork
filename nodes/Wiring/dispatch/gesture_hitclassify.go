package dispatch

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/gesturefsm"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/layoutquant"
)

// gesture_hitclassify.go — POINTER-DOWN HIT CLASSIFICATION: what a raycast hit's KIND means
// for the gesture about to start (hitClassifiers, gestPointerDown's dispatch table). This
// is a different concern from rowtables.RowTables' row-index → topology-identity lookups
// (NodeFromHit/EdgeFromHit), which this file's classifiers call into. nodeBodyRadius, the
// node-body sizing gestHome (gesture_handlers.go) uses to frame a fit, lives on
// moverreg.MoverRegistry (nodes/Wiring/moverreg) as a pure single-owner forward.

// hitClassifiers is gestPointerDown's dispatch table, keyed by the raycast hit kind. The
// switch it replaces was TERMINAL in gestPointerDown (nothing ran after it), so each case's
// `return` becomes a `return` from the handler func here — behavior-equivalent because
// nothing downstream of the switch depended on falling through to it.
var hitClassifiers = map[string]func(md *MoveDispatch, g *gesturefsm.GestureState, ev inputcodec.RawInputMsg){
	"handhold": func(md *MoveDispatch, g *gesturefsm.GestureState, ev inputcodec.RawInputMsg) {
		// Handhold grab → axis-locked (constrained) orbit. Freeze the sphere rotation frame
		// now (mirrors interaction-handlers.ts: beginSphereRotation on a handhold hit).
		g.HandholdDown = true
		nodeGeoms, centerOf := md.MR.NodeGeoms(), md.MR.CenterOfNode
		g.BeginSphereRotation(md.UI.VP.Viewpoint, func() map[string]vec3 { return layoutquant.HeldCenters(nodeGeoms, centerOf) }, ev)
	},
	"node": func(md *MoveDispatch, g *gesturefsm.GestureState, ev inputcodec.RawInputMsg) {
		if node, ok := md.RT.NodeFromHit(ev.Hit); ok {
			if c, ok := md.MR.CenterOfNode(node); ok {
				g.DragNode = node
				g.DragStartCenter = c
			}
		}
	},
	"empty": func(md *MoveDispatch, g *gesturefsm.GestureState, ev inputcodec.RawInputMsg) {
		g.EmptyDown = true
		nodeGeoms, centerOf := md.MR.NodeGeoms(), md.MR.CenterOfNode
		g.BeginSphereRotation(md.UI.VP.Viewpoint, func() map[string]vec3 { return layoutquant.HeldCenters(nodeGeoms, centerOf) }, ev)
	},
}
