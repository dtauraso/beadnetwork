package NodePhiTheta3

import (
	"context"
	"fmt"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/TiltPanel"
	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	NodeCat "github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

const nodeCount = 3

var (
	link  [nodeCount][nodeCount]chan TiltPanel.TiltVectorMsg
	built int
)

func init() {
	for from := 0; from < nodeCount; from++ {
		for to := 0; to < nodeCount; to++ {
			if from != to {
				link[from][to] = make(chan TiltPanel.TiltVectorMsg)
			}
		}
	}
}

type NodePhiTheta3 struct {
	geom *NodeCat.NodeGeometry

	Clock   clock.Clock
	SpeedCh <-chan float64

	Me       int
	Partners [2]int

	Top   Turn
	Rings Rings

	loggedPhi, loggedTheta int
	logged                 bool
}

type Turn = polarindex.Index

func vectorOf(m TiltPanel.TiltVectorMsg) Turn {
	return Turn{Phi: int(m.PhiIdx), Theta: int(m.ThetaIdx), R: int(m.RIdx)}
}

type Ring struct {
	Whole int
}

type Rings struct {
	Phi   Ring
	Theta Ring
}

func RingsFor(maxIndexPhi, maxIndexTheta int) Rings {
	return Rings{Phi: Ring{Whole: maxIndexPhi}, Theta: Ring{Whole: maxIndexTheta}}
}

func mod(x, whole int) int {
	if whole <= 0 {
		return x
	}
	return (x%whole + whole) % whole
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (r Ring) Bottom(top int) int { return mod(top+r.Whole/2, r.Whole) }

func (r Ring) DistanceTop(top, arrival int) int { return mod(abs(top-arrival), r.Whole/4) }

func (r Ring) DistanceBottom(top, arrival int) int {
	return mod(abs(r.Bottom(top)-arrival), r.Whole/4)
}

func (r Ring) OffsetCase(top, arrival int) (offset, rule int) {
	distanceTop := r.DistanceTop(top, arrival)
	distanceBottom := r.DistanceBottom(top, arrival)

	switch {
	case distanceTop == 0 && distanceBottom == 0:
		return 0, 1
	case distanceTop < r.Whole/4:
		return -1, 2
	case distanceBottom < r.Whole/4:
		return -1, 3
	}

	return 0, 0
}

func (r Ring) Offset(top, arrival int) int {
	offset, _ := r.OffsetCase(top, arrival)
	return offset
}

func (r Ring) Next(center, top, arrival int) int {
	return mod(center+r.Offset(top, arrival), r.Whole)
}

func (r Ring) AtRest(top, arrival int) bool { return r.Offset(top, arrival) == 0 }

func (n *NodePhiTheta3) breadcrumb(label, value string) {
	n.geom.Trace().Post([]NodeCat.RowEvent{{
		Kind: NodeCat.KindBreadcrumb, Label: label, Debug: 1,
		NodeRow: n.geom.Stream().NodeRow(),
		PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Text: value,
	}})
}

func (r Ring) logLine(angle string, center, top, arrival, next int) string {
	offset, rule := r.OffsetCase(top, arrival)
	return fmt.Sprintf("%s c=%d->%d top=%d bot=%d arr=%d dTop=%d dBot=%d q=%d off=%+d case=%d",
		angle, center, next, top, r.Bottom(top), arrival,
		r.DistanceTop(top, arrival), r.DistanceBottom(top, arrival),
		r.Whole/4, offset, rule)
}

func (n *NodePhiTheta3) Update(ctx context.Context) {
	clk := n.Clock.Copy()
	clk.SpeedFrom(n.SpeedCh)
	n.geom.Clocks().Use(clk)

	for {
		if ctx.Err() != nil {
			return
		}

		if clk.Speed() > 0 {
			here := Turn(n.geom.ComposedIndex())
			mine := TiltPanel.TiltVectorMsg{
				PhiIdx: int32(here.Phi), ThetaIdx: int32(here.Theta), RIdx: int32(here.R),
			}

			var arrivals [2]Turn
			sent, got := [2]bool{}, [2]bool{}
			for !(sent[0] && sent[1] && got[0] && got[1]) {
				a, b := n.Partners[0], n.Partners[1]

				var outA, outB chan<- TiltPanel.TiltVectorMsg
				if !sent[0] {
					outA = link[n.Me][a]
				}
				if !sent[1] {
					outB = link[n.Me][b]
				}
				var inA, inB <-chan TiltPanel.TiltVectorMsg
				if !got[0] {
					inA = link[a][n.Me]
				}
				if !got[1] {
					inB = link[b][n.Me]
				}

				select {
				case outA <- mine:
					sent[0] = true
				case outB <- mine:
					sent[1] = true
				case v := <-inA:
					arrivals[0], got[0] = vectorOf(v), true
				case v := <-inB:
					arrivals[1], got[1] = vectorOf(v), true
				case <-ctx.Done():
					return
				}
			}

			arrival := arrivals[0]

			offPhi := n.Rings.Phi.Offset(n.Top.Phi, arrival.Phi)
			offTheta := n.Rings.Theta.Offset(n.Top.Theta, arrival.Theta)

			if !n.logged || offPhi != n.loggedPhi {
				n.breadcrumb("phi", n.Rings.Phi.logLine("phi", here.Phi, n.Top.Phi, arrival.Phi,
					n.Rings.Phi.Next(here.Phi, n.Top.Phi, arrival.Phi)))
				n.loggedPhi = offPhi
			}
			if !n.logged || offTheta != n.loggedTheta {
				n.breadcrumb("theta", n.Rings.Theta.logLine("theta", here.Theta, n.Top.Theta, arrival.Theta,
					n.Rings.Theta.Next(here.Theta, n.Top.Theta, arrival.Theta)))
				n.loggedTheta = offTheta
			}
			n.logged = true

			if offPhi != 0 || offTheta != 0 {
				n.geom.KindPosts().PostStep(polarindex.Offset{Phi: offPhi, Theta: offTheta})
			}
		}

		if err := clk.SleepCycle(ctx); err != nil {
			return
		}
	}
}

func (a BuildArgs) Geom() *NodeCat.NodeGeometry {
	if a.Deps == nil {
		return nil
	}
	ng, _ := a.Deps.SelfDriveGeom(a.Name).(*NodeCat.NodeGeometry)
	return ng
}

var Builder = BuilderFor("NodePhiTheta3",
	func(a BuildArgs) (any, error) {
		n := &NodePhiTheta3{}
		n.Clock = a.Clock()
		n.SpeedCh = a.SpeedCh()
		n.geom = a.Geom()

		n.Me = built % nodeCount
		n.Partners = [2]int{(n.Me + 1) % nodeCount, (n.Me + 2) % nodeCount}
		built++

		here := Turn(n.geom.ComposedIndex())
		c := n.geom.Constants()
		n.Rings = RingsFor(c.MaxIndexPhi, c.MaxIndexTheta)
		n.Top = Turn{Phi: 0, Theta: 0, R: here.R}

		n.breadcrumb("built", fmt.Sprintf("name=%s at phi=%d theta=%d r=%d  rings phi=%d theta=%d",
			a.Name, here.Phi, here.Theta, here.R, n.Rings.Phi.Whole, n.Rings.Theta.Whole))

		return n, nil
	})
