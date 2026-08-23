package main

//go:generate go run .

import (
	"github.com/dtauraso/wirefold/scripts/genpaths"
	"path/filepath"

	"github.com/dtauraso/wirefold/scripts/genpaths/params"
)

func main() {
	genpaths.SetName("Categories/Node/Poles/shading")
	repoRoot, srcRoot := genpaths.Roots()

	dir := filepath.Join(srcRoot, "Node", "Poles")
	goPath := filepath.Join(dir, "shading_params.go")
	shadingParams, err := params.ParseShadingParams(repoRoot, goPath)
	if err != nil {
		genpaths.Fatalf("parse shading params: %v", err)
	}
	tsPath := filepath.Join(dir, "shading-params.ts")
	if err := params.WriteShadingParams(tsPath, shadingParams, genpaths.Name(), "Categories/Node/Poles/shading_params.go"); err != nil {
		genpaths.Fatalf("write %s: %v", tsPath, err)
	}
	genpaths.Announce(tsPath, len(shadingParams), "pole ring constants")
}
