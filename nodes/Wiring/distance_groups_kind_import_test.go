// distance_groups_kind_import_test.go — registers the ONE node kind the production
// topology needs that no existing Wiring _test.go file already registers:
// "SelectLeft" (node 10). selectleft imports gatecommon,
// which imports Wiring — an import cycle from the INTERNAL "Wiring" test package, so
// this blank import lives in the EXTERNAL "Wiring_test" package instead (same pattern
// gate_nonblocking_traversal_test.go uses for SelectRight). `go test`
// compiles both the internal and external test files of a directory into ONE binary, so
// this kind's init()-time wire.Register call is visible to distance_groups_test.go's
// (package Wiring) LoadTopology calls too. Every other kind the production topology
// needs (Input, Time, Pulse, Hold, SelectRight) is already
// registered by other _test.go files in this same test binary.
package Wiring_test

import (
	_ "github.com/dtauraso/wirefold/nodes/selectleft"
)
