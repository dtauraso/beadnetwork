package main

//go:generate go run .

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/scripts/genpaths"
)

func main() {
	genpaths.Name = "nodegeom/gen"
	repoRoot, srcRoot := genpaths.Roots()
	kinds := genpaths.Kinds(repoRoot)

	dimsPath := filepath.Join(genpaths.NetworkDir(repoRoot), "Wiring", "nodegeom", "node_dims_gen.go")
	if err := writeNodeDims(dimsPath, kinds); err != nil {
		genpaths.Fatalf("write %s: %v", dimsPath, err)
	}
	genpaths.Announce(dimsPath, len(kinds), "kinds")

	kindIDPath := filepath.Join(srcRoot, "Node", "node_kind_id_gen.go")
	if err := writeNodeKindID(kindIDPath, kinds); err != nil {
		genpaths.Fatalf("write %s: %v", kindIDPath, err)
	}
	genpaths.Announce(kindIDPath, len(kinds), "kinds")
}
