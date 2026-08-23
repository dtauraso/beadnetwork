package main

//go:generate go run .

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/Categories/NodeKinds/gen/kindscan"
	"github.com/dtauraso/wirefold/scripts/genpaths"
)

func main() {
	genpaths.SetName("Categories/Node/gen/kinds")
	repoRoot, srcRoot := genpaths.Roots()
	kinds := kindscan.Kinds(repoRoot)

	dir := filepath.Join(srcRoot, "Node")

	dimsPath := filepath.Join(dir, "geom_node_dims_gen.go")
	if err := writeNodeDims(dimsPath, kinds); err != nil {
		genpaths.Fatalf("write %s: %v", dimsPath, err)
	}
	genpaths.Announce(dimsPath, len(kinds), "kinds")

	kindIDPath := filepath.Join(dir, "node_kind_id_gen.go")
	if err := writeNodeKindID(kindIDPath, kinds); err != nil {
		genpaths.Fatalf("write %s: %v", kindIDPath, err)
	}
	genpaths.Announce(kindIDPath, len(kinds), "kinds")
}
