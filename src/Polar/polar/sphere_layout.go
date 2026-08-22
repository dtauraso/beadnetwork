package polar

import (
	"math"

	"github.com/dtauraso/wirefold/src/spatial"
)

type SceneSphere struct {
	Center spatial.Vec3
	Radius float64
}

func ContentSphereOf(centers map[string]spatial.Vec3) (center spatial.Vec3, radius float64) {
	if len(centers) == 0 {
		return spatial.Vec3{}, 100
	}
	min := spatial.Vec3{X: math.Inf(1), Y: math.Inf(1), Z: math.Inf(1)}
	max := spatial.Vec3{X: math.Inf(-1), Y: math.Inf(-1), Z: math.Inf(-1)}
	for _, p := range centers {
		if math.IsInf(p.X, 0) || math.IsNaN(p.X) {
			continue
		}
		min.X, max.X = math.Min(min.X, p.X), math.Max(max.X, p.X)
		min.Y, max.Y = math.Min(min.Y, p.Y), math.Max(max.Y, p.Y)
		min.Z, max.Z = math.Min(min.Z, p.Z), math.Max(max.Z, p.Z)
	}
	center = min.Add(max).Scale(0.5)
	r := 0.0
	for _, p := range centers {
		r = math.Max(r, p.Sub(center).Length())
	}
	return center, math.Max(r*1.1, 1)
}

func ContentFitSceneSphere(centers map[string]spatial.Vec3) SceneSphere {
	c, r := ContentSphereOf(centers)
	return SceneSphere{Center: c, Radius: r}
}
