package main

//go:generate go run .

import (
	"path/filepath"

	"github.com/dtauraso/beadnetwork/scripts/genpaths"
)

func main() {
	genpaths.SetName("Categories/Overlay/flags")
	_, srcRoot := genpaths.Roots()

	flagsTSPath := filepath.Join(srcRoot, "Overlay", "flags.ts")
	flags, err := ParseOverlayFlags(flagsTSPath)
	if err != nil {
		genpaths.Fatalf("parse overlay flags: %v", err)
	}
	dir := filepath.Join(srcRoot, "Overlay")
	statePath := filepath.Join(dir, "overlay_state.go")
	tablesPath := filepath.Join(dir, "overlay_tables_gen.go")
	if err := WriteOverlayGen(statePath, tablesPath, flags); err != nil {
		genpaths.Fatalf("write overlay gen: %v", err)
	}
	genpaths.Announce(statePath, len(flags), "overlay flags")
	genpaths.Announce(tablesPath, len(flags), "overlay flags")

	pathsDir := filepath.Join(dir, "paths")
	if err := WritePathFiles(pathsDir, flags); err != nil {
		genpaths.Fatalf("write overlay paths: %v", err)
	}
	genpaths.Announce(pathsDir, len(flags), "overlay paths")
	genpaths.Announce(filepath.Join(dir, "flag_paths_gen.go"), len(flags), "overlay paths")

}
