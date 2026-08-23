package Dispatch

import (
	"fmt"

	"github.com/dtauraso/wirefold/Categories/Scene"

	"github.com/dtauraso/wirefold/Categories/Node"
	"github.com/dtauraso/wirefold/Categories/Node/Edge/edgefile"
	"github.com/dtauraso/wirefold/Categories/Polar/polar"
	"github.com/dtauraso/wirefold/Categories/Scene/Camera"
	"github.com/dtauraso/wirefold/Categories/Scene/Scenes"
	"github.com/dtauraso/wirefold/Categories/Scene/viewstate"
)

func CreateNode(scenes *Scenes.SceneSwitch, ui *viewstate.UIState, nodeGeoms map[string]*Node.NodeGeometry, nearestTo func(viewstate.Vec3) (string, bool), kindID uint8, ndcX, ndcY float64) {
	if scenes == nil || scenes.TreeRoot == "" || scenes.Quit == nil {
		return
	}

	if !ui.SceneEditable {
		ui.RefuseStructuralEdit("this scene does not take structural edits")
		ui.EmitViewFrame(nil)
		return
	}
	kind, ok := kindForID(kindID)
	if !ok {
		ui.RefuseStructuralEdit(fmt.Sprintf("unknown kind id %d", kindID))
		ui.EmitViewFrame(nil)
		return
	}

	if ui.SceneKinds&(1<<uint(kindID)) == 0 {
		ui.RefuseStructuralEdit(fmt.Sprintf("this scene does not take %s nodes", kind))
		ui.EmitViewFrame(nil)
		return
	}

	drop, okDrop := ui.DropPointFromNDC(ndcX, ndcY)
	if !okDrop {
		ui.RefuseStructuralEdit("could not resolve where the drop landed")
		ui.EmitViewFrame(nil)
		return
	}

	src, okNear := nearestTo(drop)
	target := newNodeID(scenes.TreeRoot)
	var srcHandle, targetPort string
	if okNear {
		srcGeom, srcFound := nodeGeoms[src]
		srcKind := ""
		if srcFound {
			srcKind = srcGeom.Kind()
		}
		link, why, canLink := linkRefusalFor(src, srcKind, srcFound, kind)
		if !canLink {
			ui.RefuseStructuralEdit(why)
			ui.EmitViewFrame(nil)
			return
		}
		var okHandle bool
		if srcHandle, why, okHandle = edgefile.SourceHandleFor(scenes.TreeRoot, src, link.SrcPort, link.Broadcast); !okHandle {
			ui.RefuseStructuralEdit(why)
			ui.EmitViewFrame(nil)
			return
		}
		targetPort = link.TargetPort
	}

	c := ui.SceneSphere.Center
	off := drop.Sub(viewstate.Vec3(c))
	d := Camera.WorldDirToAngles(Camera.Vec3(off))
	sc, err := Scene.LoadSceneConstants(scenes.TreeRoot)
	if err != nil {
		ui.RefuseStructuralEdit(fmt.Sprintf("could not load scene constants: %v", err))
		ui.EmitViewFrame(nil)
		return
	}
	if err := Node.WriteNewNodeFiles(scenes.TreeRoot, target, kind, polar.Polar{R: off.Length(), Phi: d.Phi, Theta: d.Theta}, sc); err != nil {
		ui.RefuseStructuralEdit(fmt.Sprintf("could not write node %s: %v", target, err))
		ui.EmitViewFrame(nil)
		return
	}
	edges := countEdgeFiles(scenes.TreeRoot)
	if okNear {
		if err := edgefile.WriteEdgeFile(scenes.TreeRoot, src, srcHandle, target, targetPort); err != nil {
			ui.RefuseStructuralEdit(fmt.Sprintf("could not write edge %s->%s: %v", src, target, err))
			ui.EmitViewFrame(nil)
			return
		}
		edges++
	}

	if err := writeCounts(scenes.TreeRoot, largestNodeID(scenes.TreeRoot), edges); err != nil {
		ui.RefuseStructuralEdit(fmt.Sprintf("could not update counts.json: %v", err))
		ui.EmitViewFrame(nil)
		return
	}
	scenes.Quit()
}

func DeleteNode(scenes *Scenes.SceneSwitch, ui *viewstate.UIState, id string, row int) {
	if scenes == nil || scenes.TreeRoot == "" || scenes.Quit == nil {
		return
	}

	if !ui.SceneEditable {
		ui.RefuseStructuralEdit("this scene does not take structural edits")
		ui.EmitViewFrame(nil)
		return
	}
	if id == "" {
		ui.RefuseStructuralEdit(fmt.Sprintf("no node on row %d", row))
		ui.EmitViewFrame(nil)
		return
	}
	root := scenes.TreeRoot
	if err := Node.RemoveNodeDir(root, id); err != nil {
		ui.RefuseStructuralEdit(fmt.Sprintf("could not remove node %s: %v", id, err))
		ui.EmitViewFrame(nil)
		return
	}
	if err := edgefile.RemoveEdgesTo(root, id, nodeIDStringsInTree(root)); err != nil {
		ui.RefuseStructuralEdit(fmt.Sprintf("could not remove edges into %s: %v", id, err))
		ui.EmitViewFrame(nil)
		return
	}

	if err := writeCounts(root, largestNodeID(root), countEdgeFiles(root)); err != nil {
		ui.RefuseStructuralEdit(fmt.Sprintf("could not update counts.json: %v", err))
		ui.EmitViewFrame(nil)
		return
	}
	scenes.Quit()
}
