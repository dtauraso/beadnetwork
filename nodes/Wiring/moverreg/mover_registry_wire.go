package moverreg

import (
	"context"
	"sync"

	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/wire/outport"
)

func (mr *MoverRegistry) Bind(outSink map[string]*outport.Out, slotReg inputcodec.SlotRegistry, edgeRowFor func(src, dst string) (int32, bool)) {
	for edgeID, em := range mr.edgeMovers {
		if oo, ok := outSink[em.SrcID()+"."+em.SrcHandle()]; ok {
			em.SetOut(oo)
			mr.edgeOut[edgeID] = oo
		}
		if pw, ok := slotReg[em.DstID()+"."+em.DstHandle()]; ok {
			em.SetDest(pw)

			if srcNM, ok := mr.nodeGeoms[em.SrcID()]; ok {
				edgeRow := int32(-1)
				if edgeRowFor != nil {
					if r, ok := edgeRowFor(em.SrcID(), em.DstID()); ok {
						edgeRow = r
					}
				}
				srcNM.AddOutWire(pw, edgeRow)
			}
		}
	}
}

func (mr *MoverRegistry) Start(ctx context.Context) *sync.WaitGroup {
	wg := new(sync.WaitGroup)

	for _, em := range mr.edgeMovers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			em.Run(ctx)
		}()
	}
	return wg
}
