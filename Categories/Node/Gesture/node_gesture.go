package Gesture

import (
	"github.com/dtauraso/beadnetwork/Categories/Node"
	NodeDrag "github.com/dtauraso/beadnetwork/Categories/Node/Drag"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Drag"
	"github.com/dtauraso/beadnetwork/Categories/Scene/View"
)

func Grab(g *NodeDrag.Gesture, node string, center NodeDrag.Vec3) {
	g.Node = node
	g.StartCenter = center
}

func CommitDragStart(ui *View.UIState, sendMoveFn func(id string, msg Node.Msg), g *Drag.GestureState, ev Drag.RawInputMsg) {
	g.NodeDrag.GrabAt(NodeDrag.Vec3(ev.Hit.Point))

	ui.LastDraggedNode = g.NodeDrag.Node

	sendMoveFn(g.NodeDrag.Node, Node.Msg{NodeID: g.NodeDrag.Node, Body: Node.DragStart{}})
}

func ApplyDragTarget(ui *View.UIState, rootMove func(id string, target Node.Vec3) bool, ev Drag.RawInputMsg) bool {
	g := &ui.Gest
	rootMove(g.NodeDrag.Node, Node.Vec3(g.NodeDrag.TargetFor(NodeDrag.Vec3(ev.Hit.Point))))
	return true
}

func SetHover(ui *View.UIState, sendMoveFn func(id string, msg Node.Msg), node, port string, isInput bool) (changed bool) {
	if ui.HoverIs(node, port, isInput) {
		return false
	}

	ui.SetHoverUI(sendMoveFn, node, port, isInput)
	return true
}
