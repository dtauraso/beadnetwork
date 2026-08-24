package main

//go:generate go run .

import (
	"path/filepath"

	"github.com/dtauraso/beadnetwork/scripts/genpaths"
	"github.com/dtauraso/beadnetwork/scripts/genpaths/params"
)

func main() {
	genpaths.SetName("Categories/Scene/Handholds/shading")
	repoRoot, srcRoot := genpaths.Roots()

	dir := filepath.Join(srcRoot, "Scene", "Handholds")
	goPath := filepath.Join(dir, "shading_params.go")
	shadingParams, err := params.ParseShadingParams(repoRoot, goPath)
	if err != nil {
		genpaths.Fatalf("parse shading params: %v", err)
	}
	tsPath := filepath.Join(dir, "shading-params.ts")
	if err := params.WriteShadingParams(tsPath, shadingParams, genpaths.Name(), "Categories/Scene/Handholds/shading_params.go"); err != nil {
		genpaths.Fatalf("write %s: %v", tsPath, err)
	}
	genpaths.Announce(tsPath, len(shadingParams), "handhold constants")
}
