package gesture

import (
	"github.com/dtauraso/wirefold/src/Input/gesturefsm"
	"github.com/dtauraso/wirefold/src/Input/inputcodec"
	"github.com/dtauraso/wirefold/src/Node/nodemove"
	"github.com/dtauraso/wirefold/src/spatial"
)

var hitClassifiers = map[string]func(d Deps, g *gesturefsm.GestureState, ev inputcodec.RawInputMsg){
	"handhold": func(d Deps, g *gesturefsm.GestureState, ev inputcodec.RawInputMsg) {

		g.HandholdDown = true
		nodeGeoms, centerOf := d.MR.NodeGeoms(), d.MR.CenterOfNode
		g.BeginSphereRotation(d.UI.VP.Viewpoint, func() map[string]spatial.Vec3 { return nodemove.HeldCenters(nodeGeoms, centerOf) }, ev)
	},
	"node": func(d Deps, g *gesturefsm.GestureState, ev inputcodec.RawInputMsg) {
		if node, ok := d.RT.NodeFromHit(ev.Hit); ok {
			if c, ok := d.MR.CenterOfNode(node); ok {
				g.DragNode = node
				g.DragStartCenter = c
			}
		}
	},
	"empty": func(d Deps, g *gesturefsm.GestureState, ev inputcodec.RawInputMsg) {
		g.EmptyDown = true
		nodeGeoms, centerOf := d.MR.NodeGeoms(), d.MR.CenterOfNode
		g.BeginSphereRotation(d.UI.VP.Viewpoint, func() map[string]spatial.Vec3 { return nodemove.HeldCenters(nodeGeoms, centerOf) }, ev)
	},
}
