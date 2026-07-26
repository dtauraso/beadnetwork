// gen-node-defs walks nodes/*/ and emits src/schema/node-defs.ts.
// Port names and directions are derived from Go channel-typed struct fields.
// View metadata and per-port accent overrides are read from SPEC.md.
// Run: go run ../../tools/gen-node-defs (from tools/topology-vscode/)
//
// This is the SINGLE entry point for every generator pipeline in this package
// (node-defs, wire-defs, trace-kinds, node-dims/kind-id, curve/shading params,
// overlay-gen, buffer-layout). tools/check-generated.sh derives its guarded-file
// list from this one invocation's "wrote <path>" stderr lines — do not split
// this into multiple binaries; add new pipelines as new files/functions in this
// package and call them from main() below.
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// wireProp represents one wire:"prop,..." tagged field on specEdge.
type wireProp struct {
	jsonName string // from json:"..." tag
	tsType   string // from tsType:... in wire tag
	required bool   // false if "optional", true if "required"
}

// port represents one channel-typed struct field.
type port struct {
	id        string // Go field name
	direction string // "in" or "out"
	accent    string // optional hex color from SPEC.md
	edgeKind  string // optional edge kind from SPEC.md Ports table EdgeKind column
	isMulti   bool   // true when the Go type is Wiring.OutMulti
	optional  bool   // true when SPEC.md Ports table marks this port Optional=yes
}

// dataField represents a wire:"data.*" tagged struct field.
type dataField struct {
	wireTag   string // full tag value after wire:"data." prefix, e.g. "init" or "state"
	goType    string // Go type string, e.g. "[]int", "int", "string"
	fieldName string // Go struct field name (used for wire:"data.state" key derivation)
}

// viewDef holds the SPEC.md ## View section fields.
type viewDef struct {
	kind     string
	kindID   string // raw "kindId" Field/Value cell from SPEC.md's View table; "" if unassigned
	bg       string
	border   string
	text     string
	minWidth string
	// NodeTypeDef-compatible fields (used by schema/node-types consumers).
	role   string
	shape  string
	fill   string
	stroke string
	width  string
	height string
}

// kindEntry is one node kind to emit.
type kindEntry struct {
	kind        string // RF/view kind name (camelCase, from SPEC.md)
	goKind      string // Go/topology kind name (PascalCase, from Wiring.Register)
	dir         string // node package directory name under nodes/ (import path segment)
	kindID      uint8  // stable buffer KindId — assigned once, never renumbered by sort order
	view        viewDef
	ports       []port
	dataFields  []dataField
	defaultData string // raw JSON from SPEC.md ## Default data, or ""
}

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
	entries, err := os.ReadDir(nodesDir)
	if err != nil {
		fatalf("readdir nodes: %v", err)
	}

	var kinds []kindEntry
	seenGoKind := map[string]string{} // goKind → dir name that registered it
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pkgDir := filepath.Join(nodesDir, e.Name())
		if !hasRegister(pkgDir) {
			continue
		}
		ports, err := parsePortsFromAST(pkgDir)
		if err != nil {
			fatalf("parse ports %s: %v", e.Name(), err)
		}
		// Merge in ports declared on embedded structs from other local nodes/
		// packages (e.g. gatecommon.GateNode) — AST parsing only looks at
		// pkgDir's own files, so promoted fields from an embedded sibling
		// package are otherwise invisible.
		embedded, err := parseEmbeddedPorts(nodesDir, pkgDir, map[string]bool{})
		if err != nil {
			fatalf("parse embedded ports %s: %v", e.Name(), err)
		}
		ports = append(ports, embedded...)
		// Fallback: if AST found no ports (e.g. all ports are in an embedded struct
		// from another package), read them from the SPEC.md Ports table.
		if len(ports) == 0 {
			ports = parsePortsFromSpec(pkgDir)
		}
		dataFields, err := parseDataFieldsFromAST(pkgDir)
		if err != nil {
			fatalf("parse data fields %s: %v", e.Name(), err)
		}
		goKind, err := parseGoKindName(pkgDir)
		if err != nil {
			fatalf("parse go kind name %s: %v", e.Name(), err)
		}
		// A duplicate goKind across dirs produces a silent last-wins duplicate TS
		// object key in node-defs.ts; reject it here naming both dirs.
		if prev, dup := seenGoKind[goKind]; dup {
			fatalf("duplicate kind name %q registered by both %q and %q", goKind, prev, e.Name())
		}
		seenGoKind[goKind] = e.Name()
		view, accentOverrides, edgeKindOverrides, optionalPorts, specPortNames, err := parseSpecMD(pkgDir)
		if err != nil {
			// This dir has a Wiring.Register (a real node package), so a missing or
			// broken SPEC.md View section is a half-landed kind — fail loudly rather
			// than silently dropping the kind from all generated files.
			fatalf("kind %q registers a Go runtime but its SPEC.md View section is missing/broken: %v", e.Name(), err)
		}
		if view.kind == "" {
			fatalf("kind %q registers a Go runtime but its SPEC.md ## View has an empty view.kind", e.Name())
		}
		// Finding C: every Ports-table "Name" must resolve to a real AST-derived port
		// id. A typo here previously dropped its accent/edgeKind/optional override
		// silently; now it's a loud generator error.
		astPortIDs := map[string]bool{}
		for _, p := range ports {
			astPortIDs[p.id] = true
		}
		for name := range specPortNames {
			if !astPortIDs[name] {
				fatalf("kind %q: SPEC.md Ports table Name %q does not match any Go channel-typed port (got: %v)", e.Name(), name, sortedKeys(astPortIDs))
			}
		}
		// Apply accent, edgeKind overrides, and optional flags from SPEC.md Ports table.
		for i, p := range ports {
			if a, ok := accentOverrides[p.id]; ok && a != "" {
				ports[i].accent = a
			}
			if ek, ok := edgeKindOverrides[p.id]; ok && ek != "" {
				ports[i].edgeKind = ek
			}
			if optionalPorts[p.id] {
				ports[i].optional = true
			}
		}
		defaultData := parseDefaultData(pkgDir)
		kinds = append(kinds, kindEntry{kind: view.kind, goKind: goKind, dir: e.Name(), view: view, ports: ports, dataFields: dataFields, defaultData: defaultData})
	}

	// Finding B: KindId is a stable, assigned-once fact per kind (from SPEC.md's
	// View "kindId" field), NOT a sort-derived index — adding a kind must never
	// renumber an existing one. Resolve each kind's id, auto-assigning (and
	// writing back into its SPEC.md) the next free id for any kind that doesn't
	// have one yet, so a new kind is stable from the moment it's first generated.
	assignKindIDs(kinds, nodesDir)

	// Sort alphabetically by Go kind name (PascalCase spec kind) for stable,
	// human-readable emission ORDER only — this sort has no bearing on the id
	// VALUE, which comes from kindID above.
	sort.Slice(kinds, func(i, j int) bool {
		return kinds[i].goKind < kinds[j].goKind
	})

	// kinds_generated.go — blank imports that make each node package's init() (and thus
	// its Wiring.Register) run. Folded in from the former standalone tools/gen-kind-imports
	// so ONE registration scan feeds every kind-derived output; a separate binary could
	// silently diverge from this one (audit finding). check-generated.sh guards it via the
	// "wrote" line below, so no dedicated guard is needed.
	kindImportsPath := filepath.Join(repoRoot, "kinds_generated.go")
	if err := writeKindImports(kindImportsPath, kinds); err != nil {
		fatalf("write %s: %v", kindImportsPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d kinds)\n", kindImportsPath, len(kinds))

	outPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "schema", "node-defs.ts")
	if err := writeNodeDefs(outPath, kinds); err != nil {
		fatalf("write %s: %v", outPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d entries)\n", outPath, len(kinds))

	loaderPath := filepath.Join(repoRoot, "nodes", "Wiring", "loader.go")
	wireProps, err := parseWirePropsFromFile(loaderPath)
	if err != nil {
		fatalf("parse wire props from loader.go: %v", err)
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
	shadingParams, err := parseShadingParams(shadingParamsGoPath)
	if err != nil {
		fatalf("parse shading params: %v", err)
	}
	shadingParamsTsPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "schema", "shading-params.ts")
	if err := writeShadingParams(shadingParamsTsPath, shadingParams); err != nil {
		fatalf("write %s: %v", shadingParamsTsPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d constants)\n", shadingParamsTsPath, len(shadingParams))

	bufLayoutGoPath := filepath.Join(repoRoot, "Buffer", "layout.go")
	bufSchema, err := parseBufferLayout(bufLayoutGoPath)
	if err != nil {
		fatalf("parse buffer layout: %v", err)
	}
	bufLayoutGenGoPath := filepath.Join(repoRoot, "Buffer", "buffer_layout_gen.go")
	if err := writeBufferLayoutGo(bufLayoutGenGoPath, bufSchema); err != nil {
		fatalf("write %s: %v", bufLayoutGenGoPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d blocks)\n", bufLayoutGenGoPath, len(bufSchema.blocks))

	bufLayoutTSPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "schema", "buffer-layout.ts")
	if err := writeBufferLayoutTS(bufLayoutTSPath, bufSchema); err != nil {
		fatalf("write %s: %v", bufLayoutTSPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d blocks)\n", bufLayoutTSPath, len(bufSchema.blocks))

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

	inputCodecGoPath := filepath.Join(repoRoot, "nodes", "Wiring", "input_codec.go")
	inputFP, err := parseInputLayoutFingerprint(inputCodecGoPath)
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

// findRepoRoot walks up from dir until it finds a directory containing "nodes/".
func findRepoRoot(dir string) string {
	for {
		if _, err := os.Stat(filepath.Join(dir, "nodes")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// hasRegister reports whether any .go file in dir contains "Wiring.Register(".
func hasRegister(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		if bytes.Contains(data, []byte("Wiring.Register(")) {
			return true
		}
	}
	return false
}

// sortedKeys returns the keys of a string-keyed bool set in sorted order, for
// deterministic error messages.
func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// assignKindIDs resolves each kind's stable buffer KindId from its SPEC.md
// View "kindId" field, in place on kinds. A kind whose SPEC.md has no kindId
// yet is auto-assigned max(existing ids)+1 and that assignment is written
// back into its SPEC.md immediately, so the id is stable from here on —
// regenerating never reassigns it again. Fails loudly on a duplicate id or
// an id colliding with/exceeding the KindIDUnknown sentinel (0xFF).
func assignKindIDs(kinds []kindEntry, nodesDir string) {
	usedBy := map[uint8]string{} // id -> goKind that claimed it
	maxID := -1
	var unassigned []int // indices into kinds needing auto-assignment

	for i := range kinds {
		raw := strings.TrimSpace(kinds[i].view.kindID)
		if raw == "" {
			unassigned = append(unassigned, i)
			continue
		}
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 || n > 254 {
			fatalf("kind %q: SPEC.md kindId %q must be an integer in [0,254]", kinds[i].goKind, raw)
		}
		id := uint8(n)
		if prev, dup := usedBy[id]; dup {
			fatalf("kind %q and kind %q both claim KindId %d in their SPEC.md — ids must be unique and assigned once", prev, kinds[i].goKind, id)
		}
		usedBy[id] = kinds[i].goKind
		kinds[i].kindID = id
		if n > maxID {
			maxID = n
		}
	}

	for _, i := range unassigned {
		maxID++
		if maxID > 254 {
			fatalf("kind %q: no free KindId below the KindIDUnknown sentinel (0xFF)", kinds[i].goKind)
		}
		id := uint8(maxID)
		usedBy[id] = kinds[i].goKind
		kinds[i].kindID = id
		if err := writeBackKindID(nodesDir, kinds[i].dir, id); err != nil {
			fatalf("kind %q: auto-assigned KindId %d but failed to write it back into SPEC.md: %v", kinds[i].goKind, id, err)
		}
		fmt.Fprintf(os.Stderr, "gen-node-defs: auto-assigned KindId %d to new kind %q (written to nodes/%s/SPEC.md)\n", id, kinds[i].goKind, kinds[i].dir)
	}
}

// writeBackKindID inserts a "| kindId | N |" row directly above the existing
// "| kind | ... |" row in nodes/<dir>/SPEC.md's View table, so a newly
// auto-assigned id is persisted and stable on the next regeneration.
func writeBackKindID(nodesDir, dir string, id uint8) error {
	specPath := filepath.Join(nodesDir, dir, "SPEC.md")
	data, err := os.ReadFile(specPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	for i, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "| kind |") || strings.HasPrefix(trimmed, "|kind|") {
			row := fmt.Sprintf("| kindId | %d |", id)
			lines = append(lines[:i], append([]string{row}, lines[i:]...)...)
			return os.WriteFile(specPath, []byte(strings.Join(lines, "\n")), 0644)
		}
	}
	return fmt.Errorf("no '| kind |' row found in View table to anchor kindId insertion")
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "gen-node-defs: "+format+"\n", args...)
	os.Exit(1)
}
