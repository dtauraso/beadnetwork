package PhiTheta

import (
	"fmt"
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	"github.com/dtauraso/beadnetwork/Categories/Node/TiltVectors"
)

func (n *PhiTheta) probeTicks() {
	s := n.peer.S
	turn := float64(Links.TurnTicks * s)
	ticks := func(v Links.Vec3) (float64, float64) {
		r := math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
		if r < 1e-12 {
			return 0, 0
		}
		phi := math.Acos(math.Max(-1, math.Min(1, v.Y/r)))
		theta := math.Atan2(v.Z, v.X)
		return phi / (2 * math.Pi) * turn, Links.Wrap(s, theta/(2*math.Pi)*turn)
	}
	off := func(t float64) float64 { return t - math.Round(t) }
	pos := n.peer.Pos()
	n.breadcrumb("tick-probe", fmt.Sprintf("node %s s=%d spokes=%d diskθ=%v pos=(%.4f %.4f %.4f)",
		n.peer.ID, s, Links.TurnTicks*s, TiltVectors.ArrowRingDiskTheta, pos.X, pos.Y, pos.Z))
	for _, l := range n.peer.Links() {
		ref := n.peer.RefOf(l.From)
		v := n.peer.Frame.LinkVec(l, ref)
		phi, theta := ticks(v)
		sum := l.Sum()
		inDisk := math.Abs(off(theta/(turn/2))) < 1e-6
		n.breadcrumb("tick-probe", fmt.Sprintf(
			"link %s→%s start(φ %v θ %v r %v) Δ(φ %d θ %d r %d) ref(φ %v θ %v turns) vec(%.4f %.4f %.4f) vecφ=%.4f ticks (off spoke %.4f) vecθ=%.4f ticks (off spoke %.4f) in φ-spoke plane=%v",
			l.From, l.To, l.Start.Phi, l.Start.Theta, l.Start.R, sum.Phi, sum.Theta, sum.R,
			ref.Phi, ref.Theta, v.X, v.Y, v.Z, phi, off(phi), theta, off(theta), inDisk))
	}
}
