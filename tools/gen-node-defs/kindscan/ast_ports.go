package kindscan

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func parsePortsFromAST(pkgDir string) ([]Port, error) {
	fset := token.NewFileSet()
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return nil, err
	}
	pkgs := map[string][]*ast.File{}
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
		pkgName := f.Name.Name
		pkgs[pkgName] = append(pkgs[pkgName], f)
	}
	var ports []Port

	pkgNames := make([]string, 0, len(pkgs))
	for name := range pkgs {
		pkgNames = append(pkgNames, name)
	}
	sort.Strings(pkgNames)
	for _, pkgName := range pkgNames {
		files := pkgs[pkgName]
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
						dir, ok := chanDirection(field.Type)
						if !ok {
							continue
						}

						if field.Tag != nil {
							tag := strings.Trim(field.Tag.Value, "`")
							if strings.Contains(tag, `wire:"data.`) {
								continue
							}
						}

						multi := dir == "outMulti"
						outDir := dir
						if multi {
							outDir = "out"
						}
						for _, name := range field.Names {
							ports = append(ports, Port{ID: name.Name, Direction: outDir, IsMulti: multi})
						}
					}
				}
			}
		}
	}
	return ports, nil
}

func chanDirection(expr ast.Expr) (string, bool) {
	// here so a kind's ports can mix inport.In with outport.Out/outport.Broadcast.
	isWirePkg := func(pkg *ast.Ident) bool {
		return pkg.Name == "Wiring" || pkg.Name == "wire" || pkg.Name == "outport" || pkg.Name == "inport"
	}

	if star, ok := expr.(*ast.StarExpr); ok {
		if sel, ok := star.X.(*ast.SelectorExpr); ok {
			if pkg, ok := sel.X.(*ast.Ident); ok && isWirePkg(pkg) {
				switch sel.Sel.Name {
				case "In":
					return "in", true
				case "Out":
					return "out", true
				}
			}
		}
		return "", false
	}

	if sel, ok := expr.(*ast.SelectorExpr); ok {
		if pkg, ok := sel.X.(*ast.Ident); ok && isWirePkg(pkg) && sel.Sel.Name == "Broadcast" {
			return "outMulti", true
		}

		if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "Wiring" && sel.Sel.Name == "DrivenOut" {
			return "out", true
		}
	}
	return "", false
}
