package Wiring

// scene_path_safety_test.go — verifies writeQuantOffset rejects an unsafe id rather than
// escaping the tree root (see quant_offset_persist.go). The former port-anchor write sink
// (scene_anchor_persist.go) is gone — docs/bead-model/channels-not-ports.md, a port has no file
// of its own any more.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/quantoffset"
)

func TestWriteQuantOffsetRejectsTraversalID(t *testing.T) {
	root := t.TempDir()
	err := writeQuantOffset(root, "../../evil", quantoffset.QuantizedOffset{}, geom.Polar{}, 0)
	if err == nil {
		t.Fatal("expected error for traversal node id, got nil")
	}
	if _, statErr := os.Stat(filepath.Join(root, "..", "..", "evil", "meta.json")); statErr == nil {
		t.Fatal("traversal write unexpectedly created a file outside the tree root")
	}
}
