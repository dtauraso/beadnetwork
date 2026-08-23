package main

//go:generate go run .

import (
	"github.com/dtauraso/wirefold/scripts/genpaths"
	"path/filepath"

	"github.com/dtauraso/wirefold/Categories/NodeKinds/gen/kindscan"
	"github.com/dtauraso/wirefold/Categories/NodeKinds/gen/nodedefs"
)

func main() {
	genpaths.SetName("Categories/NodeKinds/gen")
	repoRoot, srcRoot := genpaths.Roots()
	kinds := kindscan.Kinds(repoRoot)

	importsPath := filepath.Join(srcRoot, "NodeKinds", "kinds_gen.go")
	if err := kindscan.WriteKindImports(importsPath, kindscan.KindsPkg(repoRoot), kinds); err != nil {
		genpaths.Fatalf("write %s: %v", importsPath, err)
	}
	genpaths.Announce(importsPath, len(kinds), "kinds")

	defsPath := filepath.Join(srcRoot, "NodeKinds", "node-defs.ts")
	if err := nodedefs.WriteNodeDefs(defsPath, kinds); err != nil {
		genpaths.Fatalf("write %s: %v", defsPath, err)
	}
	genpaths.Announce(defsPath, len(kinds), "entries")

	dimsPath := filepath.Join(srcRoot, "Node", "geom_node_dims_gen.go")
	if err := writeNodeDims(dimsPath, kinds); err != nil {
		genpaths.Fatalf("write %s: %v", dimsPath, err)
	}
	genpaths.Announce(dimsPath, len(kinds), "kinds")

	portsPath := filepath.Join(srcRoot, "NodeKinds", "portwiring", "kind_ports_gen.go")
	if err := writeKindPorts(portsPath, kinds); err != nil {
		genpaths.Fatalf("write %s: %v", portsPath, err)
	}
	genpaths.Announce(portsPath, len(kinds), "kinds")

	kindsDir := filepath.Join(srcRoot, "NodeKinds")
	if err := writeKindOwnPorts(kindsDir, kinds); err != nil {
		genpaths.Fatalf("write per-kind ports: %v", err)
	}
	genpaths.Announce(kindsDir, len(kinds), "per-kind port tables")

	kindIDPath := filepath.Join(srcRoot, "Node", "node_kind_id_gen.go")
	if err := writeNodeKindID(kindIDPath, kinds); err != nil {
		genpaths.Fatalf("write %s: %v", kindIDPath, err)
	}
	genpaths.Announce(kindIDPath, len(kinds), "kinds")
}
