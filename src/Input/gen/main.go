package main

//go:generate go run .

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/scripts/genpaths"
	"github.com/dtauraso/wirefold/src/Input/gen/inputlayout"
)

func main() {
	genpaths.Name = "Input/gen"
	_, srcRoot := genpaths.Roots()

	codecGoDir := filepath.Join(srcRoot, "Input", "inputcodec")
	inputFP, err := inputlayout.ParseInputLayoutFingerprintDir(codecGoDir)
	if err != nil {
		genpaths.Fatalf("parse input layout fingerprint: %v", err)
	}
	outPath := filepath.Join(srcRoot, "Input", "input-layout-gen.ts")
	if err := inputlayout.WriteInputLayout(outPath, inputFP); err != nil {
		genpaths.Fatalf("write %s: %v", outPath, err)
	}
	genpaths.Announce(outPath, 1+len(inputFP.KindNames)+4, "constants")

	goKindsPath := filepath.Join(srcRoot, "Input", "inputcodec", "event_kinds_gen.go")
	if err := inputlayout.WriteGoEventKinds(goKindsPath, inputFP); err != nil {
		genpaths.Fatalf("write %s: %v", goKindsPath, err)
	}
	genpaths.Announce(goKindsPath, 1, "event kinds")
}
