// vec3.go — geom's own local alias for the shared vector type, mirroring the
// alias nodes/Wiring/curve_params.go already keeps for the same reason: the
// vector type and its methods (Sub/Add/Scale/Length/Normalize/Dot/Cross) live
// in nodes/wire/geometry.go; geom keeps this short name so every existing
// vec3{...} literal in this package reads unchanged.
package geom

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// vec3 aliases wire.Vec3.
type vec3 = wire.Vec3
