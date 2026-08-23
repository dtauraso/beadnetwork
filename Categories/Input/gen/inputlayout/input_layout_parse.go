package inputlayout

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
)

type inputLayoutFingerprint struct {
	raw         string
	KindNames   []string
	kindValues  map[string]string
	updateKinds []string
}

var errFingerprintNotFound = fmt.Errorf("InputLayoutFingerprint const not found")

func ParseInputLayoutFingerprintDir(dir string) (*inputLayoutFingerprint, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	for _, p := range paths {
		fp, err := parseInputLayoutFingerprint(p)
		if errors.Is(err, errFingerprintNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		return fp, nil
	}
	return nil, fmt.Errorf("no file in %s declares InputLayoutFingerprint", dir)
}

func parseInputLayoutFingerprint(goPath string) (*inputLayoutFingerprint, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, goPath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var raw string
	found := false
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
			for i, name := range vs.Names {
				if name.Name != "InputLayoutFingerprint" {
					continue
				}
				if i >= len(vs.Values) {
					continue
				}
				lit, ok := vs.Values[i].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return nil, fmt.Errorf("InputLayoutFingerprint is not a string literal")
				}
				val, err := unquoteGoString(lit.Value)
				if err != nil {
					return nil, err
				}
				raw = val
				found = true
			}
		}
	}
	if !found {
		return nil, fmt.Errorf("%w in %s", errFingerprintNotFound, goPath)
	}

	fp := &inputLayoutFingerprint{raw: raw, kindValues: map[string]string{}}

	kindsTok := fpListToken(raw, "kinds=")
	if kindsTok == "" {
		return nil, fmt.Errorf("InputLayoutFingerprint is missing the kinds= token")
	}
	for _, pair := range strings.Split(kindsTok, ",") {
		nv := strings.SplitN(pair, ":", 2)
		if len(nv) != 2 {
			return nil, fmt.Errorf("InputLayoutFingerprint kinds= entry %q is not name:value", pair)
		}
		fp.KindNames = append(fp.KindNames, nv[0])
		fp.kindValues[nv[0]] = nv[1]
	}

	fp.updateKinds = fpList(raw, "updateKinds=")
	for _, e := range []struct {
		marker string
		list   []string
	}{
		{"updateKinds=", fp.updateKinds},
	} {
		if len(e.list) == 0 {
			return nil, fmt.Errorf("InputLayoutFingerprint is missing the %s token", e.marker)
		}
	}
	return fp, nil
}

func (f *inputLayoutFingerprint) List(marker string) []string {
	switch marker {
	case "updateKinds=":
		return f.updateKinds
	}
	return nil
}

func (f *inputLayoutFingerprint) KindValue(name string) string { return f.kindValues[name] }
