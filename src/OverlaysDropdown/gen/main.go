// The overlays generator: the overlay flag state and lookup tables, read from
// the flag names messages.ts already declares.
package main

//go:generate go run .

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/scripts/genpaths"
	"github.com/dtauraso/wirefold/src/OverlaysDropdown/gen/overlaygen"
)

func main() {
	genpaths.Name = "OverlaysDropdown/gen"
	_, srcRoot := genpaths.Roots()

	messagesTSPath := filepath.Join(srcRoot, "messages.ts")
	flags, err := overlaygen.ParseOverlayFlags(messagesTSPath)
	if err != nil {
		genpaths.Fatalf("parse overlay flags: %v", err)
	}
	dir := filepath.Join(srcRoot, "OverlaysDropdown")
	statePath := filepath.Join(dir, "overlay_state.go")
	tablesPath := filepath.Join(dir, "overlay_tables_gen.go")
	if err := overlaygen.WriteOverlayGen(statePath, tablesPath, flags); err != nil {
		genpaths.Fatalf("write overlay gen: %v", err)
	}
	genpaths.Announce(statePath, len(flags), "overlay flags")
	genpaths.Announce(tablesPath, len(flags), "overlay flags")
}
