package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// RelDir is where a scene keeps the camera's primitives, relative to the scene
// root. It matches scenepaths.CameraDirPath, which is what Go resolves against.
const RelDir = "view/camera"

// writeCameraPaths emits one file per camera primitive, each holding that
// primitive's path relative to the scene root, FROM Go's own File* constants.
//
// The renderer read these path files while Go built the same string from those
// constants — agreement by convention, not construction. Generating the
// renderer's half from Go's constants makes them one declaration, so the
// overlay failure (Go writing .json while the renderer read .bin) cannot happen
// here either.
func writeCameraPaths(cameraDir string, names []string) error {
	pathsDir := filepath.Join(cameraDir, "paths")
	if err := os.MkdirAll(pathsDir, 0o755); err != nil {
		return err
	}
	keep := map[string]bool{}
	for _, name := range names {
		keep[name] = true
		rel := RelDir + "/" + name
		if err := os.WriteFile(filepath.Join(pathsDir, name), []byte(rel), 0o644); err != nil {
			return err
		}
	}

	// A primitive that leaves Go's constants loses its path file, or the
	// renderer keeps reading a value nothing writes.
	entries, err := os.ReadDir(pathsDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() || keep[e.Name()] { // path-resolution-ok: the generator's own output directory
			continue
		}
		if err := os.Remove(filepath.Join(pathsDir, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

// cameraFileConsts reads the File* constants that name the camera's primitives.
func cameraFileConsts(path string) ([]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.CONST {
			continue
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, id := range vs.Names {
				if !strings.HasPrefix(id.Name, "File") || i >= len(vs.Values) {
					continue
				}
				lit, ok := vs.Values[i].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}
				v, err := strconv.Unquote(lit.Value)
				if err != nil {
					return nil, err
				}
				out = append(out, v)
			}
		}
	}
	return out, nil
}
