package kindscan

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func parseEmbeddedPorts(nodesDir, pkgDir string, visited map[string]bool) ([]Port, error) {
	fset := token.NewFileSet()
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return nil, err
	}
	var ports []Port
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(pkgDir, name), nil, 0)
		if err != nil {
			return nil, err
		}
		for _, decl := range f.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}
				for _, field := range structType.Fields.List {
					if len(field.Names) != 0 {
						continue
					}
					sel, ok := field.Type.(*ast.SelectorExpr)
					if !ok {
						continue
					}
					pkgIdent, ok := sel.X.(*ast.Ident)
					if !ok {
						continue
					}
					embedDir := filepath.Join(nodesDir, strings.ToLower(pkgIdent.Name))
					if visited[embedDir] {
						continue
					}
					visited[embedDir] = true
					if _, statErr := os.Stat(embedDir); statErr != nil {
						continue
					}
					embedded, err := parsePortsFromAST(embedDir)
					if err != nil {
						return nil, err
					}
					ports = append(ports, embedded...)
					more, err := parseEmbeddedPorts(nodesDir, embedDir, visited)
					if err != nil {
						return nil, err
					}
					ports = append(ports, more...)
				}
			}
		}
	}
	return ports, nil
}
