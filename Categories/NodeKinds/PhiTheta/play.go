package PhiTheta

import (
	"github.com/dtauraso/beadnetwork/Categories/Node/AngleArcs"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
)

func (n *PhiTheta) play(msg any) bool {
	switch m := msg.(type) {
	case Links.PlayFrame:
		if m.Run <= n.stopped {
			return true
		}
		if m.Stop {
			n.stopPlay(m.Run)
			return true
		}
		n.playing = m.Run
		n.postPlayOuts(m.Frac())
		if !n.peer.HasPickedIn() {
			n.send(n.peer.PlayTo(m.Run, n.peer.Pos(), m.K, m.N))
		}
		return true
	case Links.PlayAt:
		if m.Run <= n.stopped {
			return true
		}
		v, ok := n.peer.PlayIn(m.From, m.Frac())
		if !ok {
			return true
		}
		n.playing = m.Run
		at := m.P.Add(v)
		n.geom.KindPosts().PostPlace(n.peer.Frame.GridAt(at))
		n.send(n.peer.PlayTo(m.Run, at, m.K, m.N))
		return true
	}
	return false
}

func (n *PhiTheta) postPlayOuts(frac float64) {
	var out []polar.Polar
	vecs, picked := n.peer.PlayOutVecs(frac)
	for _, v := range vecs {
		out = append(out, polar.Cart2polar(v))
	}
	n.geom.KindPosts().PostStartVectors(out, picked)
	n.geom.KindPosts().PostArcs([]AngleArcs.Angle{})
}

func (n *PhiTheta) stopPlay(run int) {
	n.stopped = max(n.stopped, run)
	if n.playing == 0 {
		return
	}
	n.playing = 0
	n.geom.KindPosts().PostPlace(*n.peer.At)
	n.postVectors()
}

func (n *PhiTheta) endPlay() {
	if n.playing != 0 {
		n.stopPlay(n.playing)
	}
}
