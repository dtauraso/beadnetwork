package Gesture

import (
	"github.com/dtauraso/wirefold/Categories/Node"
	NodeDrag "github.com/dtauraso/wirefold/Categories/Node/Drag"
	"github.com/dtauraso/wirefold/Categories/Node/nodegeom"
	"github.com/dtauraso/wirefold/Categories/Scene/Drag"
	"github.com/dtauraso/wirefold/Categories/Scene/rowtables"
	"github.com/dtauraso/wirefold/Categories/Scene/viewstate"
)

func Grab(g *NodeDrag.Gesture, node string, center nodegeom.Vec3) {
	g.Node = node
	g.StartCenter = center
}

func CommitDragStart(ui *viewstate.UIState, sendMoveFn func(id string, msg Node.Msg), g *Drag.GestureState, ev Drag.RawInputMsg) {
	if hit, ok := ui.DragPlaneHit(ev); ok {
		g.NodeDrag.GrabAt(nodegeom.Vec3(hit))
	}

	ui.LastDraggedNode = g.NodeDrag.Node

	sendMoveFn(g.NodeDrag.Node, Node.Msg{NodeID: g.NodeDrag.Node, Body: Node.DragStart{}})
}

func ApplyDragTarget(ui *viewstate.UIState, rootMove func(id string, target nodegeom.Vec3) bool, ev Drag.RawInputMsg) bool {
	g := &ui.Gest
	hit, ok := ui.DragPlaneHit(ev)
	if !ok {
		return false
	}
	rootMove(g.NodeDrag.Node, g.NodeDrag.TargetFor(nodegeom.Vec3(hit)))
	return true
}

func SetHover(ui *viewstate.UIState, sendMoveFn func(id string, msg Node.Msg), RT *rowtables.RowTables, node, port string, isInput bool) (changed bool) {
	if node == ui.Sel.HoverNode && port == ui.Sel.HoverPort && isInput == ui.Sel.HoverInput {
		return false
	}

	ui.SetHoverUI(sendMoveFn, node, port, isInput)
	return true
}
