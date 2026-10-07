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
	Target polarindex.Index
	Path   []string
}

type Leads struct {
	leads []Lead
	from  polarindex.Index
	sent  bool
	path  []string
	out   map[string]chan<- Placement
	in    []<-chan Placement
}

func (g *NodeGeometry) WirePlacement(out map[string]chan<- Placement, in []<-chan Placement) {
	g.msg.leads.out, g.msg.leads.in = out, in
}

func (k *KindPosts) PostLeads(leads []Lead) {
	k.post(func(p *KindPost) { p.Leads = &leads })
}

func (g *NodeGeometry) applyLeads(leads []Lead) {
	g.msg.leads.leads = leads
	g.sendLeads(g.ComposedIndex())
}

func (g *NodeGeometry) leadsOnMove(at polarindex.Index) {
	if g.msg.leads.sent && at == g.msg.leads.from {
		return
	}
	g.sendLeads(at)
}

func (g *NodeGeometry) sendLeads(at polarindex.Index) {
	g.msg.leads.from, g.msg.leads.sent = at, true
	path := append(slices.Clone(g.msg.leads.path), g.id)
	center := NodeWorldPos(g.geom)
	for _, l := range g.msg.leads.leads {
		if slices.Contains(path, l.TargetID) {
			continue
		}
		ch, ok := g.msg.leads.out[l.TargetID]
		if !ok {
			panic(fmt.Sprintf("Node.sendLeads: node %s leads node %s but holds no placement channel to it — "+
				"the kind posting leads must wire one per partner with WirePlacement when it builds", g.id, l.TargetID))
		}
		select {
		case ch <- Placement{Target: TipIndex(center, g.SceneCenter(), l.Vec, g.Constants()), Path: path}:
		default:
			panic(fmt.Sprintf("Node.sendLeads: placement channel %s -> %s is full at %d unread — node %s's "+
				"geometry goroutine drains it every pulse (drainPlacements), so it has stopped running, or leads "+
				"are being sent in a loop the path check should have cut", g.id, l.TargetID, cap(ch), l.TargetID))
		}
	}
}

func (g *NodeGeometry) drainPlacements() {
	for _, ch := range g.msg.leads.in {
		for {
			select {
			case p := <-ch:
				g.msg.leads.path = p.Path
				g.msg.ApplyDerived(g.id, p.Target)
				g.msg.leads.path = nil
				continue
			default:
			}
			break
		}
	}
}
