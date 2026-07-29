// registry.go — RETIRED.
//
// This file held KindRegistry and Register: each node package's init() handed over an
// EMPTY struct, which nodes/Wiring's reflectBuild then filled in by reflection. Both are
// gone. A kind now registers a real constructor for itself with
// Wiring.RegisterBuilder (nodes/Wiring/build_args.go), so the struct that reaches the
// loader is already wired and a field the constructor forgets is a compile error.
//
// The file is kept as this note rather than deleted outright because "where did
// wire.Register go?" is the obvious question when reading an older commit or doc.

package wire
