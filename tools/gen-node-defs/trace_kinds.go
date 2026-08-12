// trace-kinds.ts pipeline: scans Trace/ for Kind* string constants and the BreadcrumbLabels
// var (this file). Emitting TRACE_EVENT_KINDS/BREADCRUMB_LABELS from the parsed result is
// the sibling trace_kinds_write.go.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// parseTraceKinds scans EVERY non-test *.go file under the Trace/ dir and returns
// the string values of all Kind* constants (e.g. "recv", "fire", "send", "slot").
// Scanning the whole dir (not just Trace.go) means a Kind* const declared in any
// sibling file under Trace/ is still picked up — the single-file-path guard-blindness
// class (memory: feedback_guards_hardcoding_single_file_break_on_split).
func parseTraceKinds(traceDir string) ([]string, error) {
	entries, err := os.ReadDir(traceDir)
	if err != nil {
		return nil, err
	}
	var kinds []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, filepath.Join(traceDir, name), nil, 0)
		if err != nil {
			return nil, err
		}
		for _, decl := range f.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.CONST {
				continue
			}
			for _, spec := range genDecl.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for i, nm := range vs.Names {
					if !strings.HasPrefix(nm.Name, "Kind") {
						continue
					}
					if i >= len(vs.Values) {
						continue
					}
					lit, ok := vs.Values[i].(*ast.BasicLit)
					if !ok || lit.Kind != token.STRING {
						continue
					}
					kinds = append(kinds, strings.Trim(lit.Value, `"`))
				}
			}
		}
	}
	// No sort: os.ReadDir yields filenames in lexical order and go/ast preserves
	// in-file declaration order, so the emitted slice is deterministic across runs.
	if len(kinds) == 0 {
		return nil, fmt.Errorf("no Kind* constants found in %s", traceDir)
	}
	return kinds, nil
}

// parseBreadcrumbLabels scans EVERY non-test *.go file under traceDir for the
// `BreadcrumbLabels = []string{...}` var and returns its string literals in order —
// the mirror of parseTraceKinds for the 13-value BreadcrumbLabel* enum
// (Buffer/layout.go's bufLayoutEvent.Label column, Kind==KindBreadcrumb rows only).
func parseBreadcrumbLabels(traceDir string) ([]string, error) {
	entries, err := os.ReadDir(traceDir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, filepath.Join(traceDir, name), nil, 0)
		if err != nil {
			return nil, err
		}
		for _, decl := range f.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.VAR {
				continue
			}
			for _, spec := range genDecl.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for i, nm := range vs.Names {
					if nm.Name != "BreadcrumbLabels" {
						continue
					}
					if i >= len(vs.Values) {
						continue
					}
					cl, ok := vs.Values[i].(*ast.CompositeLit)
					if !ok {
						continue
					}
					var labels []string
					for _, elt := range cl.Elts {
						lit, ok := elt.(*ast.BasicLit)
						if !ok || lit.Kind != token.STRING {
							continue
						}
						labels = append(labels, strings.Trim(lit.Value, `"`))
					}
					return labels, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("BreadcrumbLabels var not found under %s", traceDir)
}
