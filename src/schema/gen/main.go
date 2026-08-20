package main

//go:generate go run .

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/scripts/kindscan"
	"github.com/dtauraso/wirefold/scripts/genpaths"
	"github.com/dtauraso/wirefold/src/schema/gen/nodedefs"
	"github.com/dtauraso/wirefold/src/schema/gen/wiredefs"
)

func main() {
	genpaths.Name = "schema/gen"
	repoRoot, srcRoot := genpaths.Roots()
	kinds := genpaths.Kinds(repoRoot)

	generateKindImports(repoRoot, kinds)
	generateNodeDefs(srcRoot, kinds)
	generateWireDefs(repoRoot, srcRoot)
	generateScenes(repoRoot, srcRoot)
}

func generateKindImports(repoRoot string, kinds []kindscan.KindEntry) {
	path := filepath.Join(repoRoot, "kinds_generated.go")
	if err := kindscan.WriteKindImports(path, genpaths.KindsPkg(repoRoot), kinds); err != nil {
		genpaths.Fatalf("write %s: %v", path, err)
	}
	genpaths.Announce(path, len(kinds), "kinds")
}

func generateNodeDefs(srcRoot string, kinds []kindscan.KindEntry) {
	path := filepath.Join(srcRoot, "schema", "node-defs.ts")
	if err := nodedefs.WriteNodeDefs(path, kinds); err != nil {
		genpaths.Fatalf("write %s: %v", path, err)
	}
	genpaths.Announce(path, len(kinds), "entries")
}

func generateWireDefs(repoRoot, srcRoot string) {
	loaderPath := filepath.Join(srcRoot, "runtopology", "loadspec", "topo_spec.go")
	wireProps, err := wiredefs.ParseWirePropsFromFile(loaderPath)
	if err != nil {
		genpaths.Fatalf("parse wire props from loadspec/topo_spec.go: %v", err)
	}
	path := filepath.Join(srcRoot, "schema", "wire-defs.ts")
	if err := wiredefs.WriteWireDefs(path, wireProps); err != nil {
		genpaths.Fatalf("write %s: %v", path, err)
	}
	genpaths.Announce(path, len(wireProps), "entries")
}
