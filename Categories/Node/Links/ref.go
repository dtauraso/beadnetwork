package Links

import (
	"maps"
	"math"
	"slices"
)

type Turn struct{ Phi, Theta float64 }

func RefTurn(ref Nums) Turn {
	return Turn{Phi: ref.Phi / TurnTicks, Theta: ref.Theta / TurnTicks}
}

func (r Turn) In(s int) Start {
	t := float64(TurnTicks * s)
	return Start{Phi: r.Phi * t, Theta: r.Theta * t}
}

func (r Turn) Rotate(v Vec3) Vec3 {
	sb, cb := math.Sincos(2 * math.Pi * r.Theta)
	sa, ca := math.Sincos(2 * math.Pi * r.Phi)
	x, z := v.X*cb-v.Z*sb, v.X*sb+v.Z*cb
	return Vec3{X: x*ca + v.Y*sa, Y: v.Y*ca - x*sa, Z: z}
}

func (r Turn) Unrotate(v Vec3) Vec3 {
	sb, cb := math.Sincos(2 * math.Pi * r.Theta)
	sa, ca := math.Sincos(2 * math.Pi * r.Phi)
	x, y := v.X*ca-v.Y*sa, v.X*sa+v.Y*ca
	return Vec3{X: x*cb + v.Z*sb, Y: y, Z: v.Z*cb - x*sb}
}

func (f Frame) LinkVec(l Link, ref Turn) Vec3 { return ref.Rotate(f.Vec(1, l.Dir())) }

func (l *Link) DirFrom(f Frame, ref Turn) Start {
	d, a := l.Dir(), ref.In(1)
	return Measure(f, Start{Phi: d.Phi + a.Phi, Theta: d.Theta + a.Theta}, Vec3{}, f.LinkVec(*l, ref))
}

type SetRef struct {
	K string
	V float64
}

type RefSet struct {
	From string
	Ref  Turn
}

func (p *Peer) setRef(m SetRef) []Envelope {
	if m.K == KTheta {
		p.Ref.Theta = m.V
	} else {
		p.Ref.Phi = m.V
	}
	p.Changed.Ref = true
	p.refs[p.ID] = RefTurn(p.Ref)
	var out []Envelope
	for _, to := range slices.Sorted(maps.Keys(p.out)) {
		out = append(out, Envelope{To: to, Msg: RefSet{From: p.ID, Ref: p.refs[p.ID]}})
	}
	return out
}

func (p *Peer) refSet(m RefSet) []Envelope {
	p.refs[m.From] = m.Ref
	l, ok := p.own[m.From]
	if !ok {
		return nil
	}
	p.touched(m.From)
	return p.take(l)
}

func (p *Peer) RefOf(id string) Turn { return p.refs[id] }

func (p *Peer) RefVecs() (phi, theta Vec3) {
	r := p.refs[p.ID]
	return r.Rotate(p.Frame.cart(p.S, 1, 0, 0)), r.Rotate(p.Frame.cart(p.S, 1, half(p.S)/2, 0))
}
