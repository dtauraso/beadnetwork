// vector_channels.go — the tilt-vector channel allocation derive phase, lifted out of
// nodes/Wiring/build_wires.go.
package topoderive

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
)

// AllocateVectorChannels creates one dedicated node-to-node tilt-vector channel
// (tilt_vector_channel.go's TiltVectorMsg, buffered 1, latest-wins) per directed
// edge whose BOTH endpoint kinds ask for one (tiltvector.KindWantsVectorChannel — today only
// PairNode). A kind that never asks gets no entry in either map and is entirely
// unaffected. This travels ALONGSIDE the ordinary bead edge (allocateWires, which stays in
// nodes/Wiring), never replacing it — the source node keeps its existing *wire.Out for beads
// and additionally gets this channel's send end; the target node keeps its existing *wire.In
// and additionally gets this channel's receive end.
func AllocateVectorChannels(spec loadspec.TopoSpec) (vectorOutByNode, vectorInByNode map[string]chan tiltvector.TiltVectorMsg) {
	kindByID := make(map[string]string, len(spec.Nodes))
	for _, n := range spec.Nodes {
		kindByID[n.ID] = n.Type
	}
	vectorOutByNode = map[string]chan tiltvector.TiltVectorMsg{}
	vectorInByNode = map[string]chan tiltvector.TiltVectorMsg{}
	for _, e := range spec.Edges {
		if !tiltvector.KindWantsVectorChannel(kindByID[e.Source]) || !tiltvector.KindWantsVectorChannel(kindByID[e.Target]) {
			continue
		}
		sourceToTargetVectorCh := make(chan tiltvector.TiltVectorMsg, 1)
		vectorOutByNode[e.Source] = sourceToTargetVectorCh
		vectorInByNode[e.Target] = sourceToTargetVectorCh
	}
	return vectorOutByNode, vectorInByNode
}
