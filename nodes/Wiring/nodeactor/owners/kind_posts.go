package owners

type TiltIndexPost struct {
	Theta, NormalTheta, BottomTheta int32
}

type ReceivedVectorPost struct {
	Theta int32
	Set   bool
}

type RoundsPost struct {
	Rounds, Msgs int32
}

type KindPost struct {
	Tilt     *TiltIndexPost
	Received *ReceivedVectorPost
	Rounds   *RoundsPost
	Lattice  *int32
}

type KindPosts struct {
	ch chan KindPost
}

func NewKindPosts() KindPosts {
	return KindPosts{ch: make(chan KindPost, 1)}
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

func (k *KindPosts) PostTiltIndex(theta, normalTheta, bottomTheta int32) {
	k.post(func(p *KindPost) {
		p.Tilt = &TiltIndexPost{Theta: theta, NormalTheta: normalTheta, BottomTheta: bottomTheta}
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
