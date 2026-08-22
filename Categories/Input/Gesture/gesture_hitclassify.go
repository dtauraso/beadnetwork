package Gesture

import (
	"github.com/dtauraso/wirefold/Categories/Input/Drag"
	"github.com/dtauraso/wirefold/Categories/Node/nodemove"
)

var hitClassifiers = map[string]func(d Deps, g *Drag.GestureState, ev Drag.RawInputMsg){
	"handhold": func(d Deps, g *Drag.GestureState, ev Drag.RawInputMsg) {

		g.HandholdDown = true
		nodeGeoms, centerOf := d.MR.NodeGeoms(), d.MR.CenterOfNode
		g.BeginSphereRotation(d.UI.VP.Viewpoint, func() map[string]Drag.Vec3 {
			return centersForDrag(nodemove.HeldCenters(nodeGeoms, centerOfForMove(centerOf)))
		}, ev)
	},
	"node": func(d Deps, g *Drag.GestureState, ev Drag.RawInputMsg) {
		if node, ok := d.RT.NodeFromHit(ev.Hit); ok {
			if c, ok := d.MR.CenterOfNode(node); ok {
				g.DragNode = node
				g.DragStartCenter = Drag.Vec3(c)
			}
		}
	},
	"empty": func(d Deps, g *Drag.GestureState, ev Drag.RawInputMsg) {
		g.EmptyDown = true
		nodeGeoms, centerOf := d.MR.NodeGeoms(), d.MR.CenterOfNode
		g.BeginSphereRotation(d.UI.VP.Viewpoint, func() map[string]Drag.Vec3 {
			return centersForDrag(nodemove.HeldCenters(nodeGeoms, centerOfForMove(centerOf)))
		}, ev)
	},
}
