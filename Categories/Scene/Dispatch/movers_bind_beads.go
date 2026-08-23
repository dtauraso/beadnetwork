package Dispatch

import (
	beadanimation "github.com/dtauraso/wirefold/Categories/Node/BeadAnimation"
)

func (m *Movers) Bind(outSink map[string]*beadanimation.Sender, slotReg map[string]*beadanimation.BeadLine, edgeRowFor func(src, dst string) (int32, bool)) {
	for edgeID, e := range m.edges {
		var port *beadanimation.Sender
		if oo, ok := outSink[e.SrcID()+"."+e.SrcHandle()]; ok {
			e.SetOut(oo)
			m.edgeOut[edgeID] = oo
			port = oo
		}
		pw, hasDest := slotReg[e.DstID()+"."+e.DstHandle()]
		if hasDest {
			e.SetDest(pw)
		}

		srcNM, ok := m.nodeGeoms[e.SrcID()]
		if !ok {
			continue
		}

		dstKind := ""
		if dstNM, ok := m.nodeGeoms[e.DstID()]; ok {
			dstKind = dstNM.SelfKind()
		}
		srcNM.BindOutEdgeRun(e.Label(), e.DstID(), dstKind, port, pw)

		if hasDest {
			edgeRow := int32(-1)
			if edgeRowFor != nil {
				if r, ok := edgeRowFor(e.SrcID(), e.DstID()); ok {
					edgeRow = r
				}
			}
			srcNM.Anim().AddBeadLine(pw, edgeRow)
		}
	}
}
