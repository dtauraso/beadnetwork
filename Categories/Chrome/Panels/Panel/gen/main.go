package main

//go:generate go run .

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/Panel/gen/panelgen"
)

func main() {
	genName = "Categories/Chrome/Panels/Panel/gen"
	_, srcRoot := roots()

	flagsTSPath := filepath.Join(srcRoot, "Chrome", "Panels", "Panel", "flags.ts")
	flags, err := panelgen.ParsePanelFlags(flagsTSPath)
	if err != nil {
		fatalf("parse panel flags: %v", err)
	}

	dir := filepath.Join(srcRoot, "Chrome", "Panels", "Panel")
	pathsDir := filepath.Join(dir, "paths")
	if err := panelgen.WritePathFiles(pathsDir, flags); err != nil {
		fatalf("write panel paths: %v", err)
	}
	announce(pathsDir, len(flags), "panel paths")
	announce(filepath.Join(dir, "flag_paths_gen.go"), len(flags), "panel paths")

	writeWireTS(srcRoot)
}
