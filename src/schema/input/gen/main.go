package main

//go:generate go run .

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/scripts/genpaths"
	"github.com/dtauraso/wirefold/src/schema/input/gen/inputlayout"
)

func main() {
	genpaths.Name = "schema/input/gen"
	repoRoot, srcRoot := genpaths.Roots()

	wiringGoDir := filepath.Join(genpaths.NetworkDir(repoRoot), "Wiring", "inputcodec")
	inputFP, err := inputlayout.ParseInputLayoutFingerprintDir(wiringGoDir)
	if err != nil {
		genpaths.Fatalf("parse input layout fingerprint: %v", err)
	}
	outPath := filepath.Join(srcRoot, "schema", "input", "input-layout-gen.ts")
	if err := inputlayout.WriteInputLayout(outPath, inputFP); err != nil {
		genpaths.Fatalf("write %s: %v", outPath, err)
	}
	genpaths.Announce(outPath, 1+len(inputFP.KindNames)+4, "constants")
}
