package main

//go:generate go run .

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/scripts/genpaths"
	"github.com/dtauraso/wirefold/src/Scene/loadspec/gen/wiredefs"
)

func main() {
	genpaths.Name = "loadspec/gen"
	_, srcRoot := genpaths.Roots()

	specPath := filepath.Join(srcRoot, "Scene", "loadspec", "topo_spec.go")
	wireProps, err := wiredefs.ParseWirePropsFromFile(specPath)
	if err != nil {
		genpaths.Fatalf("parse wire props from topo_spec.go: %v", err)
	}

	outPath := filepath.Join(srcRoot, "Scene", "loadspec", "wire-defs.ts")
	if err := wiredefs.WriteWireDefs(outPath, wireProps); err != nil {
		genpaths.Fatalf("write %s: %v", outPath, err)
	}
	genpaths.Announce(outPath, len(wireProps), "entries")
}
