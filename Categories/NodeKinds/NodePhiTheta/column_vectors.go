package NodePhiTheta

import "github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"

type Turn = polarindex.Index

type Ring struct {
	Whole int
}

func mod(x, whole int) int {
	if whole <= 0 {
		return x
	}
	return (x%whole + whole) % whole
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (r Ring) Bottom(top int) int { return mod(top+r.Whole/2, r.Whole) }

func (r Ring) short(delta int) int {
	d := abs(delta) % r.Whole
	if d > r.Whole-d {
		return r.Whole - d
	}
	return d
}

func (r Ring) DistanceTop(top, arrival int) int { return r.short(top - arrival) }

func (r Ring) DistanceBottom(top, arrival int) int { return r.short(r.Bottom(top) - arrival) }

func (r Ring) DistanceOwn(center, top int) int { return r.short(center - top) }

func (r Ring) Offset(center, top, arrival int) int {
	if r.DistanceOwn(center, top) == r.Whole/4 {
		return 0
	}

	distanceTop := r.DistanceTop(top, arrival)
	distanceBottom := r.DistanceBottom(top, arrival)
	switch {
	case distanceTop == 0 && distanceBottom == 0:
		return 0
	case distanceTop < r.Whole/4:
		return -1
	case distanceBottom < r.Whole/4:
		return +1
	default:
		return 0
	}
}

func (r Ring) Next(center, top, arrival int) int {
	return mod(center+r.Offset(center, top, arrival), r.Whole/4+1)
}

func (r Ring) AtRest(center, top, arrival int) bool { return r.Offset(center, top, arrival) == 0 }

type Rings struct {
	Phi   Ring
	Theta Ring
}

func RingsFor(maxIndexPhi, maxIndexTheta int) Rings {
	return Rings{
		Phi:   Ring{Whole: maxIndexPhi},
		Theta: Ring{Whole: maxIndexTheta},
	}
}
