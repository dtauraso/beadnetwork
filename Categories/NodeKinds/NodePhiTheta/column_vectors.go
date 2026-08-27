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

func (r Ring) DistanceTop(top, arrival int) int { return abs(top - arrival) }

func (r Ring) DistanceBottom(top, arrival int) int { return abs(r.Bottom(top) - arrival) }

func (r Ring) Offset(top, arrival int) int {
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

func (r Ring) Quadrant(center int) int {
	q := r.Whole / 4
	if q <= 0 {
		return center
	}
	return mod(((mod(center, r.Whole)+q/2)/q)*q, r.Whole)
}

func (r Ring) Next(center, top, arrival int) int {
	if r.Offset(top, arrival) == 0 {
		return center
	}
	return r.Quadrant(center)
}

func (r Ring) AtRest(top, arrival int) bool { return r.Offset(top, arrival) == 0 }

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
