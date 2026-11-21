package Gesture

import (
	"github.com/dtauraso/beadnetwork/Categories/Node/Messaging"
	NodeDrag "github.com/dtauraso/beadnetwork/Categories/Node/Drag"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Drag"
	"github.com/dtauraso/beadnetwork/Categories/Scene/View"
)

func Grab(g *NodeDrag.Gesture, node string, center NodeDrag.Vec3) {
	g.Node = node
	g.StartCenter = center
}

func CommitDragStart(ui *View.UIState, sendMoveFn func(id string, msg Messaging.Msg), g *Drag.GestureState, ev Drag.RawInputMsg) {
	g.NodeDrag.GrabAt(NodeDrag.Vec3(ev.Hit.Point))

	ui.LastDraggedNode = g.NodeDrag.Node

	sendMoveFn(g.NodeDrag.Node, Messaging.Msg{NodeID: g.NodeDrag.Node, Body: Messaging.DragStart{}})
}

func ApplyDragTarget(ui *View.UIState, rootMove func(id string, target Messaging.Vec3) bool, ev Drag.RawInputMsg) bool {
	g := &ui.Gest
	rootMove(g.NodeDrag.Node, Messaging.Vec3(g.NodeDrag.TargetFor(NodeDrag.Vec3(ev.Hit.Point))))
	return true
}

func SetHover(ui *View.UIState, sendMoveFn func(id string, msg Messaging.Msg), node, port string, isInput bool) (changed bool) {
	if ui.HoverIs(node, port, isInput) {
		return false
	}

	ui.SetHoverUI(sendMoveFn, node, port, isInput)
	return true
}
