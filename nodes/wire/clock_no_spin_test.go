package wire

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// clock_no_spin_test.go — source guard: RealClock.SleepCycle's blocking select must carry
// NO default case. select WITH default is non-blocking, so a caller looping around it
// spins and burns a core; select WITHOUT default parks the goroutine on both channels'
// wait queues at zero CPU. This distinction is invisible to any behavioural test (a
// default-cased version would still eventually return once ctx or the tick fires), so it
// is pinned here by inspecting the actual source, the same "assert by source guard" shape
// PLAN.md calls for.
func TestSleepCycleSelectHasNoDefaultCase(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "clock.go", nil, 0)
	if err != nil {
		t.Fatalf("parse clock.go: %v", err)
	}

	var fn *ast.FuncDecl
	for _, decl := range file.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Name.Name != "SleepCycle" {
			continue
		}
		fn = fd
		break
	}
	if fn == nil {
		t.Fatal("SleepCycle function not found in clock.go")
	}

	found := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectStmt)
		if !ok {
			return true
		}
		found = true
		for _, stmt := range sel.Body.List {
			cc, ok := stmt.(*ast.CommClause)
			if !ok {
				continue
			}
			if cc.Comm == nil { // nil Comm marks the default: case
				t.Fatal("SleepCycle's select carries a default: case — this makes the wait NON-BLOCKING, which spins a core instead of parking the goroutine at zero CPU. Remove the default case.")
			}
		}
		return true
	})
	if !found {
		t.Fatal("SleepCycle contains no select statement — expected exactly one blocking select over the tick channel and ctx.Done()")
	}
}
