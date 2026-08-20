package main

//go:generate go run .

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/scripts/genpaths"
	"github.com/dtauraso/wirefold/src/Overlay/gen/overlaygen"
)

func main() {
	genpaths.Name = "Overlay/gen"
	_, srcRoot := genpaths.Roots()

	messagesTSPath := filepath.Join(srcRoot, "schema", "messages.ts")
	flags, err := overlaygen.ParseOverlayFlags(messagesTSPath)
	if err != nil {
		genpaths.Fatalf("parse overlay flags: %v", err)
	}
	dir := filepath.Join(srcRoot, "Overlay")
	statePath := filepath.Join(dir, "overlay_state.go")
	tablesPath := filepath.Join(dir, "overlay_tables_gen.go")
	if err := overlaygen.WriteOverlayGen(statePath, tablesPath, flags); err != nil {
		genpaths.Fatalf("write overlay gen: %v", err)
	}
	genpaths.Announce(statePath, len(flags), "overlay flags")
	genpaths.Announce(tablesPath, len(flags), "overlay flags")
}
