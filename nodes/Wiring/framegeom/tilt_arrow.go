package framegeom

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/camera"
)

const (
	arrowShaftRadiusFrac = 0.035
	arrowHeadLenFrac     = 0.22
	arrowHeadRadiusFrac  = 0.09
)

type TiltArrow struct {
	Received bool

	Shaft [16]float32
	Head  [16]float32
}

const ArrowRingDiskTheta = 0

func axisBasisFrom(from, axis vec3) (bx, by, bz vec3) {
	f := from.Normalize()
	t := axis.Normalize()
	cosA := f.Dot(t)
	x := vec3{X: 1, Y: 0, Z: 0}
	y := vec3{X: 0, Y: 1, Z: 0}
	z := vec3{X: 0, Y: 0, Z: 1}
	if cosA > 1-1e-9 {
		return x, y, z
	}
	if cosA < -1+1e-9 {
		perp := x
		if math.Abs(f.X) > 0.9 {
			perp = y
		}
		k := f.Cross(perp).Normalize()
		flip := func(v vec3) vec3 {
			return v.Scale(-1).Add(k.Scale(2 * k.Dot(v)))
		}
		return flip(x), flip(y), flip(z)
	}
	k := f.Cross(t).Normalize()
	sinA := math.Sqrt(1 - cosA*cosA)
	rot := func(v vec3) vec3 {
		return v.Scale(cosA).Add(k.Cross(v).Scale(sinA)).Add(k.Scale(k.Dot(v) * (1 - cosA)))
	}
	return rot(x), rot(y), rot(z)
}

func composeColumnMajor(bx, by, bz, center vec3, sx, sy, sz float64) [16]float32 {
	bx = bx.Scale(sx)
	by = by.Scale(sy)
	bz = bz.Scale(sz)
	return [16]float32{
		float32(bx.X), float32(bx.Y), float32(bx.Z), 0,
		float32(by.X), float32(by.Y), float32(by.Z), 0,
		float32(bz.X), float32(bz.Y), float32(bz.Z), 0,
		float32(center.X), float32(center.Y), float32(center.Z), 1,
	}
}

func ArrowMatrices(center vec3, length, phi float64, received bool) TiltArrow {
	axis := camera.AnglesToWorldOffset(1, phi, ArrowRingDiskTheta).Normalize()
	up := vec3{X: 0, Y: 1, Z: 0}
	bx, by, bz := axisBasisFrom(up, axis)

	shaftLen := length * (1 - arrowHeadLenFrac)
	shaftCenter := center.Add(axis.Scale(shaftLen / 2))
	shaft := composeColumnMajor(bx, by, bz, shaftCenter,
		length*arrowShaftRadiusFrac, shaftLen, length*arrowShaftRadiusFrac)

	headLen := length * arrowHeadLenFrac
	headCenter := center.Add(axis.Scale(length - headLen/2))
	head := composeColumnMajor(bx, by, bz, headCenter,
		length*arrowHeadRadiusFrac, headLen, length*arrowHeadRadiusFrac)

	return TiltArrow{Received: received, Shaft: shaft, Head: head}
}
