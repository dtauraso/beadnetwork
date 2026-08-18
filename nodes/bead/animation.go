package bead

import (
	"context"
	"encoding/binary"
	"io"
	"math"

	SF "github.com/dtauraso/wirefold/Buffer/streamframe"
	"github.com/dtauraso/wirefold/nodes/Wiring/framegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/bead/lattice"
	"github.com/dtauraso/wirefold/nodes/clock"
	"github.com/dtauraso/wirefold/nodes/rowevent"
	"github.com/dtauraso/wirefold/nodes/spatial"
)

type Animation struct {
	speedCh <-chan float64

	scalar float64

	outRuns []*BeadRun

	outEdgeRows []int32

	beadOut io.Writer

	nodeRow int32

	buildBeadFrame func(tick uint32, nodeRow int32, beads []SF.EdgeBead, events []rowevent.RowEvent) []byte
}

type BeadFrameBuilder = func(tick uint32, nodeRow int32, beads []SF.EdgeBead, events []rowevent.RowEvent) []byte

func (o *Animation) HasBeadRuns() bool { return len(o.outRuns) > 0 }

func (o *Animation) SetBeadStream(w io.Writer, nodeRow int32, buildBeadFrame func(tick uint32, nodeRow int32, beads []SF.EdgeBead, events []rowevent.RowEvent) []byte) {
	o.beadOut = w
	o.nodeRow = nodeRow
	o.buildBeadFrame = buildBeadFrame
}

func (o *Animation) SetSpeedCh(ch <-chan float64) { o.speedCh = ch }

func (o *Animation) wakePulses() int {
	s := o.scalar
	if s <= 0 {
		s = 1
	}
	n := int(math.Round(lattice.PulsesPerSlot / s))
	if n < 1 {
		return 1
	}
	return n
}

func (o *Animation) RunAnimation(ctx context.Context) {
	if !o.HasBeadRuns() {
		return
	}
	clk := clock.NewRealClock()
	for {
		if ctx.Err() != nil {
			return
		}
		if sp, ok := clock.RecvSpeedNonBlocking(o.speedCh); ok {
			o.scalar = sp
		}
		o.stepBeads(ctx, clk.Tick())
		if err := clk.SleepPulses(ctx, o.wakePulses()); err != nil {
			return
		}
	}
}

func (o *Animation) stepBeads(ctx context.Context, tick int64) {
	axisPhi, axisTheta := framegeom.TorusDefaultAxisAngles()
	beads := make([]SF.EdgeBead, 0, len(o.outRuns))
	var events []rowevent.RowEvent

	for i, pw := range o.outRuns {
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
		events = append(events, o.drainBeadEvents(pw)...)
	}
	o.writeBeadFrame(tick, beads, events)
}

func (o *Animation) drainBeadEvents(pw *BeadRun) []rowevent.RowEvent {
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

func (o *Animation) writeBeadFrame(tick int64, beads []SF.EdgeBead, events []rowevent.RowEvent) {
	if o.beadOut == nil || o.buildBeadFrame == nil {
		return
	}
	frame := o.buildBeadFrame(uint32(tick), o.nodeRow, beads, events)
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], uint32(len(frame)))
	_, _ = o.beadOut.Write(hdr[:])
	_, _ = o.beadOut.Write(frame)
}

func (o *Animation) ClearBeadRuns() {
	for _, pw := range o.outRuns {
		pw.ClearInFlight()
	}
}

func (o *Animation) AddBeadRun(pw *BeadRun, edgeRow int32) {
	o.outRuns = append(o.outRuns, pw)
	o.outEdgeRows = append(o.outEdgeRows, edgeRow)
}
