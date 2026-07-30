package Wiring

import (
	"testing"
)

// writeLoaderTreeFixture lays down a small, self-contained directory-tree topology
// (independent of the production topology/ dir) that exercises the loadTree shapes
// TestLoadTreeRoundTrip asserts on: distinct source/target handles per edge, and one
// edge label that is deliberately NEVER written to the fixture (so an absence assertion
// on it is a genuine proof, not a tautology about a string nobody could produce). There
// are no port files any more (docs/channels-not-ports.md — a port is a load-time
// channel-binding ROLE, resolved from the kind's registry, never a placed file on disk).
func writeLoaderTreeFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeTreeFile(t, root, "nodes/n1/meta.json", `{"id":"n1","type":"Input"}`)
	writeTreeFile(t, root, "nodes/n2/meta.json", `{"id":"n2","type":"Time"}`)
	writeTreeFile(t, root, "nodes/n3/meta.json", `{"id":"n3","type":"TimeEnd"}`)
	writeTreeFile(t, root, "nodes/n1/edges/n1Ton2.json", `{"label":"n1Ton2","kind":"data","sourceHandle":"Out","target":"n2","targetHandle":"FromPrev"}`)
	writeTreeFile(t, root, "nodes/n2/edges/n2Ton3.json", `{"label":"n2Ton3","kind":"data","sourceHandle":"ToNext","target":"n3","targetHandle":"FromInput"}`)
	return root
}

// TestLoadTreeRoundTrip drives loadTree against a small, self-contained fixture (NOT
// the live production topology/ dir — that dir is the visual editor's OWN save target,
// so a change-detector pinned to its exact node/edge count would break on every
// legitimate editor edit with no loader bug involved) and asserts the genuinely general
// loader behaviors: node id/type round-trip, edge source/target/handle round-trip, and
// absence of a label that was deliberately never written to the fixture.
func TestLoadTreeRoundTrip(t *testing.T) {
	root := writeLoaderTreeFixture(t)

	spec, err := loadTree(root)
	if err != nil {
		t.Fatalf("loadTree: %v", err)
	}

	if len(spec.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(spec.Nodes))
	}
	if len(spec.Edges) != 2 {
		t.Fatalf("expected 2 edges, got %d", len(spec.Edges))
	}

	nodeByID := map[string]specNode{}
	for _, n := range spec.Nodes {
		nodeByID[n.ID] = n
	}

	n1, ok := nodeByID["n1"]
	if !ok {
		t.Fatal("node \"n1\" not found")
	}
	if n1.Type != "Input" {
		t.Errorf("node \"n1\" type: got %q, want \"Input\"", n1.Type)
	}

	n2, ok := nodeByID["n2"]
	if !ok {
		t.Fatal("node \"n2\" not found")
	}
	if n2.Type != "Time" {
		t.Errorf("node \"n2\" type: got %q, want \"Time\"", n2.Type)
	}

	n3, ok := nodeByID["n3"]
	if !ok {
		t.Fatal("node \"n3\" not found")
	}
	if n3.Type != "TimeEnd" {
		t.Errorf("node \"n3\" type: got %q, want \"TimeEnd\"", n3.Type)
	}

	edgeByLabel := map[string]specEdge{}
	for _, e := range spec.Edges {
		edgeByLabel[e.Label] = e
	}

	e1, ok := edgeByLabel["n1Ton2"]
	if !ok {
		t.Fatal("edge \"n1Ton2\" not found")
	}
	if e1.Source != "n1" {
		t.Errorf("edge n1Ton2 source: got %q, want \"n1\"", e1.Source)
	}
	if e1.SourceHandle != "Out" {
		t.Errorf("edge n1Ton2 sourceHandle: got %q, want \"Out\"", e1.SourceHandle)
	}
	if e1.Target != "n2" {
		t.Errorf("edge n1Ton2 target: got %q, want \"n2\"", e1.Target)
	}
	if e1.TargetHandle != "FromPrev" {
		t.Errorf("edge n1Ton2 targetHandle: got %q, want \"FromPrev\"", e1.TargetHandle)
	}

	e2, ok := edgeByLabel["n2Ton3"]
	if !ok {
		t.Fatal("edge \"n2Ton3\" not found")
	}
	if e2.SourceHandle != "ToNext" {
		t.Errorf("edge n2Ton3 sourceHandle: got %q, want \"ToNext\"", e2.SourceHandle)
	}
	if e2.TargetHandle != "FromInput" {
		t.Errorf("edge n2Ton3 targetHandle: got %q, want \"FromInput\"", e2.TargetHandle)
	}

	// A label never written to the fixture must genuinely be absent — this is a real
	// proof here (unlike the deleted production-pinned version, which asserted absence
	// of a string that no code path could ever produce again).
	if _, ok := edgeByLabel["n1Ton3"]; ok {
		t.Error("edge \"n1Ton3\" was never written to the fixture and should not exist")
	}
}
