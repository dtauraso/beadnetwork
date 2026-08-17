package tiltring

import "fmt"

type State struct {
	Idx int32

	R *Ring

	Next     *State
	Prev     *State
	Opposite *State
	Quarter  *State
}

type Ring struct {
	Points      int32
	QuarterTurn int32
	HalfTurn    int32

	States []State
}

func NewRing(points int32) *Ring {
	if points < 4 || points%4 != 0 {
		panic(fmt.Sprintf(
			"tiltring: a lattice needs a positive multiple of four points — got %d; a quarter turn must be a whole number of states or the coplanar normal and the parallel halt name nothing",
			points))
	}
	r := &Ring{
		Points:      points,
		QuarterTurn: points / 4,
		HalfTurn:    points / 2,
		States:      make([]State, points),
	}
	for i := range r.States {
		r.States[i].Idx = int32(i)
		r.States[i].R = r
	}
	n := int32(len(r.States))
	for i := int32(0); i < n; i++ {
		r.States[i].Next = &r.States[(i+1)%n]
		r.States[i].Prev = &r.States[(i-1+n)%n]
		r.States[i].Opposite = &r.States[(i+r.HalfTurn)%n]
		r.States[i].Quarter = &r.States[(i+r.QuarterTurn)%n]
	}
	return r
}

func (r *Ring) At(idx int32) *State { return &r.States[idx] }

func (r *Ring) ArrivedState(idx int32) *State {
	if idx < 0 || idx >= r.Points {
		panic(fmt.Sprintf(
			"tiltring: a direction arriving on the vector channel must already be an index on this node's own %d-point ring (0..%d) — got %d; the sender is this same kind sending one of its own states, so an index off the ring is a defect or a partner on a different lattice, not something to fold onto this one",
			r.Points, r.Points-1, idx))
	}
	return r.At(idx)
}

func (r *Ring) SeedState(idx int32) (s *State, unknown bool) {
	for i := range r.States {
		if r.States[i].Idx == idx {
			return &r.States[i], false
		}
	}
	return &r.States[0], true
}

func (s *State) AngleLength(target *State) int32 {

	d := s.Idx - target.Idx
	if d < 0 {
		d = -d
	}
	if d > s.R.HalfTurn {
		d = s.R.Points - d
	}
	return d
}
