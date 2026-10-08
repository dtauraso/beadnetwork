package Node

import (
	"fmt"
	"slices"

	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

type Lead struct {
	TargetID string
	Vec      polar.Polar
}

type Placement struct {
	From    string
	At      Vec3
	Target  polarindex.Index
	Move    Move
	Release bool
}

type Move struct {
	Origin string
	Seq    uint64
}

type Leads struct {
	leads    []Lead
	from     Vec3
	sent     bool
	move     *Move
	placedBy *Move
	seq      uint64
	moved    map[string]uint64
	placing  *Vec3
	incoming map[string]Placement
	out      map[string]chan<- Placement
	in       []<-chan Placement
}

func (g *NodeGeometry) WirePlacement(out map[string]chan<- Placement, in []<-chan Placement) {
	g.msg.leads.out, g.msg.leads.in = out, in
}

func (k *KindPosts) PostLeads(leads []Lead) {
	k.post(func(p *KindPost) { p.Leads, p.LeadsHeld = &leads, false })
}

func (k *KindPosts) PostLeadsHeld(leads []Lead) {
	k.post(func(p *KindPost) {
		if p.Leads == nil || p.LeadsHeld {
			p.LeadsHeld = true
		}
		p.Leads = &leads
	})
}

func (g *NodeGeometry) applyLeads(leads []Lead, held bool) {
	l := &g.msg.leads
	for _, old := range l.leads {
		if !slices.ContainsFunc(leads, func(n Lead) bool { return n.TargetID == old.TargetID }) {
			g.sendPlacement(Placement{From: g.id, Release: true}, old.TargetID)
		}
	}
	l.leads = leads
	g.resolvePlacement()
	if !held {
		g.sendLeads()
		return
	}
	if g.geom.HasPlaced && l.placedBy != nil {
		l.move = l.placedBy
		g.sendLeads()
		l.move = nil
	}
}

func (g *NodeGeometry) setPlaced() {
	if p := g.msg.leads.placing; p != nil {
		g.geom.Placed, g.geom.HasPlaced = *p, true
		return
	}
	g.geom.HasPlaced = false
}

func (g *NodeGeometry) unplace() {
	g.leadCrumb("unplaced: no partner places this node, back to its index")
	g.geom.HasPlaced = false
	g.msg.PublishCenter(Vec3(NodeWorldPos(g.geom)))
	g.emitGeometry()
	g.leadsOnMove()
}

func (g *NodeGeometry) leadsOnMove() {
	l := &g.msg.leads
	if l.sent && NodeWorldPos(g.geom) == l.from {
		return
	}
	g.sendLeads()
}

func (g *NodeGeometry) sendLeads() {
	l := &g.msg.leads
	center := NodeWorldPos(g.geom)
	l.from, l.sent = center, true
	mv := l.move
	if mv == nil {
		l.seq++
		mv = &Move{Origin: g.id, Seq: l.seq}
	} else if mv.Origin == g.id {
		return
	}
	g.markMoved(*mv)
	for _, ld := range l.leads {
		tip := TipPoint(center, ld.Vec)
		g.sendPlacement(Placement{
			From:   g.id,
			At:     tip,
			Target: TipIndex(tip, g.SceneCenter(), g.Constants()),
			Move:   *mv,
		}, ld.TargetID)
	}
}

func (g *NodeGeometry) markMoved(mv Move) {
	l := &g.msg.leads
	if l.moved == nil {
		l.moved = map[string]uint64{}
	}
	l.moved[mv.Origin] = mv.Seq
}

func (g *NodeGeometry) movedIn(mv Move) bool {
	seq, ok := g.msg.leads.moved[mv.Origin]
	return ok && seq >= mv.Seq
}

func (g *NodeGeometry) sendPlacement(p Placement, to string) {
	ch, ok := g.msg.leads.out[to]
	if !ok {
		panic(fmt.Sprintf("Node.sendPlacement: node %s sends a placement to node %s but holds no placement channel to it — "+
			"the kind posting leads must wire one per partner with WirePlacement when it builds", g.id, to))
	}
	select {
	case ch <- p:
	default:
		panic(fmt.Sprintf("Node.sendPlacement: placement channel %s -> %s is full at %d unread — node %s's "+
			"geometry goroutine drains it every pulse (drainPlacements), so it has stopped running, or placements "+
			"are being sent in a loop the move check (takePlacement, movedIn) should have cut", g.id, to, cap(ch), to))
	}
}

func (g *NodeGeometry) drainPlacements() {
	for _, ch := range g.msg.leads.in {
		for {
			select {
			case p := <-ch:
				g.takePlacement(p)
				continue
			default:
			}
			break
		}
	}
}

func (g *NodeGeometry) takePlacement(p Placement) {
	l := &g.msg.leads
	if l.incoming == nil {
		l.incoming = map[string]Placement{}
	}
	if p.Release {
		g.leadCrumb(fmt.Sprintf("release from=%s", p.From))
		delete(l.incoming, p.From)
	} else if p.Move.Origin != g.id && g.movedIn(p.Move) {
		g.leadCrumb(fmt.Sprintf("ignored from=%s move=%s/%d at=%v", p.From, p.Move.Origin, p.Move.Seq, p.At))
		return
	} else {
		g.leadCrumb(fmt.Sprintf("held from=%s move=%s/%d at=%v", p.From, p.Move.Origin, p.Move.Seq, p.At))
		l.incoming[p.From] = p
	}
	g.resolvePlacement()
}

func (g *NodeGeometry) leadCrumb(text string) {
	g.trace.Post([]RowEvent{{
		Kind: KindBreadcrumb, Label: "lead", Debug: 1,
		NodeRow: g.stream.NodeRow(),
		PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Text: text,
	}})
}

func (g *NodeGeometry) pickPlacement() (Placement, bool) {
	l := &g.msg.leads
	if len(l.incoming) == 1 {
		for _, p := range l.incoming {
			return p, true
		}
	}
	if len(l.incoming) < 2 || len(l.leads) != 1 {
		return Placement{}, false
	}
	p, ok := l.incoming[l.leads[0].TargetID]
	return p, ok
}

func (g *NodeGeometry) resolvePlacement() {
	l := &g.msg.leads
	p, ok := g.pickPlacement()
	if !ok {
		if len(l.incoming) == 0 && g.geom.HasPlaced {
			g.unplace()
		}
		return
	}
	if g.geom.HasPlaced && g.geom.Placed == p.At {
		return
	}
	l.move, l.placing = &p.Move, &p.At
	mv := p.Move
	l.placedBy = &mv
	g.leadCrumb(fmt.Sprintf("placed by=%s move=%s/%d at=%v", p.From, p.Move.Origin, p.Move.Seq, p.At))
	g.msg.ApplyDerived(g.id, p.Target)
	l.move, l.placing = nil, nil
}
