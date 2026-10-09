package Node

import (
	"fmt"
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
)

type Tip struct {
	From string
	At   Vec3
}

type TipPost struct {
	TargetID string
	Vec      polar.Polar
}

type Tips struct {
	placing *Vec3
	out     map[string]chan<- Tip
	in      []<-chan Tip
}

func (g *NodeGeometry) WireTips(out map[string]chan<- Tip, in []<-chan Tip) {
	g.msg.tips.out, g.msg.tips.in = out, in
}

func (k *KindPosts) PostTips(tips []TipPost) {
	k.post(func(p *KindPost) {
		if p.Tips != nil {
			tips = append(*p.Tips, tips...)
		}
		p.Tips = &tips
	})
}

func (g *NodeGeometry) sendTips(tips []TipPost) {
	from := NodeWorldPos(g.geom)
	for _, tp := range tips {
		at := TipPoint(from, tp.Vec)
		g.tipCrumb(fmt.Sprintf("send to=%s from=%v vec(r %.3f phi %.4f theta %.4f) at=%v",
			tp.TargetID, from, tp.Vec.R, tp.Vec.Phi, tp.Vec.Theta, at))
		g.sendTip(Tip{From: g.id, At: at}, tp.TargetID)
	}
}

func (g *NodeGeometry) sendTip(t Tip, to string) {
	ch, ok := g.msg.tips.out[to]
	if !ok {
		panic(fmt.Sprintf("Node.sendTip: node %s sends a tip to node %s but holds no tip channel to it — "+
			"the kind sending tips must wire one per partner with WireTips when it builds", g.id, to))
	}
	select {
	case ch <- t:
	default:
		panic(fmt.Sprintf("Node.sendTip: tip channel %s -> %s is full at %d unread — node %s's geometry "+
			"goroutine drains it every pulse (drainTips), so it has stopped running", g.id, to, cap(ch), to))
	}
}

func (g *NodeGeometry) drainTips() {
	for _, ch := range g.msg.tips.in {
		for {
			select {
			case t := <-ch:
				g.takeTip(t)
				continue
			default:
			}
			break
		}
	}
}

func (g *NodeGeometry) takeTip(t Tip) {
	if math.IsNaN(t.At.X) || math.IsNaN(t.At.Y) || math.IsNaN(t.At.Z) {
		panic(fmt.Sprintf("Node.takeTip: node %s was sent a NaN tip by node %s", g.id, t.From))
	}
	g.tipCrumb(fmt.Sprintf("placed by=%s at=%v", t.From, t.At))
	tr := &g.msg.tips
	tr.placing = &t.At
	g.msg.CommitLocal(g.id, TipIndex(t.At, g.SceneCenter(), g.Constants()))
	tr.placing = nil
}

func (g *NodeGeometry) setPlaced() {
	if p := g.msg.tips.placing; p != nil {
		g.geom.Placed, g.geom.HasPlaced = *p, true
		return
	}
	g.geom.HasPlaced = false
}

func (g *NodeGeometry) tipCrumb(text string) {
	g.trace.Post([]RowEvent{{
		Kind: KindBreadcrumb, Label: "tip", Debug: 1,
		NodeRow: g.stream.NodeRow(),
		PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Text: text,
	}})
}
