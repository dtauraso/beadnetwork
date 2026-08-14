package owners

import (
	"encoding/binary"
	"io"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/rowevent"
)

// EdgeFrameBuilder packs one drawn edge for one tick.
type EdgeFrameBuilder = func(tick uint32, sx, sy, sz, ex, ey, ez float32, srcNodeRow int32, label string, events []rowevent.RowEvent) []byte

// outEdge is one edge LEAVING this node. The node holds the edge's own delta triple D and
// derives the far endpoint from it — Compose(ownPolar, D) — instead of keeping a copy of
// the neighbour's centre. D describes the edge; it does not say where the neighbour is.
type outEdge struct {
	label    string
	edgeRow  int32
	targetID string

	targetKind string

	out io.Writer
}

// OutEdges draws the edges leaving this node, on this node's own goroutine — the one that
// owns its position and its deltas.
type OutEdges struct {
	edges []outEdge

	nodeRow int32

	buildFrame EdgeFrameBuilder
}

func (o *OutEdges) Any() bool { return len(o.edges) > 0 }

// AddOutEdge claims the stream for one edge leaving this node.
func (o *OutEdges) AddOutEdge(label string, edgeRow int32, targetID, targetKind string, w io.Writer, nodeRow int32, buildFrame EdgeFrameBuilder) {
	o.nodeRow = nodeRow
	o.buildFrame = buildFrame
	o.edges = append(o.edges, outEdge{
		label: label, edgeRow: edgeRow, targetID: targetID, targetKind: targetKind, out: w,
	})
}

// WriteFrames draws every out-edge for this tick. Each segment is arithmetic on state this
// node owns: its own polar position plus the edge's delta. No trig per edge beyond the one
// cartesian boundary the renderer needs.
func (o *OutEdges) WriteFrames(tick int64, self nodegeom.NodeGeom, deltas *Deltas) {
	if o.buildFrame == nil || !self.HasPos {
		return
	}
	selfCenter := nodegeom.NodeWorldPos(self)
	ownPolar := polar.Cart2polar(selfCenter.Sub(self.SceneCenter))

	for _, e := range o.edges {
		if e.out == nil {
			continue
		}
		d, ok := deltas.DeltaTo(e.targetID)
		if !ok {
			continue
		}
		targetCenter := polar.Polar2cart(polar.Compose(ownPolar, d)).Add(self.SceneCenter)

		start, end := selfCenter, targetCenter
		dir := targetCenter.Sub(selfCenter)
		if dir.Length() >= 1e-9 {
			unit := dir.Normalize()
			start = selfCenter.Add(unit.Scale(nodegeom.NodeTorusOuterR(self.Kind)))
			end = targetCenter.Sub(unit.Scale(nodegeom.NodeTorusOuterR(e.targetKind)))
		}

		frame := o.buildFrame(uint32(tick),
			float32(start.X), float32(start.Y), float32(start.Z),
			float32(end.X), float32(end.Y), float32(end.Z),
			o.nodeRow, e.label, nil)
		var hdr [4]byte
		binary.LittleEndian.PutUint32(hdr[:], uint32(len(frame)))
		_, _ = e.out.Write(hdr[:])
		_, _ = e.out.Write(frame)
	}
}
