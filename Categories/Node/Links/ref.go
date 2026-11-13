package Links

import (
	"fmt"
	"maps"
	"math"
	"slices"
)

type Turn struct{ Phi, Theta float64 }

func RefTurn(ref Ref) Turn {
	return Turn{Phi: float64(ref.Phi) / TurnTicks, Theta: float64(ref.Theta) / TurnTicks}
}

func MustBeRefCount(id string, ref Ref) {
	const most = 1 << 40
	for _, v := range []int{ref.Phi, ref.Theta, ref.S} {
		if v > most || v < -most {
			panic(fmt.Sprintf("Links.MustBeRefCount: node %s holds %d in a ref file, which is a float64 read as an int64 — the reference Δφ, Δθ and s are whole numbers stored as int64, so this file was written before they were", id, v))
		}
	}
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

func (f Frame) LinkVec(l Link, ref Turn) Vec3 {
	p, d := l.Drawn()
	return ref.Rotate(f.Vec(p, d))
}

type SetRef struct {
	K string
	V int
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
	var out []Envelope
	for _, to := range slices.Sorted(maps.Keys(p.out)) {
		out = append(out, Envelope{To: to, Msg: RefSet{From: p.ID, Ref: RefTurn(*p.Ref)}})
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
	return p.take(l, true)
}

func (p *Peer) RefOf(id string) Turn {
	if id == p.ID {
		return RefTurn(*p.Ref)
	}
	return p.refs[id]
}

func (p *Peer) RefVecs() (phi, theta Vec3) {
	r := RefTurn(*p.Ref)
	return r.Rotate(p.Frame.cart(p.S, 1, 0, 0)), r.Rotate(p.Frame.cart(p.S, 1, half(p.S)/2, 0))
}
