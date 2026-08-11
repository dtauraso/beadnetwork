// vec_alias.go — local spelling for the shared vector type.
//
// vec3 used to be declared (alongside wireSegment) in curve_params.go, which moved to
// nodes/Wiring/nodegeom (god-object decomposition — the NodeGeom chokepoint moved together
// with its own dependencies, port_geometry.go/curve_params.go/shading_params.go/
// chain_length.go/node_geom.go/node_dims_gen.go). Wiring still has call sites spelling
// this name, so it is re-declared here as the SAME transparent alias (`type vec3 =
// wire.Vec3`) rather than renamed at every call site: a type alias has no identity of its
// own, so Wiring's vec3 and nodegeom's vec3 are the identical type (wire.Vec3) and values
// pass between the two packages with zero conversion. wireSegment's own alias moved with
// build.go/loader.go to nodes/Wiring/build (this task) — its last dispatch-package call
// site left with them; build.go/loader.go now spell it wire.WireSegment directly.
package dispatch

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// vec3 aliases wire.Vec3 — see nodes/wire/geometry.go.
type vec3 = wire.Vec3
