package Node

import (
	"fmt"
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
)

type Turn struct {
	From   string
	Centre bool
	At     Vec3
	APhi   float64
	ATheta float64
}

type TurnPost struct {
	TargetID string
	APhi     float64
	ATheta   float64
}

type Turns struct {
	centreID string
	centres  map[string]Vec3
	measured bool
	local    polar.Polar
	placing  *Vec3
	out      map[string]chan<- Turn
	in       []<-chan Turn
}

func (g *NodeGeometry) WireTurns(out map[string]chan<- Turn, in []<-chan Turn) {
	g.msg.turns.out, g.msg.turns.in = out, in
}

func (k *KindPosts) PostCentre(id string) {
	k.post(func(p *KindPost) { p.Centre = &id })
}

func (k *KindPosts) PostTurns(turns []TurnPost) {
	k.post(func(p *KindPost) {
		if p.Turns != nil {
			turns = append(*p.Turns, turns...)
		}
		p.Turns = &turns
	})
}

func (g *NodeGeometry) setCentre(id string) {
	t := &g.msg.turns
	t.centreID, t.measured = id, false
	g.turnCrumb(fmt.Sprintf("centre node=%q", id))
	if id == g.id {
		g.sendCentre()
	}
}

func (g *NodeGeometry) sendCentre() {
	for to := range g.msg.turns.out {
		g.sendTurn(Turn{From: g.id, Centre: true, At: NodeWorldPos(g.geom)}, to)
	}
}

func (g *NodeGeometry) sendTurns(turns []TurnPost) {
	for _, tp := range turns {
		if tp.APhi == 0 && tp.ATheta == 0 {
			continue
		}
		g.sendTurn(Turn{From: g.id, APhi: tp.APhi, ATheta: tp.ATheta}, tp.TargetID)
	}
}

func (g *NodeGeometry) sendTurn(t Turn, to string) {
	ch, ok := g.msg.turns.out[to]
	if !ok {
		panic(fmt.Sprintf("Node.sendTurn: node %s sends a turn to node %s but holds no turn channel to it — "+
			"the kind sending turns must wire one per partner with WireTurns when it builds", g.id, to))
	}
	select {
	case ch <- t:
	default:
		panic(fmt.Sprintf("Node.sendTurn: turn channel %s -> %s is full at %d unread — node %s's geometry "+
			"goroutine drains it every pulse (drainTurns), so it has stopped running", g.id, to, cap(ch), to))
	}
}

func (g *NodeGeometry) drainTurns() {
	for _, ch := range g.msg.turns.in {
		for {
			select {
			case t := <-ch:
				g.takeTurn(t)
				continue
			default:
			}
			break
		}
	}
}

func (g *NodeGeometry) takeTurn(t Turn) {
	tr := &g.msg.turns
	if t.Centre {
		if tr.centres == nil {
			tr.centres = map[string]Vec3{}
		}
		tr.centres[t.From] = t.At
		if t.From == tr.centreID {
			tr.measured = false
		}
		return
	}
	c, ok := tr.centres[tr.centreID]
	if tr.centreID == "" || tr.centreID == g.id || !ok {
		g.turnCrumb(fmt.Sprintf("turn from=%s ignored: no centre node selected (centre=%q)", t.From, tr.centreID))
		return
	}
	p := NodeWorldPos(g.geom)
	if !tr.measured {
		tr.local = polar.Cart2polar(polar.Vec3(p.Sub(c)))
		tr.measured = true
	}
	tr.local.Phi += t.APhi
	tr.local.Theta += t.ATheta
	next := c.Add(Vec3(polar.Polar2cart(tr.local)))
	g.turnCrumb(fmt.Sprintf("turn from=%s about=%s aphi=%.4f atheta=%.4f local(r %.3f phi %.4f theta %.4f) at=%v",
		t.From, tr.centreID, t.APhi, t.ATheta, tr.local.R, tr.local.Phi, tr.local.Theta, next))
	if math.IsNaN(next.X) || math.IsNaN(next.Y) || math.IsNaN(next.Z) {
		panic(fmt.Sprintf("Node.takeTurn: node %s turned to a NaN position about centre %s — local %+v", g.id, tr.centreID, tr.local))
	}
	tr.placing = &next
	g.msg.CommitLocal(g.id, TipIndex(next, g.SceneCenter(), g.Constants()))
	tr.placing = nil
}

func (g *NodeGeometry) setPlaced() {
	tr := &g.msg.turns
	if p := tr.placing; p != nil {
		g.geom.Placed, g.geom.HasPlaced = *p, true
		return
	}
	g.geom.HasPlaced, tr.measured = false, false
}

func (g *NodeGeometry) turnsOnMove() {
	if g.msg.turns.centreID == g.id {
		g.sendCentre()
	}
}

func (g *NodeGeometry) turnCrumb(text string) {
	g.trace.Post([]RowEvent{{
		Kind: KindBreadcrumb, Label: "turn", Debug: 1,
		NodeRow: g.stream.NodeRow(),
		PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Text: text,
	}})
}
