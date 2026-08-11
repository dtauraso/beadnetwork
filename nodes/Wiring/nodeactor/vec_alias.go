// vec_alias.go — local spelling for the shared vector type, mirroring package Wiring's own
// vec_alias.go (movedispatch-decomposition §20): vec3 is a transparent alias (`type X =
// wire.Y`), so nodeactor's vec3 and Wiring's vec3 are the identical type and values pass
// between the two packages with zero conversion.
package nodeactor

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// vec3 aliases wire.Vec3 — see nodes/wire/geometry.go.
type vec3 = wire.Vec3
