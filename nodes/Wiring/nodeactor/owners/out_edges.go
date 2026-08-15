package owners

import (
	"encoding/binary"
	"io"

	"github.com/dtauraso/wirefold/nodes/Wiring/edgefile"
	"github.com/dtauraso/wirefold/nodes/Wiring/edgegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/rowevent"
	"github.com/dtauraso/wirefold/nodes/spatial"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/outport"
)

type EdgeFrameBuilder = func(tick uint32, sx, sy, sz, ex, ey, ez float32, srcNodeRow, dstNodeRow int32, deltaR float32, label string, events []rowevent.RowEvent) []byte

type outEdge struct {
	label      string
	edgeRow    int32
	targetID   string
	dstNodeRow int32
	deltaR     float32

	targetKind string

	out io.Writer

	port *outport.Out
	dest *wire.PacedWire

	start, end spatial.Vec3
	steps      int
	derived    bool

	persistedDrag polar.Polar
	hasPersisted  bool
}

type OutEdges struct {
	edges []outEdge

	nodeRow int32

	buildFrame EdgeFrameBuilder

	srcID       string
	persistRoot string
}

func (o *OutEdges) SetSrcID(id string) { o.srcID = id }

func (o *OutEdges) SetPersistRoot(root string) { o.persistRoot = root }

func (o *OutEdges) edgeFor(label string) *outEdge {
	for i := range o.edges {
		if o.edges[i].label == label {
			return &o.edges[i]
		}
	}
	o.edges = append(o.edges, outEdge{label: label, edgeRow: -1})
	return &o.edges[len(o.edges)-1]
}

func (o *OutEdges) BindWire(label, targetID, targetKind string, port *outport.Out, dest *wire.PacedWire) {
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

func (o *OutEdges) DeriveGeometry(tick int64, self nodegeom.NodeGeom, deltas *Deltas) {
	if !self.HasPos {
		return
	}
	selfCenter := nodegeom.NodeWorldPos(self)
	ownPolar := polar.Cart2polar(selfCenter.Sub(self.SceneCenter))

	for i := range o.edges {
		e := &o.edges[i]
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
		if dist, _, ok := edgegeom.EdgeCenterDistAndDir(selfCenter, targetCenter); ok {
			e.steps = edgegeom.EdgeStepCount(dist, self.Kind, e.targetKind)
		}
		e.start, e.end, e.derived = start, end, true
		e.deltaR = float32(d.R)

		if e.port != nil {
			e.port.SetGeom(e.steps, start, end)
		}
		if e.dest != nil {
			e.dest.ReviseInFlightGeometry(tick, e.steps, spatial.WireSegment{Start: start, End: end})
		}
		if dragDelta, ok := deltas.DragDeltaTo(e.targetID); ok {
			o.persistDelta(e, dragDelta)
		}
	}
}

func (o *OutEdges) persistDelta(e *outEdge, dragDelta polar.Polar) {
	if o.persistRoot == "" || o.srcID == "" {
		return
	}
	if e.hasPersisted && e.persistedDrag == dragDelta {
		return
	}
	if err := edgefile.WriteEdgeDelta(o.persistRoot, o.srcID, e.label, dragDelta); err != nil {
		jsonpersist.LogPersistErr("out_edges", o.srcID+"->"+e.targetID, err)
		return
	}
	e.persistedDrag, e.hasPersisted = dragDelta, true
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
			o.nodeRow, e.dstNodeRow, e.deltaR, e.label, nil)
		var hdr [4]byte
		binary.LittleEndian.PutUint32(hdr[:], uint32(len(frame)))
		_, _ = e.out.Write(hdr[:])
		_, _ = e.out.Write(frame)
	}
}
