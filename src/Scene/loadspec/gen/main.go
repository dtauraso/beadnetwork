package main

//go:generate go run .

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/src/Scene/loadspec/gen/wiredefs"
)

func main() {
	genName = "loadspec/gen"
	_, srcRoot := roots()

	specPath := filepath.Join(srcRoot, "Scene", "loadspec", "topo_spec.go")
	wireProps, err := wiredefs.ParseWirePropsFromFile(specPath)
	if err != nil {
		fatalf("parse wire props from topo_spec.go: %v", err)
	}

	outPath := filepath.Join(srcRoot, "Scene", "loadspec", "wire-defs.ts")
	if err := wiredefs.WriteWireDefs(outPath, wireProps); err != nil {
		fatalf("write %s: %v", outPath, err)
	}
	announce(outPath, len(wireProps), "entries")
}
