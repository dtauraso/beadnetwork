package main

//go:generate go run .

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/scripts/genpaths"
	"github.com/dtauraso/wirefold/scripts/kindscan"
	"github.com/dtauraso/wirefold/src/NodeKinds/gen/nodedefs"
)

func main() {
	genpaths.Name = "NodeKinds/gen"
	repoRoot, srcRoot := genpaths.Roots()
	kinds := genpaths.Kinds(repoRoot)

	importsPath := filepath.Join(repoRoot, "kinds_generated.go")
	if err := kindscan.WriteKindImports(importsPath, genpaths.KindsPkg(repoRoot), kinds); err != nil {
		genpaths.Fatalf("write %s: %v", importsPath, err)
	}
	genpaths.Announce(importsPath, len(kinds), "kinds")

	defsPath := filepath.Join(srcRoot, "NodeKinds", "node-defs.ts")
	if err := nodedefs.WriteNodeDefs(defsPath, kinds); err != nil {
		genpaths.Fatalf("write %s: %v", defsPath, err)
	}
	genpaths.Announce(defsPath, len(kinds), "entries")
}
