package geom

type SphereEdge struct {
	Source string
	Target string
}

type SceneSphere struct {
	Center vec3
	Radius float64
}

func ContentFitSceneSphere(centers map[string]vec3) SceneSphere {
	c, r := ContentSphereOf(centers)
	return SceneSphere{Center: c, Radius: r}
}
