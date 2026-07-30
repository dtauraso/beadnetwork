// constexpr.go — a small constant-expression evaluator over Go source, used by
// parseShadingParams (and available to any future ShadingParam*/CurveParam*
// const) so a ShadingParam const can be written as a real Go expression —
// including one that references a const in ANOTHER package, like
// `wire.BeadTorusOuterR / (1 + ShadingParamBeadRingTubeRatio)` — instead of a
// hand-computed literal that is a second copy of the same fact and free to
// drift (docs/bead-lattice.md).
//
// go/types + go/packages would be the textbook way to get this exactly right,
// but this repo has no external dependencies (no go.sum) and gen-node-defs
// only ever needs to resolve a handful of untyped numeric consts reachable
// from one file — full package type-checking is a lot of machinery for that.
// This evaluator instead walks the AST directly and leans on go/constant for
// the arithmetic itself, which is the part that is genuinely easy to get
// subtly wrong (float vs. exact-rational untyped-constant semantics). It
// supports exactly what a ShadingParam* value expression needs: literals,
// parens, unary/binary ops, local identifiers, and qualified identifiers
// (pkg.Name) resolved by locating the imported package's directory under this
// module and parsing its const declarations the same way. Anything outside
// that (function calls, iota, non-const identifiers) is a hard error, not a
// silent fallback — a ShadingParam* value should always be one of these.
package main

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// constDecl is one named const's value expression plus the file it came from
// (needed later to resolve THAT expression's own identifiers/imports).
type constDecl struct {
	expr ast.Expr
	file *ast.File
}

// constEnv caches parsed packages (by directory) and this module's import
// path prefix, across however many identifiers a single evaluation needs to
// chase.
type constEnv struct {
	fset         *token.FileSet
	repoRoot     string
	modulePrefix string // e.g. "github.com/dtauraso/wirefold", from go.mod's module line
	pkgsByDir    map[string]map[string]constDecl
}

// newConstEnv builds an evaluator rooted at repoRoot, reading the module path
// out of go.mod so qualified identifiers (wire.Foo) can be mapped from their
// import path to a directory on disk.
func newConstEnv(fset *token.FileSet, repoRoot string) (*constEnv, error) {
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
	return &constEnv{
		fset:         fset,
		repoRoot:     repoRoot,
		modulePrefix: modulePrefix,
		pkgsByDir:    map[string]map[string]constDecl{},
	}, nil
}

// loadPkgConsts parses every non-test .go file directly in dir and returns a
// name -> constDecl map of its top-level const declarations. Cached per dir
// since a package's consts are often chased more than once (e.g. LocalStepR
// depends on BeadStepR, both looked up while resolving BeadTorusOuterR).
func (env *constEnv) loadPkgConsts(dir string) (map[string]constDecl, error) {
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
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
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
			// Track the most recent non-empty Values list so a const block
			// that omits Values on later lines (the iota-continuation
			// shorthand) still resolves — none of today's ShadingParam
			// dependencies use it, but a bare ast.GenDecl walk should not
			// silently mis-index if one ever does.
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

// importDir resolves the import path referenced by alias in file to a
// directory on disk, by stripping this module's prefix and joining the rest
// onto repoRoot. Only intra-module imports are supported (a ShadingParam*
// expression has no reason to reach into the standard library or a
// dependency), and gen-node-defs has none of those to resolve anyway.
func (env *constEnv) importDir(file *ast.File, alias string) (string, error) {
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

// eval evaluates expr, which was parsed as part of file living in dir, to a
// go/constant.Value — the same representation cmd/compile uses for untyped
// constant arithmetic, so a chain of untyped-float divisions round-trips to
// exactly the float64 Go itself would produce, with no lossy string hop.
func (env *constEnv) eval(dir string, file *ast.File, expr ast.Expr) (constant.Value, error) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		v := constant.MakeFromLiteral(e.Value, e.Kind, 0)
		if v.Kind() == constant.Unknown {
			return nil, fmt.Errorf("bad literal %q", e.Value)
		}
		return v, nil

	case *ast.ParenExpr:
		return env.eval(dir, file, e.X)

	case *ast.UnaryExpr:
		x, err := env.eval(dir, file, e.X)
		if err != nil {
			return nil, err
		}
		return constant.UnaryOp(e.Op, x, 0), nil

	case *ast.BinaryExpr:
		x, err := env.eval(dir, file, e.X)
		if err != nil {
			return nil, err
		}
		y, err := env.eval(dir, file, e.Y)
		if err != nil {
			return nil, err
		}
		return constant.BinaryOp(x, e.Op, y), nil

	case *ast.Ident:
		consts, err := env.loadPkgConsts(dir)
		if err != nil {
			return nil, err
		}
		decl, ok := consts[e.Name]
		if !ok {
			return nil, fmt.Errorf("identifier %q is not a const in %s", e.Name, dir)
		}
		return env.eval(dir, decl.file, decl.expr)

	case *ast.SelectorExpr:
		pkgIdent, ok := e.X.(*ast.Ident)
		if !ok {
			return nil, fmt.Errorf("unsupported selector base %T", e.X)
		}
		otherDir, err := env.importDir(file, pkgIdent.Name)
		if err != nil {
			return nil, err
		}
		consts, err := env.loadPkgConsts(otherDir)
		if err != nil {
			return nil, err
		}
		decl, ok := consts[e.Sel.Name]
		if !ok {
			return nil, fmt.Errorf("identifier %q is not a const in %s", e.Sel.Name, otherDir)
		}
		return env.eval(otherDir, decl.file, decl.expr)

	default:
		return nil, fmt.Errorf("unsupported const expression %T (%s)", expr, exprString(expr))
	}
}

// exprString renders expr back to source text for error messages, without
// pulling in go/printer for what is only ever a short diagnostic.
func exprString(expr ast.Expr) string {
	fset := token.NewFileSet()
	var buf strings.Builder
	ast.Fprint(&buf, fset, expr, nil)
	return buf.String()
}
