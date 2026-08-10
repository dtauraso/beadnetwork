// sphere_layout.go — graph-level node-position helpers for the polar layout.
// Node positions are stored as scene polar (meta.json scenePolar) about the scene sphere center.

package geom

// SphereEdge is a DIRECTED connection: Source outputs to Target.
type SphereEdge struct {
	Source string
	Target string
}

// SceneSphere is the FIRST-CLASS scene reference (polar-model.md): the fixed frame every
// node's SCENE polar (r, θ, φ) is measured about. It is NOT the derived content-sphere
// centroid (ContentSphereOf) — that moves when nodes move, which is circular. The Center
// is the single cartesian value in the system (the world anchor, persisted in sphere.json);
// Radius is long enough to fit the whole diagram and re-fits on pan. A node's world center
// is DERIVED as Center + Polar2cart(scenePolar); its scene polar is Cart2polar(world −
// Center). Panning moves Center and recomputes every node's scene polar + re-fits Radius.
type SceneSphere struct {
	Center vec3
	Radius float64
}

// ContentFitSceneSphere derives a sensible DEFAULT scene sphere from the current node
// centers (bbox midpoint + fit radius) — used only when sphere.json has no persisted sphere
// yet, so an existing scene gets a sane reference without any authored value. Once
// persisted, the stored Center is authoritative and is NOT re-derived from node positions.
func ContentFitSceneSphere(centers map[string]vec3) SceneSphere {
	c, r := ContentSphereOf(centers)
	return SceneSphere{Center: c, Radius: r}
}
