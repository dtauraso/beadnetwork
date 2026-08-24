package Node

import (
	VecB "github.com/dtauraso/beadnetwork/Categories/Node/ChannelVectors"
	TiltB "github.com/dtauraso/beadnetwork/Categories/Node/TiltVectors"
)

func WriteNodeBlock(w *ValueWriter, f NodeState) error {
	if w == nil {
		return nil
	}
	w.Begin()
	WriteNodeValues(w, f)
	TiltB.WriteTiltArrowValues(w, f.TiltArrows)
	VecB.WriteChannelVectorValues(w, f.ChannelVectors)
	return w.Flush()
}

var NodeBlockValueNames = blockValueNames()

func blockValueNames() []string {
	names := make([]string, 0, len(NodeValueNames)+len(TiltB.TiltValueNames)+len(VecB.VectorValueNames))
	names = append(names, NodeValueNames...)
	names = append(names, TiltB.TiltValueNames...)
	names = append(names, VecB.VectorValueNames...)
	return names
}
