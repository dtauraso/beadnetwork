// per_edge_travel_time_test.go — load-time validation guard for the model boundary:
// fan-in (two edges into the same input port) is rejected at load.

package Wiring

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"os"
	"path/filepath"
	"testing"

	T "github.com/dtauraso/wirefold/Trace"
)

// TestFanInRejectedAtLoad pins the model boundary: two edges targeting the SAME input
// port (fan-in) must be rejected at load, not silently share a wire. Uses SinkNode's one
// "In" port with two incident edges — the exact shape validateNoFanIn forbids.
func TestFanInRejectedAtLoad(t *testing.T) {
	const topo = `{
	  "nodes": [
	    {"id":"a","type":"SrcNode","outputs":[{"name":"Out"}]},
	    {"id":"b","type":"SrcNode","outputs":[{"name":"Out"}]},
	    {"id":"sink","type":"SinkNode","inputs":[{"name":"In"}]}
	  ],
	  "edges": [
	    {"label":"eA","kind":"data","source":"a","sourceHandle":"Out","target":"sink","targetHandle":"In"},
	    {"label":"eB","kind":"data","source":"b","sourceHandle":"Out","target":"sink","targetHandle":"In"}
	  ]
	}`
	dir := t.TempDir()
	path := filepath.Join(dir, "topo.json")
	if err := os.WriteFile(path, []byte(topo), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if _, _, _, _, err := LoadTopology(ctx, path, T.New(), wire.NewRealClock()); err == nil {
		t.Fatalf("LoadTopology accepted a fan-in topology (two edges into sink.In); want a fan-in rejection error")
	}
}
