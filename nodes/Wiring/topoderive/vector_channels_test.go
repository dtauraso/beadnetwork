// vector_channels_test.go — moved verbatim (same assertions, same name) from
// nodes/Wiring/build_load_derive_test.go, which recorded why it moved: it drives only
// AllocateVectorChannels plus encoding/json, no MoveDispatch/writeSpecTree/LoadTopology
// harness, so it did not need to stay in package Wiring.
package topoderive

import (
	"encoding/json"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
)

// TestAllocateVectorChannelsKeysSourceOutTargetIn pins AllocateVectorChannels' two return
// maps against being swapped: the edge's SOURCE node id must key into vectorOutByNode
// (its own send end) and the TARGET node id must key into vectorInByNode (its own
// receive end), both holding the SAME underlying channel (one directed channel per
// edge). Fails under a swap of AllocateVectorChannels' two return values: the source id
// is absent from vectorOutByNode (present in vectorInByNode instead) and vice versa for
// the target id.
func TestAllocateVectorChannelsKeysSourceOutTargetIn(t *testing.T) {
	const specJSON = `{
	  "nodes": [
	    {"id":"src","type":"PairNode"},
	    {"id":"dst","type":"PairNode"}
	  ],
	  "edges": [
	    {"label":"e0","kind":"data","source":"src","sourceHandle":"Out","target":"dst","targetHandle":"In"}
	  ]
	}`
	var spec loadspec.TopoSpec
	if err := json.Unmarshal([]byte(specJSON), &spec); err != nil {
		t.Fatalf("parse spec JSON: %v", err)
	}

	vectorOutByNode, vectorInByNode := AllocateVectorChannels(spec)

	outCh, ok := vectorOutByNode["src"]
	if !ok {
		t.Fatalf("vectorOutByNode missing edge SOURCE id %q", "src")
	}
	inCh, ok := vectorInByNode["dst"]
	if !ok {
		t.Fatalf("vectorInByNode missing edge TARGET id %q", "dst")
	}
	if outCh != inCh {
		t.Fatalf("vectorOutByNode[src] and vectorInByNode[dst] are different channels; want the same directed edge channel")
	}
	if _, ok := vectorOutByNode["dst"]; ok {
		t.Fatalf("vectorOutByNode unexpectedly has an entry for the TARGET id %q — source/target keys are swapped", "dst")
	}
	if _, ok := vectorInByNode["src"]; ok {
		t.Fatalf("vectorInByNode unexpectedly has an entry for the SOURCE id %q — source/target keys are swapped", "src")
	}
}
