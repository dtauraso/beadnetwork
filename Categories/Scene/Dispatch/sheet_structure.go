package Dispatch

import (
	"context"
	"fmt"
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder"
	NodeBuf "github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	SceneB "github.com/dtauraso/beadnetwork/Categories/Scene"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"
)

func whole(v float64) bool { return math.Abs(v-math.Round(v)) < 1e-6 }

func runSheetStructure(ctx context.Context, md *MoveDispatch, a Sheet.Act, sign int, typed *float64) bool {
	m := md.UI.Builder
	switch a.Op {
	case Builder.OpTickMs:
		setTickMs(ctx, md, sign, typed)
	case Builder.OpFrame:
		if typed != nil {
			setFrame(md, a.K, *typed)
		}
	case Builder.OpNodeS:
		if typed != nil && whole(*typed) && *typed >= 1 {
			md.sendLink(ctx, a.Node, Links.SetS{V: int(math.Round(*typed))})
		} else if typed != nil {
			md.UI.Builder.Refused = fmt.Sprintf("%s's s %v: not a whole number of at least 1, so it stays at %d", a.Node, *typed, m.ScaleOf(a.Node))
		}
	case Builder.OpNodeIdx:
		at := m.Reports[a.Node].At
		v := float64(sign) + map[string]float64{Links.KR: at.R, Links.KPhi: at.Phi, Links.KTheta: at.Theta}[a.K]
		if typed != nil {
			v = *typed
		}
		md.sendLink(ctx, a.Node, Links.SetIdx{K: a.K, V: v})
	case Builder.OpConnect:
		r, ok := m.Reports[a.To]
		if !ok {
			md.UI.Builder.Refused = fmt.Sprintf("%s→%s: %s has not reported its place yet, so there is nothing to measure the link to", a.Node, a.To, a.To)
			return true
		}
		md.sendLink(ctx, a.Node, Links.Connect{To: a.To, At: r.At})
	case Builder.OpDropLink:
		md.sendLink(ctx, a.Node, Links.Drop{To: a.To})
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
		f.Tphi, ok = int(math.Round(v)), ok && whole(v)
	case Builder.FrameTtheta:
		f.Ttheta, ok = int(math.Round(v)), ok && whole(v)
	case Builder.FrameS:
		s, ok = int(math.Round(v)), ok && whole(v)
	}
	if !ok {
		m.Refused = fmt.Sprintf("%s %v: must be above 0 (and whole for T_φ, T_θ and s), so the frame stays as it is", k, v)
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
	if err := Sheet.WriteScalar(Scenes.SFilePath(root), int32(s)); err != nil {
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
	if err := NodeBuf.WriteNewPlaceNode(root, id, kind, at, m.S); err != nil {
		m.Refused = fmt.Sprintf("could not write node %s: %v", id, err)
		return
	}
	if err := writeCounts(md.Scenes.TreeRoot, largestNodeID(root), countEdgeFiles(root)); err != nil {
		m.Refused = fmt.Sprintf("could not update the counts: %v", err)
		return
	}
	md.Scenes.Reload()
}

func removeLinkNode(md *MoveDispatch, id string) {
	m := &md.UI.Builder
	root := md.Scenes.NodesRoot
	if root == "" || md.Scenes.Reload == nil {
		return
	}
	for _, other := range m.Nodes {
		if other == id {
			continue
		}
		if err := Links.RemoveLinksTo(NodeBuf.NodeDir(root, other), other, id); err != nil {
			m.Refused = fmt.Sprintf("could not remove %s→%s: %v", other, id, err)
			return
		}
	}
	if err := NodeBuf.RemoveNodeDir(root, id); err != nil {
		m.Refused = fmt.Sprintf("could not remove node %s: %v", id, err)
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
			m.Refused = Scenes.SequenceName(scene, open) + " is already open — click another sequence to open it"
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

func LoadSequences(md *MoveDispatch, scene string) {
	m := &md.UI.Builder
	m.Seqs = nil
	for _, k := range Scenes.Sequences(scene) {
		m.Seqs = append(m.Seqs, Builder.Seq{K: k, Name: Scenes.SequenceName(scene, k)})
	}
	m.OpenSeq = Scenes.OpenSequence(scene)
}
