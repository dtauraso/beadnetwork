package Wiring

import (
	"strings"
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
	writeTreeFile(t, root, "nodes/1/meta.json", `{"id":"1","type":"Input"}`)
	writeTreeFile(t, root, "nodes/2/meta.json", `{"id":"2","type":"Time"}`)
	writeTreeFile(t, root, "nodes/3/meta.json", `{"id":"3","type":"TimeEnd"}`)
	writeTreeFile(t, root, "nodes/1/edges/n1Ton2.json", `{"label":"n1Ton2","kind":"data","sourceHandle":"Out","target":"2","targetHandle":"FromPrev"}`)
	writeTreeFile(t, root, "nodes/2/edges/n2Ton3.json", `{"label":"n2Ton3","kind":"data","sourceHandle":"ToNext","target":"3","targetHandle":"FromInput"}`)
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

	n1, ok := nodeByID["1"]
	if !ok {
		t.Fatal("node \"1\" not found")
	}
	if n1.Type != "Input" {
		t.Errorf("node \"1\" type: got %q, want \"Input\"", n1.Type)
	}

	n2, ok := nodeByID["2"]
	if !ok {
		t.Fatal("node \"2\" not found")
	}
	if n2.Type != "Time" {
		t.Errorf("node \"2\" type: got %q, want \"Time\"", n2.Type)
	}

	n3, ok := nodeByID["3"]
	if !ok {
		t.Fatal("node \"3\" not found")
	}
	if n3.Type != "TimeEnd" {
		t.Errorf("node \"3\" type: got %q, want \"TimeEnd\"", n3.Type)
	}

	edgeByLabel := map[string]specEdge{}
	for _, e := range spec.Edges {
		edgeByLabel[e.Label] = e
	}

	e1, ok := edgeByLabel["n1Ton2"]
	if !ok {
		t.Fatal("edge \"n1Ton2\" not found")
	}
	if e1.Source != "1" {
		t.Errorf("edge n1Ton2 source: got %q, want \"1\"", e1.Source)
	}
	if e1.SourceHandle != "Out" {
		t.Errorf("edge n1Ton2 sourceHandle: got %q, want \"Out\"", e1.SourceHandle)
	}
	if e1.Target != "2" {
		t.Errorf("edge n1Ton2 target: got %q, want \"2\"", e1.Target)
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

// TestLoadTreeRowCountIsLargestID pins the current model: ROW ID = NODE ID - 1, declared by
// the directory name, never derived by sorting. loadTree does not order spec.Nodes at all
// any more — it parses each directory name to an int and rejects a non-numeric name, a
// duplicate parsed id, and an id below 1, and spec.RowCount is the LARGEST id found (the
// buffer's row space), not len(spec.Nodes).
func TestLoadTreeRowCountIsLargestID(t *testing.T) {
	root := t.TempDir()
	ids := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"}
	for _, id := range ids {
		writeTreeFile(t, root, "nodes/"+id+"/meta.json", `{"id":"`+id+`","type":"Input"}`)
	}

	spec, err := loadTree(root)
	if err != nil {
		t.Fatalf("loadTree: %v", err)
	}
	if len(spec.Nodes) != len(ids) {
		t.Fatalf("expected %d nodes, got %d", len(ids), len(spec.Nodes))
	}
	if spec.RowCount != 11 {
		t.Fatalf("RowCount = %d, want 11 (the largest id)", spec.RowCount)
	}
	gotIDs := map[string]bool{}
	for _, n := range spec.Nodes {
		gotIDs[n.ID] = true
	}
	for _, id := range ids {
		if !gotIDs[id] {
			t.Fatalf("node id %q missing from spec.Nodes", id)
		}
	}
}

// TestLoadTreeGapLeavesRowCountAtLargestID: deleting a node (here, never creating id 5)
// leaves a hole in the id space rather than shrinking RowCount to the live node count — a
// deleted node's row must stay empty, not have later rows shift down to fill it.
func TestLoadTreeGapLeavesRowCountAtLargestID(t *testing.T) {
	root := t.TempDir()
	for _, id := range []string{"1", "2", "3", "4", "6"} { // no "5"
		writeTreeFile(t, root, "nodes/"+id+"/meta.json", `{"id":"`+id+`","type":"Input"}`)
	}
	spec, err := loadTree(root)
	if err != nil {
		t.Fatalf("loadTree: %v", err)
	}
	if len(spec.Nodes) != 5 {
		t.Fatalf("expected 5 nodes, got %d", len(spec.Nodes))
	}
	if spec.RowCount != 6 {
		t.Fatalf("RowCount = %d, want 6 (largest id, gap at row 4 for missing id 5)", spec.RowCount)
	}
}

// TestLoadTreeNonNumericNodeDirFailsLoudly: a node directory name that doesn't parse as an
// int is a load error naming the offending directory, never a silent fallback.
func TestLoadTreeNonNumericNodeDirFailsLoudly(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "nodes/abc/meta.json", `{"id":"abc","type":"Input"}`)
	_, err := loadTree(root)
	if err == nil {
		t.Fatalf("loadTree: expected an error for non-numeric node directory %q, got nil", "abc")
	}
	if !strings.Contains(err.Error(), "abc") {
		t.Fatalf("loadTree error %q does not name the offending directory %q", err, "abc")
	}
}

// TestLoadTreeZeroOrNegativeNodeIDFailsLoudly: node ids are 1-based; "0" and a negative id
// are both load errors, not silently accepted rows.
func TestLoadTreeZeroOrNegativeNodeIDFailsLoudly(t *testing.T) {
	for _, id := range []string{"0", "-1"} {
		root := t.TempDir()
		writeTreeFile(t, root, "nodes/"+id+"/meta.json", `{"id":"`+id+`","type":"Input"}`)
		_, err := loadTree(root)
		if err == nil {
			t.Fatalf("loadTree: expected an error for node id %q (ids are 1-based), got nil", id)
		}
	}
}

// TestLoadTreeDuplicateParsedIDFailsLoudly: two directory names that parse to the same int
// id (e.g. "1" and "01") is a duplicate row claim — a load error naming both directories,
// never a silent overwrite of one row by the other.
func TestLoadTreeDuplicateParsedIDFailsLoudly(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "nodes/1/meta.json", `{"id":"1","type":"Input"}`)
	writeTreeFile(t, root, "nodes/01/meta.json", `{"id":"01","type":"Input"}`)
	_, err := loadTree(root)
	if err == nil {
		t.Fatalf("loadTree: expected an error for duplicate node id (dirs \"1\" and \"01\"), got nil")
	}
}

// TestLoadTreeEdgeOrderIsLexicographicByLabel asserts a single node's edges/ listing
// (loader_tree.go's sort.Strings(edgeFiles)) orders by plain string comparison — edge-file
// names are LABELS, not numbers, so lexicographic order is correct here and must not be
// "fixed" the way node ids were.
func TestLoadTreeEdgeOrderIsLexicographicByLabel(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "nodes/1/meta.json", `{"id":"1","type":"Input"}`)
	dstID := map[string]string{"1": "2", "10": "3", "2": "4", "9": "5"}
	for _, id := range []string{"1", "10", "2", "9"} {
		writeTreeFile(t, root, "nodes/"+dstID[id]+"/meta.json", `{"id":"`+dstID[id]+`","type":"Time"}`)
		writeTreeFile(t, root, "nodes/1/edges/"+id+".json",
			`{"label":"`+id+`","kind":"data","sourceHandle":"Out","target":"`+dstID[id]+`","targetHandle":"FromPrev"}`)
	}

	spec, err := loadTree(root)
	if err != nil {
		t.Fatalf("loadTree: %v", err)
	}
	got := make([]string, len(spec.Edges))
	for i, e := range spec.Edges {
		got[i] = e.Label
	}
	// Lexicographic string order of "1","10","2","9" — NOT numeric order.
	want := []string{"1", "10", "2", "9"}
	if len(got) != len(want) {
		t.Fatalf("expected %d edges, got %d (%v)", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("edge order = %v, want %v (position %d: got %q, want %q)", got, want, i, got[i], want[i])
		}
	}
}

// TestLoadTreeNonNumericNodeIDFailsLoudly asserts a node directory name that is not a
// number is a LOAD ERROR, not a silent string-sort fallback: node identity IS the row
// index (no id sidecar), so a mis-ordered/mis-parsed row would silently render one node's
// geometry as another's.
func TestLoadTreeNonNumericNodeIDFailsLoudly(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "nodes/1/meta.json", `{"id":"1","type":"Input"}`)
	writeTreeFile(t, root, "nodes/alpha/meta.json", `{"id":"alpha","type":"Input"}`)

	_, err := loadTree(root)
	if err == nil {
		t.Fatal("loadTree with a non-numeric node directory name succeeded; want a load error")
	}
	if !strings.Contains(err.Error(), "alpha") || !strings.Contains(err.Error(), "not a numeric id") {
		t.Fatalf("loadTree error = %q, want it to name the offending directory and say it isn't numeric", err.Error())
	}
}
