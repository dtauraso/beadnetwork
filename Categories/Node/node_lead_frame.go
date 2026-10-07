package Node

import "math"

type Frame struct {
	X, Y, Z Vec3
	set     bool
}

func IdentityFrame() Frame {
	return Frame{X: Vec3{X: 1}, Y: Vec3{Y: 1}, Z: Vec3{Z: 1}, set: true}
}

func (f Frame) orIdentity() Frame {
	if !f.set {
		return IdentityFrame()
	}
	return f
}

func (f Frame) Apply(v Vec3) Vec3 {
	f = f.orIdentity()
	return f.X.Scale(v.X).Add(f.Y.Scale(v.Y)).Add(f.Z.Scale(v.Z))
}

func (f Frame) Then(g Frame) Frame {
	g = g.orIdentity()
	return Frame{X: f.Apply(g.X), Y: f.Apply(g.Y), Z: f.Apply(g.Z), set: true}
}

func DirFrame(phi, theta float64) Frame {
	sp, cp := math.Sin(phi), math.Cos(phi)
	st, ct := math.Sin(theta), math.Cos(theta)
	return Frame{
		X:   Vec3{X: cp * ct, Y: -sp, Z: cp * st},
		Y:   Vec3{X: sp * ct, Y: cp, Z: sp * st},
		Z:   Vec3{X: -st, Z: ct},
		set: true,
	}
}
