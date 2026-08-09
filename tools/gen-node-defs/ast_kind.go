// AST reads that identify a kind: its registered Go kind name (parseGoKindName)
// and its wire:"data.*" tagged fields (parseDataFieldsFromAST).
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// goTypeExprStr converts an ast.Expr to a Go type string.
func goTypeExprStr(expr ast.Expr) (string, bool) {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name, true
	case *ast.ArrayType:
		elt, ok := goTypeExprStr(t.Elt)
		if !ok {
			return "", false
		}
		return "[]" + elt, true
	case *ast.MapType:
		k, ok1 := goTypeExprStr(t.Key)
		v, ok2 := goTypeExprStr(t.Value)
		if !ok1 || !ok2 {
			return "", false
		}
		return "map[" + k + "]" + v, true
	}
	return "", false
}

// goIdentRE matches a legal TS/Go identifier. goKind is emitted as an unquoted
// TS object key in node-defs.ts, so a non-identifier name (hyphen, space, leading
// digit) would produce invalid TS; validate it at parse time and fail loudly.
var goIdentRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// parseGoKindName extracts the first string argument to Register (nodes/wire's
// wire.Register, or the pre-decompose monolithic Wiring package's Register) in pkgDir.
func parseGoKindName(pkgDir string) (string, error) {
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return "", err
	}
	// RegisterBuilder is the SELF-CONSTRUCTION registration (build_args.go): a kind
	// that builds itself no longer calls wire.Register at all, and a generator that only
	// knew the old marker silently dropped it from NODE_DEFS — the editor then loses the
	// kind while the Go side works fine.
	markers := []string{`Wiring.RegisterBuilder("`}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(pkgDir, name))
		if err != nil {
			continue
		}
		s := string(data)
		for _, marker := range markers {
			_, rest, ok := strings.Cut(s, marker)
			if !ok {
				continue
			}
			name2, _, ok2 := strings.Cut(rest, `"`)
			if !ok2 {
				continue
			}
			if !goIdentRE.MatchString(name2) {
				fatalf("kind name %q from %s in %s is not a legal identifier (must match [A-Za-z_][A-Za-z0-9_]*); it is emitted as an unquoted TS object key", name2, marker, pkgDir)
			}
			return name2, nil
		}
	}
	return "", fmt.Errorf("RegisterBuilder call not found in %s", pkgDir)
}

// parseDataFieldsFromAST reads all .go files in pkgDir and returns data fields
// derived from struct fields tagged with wire:"data.*".
func parseDataFieldsFromAST(pkgDir string) ([]dataField, error) {
	fset := token.NewFileSet()
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return nil, err
	}
	var files []*ast.File
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		fullPath := filepath.Join(pkgDir, name)
		f, err := parser.ParseFile(fset, fullPath, nil, 0)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	var fields []dataField
	for _, file := range files {
		for _, decl := range file.Decls {
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
					if field.Tag == nil {
						continue
					}
					tag := strings.Trim(field.Tag.Value, "`")
					const prefix = `wire:"data.`
					_, after, ok := strings.Cut(tag, prefix)
					if !ok {
						continue
					}
					wireKey, _, ok2 := strings.Cut(after, `"`)
					if !ok2 {
						continue
					}
					var fname string
					if len(field.Names) > 0 {
						fname = field.Names[0].Name
					}
					typeStr, ok := goTypeExprStr(field.Type)
					if !ok {
						displayName := fname
						if displayName == "" {
							displayName = "<anonymous>"
						}
						return nil, fmt.Errorf("kind %q: wire:\"data.%s\" field %q has an unsupported/unstringifiable Go type %T", filepath.Base(pkgDir), wireKey, displayName, field.Type)
					}
					fields = append(fields, dataField{wireTag: wireKey, goType: typeStr, fieldName: fname})
				}
			}
		}
	}
	return fields, nil
}
