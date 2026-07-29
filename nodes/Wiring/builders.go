// builders.go — geometry constants shared by the node/edge movers.
//
// This file used to hold the CENTRAL REFLECTION BUILD PIPELINE: reflectBuild, which took
// an empty struct from wire.Register and filled it in by matching field names and types
// (injectClosures, injectFunc), struct shape (reflectPorts/collectPorts/wirePorts) and
// struct tags (populateData). All of it is gone — every kind now constructs itself through
// Wiring.RegisterBuilder/BuildArgs (build_args.go), so a field the builder forgets is a
// compile error instead of a silent nil at runtime.
//
// Deleting it is what PROVED the migration: if any kind had still depended on reflection,
// removing these functions would not compile.

package Wiring

// verticalRingNormal and flatRingNormal are the two great-circle ring normals
// streamed on every node-geometry event so TS never hardcodes ring orientation.
// vertical: ring stands upright (normal points along +Z world axis).
// flat: ring lies flat (normal points along +Y world axis, Three y-up convention).
const (
	verticalRingNormalX, verticalRingNormalY, verticalRingNormalZ = 0.0, 0.0, 1.0
	flatRingNormalX, flatRingNormalY, flatRingNormalZ             = 0.0, 1.0, 0.0
)
