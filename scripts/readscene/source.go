package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type jsonFunc struct {
	Name string `json:"name"`
	File string `json:"file"`
	Line int    `json:"line"`
}

func receiver(d *ast.FuncDecl) string {
	if d.Recv == nil || len(d.Recv.List) == 0 {
		return ""
	}
	t := d.Recv.List[0].Type
	if s, ok := t.(*ast.StarExpr); ok {
		t = s.X
	}
	if id, ok := t.(*ast.Ident); ok {
		return id.Name + "."
	}
	return ""
}

func dumpSource(dir string) int {
	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil || len(files) == 0 {
		fmt.Fprintf(os.Stderr, "readscene -source: no Go files in %q\n", dir)
		return 1
	}
	sort.Strings(files)
	fset := token.NewFileSet()
	all := []jsonFunc{}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		src, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "readscene -source: %v\n", err)
			return 1
		}
		f, err := parser.ParseFile(fset, path, src, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "readscene -source: %v\n", err)
			return 1
		}
		for _, decl := range f.Decls {
			d, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			all = append(all, jsonFunc{Name: receiver(d) + d.Name.Name, File: filepath.ToSlash(path), Line: fset.Position(d.Pos()).Line})
		}
	}
	out := struct {
		Functions []jsonFunc `json:"functions"`
	}{all}
	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "readscene -source: %v\n", err)
		return 1
	}
	return 0
}
