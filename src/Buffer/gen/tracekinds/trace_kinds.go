package tracekinds

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func ParseTraceKinds(traceDir string) ([]string, error) {
	entries, err := os.ReadDir(traceDir)
	if err != nil {
		return nil, err
	}
	var kinds []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") { // path-resolution-ok: a generator walking its own source tree, not a scene path
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

	if len(kinds) == 0 {
		return nil, fmt.Errorf("no Kind* constants found in %s", traceDir)
	}
	return kinds, nil
}

func ParseBreadcrumbLabels(traceDir string) ([]string, error) {
	entries, err := os.ReadDir(traceDir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") { // path-resolution-ok: a generator walking its own source tree, not a scene path
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
