package owners

import (
	"encoding/binary"
	"io"

	"github.com/dtauraso/wirefold/nodes/Wiring/edgefile"
	"github.com/dtauraso/wirefold/nodes/Wiring/edgegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/polarindex"
	"github.com/dtauraso/wirefold/nodes/bead"
	"github.com/dtauraso/wirefold/nodes/bead/outport"
	"github.com/dtauraso/wirefold/nodes/rowevent"
	"github.com/dtauraso/wirefold/nodes/spatial"
)

type EdgeFrameBuilder = func(tick uint32, sx, sy, sz, ex, ey, ez float32, srcNodeRow, dstNodeRow int32, deltaR float32, dragActive uint8, label string, events []rowevent.RowEvent) []byte

type outEdge struct {
	label      string
	edgeRow    int32
	targetID   string
	dstNodeRow int32
	deltaR     float32

	targetKind string

	out io.Writer

	port *outport.Out
	dest *bead.BeadRun

	start, end spatial.Vec3
	steps      int
	derived    bool

	persistedDragIdx polarindex.Offset
	hasPersisted     bool

	ruleInactive bool
}

type OutEdges struct {
	edges []outEdge

	nodeRow int32

	buildFrame EdgeFrameBuilder

	srcID       string
	persistRoot string
	constants   polarindex.SceneConstants
}

func (o *OutEdges) SetSrcID(id string) { o.srcID = id }

func (o *OutEdges) SetPersistRoot(root string) { o.persistRoot = root }

func (o *OutEdges) SetConstants(sc polarindex.SceneConstants) { o.constants = sc }

func (o *OutEdges) edgeFor(label string) *outEdge {
	for i := range o.edges {
		if o.edges[i].label == label {
			return &o.edges[i]
		}
	}
	o.edges = append(o.edges, outEdge{label: label, edgeRow: -1})
	return &o.edges[len(o.edges)-1]
}

func (o *OutEdges) BindWire(label, targetID, targetKind string, port *outport.Out, dest *bead.BeadRun) {
	e := o.edgeFor(label)
	e.port = port
	e.dest = dest
	if e.targetID == "" {
		e.targetID = targetID
	}
	if e.targetKind == "" {
		e.targetKind = targetKind
	}
}

func (o *OutEdges) SetEdgeRuleActive(label string, active bool) {
	o.edgeFor(label).ruleInactive = !active
}

func activeByte(inactive bool) uint8 {
	if inactive {
		return 0
	}
	return 1
}

func (o *OutEdges) Any() bool { return len(o.edges) > 0 }

func (o *OutEdges) AddOutEdge(label string, edgeRow int32, targetID, targetKind string, w io.Writer, nodeRow, dstNodeRow int32, buildFrame EdgeFrameBuilder) {
	o.nodeRow = nodeRow
	o.buildFrame = buildFrame
	e := o.edgeFor(label)
	e.edgeRow = edgeRow
	e.targetID = targetID
	e.targetKind = targetKind
	e.dstNodeRow = dstNodeRow
	e.out = w
}

func (o *OutEdges) DeriveGeometry(self nodegeom.NodeGeom, deltas *Deltas) {
	if !self.HasPos {
		return
	}
	selfCenter := nodegeom.NodeWorldPos(self)
	selfIndex := nodegeom.ComposedIndexOf(self)

	for i := range o.edges {
		e := &o.edges[i]
		d, ok := deltas.DeltaTo(e.targetID)
		if !ok {
			continue
		}
		targetIndex := polarindex.Compose(selfIndex, d, o.constants)
		targetCenter := polar.Polar2cart(polarindex.ToPolar(targetIndex, o.constants)).Add(self.SceneCenter)

		start, end := selfCenter, targetCenter
		dir := targetCenter.Sub(selfCenter)
		if dir.Length() >= 1e-9 {
			unit := dir.Normalize()
			start = selfCenter.Add(unit.Scale(nodegeom.NodeTorusOuterR(self.Kind)))
			end = targetCenter.Sub(unit.Scale(nodegeom.NodeTorusOuterR(e.targetKind)))
		}
		if dist, _, ok := edgegeom.EdgeCenterDistAndDir(selfCenter, targetCenter); ok {
			e.steps = edgegeom.EdgeStepCount(dist, self.Kind, e.targetKind)
		}
		e.start, e.end, e.derived = start, end, true
		e.deltaR = float32(d.R) * float32(o.constants.ConstantR)

		if e.port != nil {
			e.port.PostGeom(e.steps, start, end)
		}
		if e.dest != nil {
			e.dest.PostGeom(e.steps, spatial.Segment{Start: start, End: end})
		}
		if dragDelta, ok := deltas.DragDeltaTo(e.targetID); ok {
			o.persistDelta(e, dragDelta)
		}
	}
}

func (o *OutEdges) persistDelta(e *outEdge, off polarindex.Offset) {
	if o.persistRoot == "" || o.srcID == "" {
		return
	}
	if e.hasPersisted && e.persistedDragIdx == off {
		return
	}
	if err := edgefile.WriteEdgeDrag(o.persistRoot, o.srcID, e.label, off); err != nil {
		jsonpersist.LogPersistErr("out_edges", o.srcID+"->"+e.targetID, err)
		return
	}
	e.persistedDragIdx, e.hasPersisted = off, true
}

func (o *OutEdges) WriteFrames(tick int64, self nodegeom.NodeGeom, deltas *Deltas) {
	if o.buildFrame == nil || !self.HasPos {
		return
	}

	for _, e := range o.edges {
		if e.out == nil || !e.derived {
			continue
		}
		start, end := e.start, e.end

		frame := o.buildFrame(uint32(tick),
			float32(start.X), float32(start.Y), float32(start.Z),
			float32(end.X), float32(end.Y), float32(end.Z),
			o.nodeRow, e.dstNodeRow, e.deltaR, activeByte(e.ruleInactive), e.label, nil)
		var hdr [4]byte
		binary.LittleEndian.PutUint32(hdr[:], uint32(len(frame)))
		_, _ = e.out.Write(hdr[:])
		_, _ = e.out.Write(frame)
	}
}
