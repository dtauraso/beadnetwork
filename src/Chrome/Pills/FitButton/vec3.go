package FitButton

import "math"

type Vec3 struct{ X, Y, Z float64 }

func (a Vec3) Sub(b Vec3) Vec3 { return Vec3{a.X - b.X, a.Y - b.Y, a.Z - b.Z} }
func (a Vec3) Add(b Vec3) Vec3 { return Vec3{a.X + b.X, a.Y + b.Y, a.Z + b.Z} }
func (a Vec3) Scale(s float64) Vec3 {
	return Vec3{a.X * s, a.Y * s, a.Z * s}
}
func (a Vec3) Length() float64 {
	return math.Sqrt(a.X*a.X + a.Y*a.Y + a.Z*a.Z)
}
func (a Vec3) Normalize() Vec3 {
	l := a.Length()
	if l == 0 {
		return Vec3{}
	}
	return Vec3{a.X / l, a.Y / l, a.Z / l}
}
func (a Vec3) Dot(b Vec3) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}
func (a Vec3) Cross(b Vec3) Vec3 {
	return Vec3{
		a.Y*b.Z - a.Z*b.Y,
		a.Z*b.X - a.X*b.Z,
		a.X*b.Y - a.Y*b.X,
	}
}
