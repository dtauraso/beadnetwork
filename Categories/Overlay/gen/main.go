package main

//go:generate go run .

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/Categories/Overlay/gen/overlaygen"
)

func main() {
	genName = "Categories/Overlay/gen"
	_, srcRoot := roots()

	flagsTSPath := filepath.Join(srcRoot, "Overlay", "flags.ts")
	flags, err := overlaygen.ParseOverlayFlags(flagsTSPath)
	if err != nil {
		fatalf("parse overlay flags: %v", err)
	}
	dir := filepath.Join(srcRoot, "Overlay")
	statePath := filepath.Join(dir, "overlay_state.go")
	tablesPath := filepath.Join(dir, "overlay_tables_gen.go")
	if err := overlaygen.WriteOverlayGen(statePath, tablesPath, flags); err != nil {
		fatalf("write overlay gen: %v", err)
	}
	announce(statePath, len(flags), "overlay flags")
	announce(tablesPath, len(flags), "overlay flags")

	pathsDir := filepath.Join(dir, "paths")
	if err := overlaygen.WritePathFiles(pathsDir, flags); err != nil {
		fatalf("write overlay paths: %v", err)
	}
	announce(pathsDir, len(flags), "overlay paths")
	announce(filepath.Join(dir, "flag_paths_gen.go"), len(flags), "overlay paths")
}
