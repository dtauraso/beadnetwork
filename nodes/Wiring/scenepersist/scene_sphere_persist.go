// scene_sphere_persist.go — pure read/write helpers for the first-class SCENE SPHERE
// (sphere_layout.go sceneSphere) in view/sphere.json, lifted out of
// nodes/Wiring/scene_sphere_persist.go (which keeps the MoveDispatch-facing
// LoadSceneSphere method and the sceneSpherePersister type — see that file's own header).
// sphere.json has exactly one writer (WriteSceneSphere), so each write is a fresh
// whole-file marshal — no read-modify-write.
//
// On-disk shape:
//
//	{ "sceneSphere": { "center": [x,y,z], "radius": n } }
//
// Pointer fields distinguish "absent" from a legitimate zero so a partial object is
// rejected (→ content-fit default) rather than silently read as a degenerate sphere.
package scenepersist

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

type sceneSphereJSON struct {
	Center *[3]float64 `json:"center"`
	Radius *float64    `json:"radius"`
}

// LoadSceneSphere reads the persisted scene sphere from sphere.json. ok is false when it
// yields no complete sphere — callers then content-fit.
func LoadSceneSphere(topologyPath string) (geom.SceneSphere, bool) {
	var sj sceneSphereJSON
	jsonpersist.ReadJSONBestEffort(scenepaths.SphereFilePath(topologyPath), &sj)
	if sj.Center == nil || sj.Radius == nil {
		return geom.SceneSphere{}, false
	}
	return geom.SceneSphere{
		Center: wire.Vec3{X: sj.Center[0], Y: sj.Center[1], Z: sj.Center[2]},
		Radius: *sj.Radius,
	}, true
}

// WriteSceneSphere writes the scene sphere as the whole content of sphereJSONPath
// (sphere.json) — the sole writer of that file, so no read-modify-write is needed.
func WriteSceneSphere(sphereJSONPath string, s geom.SceneSphere) error {
	center := [3]float64{s.Center.X, s.Center.Y, s.Center.Z}
	radius := s.Radius
	return jsonpersist.WriteJSONAtomic(sphereJSONPath, sceneSphereJSON{Center: &center, Radius: &radius})
}
