package beadcrud

import (
	"math"

	"github.com/dtauraso/wirefold/src/Node/spatial"
)

type vec3 = spatial.Vec3

type BeadCrudVerdict int

const (
	BeadCrudNone BeadCrudVerdict = iota

	BeadCrudAdd

	BeadCrudRemove
)

func BeadCrudDecide(beadSource, beadCentre, nodeDestination, dragVector vec3, beadLen float64) (BeadCrudVerdict, vec3) {
	third := nodeDestination.Sub(beadSource)
	tl := third.Length()
	switch {
	case tl < beadLen:
		return BeadCrudRemove, third
	case tl > beadLen:
		beadVec := beadCentre.Sub(beadSource)
		bl, dl := beadVec.Length(), dragVector.Length()
		if bl < 1e-12 || dl < 1e-12 {

			return BeadCrudNone, third
		}
		cosA := beadVec.Dot(dragVector) / (bl * dl)
		if cosA < 0 {

			return BeadCrudNone, third
		}
		return BeadCrudAdd, third
	default:
		return BeadCrudNone, third
	}
}

func BeadCrudImpliedCentre(verdict BeadCrudVerdict, beadCentre, aimDir vec3, beadLen float64) (vec3, bool) {
	switch verdict {
	case BeadCrudRemove:
		return beadCentre, true
	case BeadCrudAdd:
		newBeadCentre := beadCentre.Sub(aimDir.Scale(beadLen))
		nodeCentre := newBeadCentre.Sub(aimDir.Scale(beadLen))
		return nodeCentre, true
	default:
		return vec3{}, false
	}
}

type BeadCrudDiag struct {
	NeighborID  string
	ThirdLen    float64
	BeadLen     float64
	Verdict     BeadCrudVerdict
	CosAngle    float64
	GateBlocked bool
	SourceDist  float64
	Implied     vec3
	ImpliedOK   bool
}

func BeadCrudDiagnose(neighborID string, beadSource, beadCentre, aimDir, prevPos, nodeDestination, dragVector vec3, beadLen float64) BeadCrudDiag {
	third := nodeDestination.Sub(beadSource)
	tl := third.Length()
	d := BeadCrudDiag{
		NeighborID: neighborID,
		ThirdLen:   tl,
		BeadLen:    beadLen,
		CosAngle:   math.NaN(),
		SourceDist: beadSource.Sub(prevPos).Length(),
	}
	switch {
	case tl < beadLen:
		d.Verdict = BeadCrudRemove
	case tl > beadLen:
		beadVec := beadCentre.Sub(beadSource)
		bl, dl := beadVec.Length(), dragVector.Length()
		if bl < 1e-12 || dl < 1e-12 {
			d.Verdict = BeadCrudNone
		} else {
			cosA := beadVec.Dot(dragVector) / (bl * dl)
			d.CosAngle = cosA
			if cosA < 0 {
				d.Verdict = BeadCrudNone
				d.GateBlocked = true
			} else {
				d.Verdict = BeadCrudAdd
			}
		}
	default:
		d.Verdict = BeadCrudNone
	}
	if implied, ok := BeadCrudImpliedCentre(d.Verdict, beadCentre, aimDir, beadLen); ok {
		d.Implied, d.ImpliedOK = implied, true
	}
	return d
}

type BeadCrudResult struct {
	NeighborID string
	Verdict    BeadCrudVerdict
	Implied    vec3
}

func ResolveBeadCrudMove(beads []TouchingBead, prevPos, nodeDestination vec3, beadLen float64) (committed vec3, results []BeadCrudResult) {
	dragVector := nodeDestination.Sub(prevPos)
	for _, b := range beads {
		verdict, _ := BeadCrudDecide(b.Source, b.Centre, nodeDestination, dragVector, beadLen)
		if verdict == BeadCrudNone {
			continue
		}
		implied, ok := BeadCrudImpliedCentre(verdict, b.Centre, b.AimDir, beadLen)
		if !ok {
			continue
		}
		results = append(results, BeadCrudResult{NeighborID: b.NeighborID, Verdict: verdict, Implied: implied})
	}
	if len(results) == 0 {
		return prevPos, results
	}

	best := results[0]
	bestD := best.Implied.Sub(prevPos).Length()
	for _, r := range results[1:] {
		if d := r.Implied.Sub(prevPos).Length(); d < bestD {
			best, bestD = r, d
		}
	}
	return best.Implied, results
}
