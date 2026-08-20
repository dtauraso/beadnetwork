package constexpr

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type constDecl struct {
	expr ast.Expr
	file *ast.File
}

type Env struct {
	fset         *token.FileSet
	repoRoot     string
	modulePrefix string
	pkgsByDir    map[string]map[string]constDecl
}

func NewEnv(fset *token.FileSet, repoRoot string) (*Env, error) {
	modBytes, err := os.ReadFile(filepath.Join(repoRoot, "go.mod"))
	if err != nil {
		return nil, fmt.Errorf("read go.mod: %w", err)
	}
	var modulePrefix string
	for _, line := range strings.Split(string(modBytes), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			modulePrefix = strings.TrimSpace(strings.TrimPrefix(line, "module"))
			break
		}
	}
	if modulePrefix == "" {
		return nil, fmt.Errorf("go.mod has no module line")
	}
	return &Env{
		fset:         fset,
		repoRoot:     repoRoot,
		modulePrefix: modulePrefix,
		pkgsByDir:    map[string]map[string]constDecl{},
	}, nil
}

func (env *Env) loadPkgConsts(dir string) (map[string]constDecl, error) {
	if cached, ok := env.pkgsByDir[dir]; ok {
		return cached, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}
	consts := map[string]constDecl{}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") { // path-resolution-ok: a generator walking its own source tree, not a scene path
			continue
		}
		f, err := parser.ParseFile(env.fset, filepath.Join(dir, name), nil, 0)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", filepath.Join(dir, name), err)
		}
		for _, decl := range f.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.CONST {
				continue
			}

			var lastValues []ast.Expr
			for _, spec := range genDecl.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				if len(vs.Values) > 0 {
					lastValues = vs.Values
				}
				for i, ident := range vs.Names {
					if i >= len(lastValues) {
						continue
					}
					consts[ident.Name] = constDecl{expr: lastValues[i], file: f}
				}
			}
		}
	}
	env.pkgsByDir[dir] = consts
	return consts, nil
}

func (env *Env) importDir(file *ast.File, alias string) (string, error) {
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		name := path[strings.LastIndex(path, "/")+1:]
		if imp.Name != nil {
			name = imp.Name.Name
		}
		if name != alias {
			continue
		}
		if !strings.HasPrefix(path, env.modulePrefix) {
			return "", fmt.Errorf("import %q is outside module %q, not supported by the constant evaluator", path, env.modulePrefix)
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(path, env.modulePrefix), "/")
		return filepath.Join(env.repoRoot, filepath.FromSlash(rel)), nil
	}
	return "", fmt.Errorf("no import aliased %q in %s", alias, env.fset.Position(file.Package).Filename)
}
