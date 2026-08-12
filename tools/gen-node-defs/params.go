// curve-params.ts / shading-params.ts emitters: mechanically mirror
// CurveParam*/ShadingParam* Go constants into TS exports.
//
// Split into params_curve.go (curveParam parse+emit) and params_shading.go
// (shadingParam parse+emit); this file keeps the naming-convention helper
// shared by both.
package main

// camelToScreamingSnake converts PascalCase/camelCase to SCREAMING_SNAKE_CASE.
// e.g. CurveParamBulgeFactor → CURVE_PARAM_BULGE_FACTOR
// Inserts '_' before an uppercase letter only when the PRECEDING rune was
// lowercase, so abbreviations like "BeadID" → "BEAD_ID" and "CX" → "CX"
// stay intact (consecutive uppercase letters are NOT split).
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
