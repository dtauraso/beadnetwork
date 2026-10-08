package Node

import "fmt"

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
