package nodeactor

import (
	"math"

	"github.com/dtauraso/wirefold/Categories/Node/nodeactor/owners"
	"github.com/dtauraso/wirefold/Categories/Node/nodegeom"
	"github.com/dtauraso/wirefold/Categories/Node/rulenode"
	streamframe "github.com/dtauraso/wirefold/Categories/Scene/Vectors"
)

func (m *NodeGeometry) ChannelVectorsIn() chan bool { return m.channels.In() }

func (m *NodeGeometry) pollChannelVectors() {
	_, turnedOn := m.channels.TakeOn()
	if turnedOn {
		m.channels.Forget()
	}
	center := nodegeom.NodeWorldPos(m.geom)
	if !m.channels.NeedsBroadcast(owners.Vec3(center)) {
		return
	}
	if rn := m.RuleNode(); rn != nil {
		select {
		case rn.CenterIn() <- rulenode.Vec3(center):
		default:
		}
	}
}

func (m *NodeGeometry) channelVectors() []streamframe.ChannelVector {
	if !m.channels.On() {
		return nil
	}
	self := nodegeom.NodeWorldPos(m.geom)
	peers := m.channels.PeerCenters()
	out := make([]streamframe.ChannelVector, 0, 2*len(peers))
	for _, peer := range peers {
		if shaft, head, ok := ChannelArrow(Vec3(self), Vec3(peer)); ok {
			out = append(out, streamframe.ChannelVector{Shaft: shaft, Head: head})
		}
		if shaft, head, ok := ChannelArrow(Vec3(peer), Vec3(self)); ok {
			out = append(out, streamframe.ChannelVector{Shaft: shaft, Head: head})
		}
	}
	return out
}

const (
	ChannelLineRadius = nodegeom.ShadingParamChannelLineRadius
	ChannelHeadRadius = nodegeom.ShadingParamChannelHeadRadius
	ChannelHeadLength = nodegeom.ShadingParamChannelHeadLength
)

func ChannelArrow(from, to Vec3) (shaft, head [16]float32, ok bool) {
	dir := to.Sub(from)
	length := dir.Length()
	if length < 1e-9 {
		return shaft, head, false
	}
	axis := dir.Normalize()
	bx, by, bz := axisBasisFrom(Vec3{X: 0, Y: 1, Z: 0}, axis)

	shaft = composeColumnMajor(bx, by, bz, from.Add(axis.Scale(length/2)), 1, length, 1)
	head = composeColumnMajor(bx, by, bz, to.Sub(axis.Scale(ChannelHeadLength)), 1, 1, 1)
	return shaft, head, true
}

func axisBasisFrom(from, axis Vec3) (bx, by, bz Vec3) {
	f := from.Normalize()
	t := axis.Normalize()
	cosA := f.Dot(t)
	x := Vec3{X: 1, Y: 0, Z: 0}
	y := Vec3{X: 0, Y: 1, Z: 0}
	z := Vec3{X: 0, Y: 0, Z: 1}
	if cosA > 1-1e-9 {
		return x, y, z
	}
	if cosA < -1+1e-9 {
		perp := x
		if math.Abs(f.X) > 0.9 {
			perp = y
		}
		k := f.Cross(perp).Normalize()
		flip := func(v Vec3) Vec3 {
			return v.Scale(-1).Add(k.Scale(2 * k.Dot(v)))
		}
		return flip(x), flip(y), flip(z)
	}
	k := f.Cross(t).Normalize()
	sinA := math.Sqrt(1 - cosA*cosA)
	rot := func(v Vec3) Vec3 {
		return v.Scale(cosA).Add(k.Cross(v).Scale(sinA)).Add(k.Scale(k.Dot(v) * (1 - cosA)))
	}
	return rot(x), rot(y), rot(z)
}

func composeColumnMajor(bx, by, bz, center Vec3, sx, sy, sz float64) [16]float32 {
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
