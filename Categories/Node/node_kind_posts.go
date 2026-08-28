package Node

import "github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"

type TiltIndexPost struct {
	Theta int32
}

type ReceivedVectorPost struct {
	Theta int32
	Set   bool
}

type RoundsPost struct {
	Rounds, Msgs int32
}

type PartnerVectorPost struct {
	PartnerID string
	Vec       polarindex.Offset
}

type KindPost struct {
	Tilt        *TiltIndexPost
	Received    *ReceivedVectorPost
	Rounds      *RoundsPost
	Lattice     *int32
	Center      *polarindex.Index
	FromPartner *PartnerVectorPost
}

type KindPosts struct {
	ch    chan KindPost
	steps chan polarindex.Offset
	done  chan struct{}
}

func NewKindPosts() KindPosts {
	return KindPosts{
		ch:    make(chan KindPost, 1),
		steps: make(chan polarindex.Offset),
		done:  make(chan struct{}),
	}
}

func (k *KindPosts) Close() {
	if k.done == nil {
		return
	}
	select {
	case <-k.done:
	default:
		close(k.done)
	}
}

func (k *KindPosts) post(mut func(*KindPost)) {
	if k.ch == nil {
		return
	}
	cur := KindPost{}
	select {
	case cur = <-k.ch:
	default:
	}
	mut(&cur)
	select {
	case k.ch <- cur:
	default:
	}
}

func (k *KindPosts) PostTiltIndex(theta int32) {
	k.post(func(p *KindPost) {
		p.Tilt = &TiltIndexPost{Theta: theta}
	})
}

func (k *KindPosts) PostReceivedVector(theta int32, set bool) {
	k.post(func(p *KindPost) { p.Received = &ReceivedVectorPost{Theta: theta, Set: set} })
}

func (k *KindPosts) PostRoundsToParallel(rounds, msgs int32) {
	k.post(func(p *KindPost) { p.Rounds = &RoundsPost{Rounds: rounds, Msgs: msgs} })
}

func (k *KindPosts) PostLatticePoints(points int32) {
	k.post(func(p *KindPost) { p.Lattice = &points })
}

func (k *KindPosts) PostCenter(center polarindex.Index) {
	k.post(func(p *KindPost) { p.Center = &center })
}

func (k *KindPosts) PostVectorFrom(partnerID string, vec polarindex.Offset) {
	k.post(func(p *KindPost) { p.FromPartner = &PartnerVectorPost{PartnerID: partnerID, Vec: vec} })
}

func (k *KindPosts) PostStep(step polarindex.Offset) {
	if k.steps == nil {
		return
	}
	select {
	case k.steps <- step:
	case <-k.done:
	}
}

func (k *KindPosts) TakeStep() (polarindex.Offset, bool) {
	if k.steps == nil {
		return polarindex.Offset{}, false
	}
	select {
	case s := <-k.steps:
		return s, true
	default:
		return polarindex.Offset{}, false
	}
}

func (k *KindPosts) Take() (KindPost, bool) {
	if k.ch == nil {
		return KindPost{}, false
	}
	select {
	case p := <-k.ch:
		return p, true
	default:
		return KindPost{}, false
	}
}
