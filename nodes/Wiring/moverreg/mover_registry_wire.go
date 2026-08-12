package moverreg

import (
	"context"
	"sync"

	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/wire/outport"
)

func (mr *MoverRegistry) Bind(outSink map[string]*outport.Out, slotReg inputcodec.SlotRegistry) {
	for edgeID, em := range mr.edgeMovers {
		var o *outport.Out
		if oo, ok := outSink[em.SrcID()+"."+em.SrcHandle()]; ok {
			o = oo
			em.SetOut(oo)
			mr.edgeOut[edgeID] = oo
		}
		if pw, ok := slotReg[em.DstID()+"."+em.DstHandle()]; ok {
			em.SetDest(pw)

			if srcNM, ok := mr.nodeGeoms[em.SrcID()]; ok {
				srcNM.AddOutWire(pw, em.DstID(), o, em.SendSteps)
			}
		}
	}
}

func (mr *MoverRegistry) Start(ctx context.Context) *sync.WaitGroup {
	wg := new(sync.WaitGroup)

	for _, nm := range mr.nodeMovers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			nm.Run(ctx)
		}()
	}
	for _, anim := range mr.nodeAnimations {
		wg.Add(1)
		go func() {
			defer wg.Done()
			anim.Run(ctx)
		}()
	}
	for _, em := range mr.edgeMovers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			em.Run(ctx)
		}()
	}
	return wg
}

func (mr *MoverRegistry) FinalizeActors(speedSinks *[]chan float64) {
	mr.nodeMovers = map[string]*nodeactor.NodeMover{}
	mr.nodeAnimations = map[string]*nodeactor.NodeAnimation{}
	for id, ng := range mr.nodeGeoms {
		if mr.selfDriveClaimed[id] {
			continue
		}
		mr.nodeMovers[id] = nodeactor.NewNodeMover(ng)

		anim := ng.Animation()
		if speedSinks != nil {
			animSpeedCh := make(chan float64, 1)
			anim.SetSpeedCh(animSpeedCh)
			*speedSinks = append(*speedSinks, animSpeedCh)
		}
		mr.nodeAnimations[id] = anim
	}
}
