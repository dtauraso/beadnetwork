// Pipeline phase functions that emit layout/geometry-adjacent generated files: curve/shading
// params, the overlay-state Go file, the buffer layout Go+TS pair, frame tags, and the TS->Go
// input-layout fingerprint mirror. Split out of main.go by concern — main.go keeps only the
// entry point and the call sequence (see its header comment).
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dtauraso/wirefold/tools/gen-node-defs/buflayout"
)

func generateCurveParams(repoRoot string) {
	curveParamsGoPath := filepath.Join(repoRoot, "nodes", "Wiring", "nodegeom", "curve_params.go")
	curveParams, err := parseCurveParams(curveParamsGoPath)
	if err != nil {
		fatalf("parse curve params: %v", err)
	}
	curveParamsTsPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "schema", "curve-params.ts")
	if err := writeCurveParams(curveParamsTsPath, curveParams); err != nil {
		fatalf("write %s: %v", curveParamsTsPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d constants)\n", curveParamsTsPath, len(curveParams))
}

func generateOverlayGen(repoRoot string) {
	messagesTSPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "messages.ts")
	overlayFlags, err := parseOverlayFlags(messagesTSPath)
	if err != nil {
		fatalf("parse overlay flags: %v", err)
	}
	overlayGenGoPath := filepath.Join(repoRoot, "nodes", "Wiring", "viewstate", "overlay_state.go")
	if err := writeOverlayGen(overlayGenGoPath, overlayFlags); err != nil {
		fatalf("write %s: %v", overlayGenGoPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d overlay flags)\n", overlayGenGoPath, len(overlayFlags))
}

func generateShadingParams(repoRoot string) {
	shadingParamsGoPath := filepath.Join(repoRoot, "nodes", "Wiring", "nodegeom", "shading_params.go")
	shadingParams, err := parseShadingParams(repoRoot, shadingParamsGoPath)
	if err != nil {
		fatalf("parse shading params: %v", err)
	}
	shadingParamsTsPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "schema", "shading-params.ts")
	if err := writeShadingParams(shadingParamsTsPath, shadingParams); err != nil {
		fatalf("write %s: %v", shadingParamsTsPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d constants)\n", shadingParamsTsPath, len(shadingParams))
}

func generateBufferLayout(repoRoot string) {
	bufferDir := filepath.Join(repoRoot, "Buffer")
	bufSchema, err := buflayout.ParseBufferLayoutDir(bufferDir)
	if err != nil {
		fatalf("parse buffer layout: %v", err)
	}
	bufLayoutGenGoPath := filepath.Join(repoRoot, "Buffer", "buffer_layout_gen.go")
	if err := buflayout.WriteBufferLayoutGo(bufLayoutGenGoPath, bufSchema); err != nil {
		fatalf("write %s: %v", bufLayoutGenGoPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d blocks)\n", bufLayoutGenGoPath, len(bufSchema.Blocks))

	bufLayoutTSPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "schema", "buffer-layout.ts")
	if err := buflayout.WriteBufferLayoutTS(bufLayoutTSPath, bufSchema); err != nil {
		fatalf("write %s: %v", bufLayoutTSPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d blocks)\n", bufLayoutTSPath, len(bufSchema.Blocks))
}

func generateFrameTags(repoRoot string) {
	frameTagsGoPath := filepath.Join(repoRoot, "Buffer", "frame_tags.go")
	frameTagsHeader, frameTagConsts, err := parseFrameTags(frameTagsGoPath)
	if err != nil {
		fatalf("parse frame tags: %v", err)
	}
	frameTagsTSPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "schema", "frame-tags.ts")
	if err := writeFrameTags(frameTagsTSPath, frameTagsHeader, frameTagConsts); err != nil {
		fatalf("write %s: %v", frameTagsTSPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d constants)\n", frameTagsTSPath, len(frameTagConsts))
}

// generateInputLayout scans the package for whichever file declares InputLayoutFingerprint
// rather than naming one (it moved from input_codec.go to input_fingerprint.go when that
// file was split by job — memory/feedback_guards_hardcoding_single_file_break_on_split.md —
// and again from nodes/Wiring to nodes/Wiring/inputcodec when the TS->Go decode cluster was
// lifted into its own package).
func generateInputLayout(repoRoot string) {
	wiringGoDir := filepath.Join(repoRoot, "nodes", "Wiring", "inputcodec")
	inputFP, err := parseInputLayoutFingerprintDir(wiringGoDir)
	if err != nil {
		fatalf("parse input layout fingerprint: %v", err)
	}
	inputLayoutTSPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "schema", "input-layout-gen.ts")
	if err := writeInputLayout(inputLayoutTSPath, inputFP); err != nil {
		fatalf("write %s: %v", inputLayoutTSPath, err)
	}
	numConsts := 1 /* fingerprint */ + len(inputFP.kindNames) + 4 /* enum arrays */
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d constants)\n", inputLayoutTSPath, numConsts)
}
