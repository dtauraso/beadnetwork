// vec_alias.go — local spellings for the shared vector/segment types.
//
// vec3/wireSegment used to be declared in curve_params.go, which moved to
// nodes/Wiring/nodegeom (god-object decomposition — the NodeGeom chokepoint moved together
// with its own dependencies, port_geometry.go/curve_params.go/shading_params.go/
// chain_length.go/node_geom.go/node_dims_gen.go). Wiring still has ~200 call sites
// spelling these two names, so they are re-declared here as the SAME transparent aliases
// (`type X = wire.Y`) rather than renamed at every call site: a type alias has no identity
// of its own, so Wiring's vec3 and nodegeom's vec3 are the identical type
// (wire.Vec3/wire.WireSegment) and values pass between the two packages with zero
// conversion.
package dispatch

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// vec3 aliases wire.Vec3 — see nodes/wire/geometry.go.
type vec3 = wire.Vec3

// wireSegment aliases wire.WireSegment — see nodes/wire/geometry.go.
type wireSegment = wire.WireSegment
