package Dispatch

import (
	"context"
	"fmt"
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder"
	NodeBuf "github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"
	"github.com/dtauraso/beadnetwork/Categories/Scene/View"
)

func whole(v float64) bool { return v == math.Trunc(v) && !math.IsInf(v, 0) }

func runSheetStructure(ctx context.Context, md *MoveDispatch, a Sheet.Act, sign int, typed *float64) bool {
	m := md.UI.Builder
	switch a.Op {
	case Builder.OpTickMs:
		setTickMs(ctx, md, sign, typed)
	case Builder.OpFrame:
		if typed != nil && a.K == Builder.FrameS {
			setSceneS(ctx, md, *typed)
		} else if typed != nil {
			setFrame(md, a.K, *typed)
		}
	case Builder.OpNodeIdx:
		at := m.Col(a.Node).At
		v := float64(sign) + map[string]float64{Links.KR: at.R, Links.KPhi: at.Phi, Links.KTheta: at.Theta}[a.K]
		if typed != nil {
			v = *typed
		}
		md.sendLink(ctx, a.Node, Links.SetIdx{K: a.K, V: v})
	case Builder.OpNodeRef:
		ref := m.Col(a.Node).Ref
		v := sign + map[string]int{Links.KPhi: ref.Phi, Links.KTheta: ref.Theta}[a.K]
		if typed != nil {
			if !whole(*typed) {
				md.UI.Builder.Refused = fmt.Sprintf("ref %s %v: not a whole number, so it stays at %d", a.Node, *typed, v-sign)
				break
			}
			v = int(*typed)
		}
		md.sendLink(ctx, a.Node, Links.SetRef{K: a.K, V: v})
	case Builder.OpConnect:
		g, ok := m.Grids[a.Node]
		r := g.ShownCol()
		if !ok {
			md.UI.Builder.Refused = fmt.Sprintf("%s→%s: %s has not reported its place yet, so there is nothing to measure the link from", a.Node, a.To, a.Node)
			return true
		}
		md.sendLink(ctx, a.To, Links.Connect{From: a.Node, At: r.At})
	case Builder.OpDisconnect:
		removeLink(md, a.Node, a.To)
	case Builder.OpNewFold:
		md.UI.Builder.NewOpen = !md.UI.Builder.NewOpen
		md.UI.Builder.NetworkOpen = false
	case Builder.OpNewCenter:
		md.UI.Builder.New.Center = a.Node
	case Builder.OpNewAt:
		md.UI.Builder.NewEdit(a, sign, typed)
	case Builder.OpNewAdd:
		addLinkNodeAt(md, m.NewPlace())
	case Builder.OpRemoveNode:
		removeLinkNode(md, a.Node)
	default:
		return false
	}
	return true
}

func addLinkNodeAt(md *MoveDispatch, at Links.Place) {
	m := &md.UI.Builder
	root := md.Scenes.NodesRoot
	if root == "" || md.Scenes.Reload == nil || len(m.Nodes) == 0 {
		return
	}
	kind := md.MR.NodeGeoms()[m.Nodes[0]].Kind()
	id := newNodeID(root)
	if err := NodeBuf.WriteNewPlaceNode(root, id, kind, at); err != nil {
		m.Refused = fmt.Sprintf("could not write node %s: %v", id, err)
		return
	}
	if err := writeCounts(md.Scenes.TreeRoot, largestNodeID(root), countEdgeFiles(root)); err != nil {
		m.Refused = fmt.Sprintf("could not update the counts: %v", err)
		return
	}
	md.Scenes.Reload()
}

func topStepOp(md *MoveDispatch, op string, k int) {
	m := &md.UI.Builder
	scene := md.Scenes.TreeRoot
	if scene == "" || md.Scenes.Reload == nil {
		m.Refused = "this scene cannot change its top steps: it was opened with no tree root to reload from"
		return
	}
	tops := Scenes.TopSteps(scene)
	open, j := m.OpenTop, k+map[string]int{Builder.OpTreeEarlier: -1, Builder.OpTreeLater: 1}[op]
	switch {
	case j < 1 || j > len(tops):
		m.Refused = fmt.Sprintf("step %d is already the %s step, so it stays", k, map[bool]string{true: "first", false: "last"}[j < 1])
		return
	case op == Builder.OpTreeDrop && len(tops) < 2:
		m.Refused = fmt.Sprintf("step %d is the only step, so it stays", k)
		return
	}
	if err := m.WriteSteps(); err != nil {
		View.LogPersistErr("sheet_top_step", m.Root, err)
	}
	var err error
	switch op {
	case Builder.OpTreeAdd, Builder.OpTreeCopy:
		var to int
		to, err = Scenes.InsertTopStep(scene, k, map[bool]string{true: Builder.DirSteps}[op == Builder.OpTreeAdd])
		if err == nil && open >= to {
			open++
		}
	case Builder.OpTreeDrop:
		err = Scenes.RemoveTopStep(scene, k)
		if k == open {
			openTop(md, scene, []int{max(1, open-1)})
			return
		}
		if k < open {
			open--
		}
	case Builder.OpTreeEarlier, Builder.OpTreeLater:
		err = Scenes.SwapTopSteps(scene, k, j)
		if open == k {
			open = j
		} else if open == j {
			open = k
		}
	}
	if err != nil {
		m.Refused = fmt.Sprintf("could not change step %d: %v", k, err)
		return
	}
	if open == m.OpenTop {
		m.ReadTree()
		return
	}
	p := m.OpenPath()
	p[0] = open
	openTop(md, scene, p)
}
