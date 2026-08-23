package main

//go:generate go run .

import (
	"github.com/dtauraso/wirefold/scripts/genpaths"
	"path/filepath"

	"github.com/dtauraso/wirefold/scripts/genpaths/params"
)

func main() {
	genpaths.SetName("Categories/Ring/NodeShape/gen")
	repoRoot, srcRoot := genpaths.Roots()

	dir := filepath.Join(srcRoot, "Ring", "NodeShape")
	goPath := filepath.Join(dir, "shading_params.go")
	shadingParams, err := params.ParseShadingParams(repoRoot, goPath)
	if err != nil {
		genpaths.Fatalf("parse shading params: %v", err)
	}
	tsPath := filepath.Join(dir, "shading-params.ts")
	if err := params.WriteShadingParams(tsPath, shadingParams, genpaths.Name(), "Categories/Ring/NodeShape/shading_params.go"); err != nil {
		genpaths.Fatalf("write %s: %v", tsPath, err)
	}
	genpaths.Announce(tsPath, len(shadingParams), "node shape constants")
}
