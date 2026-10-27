package Node

import (
	"github.com/dtauraso/beadnetwork/Categories/Node/AngleArcs"
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
	WriteTickValues(w, f.TickMarks)
	AngleArcs.WriteValues(w, f.AngleArcs)
	return w.Flush()
}

var NodeBlockValueNames = blockValueNames()

func blockValueNames() []string {
	names := make([]string, 0, len(NodeValueNames)+len(TiltB.TiltValueNames)+len(VecB.VectorValueNames)+len(TickValueNames)+len(AngleArcs.ValueNames))
	names = append(names, NodeValueNames...)
	names = append(names, TiltB.TiltValueNames...)
	names = append(names, VecB.VectorValueNames...)
	names = append(names, TickValueNames...)
	names = append(names, AngleArcs.ValueNames...)
	return names
}
