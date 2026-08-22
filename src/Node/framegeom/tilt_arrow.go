package framegeom

import (
	"math"

	"github.com/dtauraso/wirefold/src/spatial"

	"github.com/dtauraso/wirefold/src/Camera"
	"github.com/dtauraso/wirefold/src/Node/nodegeom"
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

func axisBasisFrom(from, axis spatial.Vec3) (bx, by, bz spatial.Vec3) {
	f := from.Normalize()
	t := axis.Normalize()
	cosA := f.Dot(t)
	x := spatial.Vec3{X: 1, Y: 0, Z: 0}
	y := spatial.Vec3{X: 0, Y: 1, Z: 0}
	z := spatial.Vec3{X: 0, Y: 0, Z: 1}
	if cosA > 1-1e-9 {
		return x, y, z
	}
	if cosA < -1+1e-9 {
		perp := x
		if math.Abs(f.X) > 0.9 {
			perp = y
		}
		k := f.Cross(perp).Normalize()
		flip := func(v spatial.Vec3) spatial.Vec3 {
			return v.Scale(-1).Add(k.Scale(2 * k.Dot(v)))
		}
		return flip(x), flip(y), flip(z)
	}
	k := f.Cross(t).Normalize()
	sinA := math.Sqrt(1 - cosA*cosA)
	rot := func(v spatial.Vec3) spatial.Vec3 {
		return v.Scale(cosA).Add(k.Cross(v).Scale(sinA)).Add(k.Scale(k.Dot(v) * (1 - cosA)))
	}
	return rot(x), rot(y), rot(z)
}

func composeColumnMajor(bx, by, bz, center spatial.Vec3, sx, sy, sz float64) [16]float32 {
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

const (
	ChannelLineRadius = nodegeom.ShadingParamChannelLineRadius
	ChannelHeadRadius = nodegeom.ShadingParamChannelHeadRadius
	ChannelHeadLength = nodegeom.ShadingParamChannelHeadLength
)

func ChannelArrow(from, to spatial.Vec3) (shaft, head [16]float32, ok bool) {
	dir := to.Sub(from)
	length := dir.Length()
	if length < 1e-9 {
		return shaft, head, false
	}
	axis := dir.Normalize()
	bx, by, bz := axisBasisFrom(spatial.Vec3{X: 0, Y: 1, Z: 0}, axis)

	shaft = composeColumnMajor(bx, by, bz, from.Add(axis.Scale(length/2)), 1, length, 1)
	head = composeColumnMajor(bx, by, bz, to.Sub(axis.Scale(ChannelHeadLength)), 1, 1, 1)
	return shaft, head, true
}

func ArrowMatrices(center spatial.Vec3, length, phi float64, received bool) TiltArrow {
	axis := Camera.AnglesToWorldOffset(1, phi, ArrowRingDiskTheta).Normalize()
	up := spatial.Vec3{X: 0, Y: 1, Z: 0}
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
