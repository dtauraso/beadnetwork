package scenestructure

import (
	"fmt"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/countspersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/edgefile"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/camera"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/moverreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/nodefiles"
	"github.com/dtauraso/wirefold/nodes/Wiring/rowtables"
	"github.com/dtauraso/wirefold/nodes/Wiring/sceneswitch"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

func CreateNode(scenes *sceneswitch.SceneSwitch, ui *viewstate.UIState, mr *moverreg.MoverRegistry, kindID uint8, ndcX, ndcY float64, tr *T.Trace) {
	if scenes == nil || scenes.TreeRoot == "" || scenes.Quit == nil {
		return
	}

	if !ui.SceneEditable {
		ui.RefuseStructuralEdit("this scene does not take structural edits")
		ui.EmitViewFrame(nil)
		return
	}
	kind, ok := loadspec.KindForID(kindID)
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

	src, okNear := mr.NearestNodeTo(drop)
	target := loadspec.NewNodeID(scenes.TreeRoot)
	var srcPort, targetPort string
	if okNear {
		var why string
		var canLink bool
		if srcPort, targetPort, why, canLink = mr.LinkRefusal(src, kind); !canLink {
			ui.RefuseStructuralEdit(why)
			ui.EmitViewFrame(nil)
			return
		}
	}

	c := ui.SceneSphere.Center
	off := drop.Sub(c)
	d := camera.WorldDirToAngles(off)
	if err := nodefiles.WriteNewNodeFiles(scenes.TreeRoot, target, kind, off.Length(), d.Theta, d.Phi); err != nil {
		ui.RefuseStructuralEdit(fmt.Sprintf("could not write node %s: %v", target, err))
		ui.EmitViewFrame(nil)
		return
	}
	edges := loadspec.CountEdgeFiles(scenes.TreeRoot)
	if okNear {
		if err := edgefile.WriteEdgeFile(scenes.TreeRoot, src, srcPort, target, targetPort); err != nil {
			ui.RefuseStructuralEdit(fmt.Sprintf("could not write edge %s->%s: %v", src, target, err))
			ui.EmitViewFrame(nil)
			return
		}
		edges++
	}

	if err := countspersist.WriteCounts(scenes.TreeRoot, loadspec.LargestNodeID(scenes.TreeRoot), edges); err != nil {
		ui.RefuseStructuralEdit(fmt.Sprintf("could not update counts.json: %v", err))
		ui.EmitViewFrame(nil)
		return
	}
	scenes.Quit()
}

func DeleteNode(scenes *sceneswitch.SceneSwitch, ui *viewstate.UIState, rt *rowtables.RowTables, row int, tr *T.Trace) {
	if scenes == nil || scenes.TreeRoot == "" || scenes.Quit == nil {
		return
	}

	if !ui.SceneEditable {
		ui.RefuseStructuralEdit("this scene does not take structural edits")
		ui.EmitViewFrame(nil)
		return
	}
	id, ok := rt.LookupNodeRow(row)
	if !ok {
		ui.RefuseStructuralEdit(fmt.Sprintf("no node on row %d", row))
		ui.EmitViewFrame(nil)
		return
	}
	root := scenes.TreeRoot
	if err := nodefiles.RemoveNodeDir(root, id); err != nil {
		ui.RefuseStructuralEdit(fmt.Sprintf("could not remove node %s: %v", id, err))
		ui.EmitViewFrame(nil)
		return
	}
	if err := edgefile.RemoveEdgesTo(root, id, loadspec.NodeIDStringsInTree(root)); err != nil {
		ui.RefuseStructuralEdit(fmt.Sprintf("could not remove edges into %s: %v", id, err))
		ui.EmitViewFrame(nil)
		return
	}

	if err := countspersist.WriteCounts(root, loadspec.LargestNodeID(root), loadspec.CountEdgeFiles(root)); err != nil {
		ui.RefuseStructuralEdit(fmt.Sprintf("could not update counts.json: %v", err))
		ui.EmitViewFrame(nil)
		return
	}
	scenes.Quit()
}
