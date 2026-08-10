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

// parsePortsFromAST reads all .go files in pkgDir and returns ports derived
// from channel-typed struct fields. Fields with wire:"data.*" tags are skipped.
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
	// Iterate package names in sorted order so the emitted port order is
	// deterministic even when a dir contains two package names (map iteration
	// order is otherwise random and would flip-flop check-generated).
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
						// Skip wire:"data.*" fields.
						if field.Tag != nil {
							tag := strings.Trim(field.Tag.Value, "`")
							if strings.Contains(tag, `wire:"data.`) {
								continue
							}
						}
						// Get field name(s).
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

// parseEmbeddedPorts scans pkgDir's struct declarations for anonymous
// (embedded) fields whose type is a selector into another local nodes/<pkg>
// package (e.g. gatecommon.GateNode), and returns the channel-typed ports
// declared on that embedded package's own structs (recursively, guarded by
// visited to avoid cycles). This lets a wrapper kind's SPEC.md-independent
// AST port discovery still pick up promoted fields from a shared embedded
// struct package.
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
						continue // not embedded
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
						continue // not a local nodes/ package (e.g. Wiring itself)
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

// chanDirection returns ("in", true) for *Wiring.In/*wire.In, ("out", true) for
// *Wiring.Out/*wire.Out/Wiring.DrivenOut, ("outMulti", true) for
// Wiring.Broadcast/wire.Broadcast, and ("", false) for anything else. Both package idents
// are recognized: In/Out/Broadcast moved from nodes/Wiring to the leaf nodes/wire package
// (task/wiring-decompose), but this parser has to keep parsing pre-move source shapes too
// (older SPEC.md fixtures, generator tests). Wiring.DrivenOut (nodes/Wiring/driven_out.go)
// is a bare (non-pointer) selector, like Broadcast — a node kind's DriveHeld-driven output
// port (BuildArgs.DriveOut) still counts as an "out" port for SPEC.md/NODE_DEFS purposes;
// it is a different WRITE-SIDE ownership shape (docs/investigations/interior-stream-framing.md), not a
// different port direction.
func chanDirection(expr ast.Expr) (string, bool) {
	isWirePkg := func(pkg *ast.Ident) bool { return pkg.Name == "Wiring" || pkg.Name == "wire" }
	// *Wiring.In / *wire.In or *Wiring.Out / *wire.Out — pointer to selector
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
	// Wiring.Broadcast / wire.Broadcast — bare selector (type alias, no pointer): a
	// broadcast output port (one logical output emitting the same value onto N
	// independent wires).
	if sel, ok := expr.(*ast.SelectorExpr); ok {
		if pkg, ok := sel.X.(*ast.Ident); ok && isWirePkg(pkg) && sel.Sel.Name == "Broadcast" {
			return "outMulti", true
		}
		// Wiring.DrivenOut — same bare-selector shape, "nodes/Wiring" only (it is
		// defined there, not in nodes/wire): a DriveHeld-driven single output port.
		if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "Wiring" && sel.Sel.Name == "DrivenOut" {
			return "out", true
		}
	}
	return "", false
}
