package Gesture

import (
	"github.com/dtauraso/beadnetwork/Categories/Node"
	NodeDrag "github.com/dtauraso/beadnetwork/Categories/Node/Drag"
	NodeGesture "github.com/dtauraso/beadnetwork/Categories/Node/Gesture"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Drag"
)

var beginByHitKind = map[string]func(d Deps, g *Drag.GestureState, ev Drag.RawInputMsg){
	"handhold": func(d Deps, g *Drag.GestureState, ev Drag.RawInputMsg) {

		g.HandholdDown = true
		nodeGeoms, centerOf := d.MR.NodeGeoms(), d.MR.CenterOf
		g.BeginSphereRotation(d.UI.VP.Viewpoint, d.UI.SceneSphere, func() map[string]Drag.Vec3 {
			return centersForDrag(Node.HeldCenters(nodeGeoms, centerOfForMove(centerOf)))
		}, ev)
	},
	"node": func(d Deps, g *Drag.GestureState, ev Drag.RawInputMsg) {
		if !d.UI.SceneNodesDraggable {

			return
		}
		if node, ok := d.RT.NodeFromHit(ev.Hit); ok {
			if c, ok := d.MR.CenterOf(node); ok {
				NodeGesture.Grab(&g.NodeDrag, node, NodeDrag.Vec3(c))
			}
		}
	},
	"empty": func(d Deps, g *Drag.GestureState, ev Drag.RawInputMsg) {
		g.EmptyDown = true
		nodeGeoms, centerOf := d.MR.NodeGeoms(), d.MR.CenterOf
		g.BeginSphereRotation(d.UI.VP.Viewpoint, d.UI.SceneSphere, func() map[string]Drag.Vec3 {
			return centersForDrag(Node.HeldCenters(nodeGeoms, centerOfForMove(centerOf)))
		}, ev)
	},
}
