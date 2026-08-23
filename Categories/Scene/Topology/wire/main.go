package main

//go:generate go run .

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/scripts/genpaths"
)

func main() {
	genpaths.SetName("Categories/Scene/Topology/wire")
	_, srcRoot := genpaths.Roots()

	specPath := filepath.Join(srcRoot, "Scene", "Topology", "topo_spec.go")
	wireProps, err := ParseWirePropsFromFile(specPath)
	if err != nil {
		genpaths.Fatalf("parse wire props from topo_spec.go: %v", err)
	}

	outPath := filepath.Join(srcRoot, "Scene", "Topology", "wire-defs.ts")
	if err := WriteWireDefs(outPath, wireProps); err != nil {
		genpaths.Fatalf("write %s: %v", outPath, err)
	}
	genpaths.Announce(outPath, len(wireProps), "entries")
}
