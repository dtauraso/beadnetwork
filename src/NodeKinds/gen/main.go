package main

//go:generate go run .

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/src/NodeKinds/gen/kindscan"
	"github.com/dtauraso/wirefold/src/NodeKinds/gen/nodedefs"
)

func main() {
	genName = "NodeKinds/gen"
	repoRoot, srcRoot := roots()
	kinds := kindscan.Kinds(repoRoot)

	importsPath := filepath.Join(srcRoot, "NodeKinds", "kinds_gen.go")
	if err := kindscan.WriteKindImports(importsPath, kindscan.KindsPkg(repoRoot), kinds); err != nil {
		fatalf("write %s: %v", importsPath, err)
	}
	announce(importsPath, len(kinds), "kinds")

	defsPath := filepath.Join(srcRoot, "NodeKinds", "node-defs.ts")
	if err := nodedefs.WriteNodeDefs(defsPath, kinds); err != nil {
		fatalf("write %s: %v", defsPath, err)
	}
	announce(defsPath, len(kinds), "entries")
}
