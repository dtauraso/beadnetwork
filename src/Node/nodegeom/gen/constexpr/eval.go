package constexpr

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"strings"
)

func (env *Env) Eval(dir string, file *ast.File, expr ast.Expr) (constant.Value, error) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		v := constant.MakeFromLiteral(e.Value, e.Kind, 0)
		if v.Kind() == constant.Unknown {
			return nil, fmt.Errorf("bad literal %q", e.Value)
		}
		return v, nil

	case *ast.ParenExpr:
		return env.Eval(dir, file, e.X)

	case *ast.UnaryExpr:
		x, err := env.Eval(dir, file, e.X)
		if err != nil {
			return nil, err
		}
		return constant.UnaryOp(e.Op, x, 0), nil

	case *ast.BinaryExpr:
		x, err := env.Eval(dir, file, e.X)
		if err != nil {
			return nil, err
		}
		y, err := env.Eval(dir, file, e.Y)
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
		return env.Eval(dir, decl.file, decl.expr)

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
		return env.Eval(otherDir, decl.file, decl.expr)

	default:
		return nil, fmt.Errorf("unsupported const expression %T (%s)", expr, exprString(expr))
	}
}

func exprString(expr ast.Expr) string {
	fset := token.NewFileSet()
	var buf strings.Builder
	ast.Fprint(&buf, fset, expr, nil)
	return buf.String()
}
