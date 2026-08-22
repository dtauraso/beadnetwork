package beadanimation

import (
	"context"
	"time"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"
	clock "github.com/dtauraso/wirefold/src/Clock"
	"github.com/dtauraso/wirefold/src/Node/BeadAnimation/lattice"
	SF "github.com/dtauraso/wirefold/src/Node/Edge"
	"github.com/dtauraso/wirefold/src/Node/framegeom"
	"github.com/dtauraso/wirefold/src/Node/nodegeom"
)

type BeadAnimation struct {
	sleepCh <-chan int64

	sleepMs int64

	outLines []*BeadLine

	outEdgeRows []int32

	nodeRow int32

	buildBeadFrame BeadFrameBuilder

	tracePath string
}

type BeadFrameBuilder = func(tick uint32, nodeRow int32, beads []SF.EdgeBead)

func (o *BeadAnimation) HasBeadLines() bool { return len(o.outLines) > 0 }

func (o *BeadAnimation) SetBeadStream(nodeRow int32, buildBeadFrame BeadFrameBuilder, sceneRoot string) {
	o.tracePath = tracePath(sceneRoot, nodeRow)
	o.nodeRow = nodeRow
	o.buildBeadFrame = buildBeadFrame
}

func (o *BeadAnimation) SetSleepCh(ch <-chan int64) { o.sleepCh = ch }

func (o *BeadAnimation) wakeAfter() time.Duration {
	if o.sleepMs == SliderPanel.Paused {
		return 0
	}
	ms := o.sleepMs
	if ms < 1 {
		ms = int64(lattice.PulsesPerSlot) * clock.MsPerTick
	}
	return time.Duration(ms) * time.Millisecond
}

func (o *BeadAnimation) RunBeadAnimation(ctx context.Context) {
	if !o.HasBeadLines() {
		return
	}
	clk := clock.NewRealClock()
	if o.sleepMs == 0 {
		o.sleepMs = SliderPanel.SleepMs(clock.SpeedNumScale, 1)
	}
	for {
		if ctx.Err() != nil {
			return
		}
		wait := o.wakeAfter()
		if wait > 0 {
			o.stepBeads(ctx, clk.Tick())
		}

		ms, changed, err := clock.SleepForOrChange(ctx, wait, o.sleepCh)
		if err != nil {
			return
		}
		if changed {
			o.sleepMs = ms
		}
	}
}

func (o *BeadAnimation) stepBeads(ctx context.Context, tick int64) {
	axisPhi, axisTheta := framegeom.TorusDefaultAxisAngles()
	beads := make([]SF.EdgeBead, 0, len(o.outLines))
	var events []RowEvent

	for i, bl := range o.outLines {
		bl.DriveOneStep(ctx, tick)

		edgeRow := int32(-1)
		if i < len(o.outEdgeRows) {
			edgeRow = o.outEdgeRows[i]
		}
		for _, r := range bl.LiveBeadRows() {
			pos := Vec3{X: r.X, Y: r.Y, Z: r.Z}
			beads = append(beads, SF.EdgeBead{
				X: float32(r.X), Y: float32(r.Y), Z: float32(r.Z),
				Value: int32(r.Val), EdgeRow: edgeRow,
				RingMatrix: framegeom.RingInstanceMatrixColumnMajor(
					framegeom.Vec3(pos), nodegeom.ShadingParamBeadRadius, axisPhi, axisTheta),
			})
		}
		events = append(events, o.drainBeadEvents(bl)...)
	}
	o.writeBeadFrame(tick, beads, events)
}

func (o *BeadAnimation) drainBeadEvents(bl *BeadLine) []RowEvent {
	var events []RowEvent
	for _, pe := range bl.DrainPendingEvents() {
		events = append(events, RowEvent{
			Kind: pe.Kind, NodeRow: o.nodeRow, PortRow: -1,
			TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
			Value: int32(pe.Value), Bead: pe.Gen,
			X: pe.X, Y: pe.Y, Z: pe.Z,
		})
	}
	for _, ev := range bl.DrainBreadcrumbEvents() {
		ev.NodeRow = o.nodeRow
		ev.PortRow = -1
		ev.TargetRow = -1
		events = append(events, ev)
	}
	return events
}

func (o *BeadAnimation) writeBeadFrame(tick int64, beads []SF.EdgeBead, events []RowEvent) {
	appendTrace(o.tracePath, events)
	if o.buildBeadFrame == nil {
		return
	}
	o.buildBeadFrame(uint32(tick), o.nodeRow, beads)
}

func (o *BeadAnimation) ClearBeadLines() {
	for _, bl := range o.outLines {
		bl.ClearInFlight()
	}
}

func (o *BeadAnimation) AddBeadLine(bl *BeadLine, edgeRow int32) {
	o.outLines = append(o.outLines, bl)
	o.outEdgeRows = append(o.outEdgeRows, edgeRow)
}
