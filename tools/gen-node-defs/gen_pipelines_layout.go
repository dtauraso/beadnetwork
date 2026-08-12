package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dtauraso/wirefold/tools/gen-node-defs/buflayout"
	"github.com/dtauraso/wirefold/tools/gen-node-defs/inputlayout"
	"github.com/dtauraso/wirefold/tools/gen-node-defs/overlaygen"
	"github.com/dtauraso/wirefold/tools/gen-node-defs/params"
)

func generateCurveParams(repoRoot string) {
	curveParamsGoPath := filepath.Join(repoRoot, "nodes", "Wiring", "nodegeom", "curve_params.go")
	curveParams, err := params.ParseCurveParams(curveParamsGoPath)
	if err != nil {
		fatalf("parse curve params: %v", err)
	}
	curveParamsTsPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "schema", "curve-params.ts")
	if err := params.WriteCurveParams(curveParamsTsPath, curveParams); err != nil {
		fatalf("write %s: %v", curveParamsTsPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d constants)\n", curveParamsTsPath, len(curveParams))
}

func generateOverlayGen(repoRoot string) {
	messagesTSPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "messages.ts")
	overlayFlags, err := overlaygen.ParseOverlayFlags(messagesTSPath)
	if err != nil {
		fatalf("parse overlay flags: %v", err)
	}
	viewstateDir := filepath.Join(repoRoot, "nodes", "Wiring", "viewstate")
	overlayGenGoPath := filepath.Join(viewstateDir, "overlay_state.go")
	overlayTablesGoPath := filepath.Join(viewstateDir, "overlay_tables_gen.go")
	if err := overlaygen.WriteOverlayGen(overlayGenGoPath, overlayTablesGoPath, overlayFlags); err != nil {
		fatalf("write overlay gen: %v", err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d overlay flags)\n", overlayGenGoPath, len(overlayFlags))
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d overlay flags)\n", overlayTablesGoPath, len(overlayFlags))
}

func generateShadingParams(repoRoot string) {
	shadingParamsGoPath := filepath.Join(repoRoot, "nodes", "Wiring", "nodegeom", "shading_params.go")
	shadingParams, err := params.ParseShadingParams(repoRoot, shadingParamsGoPath)
	if err != nil {
		fatalf("parse shading params: %v", err)
	}
	shadingParamsTsPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "schema", "shading-params.ts")
	if err := params.WriteShadingParams(shadingParamsTsPath, shadingParams); err != nil {
		fatalf("write %s: %v", shadingParamsTsPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d constants)\n", shadingParamsTsPath, len(shadingParams))
}

func generateBufferLayout(repoRoot string) {
	bufferDir := filepath.Join(repoRoot, "Buffer", "bufschema")
	bufSchema, err := buflayout.ParseBufferLayoutDir(bufferDir)
	if err != nil {
		fatalf("parse buffer layout: %v", err)
	}
	bufLayoutGenGoPath := filepath.Join(repoRoot, "Buffer", "buffer_layout_gen.go")
	bufLayoutGenGoRowsPath := filepath.Join(repoRoot, "Buffer", "buffer_layout_gen_rows.go")
	bufLayoutGenGoRows2Path := filepath.Join(repoRoot, "Buffer", "buffer_layout_gen_rows2.go")
	bufLayoutGenGoSingletonsPath := filepath.Join(repoRoot, "Buffer", "buffer_layout_gen_singletons.go")
	if err := buflayout.WriteBufferLayoutGo(bufLayoutGenGoPath, bufLayoutGenGoRowsPath, bufLayoutGenGoRows2Path, bufLayoutGenGoSingletonsPath, bufSchema); err != nil {
		fatalf("write buffer layout go: %v", err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d blocks)\n", bufLayoutGenGoPath, len(bufSchema.Blocks))
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d blocks)\n", bufLayoutGenGoRowsPath, len(bufSchema.Blocks))
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d blocks)\n", bufLayoutGenGoRows2Path, len(bufSchema.Blocks))
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d blocks)\n", bufLayoutGenGoSingletonsPath, len(bufSchema.Blocks))

	schemaDir := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "schema")
	bufLayoutTSPath := filepath.Join(schemaDir, "buffer-layout.ts")
	bufLayoutTSRowsPath := filepath.Join(schemaDir, "buffer-layout-rows-gen.ts")
	bufLayoutTSRows2Path := filepath.Join(schemaDir, "buffer-layout-rows2-gen.ts")
	bufLayoutTSSingletonsPath := filepath.Join(schemaDir, "buffer-layout-singletons-gen.ts")
	if err := buflayout.WriteBufferLayoutTS(bufLayoutTSPath, bufLayoutTSRowsPath, bufLayoutTSRows2Path, bufLayoutTSSingletonsPath, bufSchema); err != nil {
		fatalf("write buffer layout ts: %v", err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d blocks)\n", bufLayoutTSPath, len(bufSchema.Blocks))
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d blocks)\n", bufLayoutTSRowsPath, len(bufSchema.Blocks))
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d blocks)\n", bufLayoutTSRows2Path, len(bufSchema.Blocks))
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d blocks)\n", bufLayoutTSSingletonsPath, len(bufSchema.Blocks))
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

func generateInputLayout(repoRoot string) {
	wiringGoDir := filepath.Join(repoRoot, "nodes", "Wiring", "inputcodec")
	inputFP, err := inputlayout.ParseInputLayoutFingerprintDir(wiringGoDir)
	if err != nil {
		fatalf("parse input layout fingerprint: %v", err)
	}
	inputLayoutTSPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "schema", "input-layout-gen.ts")
	if err := inputlayout.WriteInputLayout(inputLayoutTSPath, inputFP); err != nil {
		fatalf("write %s: %v", inputLayoutTSPath, err)
	}
	numConsts := 1 + len(inputFP.KindNames) + 4
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d constants)\n", inputLayoutTSPath, numConsts)
}
