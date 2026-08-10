// gen-node-defs walks nodes/*/ and emits src/schema/node-defs.ts.
// Port names and directions are derived from Go channel-typed struct fields.
// View metadata and per-port accent overrides are read from SPEC.md.
// Run: go run ../../tools/gen-node-defs (from tools/topology-vscode/)
//
// This is the SINGLE entry point for every generator pipeline in this package
// (node-defs, wire-defs, trace-kinds, node-dims/kind-id, curve/shading params,
// overlay-gen, buffer-layout). tools/buffer-schema/check-generated.sh derives its guarded-file
// list from this one invocation's "wrote <path>" stderr lines — do not split
// this into multiple binaries; add new pipelines as new files/functions in this
// package and call them from main() below.
//
// main() itself holds only the entry point and the call sequence; kind
// collection, kind-id assignment, and the shared KindEntry/Port/etc. types
// live in tools/gen-node-defs/kindscan, and repo-root resolution in
// repo_root.go.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/dtauraso/wirefold/tools/gen-node-defs/buflayout"
	"github.com/dtauraso/wirefold/tools/gen-node-defs/kindscan"
)

func main() {
	// Generator is invoked from tools/topology-vscode/ via npm run gen:node-defs.
	// Resolve repo root relative to this file's location at compile time is not
	// possible; instead, walk up from cwd until we find a "nodes/" directory.
	cwd, err := os.Getwd()
	if err != nil {
		fatalf("getwd: %v", err)
	}
	repoRoot := findRepoRoot(cwd)
	if repoRoot == "" {
		fatalf("could not locate repo root (no nodes/ dir found from %s)", cwd)
	}

	nodesDir := filepath.Join(repoRoot, "nodes")
	kinds := kindscan.CollectKinds(nodesDir)

	// Finding B: KindId is a stable, assigned-once fact per kind (from SPEC.md's
	// View "kindId" field), NOT a sort-derived index — adding a kind must never
	// renumber an existing one. Resolve each kind's id, auto-assigning (and
	// writing back into its SPEC.md) the next free id for any kind that doesn't
	// have one yet, so a new kind is stable from the moment it's first generated.
	kindscan.AssignKindIDs(kinds, nodesDir)

	// Sort alphabetically by Go kind name (PascalCase spec kind) for stable,
	// human-readable emission ORDER only — this sort has no bearing on the id
	// VALUE, which comes from kindID above.
	sort.Slice(kinds, func(i, j int) bool {
		return kinds[i].GoKind < kinds[j].GoKind
	})

	// kinds_generated.go — blank imports that make each node package's init() (and thus
	// its wire.Register call) run. Folded in from the former standalone tools/gen-kind-imports
	// so ONE registration scan feeds every kind-derived output; a separate binary could
	// silently diverge from this one (audit finding). check-generated.sh guards it via the
	// "wrote" line below, so no dedicated guard is needed.
	kindImportsPath := filepath.Join(repoRoot, "kinds_generated.go")
	if err := kindscan.WriteKindImports(kindImportsPath, kinds); err != nil {
		fatalf("write %s: %v", kindImportsPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d kinds)\n", kindImportsPath, len(kinds))

	outPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "schema", "node-defs.ts")
	if err := writeNodeDefs(outPath, kinds); err != nil {
		fatalf("write %s: %v", outPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d entries)\n", outPath, len(kinds))

	loaderPath := filepath.Join(repoRoot, "nodes", "Wiring", "topo_spec.go")
	wireProps, err := parseWirePropsFromFile(loaderPath)
	if err != nil {
		fatalf("parse wire props from topo_spec.go: %v", err)
	}
	wireDefsPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "schema", "wire-defs.ts")
	if err := writeWireDefs(wireDefsPath, wireProps); err != nil {
		fatalf("write %s: %v", wireDefsPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d entries)\n", wireDefsPath, len(wireProps))

	traceDir := filepath.Join(repoRoot, "Trace")
	traceKinds, err := parseTraceKinds(traceDir)
	if err != nil {
		fatalf("parse trace kinds: %v", err)
	}
	breadcrumbLabels, err := parseBreadcrumbLabels(traceDir)
	if err != nil {
		fatalf("parse breadcrumb labels: %v", err)
	}
	traceKindsPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "schema", "trace-kinds.ts")
	if err := writeTraceKinds(traceKindsPath, traceKinds, breadcrumbLabels); err != nil {
		fatalf("write %s: %v", traceKindsPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d kinds)\n", traceKindsPath, len(traceKinds))

	nodeDimsGoPath := filepath.Join(repoRoot, "nodes", "Wiring", "node_dims_gen.go")
	if err := writeNodeDims(nodeDimsGoPath, kinds); err != nil {
		fatalf("write %s: %v", nodeDimsGoPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d kinds)\n", nodeDimsGoPath, len(kinds))

	nodeKindIDGoPath := filepath.Join(repoRoot, "Buffer", "node_kind_id_gen.go")
	if err := writeNodeKindID(nodeKindIDGoPath, kinds); err != nil {
		fatalf("write %s: %v", nodeKindIDGoPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d kinds)\n", nodeKindIDGoPath, len(kinds))

	curveParamsGoPath := filepath.Join(repoRoot, "nodes", "Wiring", "curve_params.go")
	curveParams, err := parseCurveParams(curveParamsGoPath)
	if err != nil {
		fatalf("parse curve params: %v", err)
	}
	curveParamsTsPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "schema", "curve-params.ts")
	if err := writeCurveParams(curveParamsTsPath, curveParams); err != nil {
		fatalf("write %s: %v", curveParamsTsPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d constants)\n", curveParamsTsPath, len(curveParams))

	messagesTSPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "messages.ts")
	overlayFlags, err := parseOverlayFlags(messagesTSPath)
	if err != nil {
		fatalf("parse overlay flags: %v", err)
	}
	overlayGenGoPath := filepath.Join(repoRoot, "nodes", "Wiring", "overlay_gen.go")
	if err := writeOverlayGen(overlayGenGoPath, overlayFlags); err != nil {
		fatalf("write %s: %v", overlayGenGoPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d overlay flags)\n", overlayGenGoPath, len(overlayFlags))

	shadingParamsGoPath := filepath.Join(repoRoot, "nodes", "Wiring", "shading_params.go")
	shadingParams, err := parseShadingParams(repoRoot, shadingParamsGoPath)
	if err != nil {
		fatalf("parse shading params: %v", err)
	}
	shadingParamsTsPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "schema", "shading-params.ts")
	if err := writeShadingParams(shadingParamsTsPath, shadingParams); err != nil {
		fatalf("write %s: %v", shadingParamsTsPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d constants)\n", shadingParamsTsPath, len(shadingParams))

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

	// Scan the package for whichever file declares InputLayoutFingerprint rather than
	// naming one (it moved from input_codec.go to input_fingerprint.go when that file was
	// split by job — memory/feedback_guards_hardcoding_single_file_break_on_split.md — and
	// again from nodes/Wiring to nodes/Wiring/inputcodec when the TS->Go decode cluster
	// was lifted into its own package).
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

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "gen-node-defs: "+format+"\n", args...)
	os.Exit(1)
}
