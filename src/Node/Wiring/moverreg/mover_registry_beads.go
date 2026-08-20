package moverreg

import (
	"context"
	"sync"

	"github.com/dtauraso/wirefold/src/Node/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/src/Node/wire/outport"
)

func (mr *MoverRegistry) Bind(outSink map[string]*outport.Out, slotReg inputcodec.SlotRegistry, edgeRowFor func(src, dst string) (int32, bool)) {
	for edgeID, e := range mr.edges {
		var port *outport.Out
		if oo, ok := outSink[e.SrcID()+"."+e.SrcHandle()]; ok {
			e.SetOut(oo)
			mr.edgeOut[edgeID] = oo
			port = oo
		}
		pw, hasDest := slotReg[e.DstID()+"."+e.DstHandle()]
		if hasDest {
			e.SetDest(pw)
		}

		srcNM, ok := mr.nodeGeoms[e.SrcID()]
		if !ok {
			continue
		}

		dstKind := ""
		if dstNM, ok := mr.nodeGeoms[e.DstID()]; ok {
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
			srcNM.AddBeadRun(pw, edgeRow)
		}
	}
}

func (mr *MoverRegistry) Start(ctx context.Context) *sync.WaitGroup {
	for _, nm := range mr.nodeGeoms {
		nm.DeriveOutEdgeGeometryOnce()
	}

	wg := new(sync.WaitGroup)
	wg.Add(len(mr.nodeGeoms))
	for _, nm := range mr.nodeGeoms {
		go func() {
			defer wg.Done()
			nm.RunGeometry(ctx)
		}()
	}
	return wg
}
