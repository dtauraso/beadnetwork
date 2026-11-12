package Dispatch

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder"
	NodeBuf "github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	SceneB "github.com/dtauraso/beadnetwork/Categories/Scene"
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
		at := m.Reports[a.Node].At
		v := float64(sign) + map[string]float64{Links.KR: at.R, Links.KPhi: at.Phi, Links.KTheta: at.Theta}[a.K]
		if typed != nil {
			v = *typed
		}
		md.sendLink(ctx, a.Node, Links.SetIdx{K: a.K, V: v})
	case Builder.OpNodeRef:
		ref := m.Reports[a.Node].Ref
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
		r, ok := m.Reports[a.Node]
		if !ok {
			md.UI.Builder.Refused = fmt.Sprintf("%s→%s: %s has not reported its place yet, so there is nothing to measure the link from", a.Node, a.To, a.Node)
			return true
		}
		md.sendLink(ctx, a.To, Links.Connect{From: a.Node, At: r.At})
	case Builder.OpDisconnect:
		removeLink(md, a.Node, a.To)
	case Builder.OpAddNode:
		addLinkNode(md)
	case Builder.OpRemoveNode:
		removeLinkNode(md, a.Node)
	default:
		return false
	}
	return true
}

func setFrame(md *MoveDispatch, k string, v float64) {
	m := &md.UI.Builder
	f, s := m.Frame, m.S
	ok := v > 0
	switch k {
	case Builder.FrameRho:
		f.Rho = v
	case Builder.FrameCr:
		f.Cr = v
	case Builder.FrameTphi:
		f.Tphi, ok = int(v), ok && whole(v)
	case Builder.FrameTtheta:
		f.Ttheta, ok = int(v), ok && whole(v)
	}
	if !ok {
		m.Refused = fmt.Sprintf("%s %v: must be above 0 (and whole for T_φ and T_θ), so the frame stays as it is", k, v)
		return
	}
	root := md.Scenes.TreeRoot
	if root == "" || md.Scenes.Reload == nil {
		return
	}
	if err := SceneB.WriteFrame(root, f.Tphi, f.Ttheta, f.Cr, f.Rho); err != nil {
		m.Refused = fmt.Sprintf("could not write the frame: %v", err)
		return
	}
	if err := Sheet.WriteScalar(Scenes.OpenSFilePath(root), int32(s)); err != nil {
		m.Refused = fmt.Sprintf("could not write s: %v", err)
		return
	}
	md.Scenes.Reload()
}

func addLinkNode(md *MoveDispatch) { addLinkNodeAt(md, Links.Place{}) }

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

func runSheetSequence(md *MoveDispatch, a Sheet.Act) bool {
	m := &md.UI.Builder
	scene := md.Scenes.TreeRoot
	if scene == "" || md.Scenes.Reload == nil {
		if a.Op == Builder.OpSeqOpen || a.Op == Builder.OpSeqAdd || a.Op == Builder.OpSeqDrop {
			m.Refused = "this scene cannot switch sequences: it was opened with no tree root to reload from"
			return true
		}
		return false
	}
	open := m.OpenSeq
	switch a.Op {
	case Builder.OpSeqOpen:
		if a.I == open {
			if err := m.LevelOp(Sheet.Act{Op: Builder.OpLevelGo, I: m.TopStep()}); err != nil {
				View.LogPersistErr("sheet_step_level", m.Root, err)
			}
			return true
		}
		open = a.I
	case Builder.OpSeqAdd:
		to, err := Scenes.CopySequence(scene, open)
		if err != nil {
			m.Refused = fmt.Sprintf("could not copy sequence %d: %v", open, err)
			return true
		}
		open = to
	case Builder.OpSeqDrop:
		if len(m.Seqs) < 2 {
			return true
		}
		if err := Scenes.RemoveSequence(scene, open); err != nil {
			m.Refused = fmt.Sprintf("could not delete sequence %d: %v", open, err)
			return true
		}
		open = max(1, open-1)
	default:
		return false
	}
	if err := Scenes.WriteOpenSequence(scene, open); err != nil {
		m.Refused = fmt.Sprintf("could not open sequence %d: %v", open, err)
		return true
	}
	nodes := Scenes.NodesRoot(scene)
	if err := writeCounts(scene, largestNodeID(nodes), countEdgeFiles(nodes)); err != nil {
		m.Refused = fmt.Sprintf("could not update the counts: %v", err)
		return true
	}
	md.Scenes.Reload()
	return true
}

func nameSequence(md *MoveDispatch, k int, name string) {
	scene := md.Scenes.TreeRoot
	if scene == "" {
		md.UI.Builder.Refused = "this scene has no sequences folder, so the name was not kept"
		return
	}
	if name = strings.TrimSpace(name); name == "" {
		name = "sequence " + strconv.Itoa(k)
	}
	if err := Scenes.WriteSequenceName(scene, k, name); err != nil {
		View.LogPersistErr("sheet_seq_name", scene, err)
		md.UI.Builder.Refused = fmt.Sprintf("could not rename sequence %d: %v", k, err)
	}
	LoadSequences(md, scene)
}

func LoadSequences(md *MoveDispatch, scene string) {
	m := &md.UI.Builder
	m.Seqs = nil
	for _, k := range Scenes.Sequences(scene) {
		m.Seqs = append(m.Seqs, Builder.Seq{K: k, Name: Scenes.SequenceName(scene, k)})
	}
	m.OpenSeq = Scenes.OpenSequence(scene)
}
