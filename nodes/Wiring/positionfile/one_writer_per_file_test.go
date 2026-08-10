package positionfile

// one_writer_per_file_test.go — pins fact #2 of the one-file-per-goroutine split
//: position.json has its ON-DISK NAME literal spelled out in exactly ONE place in
// production (non-test) source — the single path-building function for that file
// (FilePath). The sibling assertions for camera.json/overlays.json/sphere.json live in
// nodes/Wiring/scenepaths/one_writer_per_file_test.go, since those three literals live
// there. Every writer, loader and persister-arming call site in nodes/Wiring reaches the
// file only through this package's FilePath; nothing else is allowed to spell the filename
// itself.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// countLiteralOccurrences counts how many non-test .go files in this package's directory
// contain the exact literal (e.g. `"position.json"`), and how many total occurrences there
// are — a second definition of the same filename anywhere in production source pushes this
// above 1.
func countLiteralOccurrences(t *testing.T, literal string) int {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	count := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || filepath.Ext(name) != ".go" || strings.HasSuffix(name, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		count += strings.Count(string(raw), literal)
	}
	return count
}

// TestEachSplitFileNameIsSpelledExactlyOnce asserts each new file's on-disk name literal
// appears in production source exactly once — proof there is exactly one place that can
// ever construct a path to it, and therefore exactly one writer.
func TestEachSplitFileNameIsSpelledExactlyOnce(t *testing.T) {
	names := []string{
		`"position.json"`,
	}
	for _, name := range names {
		got := countLiteralOccurrences(t, name)
		if got != 1 {
			t.Fatalf("filename literal %s appears %d time(s) in production source, want exactly 1 "+
				"(one path-building function) — a second spelling is how a second writer sneaks in", name, got)
		}
	}
}
