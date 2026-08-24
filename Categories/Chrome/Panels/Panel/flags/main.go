package main

//go:generate go run .

import (
	"path/filepath"

	"github.com/dtauraso/beadnetwork/scripts/genpaths"
)

func main() {
	genpaths.SetName("Categories/Chrome/Panels/Panel/flags")
	_, srcRoot := genpaths.Roots()

	flagsTSPath := filepath.Join(srcRoot, "Chrome", "Panels", "Panel", "flags.ts")
	flags, err := ParsePanelFlags(flagsTSPath)
	if err != nil {
		genpaths.Fatalf("parse panel flags: %v", err)
	}

	dir := filepath.Join(srcRoot, "Chrome", "Panels", "Panel")
	pathsDir := filepath.Join(dir, "paths")
	if err := WritePathFiles(pathsDir, flags); err != nil {
		genpaths.Fatalf("write panel paths: %v", err)
	}
	genpaths.Announce(pathsDir, len(flags), "panel paths")
	genpaths.Announce(filepath.Join(dir, "flag_paths_gen.go"), len(flags), "panel paths")

}
