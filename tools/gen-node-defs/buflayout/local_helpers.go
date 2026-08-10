package buflayout

import (
	"fmt"
	"os"
)

// fatalf mirrors gen-node-defs' package-main fatalf: report and exit(1). Duplicated
// (not imported — package main can't be imported) because this package's build-time
// failures must abort the whole generator run exactly like every other pipeline's.
func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "gen-node-defs: "+format+"\n", args...)
	os.Exit(1)
}

// camelToScreamingSnake converts PascalCase/camelCase to SCREAMING_SNAKE_CASE.
// e.g. CurveParamBulgeFactor → CURVE_PARAM_BULGE_FACTOR
// Inserts '_' before an uppercase letter only when the PRECEDING rune was
// lowercase, so abbreviations like "BeadID" → "BEAD_ID" and "CX" → "CX"
// stay intact (consecutive uppercase letters are NOT split). Duplicated from
// params.go's identical helper for the same reason as fatalf above.
func camelToScreamingSnake(s string) string {
	runes := []rune(s)
	var out []rune
	for i, r := range runes {
		if i > 0 && r >= 'A' && r <= 'Z' {
			prev := runes[i-1]
			if prev >= 'a' && prev <= 'z' {
				out = append(out, '_')
			}
		}
		if r >= 'a' && r <= 'z' {
			out = append(out, r-32) // to upper
		} else {
			out = append(out, r)
		}
	}
	return string(out)
}
