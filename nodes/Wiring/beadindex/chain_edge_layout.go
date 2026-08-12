package beadindex

import (
	"fmt"
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/edgegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/spatial"
	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

type Pulse struct {
	T     float64
	Steps int
	Val   int32
}

func ChainEdgeGeometry(selfCenter, targetCenter spatial.Vec3, selfTorusR float64, selfKind, targetKind string) (dist float64, dir spatial.Vec3, count int, ok bool) {
	dist, dir, ok = edgegeom.EdgeCenterDistAndDir(selfCenter, targetCenter)
	if !ok {
		return dist, dir, 0, false
	}
	count = edgegeom.EdgeStepCount(dist, selfKind, targetKind)
	return dist, dir, count, true
}

func ChainBeadRows(dir, chainSep spatial.Vec3, base, step float64, count int, resolved []spatial.Vec3, resolvedValid []bool, pulses []Pulse) (ox, oy, oz []float32, lit []uint8, litVal []int32) {

	for i := 0; i < count; i++ {
		var p spatial.Vec3
		if i < len(resolvedValid) && resolvedValid[i] && i < len(resolved) {

			p = resolved[i]
		} else {

			p = dir.Scale(BeadPlacementOffset(base, step, i))
		}

		p = p.Add(chainSep)
		ox = append(ox, float32(p.X))
		oy = append(oy, float32(p.Y))
		oz = append(oz, float32(p.Z))
		lit = append(lit, 0)
		litVal = append(litVal, 0)
	}

	for _, pl := range pulses {
		p := dir.Scale(PulsePlacementOffset(base, step, pl.T, pl.Steps))
		p = p.Add(chainSep)
		ox = append(ox, float32(p.X))
		oy = append(oy, float32(p.Y))
		oz = append(oz, float32(p.Z))
		lit = append(lit, 1)
		litVal = append(litVal, pl.Val)
	}
	return ox, oy, oz, lit, litVal
}

func ChainAimBreadcrumbText(to string, count int, dist float64, dir spatial.Vec3) string {
	liveTheta := math.Acos(polar.Clamp(dir.Y, -1, 1))
	livePhi := math.Atan2(dir.Z, dir.X)
	return fmt.Sprintf(
		"to=%s count=%d K=%d liveDir=(theta=%.4f,phi=%.4f)",
		to, count, int(math.Round(dist/lattice.BeadStepR)), liveTheta, livePhi)
}
