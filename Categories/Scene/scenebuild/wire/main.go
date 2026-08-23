package main

//go:generate go run .

import (
	"github.com/dtauraso/wirefold/scripts/genpaths"
	"path/filepath"
)

func main() {
	genpaths.SetName("Categories/Scene/scenebuild/wire")
	_, srcRoot := genpaths.Roots()

	specPath := filepath.Join(srcRoot, "Scene", "scenebuild", "topo_spec.go")
	wireProps, err := ParseWirePropsFromFile(specPath)
	if err != nil {
		genpaths.Fatalf("parse wire props from topo_spec.go: %v", err)
	}

	outPath := filepath.Join(srcRoot, "Scene", "scenebuild", "wire-defs.ts")
	if err := WriteWireDefs(outPath, wireProps); err != nil {
		genpaths.Fatalf("write %s: %v", outPath, err)
	}
	genpaths.Announce(outPath, len(wireProps), "entries")
}
