// distance_groups_kind_import_test.go — registers every node kind the production topology
// (distance_groups_test.go/distance_groups_scene_test.go) needs. selectleft imports
// gatecommon, which imports Wiring — an import cycle from the INTERNAL "Wiring" test
// package, so these blank imports live in the EXTERNAL "Wiring_test" package instead (same
// pattern gate_nonblocking_traversal_test.go uses for SelectRight). `go test` compiles both
// the internal and external test files of a directory into ONE binary, so each kind's
// init()-time wire.Register call is visible to distance_groups_test.go's (package Wiring)
// LoadTopology calls too.
//
// Before docs/planning/movedispatch-decomposition.md §34, most of this set (Input, Time,
// TimeEnd, TimeStart, PulseLeft, PulseRight) arrived incidentally via
// speed_delivery_full_set_test.go/vector_channel_threading_test.go's own blank imports,
// which lived in this same directory/test binary. Moving those two files to
// nodes/Wiring/build (they never named MoveDispatch) took their blank imports with them and
// broke distance_groups_test.go/distance_groups_scene_test.go's own LoadTopology calls —
// caught by re-running `go test ./...`, not by any guard. This file now names the full set
// explicitly instead of depending on another test file's incidental import list.
package dispatch_test

import (
	_ "github.com/dtauraso/wirefold/nodes/PulseLeft"
	_ "github.com/dtauraso/wirefold/nodes/PulseRight"
	_ "github.com/dtauraso/wirefold/nodes/Time"
	_ "github.com/dtauraso/wirefold/nodes/TimeEnd"
	_ "github.com/dtauraso/wirefold/nodes/TimeStart"
	_ "github.com/dtauraso/wirefold/nodes/input"
	_ "github.com/dtauraso/wirefold/nodes/pulse"
	_ "github.com/dtauraso/wirefold/nodes/selectleft"
	_ "github.com/dtauraso/wirefold/nodes/selectright"
)
