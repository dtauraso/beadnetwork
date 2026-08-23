package Gesture

import (
	"github.com/dtauraso/wirefold/Categories/Scene/Drag"
	"github.com/dtauraso/wirefold/Categories/Node"
	"github.com/dtauraso/wirefold/Categories/Node/nodegeom"
)

var hitClassifiers = map[string]func(d Deps, g *Drag.GestureState, ev Drag.RawInputMsg){
	"handhold": func(d Deps, g *Drag.GestureState, ev Drag.RawInputMsg) {

		g.HandholdDown = true
		nodeGeoms, centerOf := d.MR.NodeGeoms(), d.MR.CenterOf
		g.BeginSphereRotation(d.UI.VP.Viewpoint, func() map[string]Drag.Vec3 {
			return centersForDrag(Node.HeldCenters(nodeGeoms, centerOfForMove(centerOf)))
		}, ev)
	},
	"node": func(d Deps, g *Drag.GestureState, ev Drag.RawInputMsg) {
		if node, ok := d.RT.NodeFromHit(ev.Hit); ok {
			if c, ok := d.MR.CenterOf(node); ok {
				g.NodeDrag.Node = node
				g.NodeDrag.StartCenter = nodegeom.Vec3(c)
			}
		}
	},
	"empty": func(d Deps, g *Drag.GestureState, ev Drag.RawInputMsg) {
		g.EmptyDown = true
		nodeGeoms, centerOf := d.MR.NodeGeoms(), d.MR.CenterOf
		g.BeginSphereRotation(d.UI.VP.Viewpoint, func() map[string]Drag.Vec3 {
			return centersForDrag(Node.HeldCenters(nodeGeoms, centerOfForMove(centerOf)))
		}, ev)
	},
}
