package kindscan

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// fatalf prints a gen-node-defs error and exits — mirrors main.go's fatalf.
// Kept as its own small copy rather than a cross-package call: collectKinds
// is meant to fail loudly on a half-landed kind, and that behavior belongs to
// this package, not to a callback threaded in from main.
func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "gen-node-defs: "+format+"\n", args...)
	os.Exit(1)
}

// CollectKinds walks nodesDir and returns one KindEntry per node package that
// registers a Go runtime (hasRegister), with ports/dataFields/view resolved
// from its AST and SPEC.md. Order is directory-read order; the caller sorts
// and assigns stable KindIds afterward.
func CollectKinds(nodesDir string) []KindEntry {
	entries, err := os.ReadDir(nodesDir)
	if err != nil {
		fatalf("readdir nodes: %v", err)
	}

	var kinds []KindEntry
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
			// This dir has a wire.Register call (a real node package), so a missing or
			// broken SPEC.md View section is a half-landed kind — fail loudly rather
			// than silently dropping the kind from all generated files.
			fatalf("kind %q registers a Go runtime but its SPEC.md View section is missing/broken: %v", e.Name(), err)
		}
		if view.Kind == "" {
			fatalf("kind %q registers a Go runtime but its SPEC.md ## View has an empty view.kind", e.Name())
		}
		// Finding C: every Ports-table "Name" must resolve to a real AST-derived port
		// id. A typo here previously dropped its accent/edgeKind/optional override
		// silently; now it's a loud generator error.
		astPortIDs := map[string]bool{}
		for _, p := range ports {
			astPortIDs[p.ID] = true
		}
		for name := range specPortNames {
			if !astPortIDs[name] {
				fatalf("kind %q: SPEC.md Ports table Name %q does not match any Go channel-typed port (got: %v)", e.Name(), name, sortedKeys(astPortIDs))
			}
		}
		// Apply accent, edgeKind overrides, and optional flags from SPEC.md Ports table.
		for i, p := range ports {
			if a, ok := accentOverrides[p.ID]; ok && a != "" {
				ports[i].Accent = a
			}
			if ek, ok := edgeKindOverrides[p.ID]; ok && ek != "" {
				ports[i].EdgeKind = ek
			}
			if optionalPorts[p.ID] {
				ports[i].Optional = true
			}
		}
		defaultData := parseDefaultData(pkgDir)
		kinds = append(kinds, KindEntry{Kind: view.Kind, GoKind: goKind, Dir: e.Name(), View: view, Ports: ports, DataFields: dataFields, DefaultData: defaultData})
	}
	return kinds
}

// hasRegister reports whether any .go file in dir registers a node kind, which is now
// exactly one marker: the self-construction registration in build_args.go. The older
// empty-struct registration and the pre-decompose monolithic form are both retired.
func hasRegister(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		// PRODUCTION files only, matching parseGoKindName's scope: nodes/Wiring's own
		// tests call wire.Register to register throwaway fixture kinds (aimed_ports_test.go
		// etc.) — counting those would wrongly mark nodes/Wiring itself as a registered
		// node kind now that those calls are package-qualified (wire.Register, not the old
		// bare unqualified Register()).
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		// Register moved from nodes/Wiring to the leaf nodes/wire package
		// (task/wiring-decompose); node packages now call wire.Register.
		if bytes.Contains(data, []byte("Wiring.RegisterBuilder(")) {
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
