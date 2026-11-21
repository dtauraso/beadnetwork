package main

//go:generate go run .

import (
	"path/filepath"

	"github.com/dtauraso/beadnetwork/scripts/genpaths"
	"github.com/dtauraso/beadnetwork/scripts/genpaths/params"
)

func main() {
	genpaths.SetName("Categories/Node/Geom/shading")
	repoRoot, srcRoot := genpaths.Roots()

	dir := filepath.Join(srcRoot, "Node", "Geom")
	goPath := filepath.Join(dir, "shading_params.go")
	shParams, err := params.ParseShadingParams(repoRoot, goPath)
	if err != nil {
		genpaths.Fatalf("parse shading params: %v", err)
	}
	tsPath := filepath.Join(dir, "shading-params.ts")
	if err := params.WriteShadingParams(tsPath, shParams, genpaths.Name(), "Categories/Node/Geom/shading_params.go"); err != nil {
		genpaths.Fatalf("write %s: %v", tsPath, err)
	}
	genpaths.Announce(tsPath, len(shParams), "node constants")
}
