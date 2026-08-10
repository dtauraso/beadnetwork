package Buffer

// bufLayoutScene defines the scene-sphere column block (always 1 row).
// Matched from KindSceneSphere trace events. The scene sphere is the persisted, first-class
// world anchor every node's scene polar is measured about (nodes/Wiring/sphere_layout.go
// sceneSphere) — established ONCE at load and never moved. Replaces the TS-side
// contentSphereFromCenters (a derived, non-authoritative content-sphere centroid recomputed
// from live node positions every frame) as the sphere NavGuides draws its polar tori around.
type bufLayoutScene struct {
	CX     float32 `buf:"f32"` // scene-sphere center x (world)
	CY     float32 `buf:"f32"` // scene-sphere center y (world)
	CZ     float32 `buf:"f32"` // scene-sphere center z (world)
	Radius float32 `buf:"f32"` // scene-sphere radius
}
