package Node

import (
	"fmt"
	"slices"

	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

type Lead struct {
	TargetID string
	Vec      polarindex.Offset
}

type Placement struct {
	From    string
	Target  polarindex.Index
	Path    []string
	Release bool
}

type Leads struct {
	leads    []Lead
	from     polarindex.Index
	sent     bool
	path     []string
	incoming map[string]Placement
	out      map[string]chan<- Placement
	in       []<-chan Placement
}

func (g *NodeGeometry) WirePlacement(out map[string]chan<- Placement, in []<-chan Placement) {
	g.msg.leads.out, g.msg.leads.in = out, in
}

func (k *KindPosts) PostLeads(leads []Lead) {
	k.post(func(p *KindPost) { p.Leads = &leads })
}

func (g *NodeGeometry) applyLeads(leads []Lead) {
	l := &g.msg.leads
	for _, old := range l.leads {
		if !slices.ContainsFunc(leads, func(n Lead) bool { return n.TargetID == old.TargetID }) {
			g.sendPlacement(Placement{From: g.id, Release: true}, old.TargetID)
		}
	}
	l.leads = leads
	g.resolvePlacement()
	g.sendLeads(g.ComposedIndex())
}

func (g *NodeGeometry) leadsOnMove(at polarindex.Index) {
	l := &g.msg.leads
	if l.sent && at == l.from {
		return
	}
	g.sendLeads(at)
}

func (g *NodeGeometry) sendLeads(at polarindex.Index) {
	l := &g.msg.leads
	l.from, l.sent = at, true
	path := append(slices.Clone(l.path), g.id)
	center := NodeWorldPos(g.geom)
	for _, ld := range l.leads {
		if slices.Contains(path, ld.TargetID) {
			continue
		}
		g.sendPlacement(Placement{
			From:   g.id,
			Target: TipIndex(center, g.SceneCenter(), ld.Vec, g.Constants()),
			Path:   path,
		}, ld.TargetID)
	}
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
			"are being sent in a loop the path check should have cut", g.id, to, cap(ch), to))
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
		delete(l.incoming, p.From)
	} else {
		l.incoming[p.From] = p
	}
	g.resolvePlacement()
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
	p, ok := g.pickPlacement()
	if !ok {
		return
	}
	l := &g.msg.leads
	if p.Target == g.ComposedIndex() {
		return
	}
	l.path = p.Path
	g.msg.ApplyDerived(g.id, p.Target)
	l.path = nil
}
