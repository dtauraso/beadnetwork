package main

//go:generate go run .

import (
	"github.com/dtauraso/wirefold/scripts/genpaths"
	"path/filepath"

	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/Panel/gen/panelgen"
)

func main() {
	genpaths.SetName("Categories/Chrome/Panels/Panel/gen")
	_, srcRoot := genpaths.Roots()

	flagsTSPath := filepath.Join(srcRoot, "Chrome", "Panels", "Panel", "flags.ts")
	flags, err := panelgen.ParsePanelFlags(flagsTSPath)
	if err != nil {
		genpaths.Fatalf("parse panel flags: %v", err)
	}

	dir := filepath.Join(srcRoot, "Chrome", "Panels", "Panel")
	pathsDir := filepath.Join(dir, "paths")
	if err := panelgen.WritePathFiles(pathsDir, flags); err != nil {
		genpaths.Fatalf("write panel paths: %v", err)
	}
	genpaths.Announce(pathsDir, len(flags), "panel paths")
	genpaths.Announce(filepath.Join(dir, "flag_paths_gen.go"), len(flags), "panel paths")

	writeWireTS(srcRoot)
}
