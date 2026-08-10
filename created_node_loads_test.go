package main

// created_node_loads_test.go — THE TREE A CREATE WRITES MUST LOAD.
//
// A node palette create ends the run and the host respawns against the changed tree, so a
// node written with the wrong thing in it does not fail at the drop: it fails at the RELOAD,
// as a process that exits during load. On screen that reads as the editor FREEZING —
// navigation is Go-owned, so no camera means no zoom, pan or rotate — with nothing on screen
// saying the created node was the cause. That is exactly how it was reported.
//
// It lives in the ROOT package, not in Wiring's own tests, because it needs the kind
// REGISTRY: Registry is filled by each kind package's init() through kinds_generated.go's
// blank imports, and Wiring cannot import those (they import Wiring). A version of this test
// inside Wiring sees an empty registry and cannot resolve a port at all.
//
// It drives the real writers and the real loader over a real copy of the pair tree, which is
// the only shape of check that would have noticed
// (memory/feedback_headless_repro_verifies_persistence.md).

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	T "github.com/dtauraso/wirefold/Trace"
	Wiring "github.com/dtauraso/wirefold/nodes/Wiring"
	"github.com/dtauraso/wirefold/nodes/Wiring/countspersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/edgefile"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

// copyTreeForCreate copies a topology tree into t.TempDir() so the test can write into it
// without touching the repo's own.
func copyTreeForCreate(t *testing.T, src string) string {
	t.Helper()
	dst := t.TempDir()
	err := filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(src, p)
		if relErr != nil {
			return relErr
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, readErr := os.ReadFile(p)
		if readErr != nil {
			return readErr
		}
		return os.WriteFile(target, b, 0o644)
	})
	if err != nil {
		t.Fatalf("copy %s: %v", src, err)
	}
	return dst
}

// firstPort is the same rule CreateNode resolves an edge's handles by: a kind's FIRST
// declared port in that direction.
func firstPort(t *testing.T, kind string, dir portwiring.PortDir) string {
	t.Helper()
	b, ok := Wiring.Registry[kind]
	if !ok {
		t.Fatalf("kind %q is not registered — kinds_generated.go may be stale", kind)
	}
	for _, p := range b.Ports {
		if p.Dir == dir {
			return p.Name
		}
	}
	t.Fatalf("kind %q declares no port in that direction", kind)
	return ""
}

// nextNodeID is one past the largest id in the tree — the same rule CreateNode allocates by,
// so the test's node lands where a real one would.
func nextNodeID(t *testing.T, root string) string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, "nodes"))
	if err != nil {
		t.Fatalf("read nodes dir: %v", err)
	}
	best := 0
	for _, e := range entries {
		if n, convErr := strconv.Atoi(e.Name()); convErr == nil && n > best {
			best = n
		}
	}
	return strconv.Itoa(best + 1)
}

// countEdges counts every edge file in the tree — counts.json's "edges", a plain count.
func countEdges(t *testing.T, root string) int {
	t.Helper()
	nodes, err := os.ReadDir(filepath.Join(root, "nodes"))
	if err != nil {
		t.Fatalf("read nodes dir: %v", err)
	}
	total := 0
	for _, n := range nodes {
		files, readErr := os.ReadDir(filepath.Join(root, "nodes", n.Name(), "edges"))
		if readErr != nil {
			continue
		}
		total += len(files)
	}
	return total
}

func TestCreatedNodeTreeStillLoads(t *testing.T) {
	root := copyTreeForCreate(t, "topology-pair")

	// The new id is DERIVED, never a literal: this copies the live pair tree, which grows a
	// node every time one is actually created in the editor. A hardcoded "3" passed until
	// the tree had a 3 of its own, and then failed as a fan-in — a confusing way to be told
	// the fixture had moved.
	newID := nextNodeID(t, root)
	const newKind = "NormalSum"
	if err := Wiring.WriteNewNodeFiles(root, newID, newKind, 120, 1.4, 0.3); err != nil {
		t.Fatalf("WriteNewNodeFiles: %v", err)
	}

	// THE PORTS COME FROM THE KINDS, exactly as CreateNode resolves them. Hardcoding
	// "Out"/"In" here would make this test pass while the live path wrote an edge to a port
	// the target does not have — which is the bug it exists to catch.
	srcPort := firstPort(t, "PairNode", portwiring.PortOut)
	targetPort := firstPort(t, newKind, portwiring.PortIn)
	if targetPort == "In" {
		t.Fatalf("%s's first input is called In, so this test can no longer tell a resolved port from the old hardcoded one — point it at a kind whose ports are named something else", newKind)
	}
	if err := edgefile.WriteEdgeFile(root, "1", srcPort, newID, targetPort); err != nil {
		t.Fatalf("WriteEdgeFile: %v", err)
	}
	// Counts derived the same way, for the same reason: nodes is the LARGEST id (the row
	// count under ROW ID = NODE ID - 1), and this tree just grew by one node and one edge.
	newRows, err := strconv.Atoi(newID)
	if err != nil {
		t.Fatalf("new id %q is not a number", newID)
	}
	if err := countspersist.WriteCounts(root, newRows, countEdges(t, root)); err != nil {
		t.Fatalf("WriteCounts: %v", err)
	}

	// THE RELOAD — the moment the live editor actually fails at, so the moment to assert.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tr := T.New()
	tr.SetSink(os.Stderr)
	if _, _, _, _, err := Wiring.LoadTopology(ctx, root, tr, clock.NewRealClock()); err != nil {
		t.Fatalf("a tree with a created node did not load: %v\n"+
			"That is what the editor sees after a drop: the run ends, the host respawns, and the "+
			"new process exits during load — no camera, so no zoom, pan or rotate.", err)
	}
}
