package Dispatch

import (
	"context"
	"fmt"
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	NodeBuf "github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	SceneB "github.com/dtauraso/beadnetwork/Categories/Scene"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"
)

func whole(v float64) bool { return math.Abs(v-math.Round(v)) < 1e-6 }

func runSheetStructure(ctx context.Context, md *MoveDispatch, a Sheet.Act, sign int, typed *float64) bool {
	m := md.UI.Sheet.Model
	switch a.Op {
	case Sheet.OpFrame:
		if typed != nil {
			setFrame(md, a.K, *typed)
		}
	case Sheet.OpNodeS:
		if typed != nil && whole(*typed) && *typed >= 1 {
			md.sendLink(ctx, a.Node, Links.SetS{V: int(math.Round(*typed))})
		} else if typed != nil {
			md.UI.Sheet.Model.Refused = fmt.Sprintf("%s's s %v: not a whole number of at least 1, so it stays at %d", a.Node, *typed, m.ScaleOf(a.Node))
		}
	case Sheet.OpNodeIdx:
		at := m.Reports[a.Node].At
		v := float64(sign) + map[string]float64{Links.KR: at.R, Links.KPhi: at.Phi, Links.KTheta: at.Theta}[a.K]
		if typed != nil {
			v = *typed
		}
		md.sendLink(ctx, a.Node, Links.SetIdx{K: a.K, V: v})
	case Sheet.OpConnect:
		md.sendLink(ctx, a.Node, Links.Connect{To: a.To})
	case Sheet.OpDropLink:
		md.sendLink(ctx, a.Node, Links.Drop{To: a.To})
	case Sheet.OpAddNode:
		addLinkNode(md)
	case Sheet.OpRemoveNode:
		removeLinkNode(md, a.Node)
	default:
		return false
	}
	return true
}

func setFrame(md *MoveDispatch, k string, v float64) {
	m := &md.UI.Sheet.Model
	f, s := m.Frame, m.S
	ok := v > 0
	switch k {
	case Sheet.FrameRho:
		f.Rho = v
	case Sheet.FrameCr:
		f.Cr = v
	case Sheet.FrameTphi:
		f.Tphi, ok = int(math.Round(v)), ok && whole(v)
	case Sheet.FrameTtheta:
		f.Ttheta, ok = int(math.Round(v)), ok && whole(v)
	case Sheet.FrameS:
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

func addLinkNode(md *MoveDispatch) {
	m := &md.UI.Sheet.Model
	root := md.Scenes.TreeRoot
	if root == "" || md.Scenes.Reload == nil || len(m.Nodes) == 0 {
		return
	}
	kind := md.MR.NodeGeoms()[m.Nodes[0]].Kind()
	id := newNodeID(root)
	if err := NodeBuf.WriteNewPlaceNode(root, id, kind, Links.Place{}, m.S); err != nil {
		m.Refused = fmt.Sprintf("could not write node %s: %v", id, err)
		return
	}
	if err := writeCounts(root, largestNodeID(root), countEdgeFiles(root)); err != nil {
		m.Refused = fmt.Sprintf("could not update the counts: %v", err)
		return
	}
	md.Scenes.Reload()
}

func removeLinkNode(md *MoveDispatch, id string) {
	m := &md.UI.Sheet.Model
	root := md.Scenes.TreeRoot
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
	if err := writeCounts(root, largestNodeID(root), countEdgeFiles(root)); err != nil {
		m.Refused = fmt.Sprintf("could not update the counts: %v", err)
		return
	}
	md.Scenes.Reload()
}
