package Links

import (
	"fmt"
	"maps"
)

type Changes struct {
	Place   bool
	S       bool
	Links   map[string]bool
	Dropped []string
	Refused string
}

type Peer struct {
	ID    string
	S     int
	At    Place
	Frame Frame

	out      map[string]*Link
	order    []string
	in       map[string]bool
	inPicked map[string]bool
	places   map[string]Place
	round    int
	point    Vec3
	placed   bool
	held     []any

	Changed Changes
}

type Seed struct {
	Frame  Frame
	Places map[string]Place
	Scales map[string]int
	Links  []Link
	TickMs int
}

func NewPeer(id string, seed Seed) *Peer {
	p := &Peer{ID: id, S: seed.Scales[id], At: seed.Places[id], Frame: seed.Frame, placed: true,
		out: map[string]*Link{}, in: map[string]bool{}, inPicked: map[string]bool{}, places: map[string]Place{}}
	maps.Copy(p.places, seed.Places)
	for _, l := range seed.Links {
		switch {
		case l.From == id:
			c := l
			p.out[l.To] = &c
			p.order = append(p.order, l.To)
		case l.To == id:
			p.in[l.From] = true
			p.inPicked[l.From] = l.Picked
		}
	}
	p.Changed.Links = map[string]bool{}
	return p
}

func (p *Peer) Links() []Link {
	out := make([]Link, 0, len(p.order))
	for _, to := range p.order {
		out = append(out, *p.out[to])
	}
	return out
}

func (p *Peer) Pos() Vec3 { return p.Frame.At(p.At) }

func (p *Peer) posOf(id string) Vec3 { return p.Frame.At(p.places[id]) }

func (p *Peer) TakeChanges() Changes {
	c := p.Changed
	p.Changed = Changes{Links: map[string]bool{}}
	return c
}

func (p *Peer) touched(to string) { p.Changed.Links[to] = true }

func (p *Peer) Placed() bool { return p.placed }

func (p *Peer) vecOf(l *Link) Vec3 { return p.Frame.Vec(p.S, l.Dir(p.S)) }

func (p *Peer) hasPickedIn() bool {
	for _, on := range p.inPicked {
		if on {
			return true
		}
	}
	return false
}

func (p *Peer) picked() []*Link {
	var out []*Link
	for _, to := range p.order {
		if l := p.out[to]; l.Picked {
			out = append(out, l)
		}
	}
	return out
}

func (p *Peer) Handle(msg any) []Envelope {
	switch m := msg.(type) {
	case SetStart:
		return p.setStart(m)
	case Press:
		return p.press(m)
	case SetIncrement:
		p.setIncrement(m)
	case SetLocking:
		p.setLocking(m)
	case StepAll:
		return p.stepAll(m)
	case Put:
		if !m.Step {
			return p.take(m)
		}
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
		p.pickSet(m)
	case Unpick:
		return p.unpick(m)
	case PickDropped:
		delete(p.inPicked, m.From)
	case SetIdx:
		return p.setIdx(m)
	case SetS:
		return p.setS(m)
	case Connect:
		return p.connect(m)
	case Drop:
		return p.drop(m)
	case Linked:
		return p.linked(m)
	case Unlinked:
		delete(p.in, m.From)
		delete(p.inPicked, m.From)
	default:
		panic(fmt.Sprintf("Links.Peer.Handle: node %s was handed a %T, which no case takes — every message a node can be sent has a case here, so an unhandled one is a sender naming the wrong thing", p.ID, msg))
	}
	return nil
}
