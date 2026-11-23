package Links

import (
	"fmt"
	"maps"
	"slices"
)

func (p *Peer) note(format string, args ...any) {
	p.Notes = append(p.Notes, fmt.Sprintf(format, args...))
}

func (p *Peer) TakeNotes() []string {
	n := p.Notes
	p.Notes = nil
	return n
}

type Changes struct {
	Place   bool
	S       bool
	Ref     bool
	Links   map[string]bool
	Refused string
}

type Peer struct {
	ID    string
	S     int
	At    *Place
	Frame Frame
	Ref   *Ref

	refs      map[string]Turn
	own       map[string]*Link
	order     []string
	out       map[string]bool
	outPicked map[string]bool
	round     int
	point     Vec3
	placed    bool
	heard     map[string]int
	held      []any
	grid      Grid
	geom      GridGeom
	sent      map[sentKey]OutLink
	fresh     bool
	drawn     map[string]Vec3
	places    map[string]Vec3

	Changed Changes
	Notes   []string
}

type Seed struct {
	Frame  Frame
	Places map[string]Place
	S      int
	Refs   map[string]Ref
	Links  []Link
	TickMs int
}

func NewPeer(id string, seed Seed) *Peer {
	at, ref := seed.Places[id], seed.Refs[id]
	p := &Peer{ID: id, S: ShownS(seed.S), At: &at, Ref: &ref, Frame: seed.Frame, placed: true,
		own: map[string]*Link{}, out: map[string]bool{}, outPicked: map[string]bool{}, heard: map[string]int{}, drawn: map[string]Vec3{}}
	p.refs, p.places = map[string]Turn{}, map[string]Vec3{}
	p.Ref.S = p.S
	for n, at := range seed.Places {
		if n != id {
			p.places[n] = seed.Frame.At(at)
		}
	}
	for n, r := range seed.Refs {
		if n != id {
			p.refs[n] = RefTurn(r)
		}
	}
	for _, l := range seed.Links {
		switch {
		case l.To == id:
			c := l
			c.Start.Signed()
			c.SetS(p.S)
			p.own[l.From] = &c
			p.order = append(p.order, l.From)
		case l.From == id:
			p.out[l.To] = true
			p.outPicked[l.To] = l.Picked
		}
	}
	p.Changed.Links = map[string]bool{}
	return p
}

func (p *Peer) Links() []Link {
	out := make([]Link, 0, len(p.order))
	for _, from := range p.order {
		out = append(out, *p.own[from])
	}
	return out
}

func (p *Peer) Pos() Vec3 { return p.Frame.At(*p.At) }

func (p *Peer) outVec(to string) Vec3 {
	if c := p.grid.at(p.grid.Shown); c != nil {
		for _, o := range c.Out {
			if o.Link.To == to {
				return p.Frame.LinkVec(o.Link, RefTurn(*p.Ref))
			}
		}
	}
	return Vec3{}
}

func (p *Peer) Offset(from string) Vec3 { return Vec3{}.Sub(p.vecOf(p.own[from])) }

func (p *Peer) OutVecs() (vecs []Vec3, picked []bool) {
	for _, to := range slices.Sorted(maps.Keys(p.out)) {
		vecs = append(vecs, p.outVec(to))
		picked = append(picked, p.outPicked[to])
	}
	return vecs, picked
}

func (p *Peer) PickedOuts() map[string]bool { return maps.Clone(p.outPicked) }

func (p *Peer) TakeChanges() Changes {
	c := p.Changed
	p.Changed = Changes{Links: map[string]bool{}}
	return c
}

func (p *Peer) touched(from string) { p.Changed.Links[from] = true }

func (p *Peer) Ready() bool {
	if !p.placed || len(p.held) > 0 {
		return false
	}
	for _, id := range p.neighbours() {
		if p.heard[id] != p.round {
			return false
		}
	}
	return true
}

func (p *Peer) vecOf(l *Link) Vec3 { return p.Frame.LinkVec(*l, p.refs[l.From]) }

func (p *Peer) pickedIn() *Link {
	for _, from := range p.order {
		if l := p.own[from]; l.Picked {
			return l
		}
	}
	return nil
}

func (p *Peer) neighbours() []string {
	ids := map[string]bool{}
	for from := range p.own {
		ids[from] = true
	}
	for to := range p.out {
		ids[to] = true
	}
	return slices.Sorted(maps.Keys(ids))
}

func (p *Peer) Handle(msg any) []Envelope {
	switch m := msg.(type) {
	case StepEdit:
		return p.stepEdit(m)
	case ShowStep:
		return p.showStep(m)
	case OutLink:
		p.outLink(m)
	case WantOut:
		p.wantOut(m)
	case StepAll:
		return p.stepAll(m)
	case Put:
		if m.Round > p.round {
			p.held = append(p.held, m)
			return nil
		}
		return p.recover(m)
	case Moved:
		return p.moved(m)
	case PickReq:
		return p.pickReq(m)
	case PickSet:
		return p.pickSet(m)
	case Unpick:
		return p.unpick(m)
	case SetS:
		return p.setS(m)
	case SetRef:
		return p.setRef(m)
	case RefSet:
		return p.refSet(m)
	case Connect:
		return p.connect(m)
	case Linked:
		p.out[m.From] = true
		p.heard[m.From] = p.round
	default:
		panic(fmt.Sprintf("Links.Peer.Handle: node %s was handed a %T, which no case takes — every message a node can be sent has a case here, so an unhandled one is a sender naming the wrong thing", p.ID, msg))
	}
	return nil
}
