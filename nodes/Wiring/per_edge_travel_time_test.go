// per_edge_travel_time_test.go — load-time validation guard for the model boundary:
// fan-in (two edges into the same input port) is rejected at load.

package Wiring

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"strings"
	"testing"

	T "github.com/dtauraso/wirefold/Trace"
)

// TestFanInRejectedAtLoad pins the model boundary: two edges targeting the SAME input
// port (fan-in) must be rejected at load, not silently share a wire. Uses SinkNode's one
// "In" port with two incident edges — the exact shape validateNoFanIn forbids.
//
// One thing here is load-bearing, and was MISSING in the version of this test a
// mutation audit found vacuous (it passed with validateNoFanIn deleted outright):
//
//   - The assertion reads the error's CONTENT, not just err != nil. More than one
//     mechanism can fail this fixture (validateNoFanIn, an allocateWires panic); only the
//     message distinguishes them.
func TestFanInRejectedAtLoad(t *testing.T) {
	const topo = `{
	  "nodes": [
	    {"id":"1","type":"SrcNode","outputs":[{"name":"Out"}]},
	    {"id":"2","type":"SrcNode","outputs":[{"name":"Out"}]},
	    {"id":"3","type":"SinkNode","inputs":[{"name":"In"}]}
	  ],
	  "edges": [
	    {"label":"eA","kind":"data","source":"1","sourceHandle":"Out","target":"3","targetHandle":"In"},
	    {"label":"eB","kind":"data","source":"2","sourceHandle":"Out","target":"3","targetHandle":"In"}
	  ]
	}`
	root := writeSpecTree(t, t.TempDir(), topo)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, _, _, err := LoadTopology(ctx, root, T.New(), wire.NewRealClock())
	if err == nil {
		t.Fatalf("LoadTopology accepted a fan-in topology (two edges into sink.In); want a fan-in rejection error")
	}
	if !strings.Contains(err.Error(), "fan-in not allowed") {
		t.Fatalf("LoadTopology rejected the topology for the WRONG reason: %v\nwant an error naming the fan-in violation (two edges into sink.In)", err)
	}
}
