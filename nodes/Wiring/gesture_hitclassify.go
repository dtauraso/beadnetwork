package Wiring

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
)

// gesture_hitclassify.go — POINTER-DOWN HIT CLASSIFICATION: what a raycast hit's KIND means
// for the gesture about to start (hitClassifiers, gestPointerDown's dispatch table), plus
// nodeBodyRadius, the node-body sizing gestHome (gesture_handlers.go) uses to frame a fit.
// This is a different concern from gesture_hit.go's row-index → topology-identity lookups
// (nodeFromHit/edgeFromHit), which this file's classifiers call into.

// hitClassifiers is gestPointerDown's dispatch table, keyed by the raycast hit kind. The
// switch it replaces was TERMINAL in gestPointerDown (nothing ran after it), so each case's
// `return` becomes a `return` from the handler func here — behavior-equivalent because
// nothing downstream of the switch depended on falling through to it.
var hitClassifiers = map[string]func(md *MoveDispatch, g *gestureState, ev inputcodec.RawInputMsg){
	"handhold": func(md *MoveDispatch, g *gestureState, ev inputcodec.RawInputMsg) {
		// Handhold grab → axis-locked (constrained) orbit. Freeze the sphere rotation frame
		// now (mirrors interaction-handlers.ts: beginSphereRotation on a handhold hit).
		g.handholdDown = true
		md.beginSphereRotation(ev)
	},
	"node": func(md *MoveDispatch, g *gestureState, ev inputcodec.RawInputMsg) {
		if node, ok := md.nodeFromHit(ev.Hit); ok {
			if c, ok := md.mr.centerOfNode(node); ok {
				g.dragNode = node
				g.dragStartCenter = c
			}
		}
	},
	"empty": func(md *MoveDispatch, g *gestureState, ev inputcodec.RawInputMsg) {
		g.emptyDown = true
		md.beginSphereRotation(ev)
	},
}

// nodeBodyRadius is the node's body sphere radius used to size the home fit. It reuses the
// SAME nodeRadius the pre-branch HomeButton framed with (geometry-helpers.ts nodeRadius ←
// getNodeGeometry(id).radius, the streamed radius the buffer also renders), i.e. the shared
// nodegeom/port_geometry.go nodegeom.NodeRadius(kind) = min(width,height)/CurveParamNodeRadiusDivisor with the
// (110,60) default for an unknown kind. Framing an unknown-kind node as a zero-size POINT
// (the earlier behavior) tightened the fit vs the pre-branch, which framed it at radius 15.
func (md *MoveDispatch) nodeBodyRadius(id string) float64 {
	return nodegeom.NodeRadius(md.NodeKind(id))
}
