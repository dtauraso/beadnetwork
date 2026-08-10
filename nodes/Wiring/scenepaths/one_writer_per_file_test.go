package scenepaths

// one_writer_per_file_test.go — pins fact #2 of the one-file-per-goroutine split for the
// three scene-level files this package resolves: camera.json, overlays.json, sphere.json
// each has its ON-DISK NAME literal spelled out in exactly ONE place in production
// (non-test) source — the single path-building function for that file
// (CameraFilePath / OverlaysFilePath / SphereFilePath). Every writer, loader and
// persister-arming call site in nodes/Wiring reaches the file only through that one
// function; nothing else is allowed to spell the filename itself. The sibling assertion for
// position.json (owned by nodes/Wiring's node_mover.go, via nodes/Wiring/positionfile)
// lives in nodes/Wiring/positionfile/one_writer_per_file_test.go.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// countLiteralOccurrences counts how many non-test .go files in this package's directory
// contain the exact literal (e.g. `"camera.json"`), and how many total occurrences there
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
		`"camera.json"`,
		`"overlays.json"`,
		`"sphere.json"`,
	}
	for _, name := range names {
		got := countLiteralOccurrences(t, name)
		if got != 1 {
			t.Fatalf("filename literal %s appears %d time(s) in production source, want exactly 1 "+
				"(one path-building function) — a second spelling is how a second writer sneaks in", name, got)
		}
	}
}
