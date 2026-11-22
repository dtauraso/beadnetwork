package Dispatch

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet/Scalar"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder"
	SceneB "github.com/dtauraso/beadnetwork/Categories/Scene"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"
	Flags "github.com/dtauraso/beadnetwork/Categories/Scene/View/Flags"
)

func applyOverlayAct(md *MoveDispatch, a Sheet.Act) {
	if a.Op == Pills.OpOverlayBoth {
		flags := strings.Split(a.K, ",")
		allOn := !slices.ContainsFunc(flags, func(f string) bool {
			read, ok := Flags.OverlayFlagRead[f]
			return !ok || !read(&md.UI.OV)
		})
		Flags.SetCount(&md.UI.OV, &md.UI, flags, !allOn)
	} else {
		Flags.ToggleFlag(&md.UI.OV, &md.UI, a.K)
	}
	md.UI.PersistOverlays(md.UI.OV)
}

func setFrame(md *MoveDispatch, k string, v float64) {
	applyFrame(md, map[string]float64{k: v})
}

func applyFrame(md *MoveDispatch, edits map[string]float64) {
	m := &md.UI.Builder
	f, s := m.Frame, m.S
	keys := make([]string, 0, len(edits))
	for k := range edits {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := edits[k]
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
	}
	root := md.Scenes.TreeRoot
	if root == "" || md.Scenes.Reload == nil {
		return
	}
	if err := SceneB.WriteFrame(root, f.Tphi, f.Ttheta, f.Cr, f.Rho); err != nil {
		m.Refused = fmt.Sprintf("could not write the frame: %v", err)
		return
	}
	if err := Scalar.Write(Scenes.OpenSFilePath(root), int32(s)); err != nil {
		m.Refused = fmt.Sprintf("could not write s: %v", err)
		return
	}
	md.Scenes.Reload()
}

func runFramePop(md *MoveDispatch, a Sheet.Act, typed *float64) bool {
	m := &md.UI.Builder
	switch a.Op {
	case Builder.OpFramePop, Builder.OpStepsTab:
		if a.Op == Builder.OpFramePop && m.FramePop {
			break
		}
		m.FramePop, m.FrameDraft, m.NewOpen = a.Op == Builder.OpFramePop, nil, false
		if err := m.SetNetworkOpen(false); err != nil {
			m.Refused = fmt.Sprintf("could not save that the Network view is closed: %v", err)
		}
	case Builder.OpFrameDraft:
		if typed != nil {
			m.StageFrame(a.K, *typed)
		}
	case Builder.OpFrameApply:
		draft := m.FrameDraft
		m.FramePop, m.FrameDraft = false, nil
		if len(draft) > 0 {
			applyFrame(md, draft)
		}
	default:
		return false
	}
	return true
}
