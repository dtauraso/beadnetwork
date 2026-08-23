package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "gen-node-defs: "+format+"\n", args...)
	os.Exit(1)
}

func CollectKinds(nodesDir string) []KindEntry {
	entries, err := os.ReadDir(nodesDir)
	if err != nil {
		fatalf("readdir nodes: %v", err)
	}

	var kinds []KindEntry
	seenGoKind := map[string]string{}
	for _, e := range entries {
		if !e.IsDir() { // path-resolution-ok: a candidate kind package, not a scene path
			continue
		}
		pkgDir := filepath.Join(nodesDir, e.Name())
		if !hasRegister(pkgDir) {
			continue
		}
		ports := parsePortsFromSpec(pkgDir)
		dataFields, err := parseDataFieldsFromAST(pkgDir)
		if err != nil {
			fatalf("parse data fields %s: %v", e.Name(), err)
		}
		goKind, err := parseGoKindName(pkgDir)
		if err != nil {
			fatalf("parse go kind name %s: %v", e.Name(), err)
		}

		if prev, dup := seenGoKind[goKind]; dup {
			fatalf("duplicate kind name %q registered by both %q and %q", goKind, prev, e.Name())
		}
		seenGoKind[goKind] = e.Name()
		view, accentOverrides, edgeKindOverrides, optionalPorts, err := parseSpecMD(pkgDir)
		if err != nil {

			fatalf("kind %q registers a Go runtime but its SPEC.md View section is missing/broken: %v", e.Name(), err)
		}
		if view.Kind == "" {
			fatalf("kind %q registers a Go runtime but its SPEC.md ## View has an empty view.kind", e.Name())
		}

		checkPortRequests(e.Name(), pkgDir, ports)

		if len(ports) == 0 {
			fatalf("kind %q registers a Go runtime but its SPEC.md has no ## Ports rows: the table is the only declaration of a kind's inputs and outputs, so a kind without one binds nothing", e.Name())
		}

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

func hasRegister(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {

		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") { // path-resolution-ok: a package directory listing, not a scene path
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}

		if bytes.Contains(data, []byte("BuilderFor(")) {
			return true
		}
	}
	return false
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
