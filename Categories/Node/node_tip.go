package Node

import (
	"fmt"
	"math"
	"sort"

	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
)

type Tip struct {
	From   string
	Centre Vec3
	Vec    polar.Polar
	Shared bool
	Drop   bool
}

type TipPost struct {
	TargetID string
	Vec      polar.Polar
	With     string
	Drop     bool
}

type Tips struct {
	placing *Vec3
	out     map[string]chan<- Tip
	in      []<-chan Tip
	told    map[string]Tip
}

func (g *NodeGeometry) WireTips(out map[string]chan<- Tip, in []<-chan Tip) {
	g.msg.tips.out, g.msg.tips.in = out, in
}

func (k *KindPosts) PostTips(tips []TipPost) {
	k.post(func(p *KindPost) { p.Tips = &tips })
}

func (g *NodeGeometry) sendTips(tips []TipPost) {
	from := NodeWorldPos(g.geom)
	for _, tp := range tips {
		g.tipCrumb(fmt.Sprintf("send to=%s with=%q drop=%t from=%v vec(r %.3f phi %.4f theta %.4f)",
			tp.TargetID, tp.With, tp.Drop, from, tp.Vec.R, tp.Vec.Phi, tp.Vec.Theta))
		g.sendTip(Tip{From: g.id, Centre: from, Vec: tp.Vec, Shared: tp.With != "", Drop: tp.Drop}, tp.TargetID)
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
	heard := false
	for _, ch := range g.msg.tips.in {
		for {
			select {
			case t := <-ch:
				g.hearTip(t)
				heard = true
				continue
			default:
			}
			break
		}
	}
	if heard {
		g.takeTips()
	}
}

func (g *NodeGeometry) hearTip(t Tip) {
	if math.IsNaN(t.Centre.X) || math.IsNaN(t.Centre.Y) || math.IsNaN(t.Centre.Z) {
		panic(fmt.Sprintf("Node.hearTip: node %s was sent a NaN centre by node %s", g.id, t.From))
	}
	tr := &g.msg.tips
	if tr.told == nil {
		tr.told = map[string]Tip{}
	}
	if t.Drop {
		delete(tr.told, t.From)
		return
	}
	tr.told[t.From] = t
}

func (g *NodeGeometry) takeTips() {
	tr := &g.msg.tips
	ids := make([]string, 0, len(tr.told))
	for id := range tr.told {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	placers := make([]Vec3, 0, len(ids))
	for _, id := range ids {
		placers = append(placers, tr.told[id].Centre)
	}
	g.tilt.SetStartVectors(placers)

	at, miss := tipFrom(tr.told, ids)
	g.tipCrumb(fmt.Sprintf("placed by=%v at=%v miss=%q", ids, at, miss))
	if miss != "" {
		return
	}
	tr.placing = &at
	g.msg.CommitLocal(g.id, TipIndex(at, g.SceneCenter(), g.Constants()))
	tr.placing = nil
}

func tipFrom(told map[string]Tip, ids []string) (Vec3, string) {
	var shared []Tip
	for _, id := range ids {
		if t := told[id]; !t.Shared {
			return TipPoint(t.Centre, t.Vec), ""
		} else {
			shared = append(shared, t)
		}
	}
	if len(shared) < 2 {
		return Vec3{}, fmt.Sprintf("%d of the 2 centres it turns about have spoken", len(shared))
	}
	a, b := shared[0], shared[1]
	gap := b.Centre.Sub(a.Centre)
	half := gap.Length() / 2
	rho2 := a.Vec.R*a.Vec.R - half*half
	if half == 0 || rho2 < 0 {
		return Vec3{}, fmt.Sprintf("spheres miss: |D| = %.3f, R = %.3f", 2*half, a.Vec.R)
	}
	on := polar.Polar{R: math.Sqrt(rho2), Phi: a.Vec.Phi, Theta: a.Vec.Theta}
	return a.Centre.Add(gap.Scale(0.5)).Add(Vec3(polar.Polar2cart(on))), ""
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
