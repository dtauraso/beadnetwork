package Wiring

// scene_path_safety_test.go — verifies safeTreePathComponent rejects path-traversal
// values and that writeQuantOffset rejects an unsafe id rather than escaping the tree
// root (see quant_offset_persist.go). The former port-anchor write sink
// (scene_anchor_persist.go) is gone — docs/channels-not-ports.md, a port has no file
// of its own any more.

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSafeTreePathComponent(t *testing.T) {
	cases := []struct {
		s    string
		want bool
	}{
		{"", false},
		{".", false},
		{"..", false},
		{"../x", false},
		{"a/b", false},
		{"/abs", false},
		{`a\b`, false},
		{"n1", true},
		{"portOut", true},
		{"node_2", true},
	}
	for _, c := range cases {
		if got := safeTreePathComponent(c.s); got != c.want {
			t.Errorf("safeTreePathComponent(%q) = %v, want %v", c.s, got, c.want)
		}
	}
}

func TestWriteQuantOffsetRejectsTraversalID(t *testing.T) {
	root := t.TempDir()
	err := writeQuantOffset(root, "../../evil", quantizedOffset{}, polar{}, 0)
	if err == nil {
		t.Fatal("expected error for traversal node id, got nil")
	}
	if _, statErr := os.Stat(filepath.Join(root, "..", "..", "evil", "meta.json")); statErr == nil {
		t.Fatal("traversal write unexpectedly created a file outside the tree root")
	}
}
