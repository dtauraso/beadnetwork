package buflayout

import (
	"fmt"
	"os"
)

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "gen-node-defs: "+format+"\n", args...)
	os.Exit(1)
}

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
			out = append(out, r-32)
		} else {
			out = append(out, r)
		}
	}
	return string(out)
}
