package owners

import (
	"context"
	"encoding/binary"
	"io"

	SF "github.com/dtauraso/wirefold/Buffer/streamframe"
	"github.com/dtauraso/wirefold/nodes/Wiring/framegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/clock"
	"github.com/dtauraso/wirefold/nodes/rowevent"
	"github.com/dtauraso/wirefold/nodes/spatial"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

type Outs struct {
	outWires []*wire.PacedWire

	outEdgeRows []int32

	beadOut io.Writer

	nodeRow int32

	buildBeadFrame func(tick uint32, nodeRow int32, beads []SF.EdgeBead, events []rowevent.RowEvent) []byte
}

type BeadFrameBuilder = func(tick uint32, nodeRow int32, beads []SF.EdgeBead, events []rowevent.RowEvent) []byte

func (o *Outs) HasOutWires() bool { return len(o.outWires) > 0 }

func (o *Outs) SetBeadStream(w io.Writer, nodeRow int32, buildBeadFrame func(tick uint32, nodeRow int32, beads []SF.EdgeBead, events []rowevent.RowEvent) []byte) {
	o.beadOut = w
	o.nodeRow = nodeRow
	o.buildBeadFrame = buildBeadFrame
}

func (o *Outs) RunAnimation(ctx context.Context) {
	if !o.HasOutWires() {
		return
	}
	clk := clock.NewRealClock()
	for {
		if ctx.Err() != nil {
			return
		}
		o.stepBeads(ctx, clk.Tick())
		if err := clk.SleepPulse(ctx); err != nil {
			return
		}
	}
}

func (o *Outs) stepBeads(ctx context.Context, tick int64) {
	axisPhi, axisTheta := framegeom.TorusDefaultAxisAngles()
	beads := make([]SF.EdgeBead, 0, len(o.outWires))
	var events []rowevent.RowEvent

	for i, pw := range o.outWires {
		pw.DriveOneStep(ctx, tick)

		edgeRow := int32(-1)
		if i < len(o.outEdgeRows) {
			edgeRow = o.outEdgeRows[i]
		}
		for _, r := range pw.LiveBeadRows() {
			pos := spatial.Vec3{X: r.X, Y: r.Y, Z: r.Z}
			beads = append(beads, SF.EdgeBead{
				X: float32(r.X), Y: float32(r.Y), Z: float32(r.Z),
				Value: int32(r.Val), EdgeRow: edgeRow,
				RingMatrix: framegeom.RingInstanceMatrixColumnMajor(
					pos, nodegeom.ShadingParamBeadRadius, axisPhi, axisTheta),
			})
		}
		events = append(events, o.drainWireEvents(pw)...)
	}
	o.writeBeadFrame(tick, beads, events)
}

func (o *Outs) drainWireEvents(pw *wire.PacedWire) []rowevent.RowEvent {
	var events []rowevent.RowEvent
	for _, pe := range pw.DrainPendingEvents() {
		events = append(events, rowevent.RowEvent{
			Kind: pe.Kind, NodeRow: o.nodeRow, PortRow: -1,
			TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
			Value: int32(pe.Value), Bead: pe.Gen,
			X: pe.X, Y: pe.Y, Z: pe.Z, F: pe.T,
		})
	}
	for _, ev := range pw.DrainBreadcrumbEvents() {
		ev.NodeRow = o.nodeRow
		ev.PortRow = -1
		ev.TargetRow = -1
		events = append(events, ev)
	}
	return events
}

func (o *Outs) writeBeadFrame(tick int64, beads []SF.EdgeBead, events []rowevent.RowEvent) {
	if o.beadOut == nil || o.buildBeadFrame == nil {
		return
	}
	frame := o.buildBeadFrame(uint32(tick), o.nodeRow, beads, events)
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], uint32(len(frame)))
	_, _ = o.beadOut.Write(hdr[:])
	_, _ = o.beadOut.Write(frame)
}

func (o *Outs) ClearOutWires() {
	for _, pw := range o.outWires {
		pw.ClearInFlight()
	}
}

func (o *Outs) AddOutWire(pw *wire.PacedWire, edgeRow int32) {
	o.outWires = append(o.outWires, pw)
	o.outEdgeRows = append(o.outEdgeRows, edgeRow)
}
