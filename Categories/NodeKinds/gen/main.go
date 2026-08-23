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

	kindsDir := filepath.Join(srcRoot, "NodeKinds")
	if err := writeKindOwnPorts(kindsDir, kinds); err != nil {
		genpaths.Fatalf("write per-kind ports: %v", err)
	}
	genpaths.Announce(kindsDir, len(kinds), "per-kind port tables")
}
