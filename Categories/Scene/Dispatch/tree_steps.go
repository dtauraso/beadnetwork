package Dispatch

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"
	"github.com/dtauraso/beadnetwork/Categories/Scene/View"
)

func runSheetTree(ctx context.Context, md *MoveDispatch, a Sheet.Act) bool {
	m := &md.UI.Builder
	p := Builder.ParsePath(a.Node)
	switch a.Op {
	case Builder.OpTreeFold:
		if err := m.ToggleTreeFold(p); err != nil {
			View.LogPersistErr("sheet_tree_fold", m.Scene, err)
		}
	case Builder.OpTreeGo:
		if m.Crosses(p) {
			crossTo(md, p)
			return true
		}
		if err := m.TreeGo(p); err != nil {
			View.LogPersistErr("sheet_tree_go", m.Root, err)
		}
		md.ShowStep(ctx)
	default:
		return false
	}
	return true
}

func crossTo(md *MoveDispatch, p []int) {
	m := &md.UI.Builder
	scene := md.Scenes.TreeRoot
	if scene == "" || md.Scenes.Reload == nil {
		m.Refused = "this scene cannot open another top step: it was opened with no tree root to reload from"
		return
	}
	if err := m.WriteSteps(); err != nil {
		View.LogPersistErr("sheet_tree_cross", m.Root, err)
	}
	if len(p) == 1 {
		p = append(p, Scenes.OpenStepPath(Scenes.TopStepDir(scene, p[0]))...)
	}
	if err := Scenes.WriteOpenStep(scene, p); err != nil {
		m.Refused = fmt.Sprintf("could not open step %s: %v", Builder.PathKey(p), err)
		return
	}
	nodes := Scenes.NodesRoot(scene)
	if err := writeCounts(scene, largestNodeID(nodes), countEdgeFiles(nodes)); err != nil {
		m.Refused = fmt.Sprintf("could not update the counts: %v", err)
		return
	}
	md.Scenes.Reload()
}

func LoadOpenStep(md *MoveDispatch, scene string) {
	m := &md.UI.Builder
	m.Scene, m.OpenTop = scene, Scenes.OpenTopStep(scene)
	p := Scenes.OpenStepPath(scene)
	if m.OpenTop > 0 {
		if len(p) > 1 && p[0] == m.OpenTop {
			p = p[1:]
		} else {
			p = Scenes.OpenStepPath(m.Root)
		}
	}
	if len(p) == 0 {
		return
	}
	if _, err := os.Stat(filepath.Join(m.Root, Builder.LevelOf(p))); err != nil {
		return
	}
	m.Level, m.ShownStep = Builder.LevelOf(p[:len(p)-1]), p[len(p)-1]
}

func runTreeStepOp(ctx context.Context, md *MoveDispatch, a Sheet.Act) bool {
	if !Builder.IsTreeStepOp(a) {
		return false
	}
	m := &md.UI.Builder
	p := Builder.ParsePath(a.Node)
	switch {
	case len(p) == 0:
	case m.OpenTop > 0 && len(p) == 1 && a.Op != Builder.OpTreeInner:
		topStepOp(md, a.Op, p[0])
	case m.Crosses(p):
		m.Refused = fmt.Sprintf("step %s is inside top step %d; open that one first", Builder.PathKey(p), p[0])
	case a.Op == Builder.OpTreeInner:
		to, err := m.AddInner(p)
		if err == nil {
			err = m.TreeGo(to)
		}
		if err != nil {
			View.LogPersistErr("sheet_tree_inner", m.Root, err)
		}
		md.ShowStep(ctx)
	default:
		err := m.TreeGo(p)
		if err == nil {
			m.Refused, err = m.StepDirOp(Sheet.Act{Op: Builder.DirOpOf(a.Op), I: p[len(p)-1]})
		}
		if err != nil {
			View.LogPersistErr("sheet_tree_step", m.Root, err)
		}
		md.ShowStep(ctx)
	}
	return true
}

var seqOpOf = map[string]string{Builder.OpTreeAdd: Builder.OpSeqAdd, Builder.OpTreeCopy: Builder.OpSeqCopy, Builder.OpTreeDrop: Builder.OpSeqDrop}

func topStepOp(md *MoveDispatch, op string, k int) {
	m := &md.UI.Builder
	if s, ok := seqOpOf[op]; ok {
		runSheetSequence(md, Sheet.Act{Op: s, I: k})
		m.ReadTree()
		return
	}
	scene := md.Scenes.TreeRoot
	j := k + map[string]int{Builder.OpTreeEarlier: -1, Builder.OpTreeLater: 1}[op]
	if j < 1 || j > len(Scenes.TopSteps(scene)) {
		m.Refused = fmt.Sprintf("step %d is already the %s step, so it stays", k, map[bool]string{true: "first", false: "last"}[j < 1])
		return
	}
	if scene == "" || md.Scenes.Reload == nil {
		m.Refused = "this scene cannot reorder its top steps: it was opened with no tree root to reload from"
		return
	}
	open := m.OpenTop
	if open == k || open == j {
		if err := m.WriteSteps(); err != nil {
			View.LogPersistErr("sheet_tree_swap", m.Root, err)
		}
	}
	if err := Scenes.SwapTopSteps(scene, k, j); err != nil {
		m.Refused = fmt.Sprintf("could not move step %d: %v", k, err)
		return
	}
	if open != k && open != j {
		LoadSequences(md, scene)
		m.ReadTree()
		return
	}
	p := m.OpenPath()
	p[0] = map[bool]int{true: j, false: k}[open == k]
	if err := Scenes.WriteOpenStep(scene, p); err != nil {
		m.Refused = fmt.Sprintf("could not keep step %d open: %v", p[0], err)
		return
	}
	md.Scenes.Reload()
}
