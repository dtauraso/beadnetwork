package Links

import (
	"fmt"
	"maps"
	"slices"
)

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
	SNote string
	At    Place
	Frame Frame
	Ref   Ticks

	refs      map[string]Turn
	own       map[string]*Link
	order     []string
	out       map[string]bool
	outPicked map[string]bool
	places    map[string]Place
	round     int
	point     Vec3
	placed    bool
	heard     map[string]int
	held      []any

	Changed Changes
}

type Seed struct {
	Frame  Frame
	Places map[string]Place
	Scales map[string]int
	Refs   map[string]Ticks
	Links  []Link
	TickMs int
}

func NewPeer(id string, seed Seed) *Peer {
	p := &Peer{ID: id, S: seed.Scales[id], At: seed.Places[id], Frame: seed.Frame, placed: true,
		own: map[string]*Link{}, out: map[string]bool{}, outPicked: map[string]bool{}, places: map[string]Place{}, heard: map[string]int{}}
	maps.Copy(p.places, seed.Places)
	p.Ref, p.refs = seed.Refs[id], map[string]Turn{}
	for n, r := range seed.Refs {
		p.refs[n] = RefTurn(seed.Scales[n], r)
	}
	for _, l := range seed.Links {
		switch {
		case l.To == id:
			c := l
			c.Start.Phi, c.Start.Theta = FromSigned(p.S, c.Start.Phi), FromSigned(p.S, c.Start.Theta)
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

func (p *Peer) Pos() Vec3 { return p.Frame.At(p.At) }

func (p *Peer) posOf(id string) Vec3 { return p.Frame.At(p.places[id]) }

func (p *Peer) Offset(id string) Vec3 { return p.posOf(id).Sub(p.Pos()) }

func (p *Peer) OutVecs() (vecs []Vec3, picked []bool) {
	for _, to := range slices.Sorted(maps.Keys(p.out)) {
		vecs = append(vecs, p.posOf(to).Sub(p.Pos()))
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

func (p *Peer) vecOf(l *Link) Vec3 { return p.Frame.LinkVec(p.S, *l, p.refs[l.From]) }

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
	case SetLink:
		return p.setLink(m)
	case StepAll:
		return p.stepAll(m)
	case Put:
		if m.Round > p.round {
			p.held = append(p.held, m)
			return nil
		}
		return p.recover(m)
	case Moved:
		p.moved(m)
	case PickReq:
		return p.pickReq(m)
	case PickSet:
		p.outPicked[m.From] = m.On
	case Unpick:
		return p.unpick(m)
	case SetIdx:
		return p.setIdx(m)
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
		p.places[m.From] = m.At
		p.heard[m.From] = p.round
	default:
		panic(fmt.Sprintf("Links.Peer.Handle: node %s was handed a %T, which no case takes — every message a node can be sent has a case here, so an unhandled one is a sender naming the wrong thing", p.ID, msg))
	}
	return nil
}
