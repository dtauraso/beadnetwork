package edgemover

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/rowevent"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
	"github.com/dtauraso/wirefold/nodes/wire/outport"

	T "github.com/dtauraso/wirefold/Trace"
)

type EdgeMover struct {
	edgeID string
	srcID  string
	dstID  string
	srcH   string
	dstH   string

	srcGeom nodegeom.NodeGeom
	dstGeom nodegeom.NodeGeom
	out     *outport.Out
	dest    *wire.PacedWire

	extIn chan movemsg.Msg
	srcIn chan movemsg.Msg
	dstIn chan movemsg.Msg

	stepsIn chan int

	steps int
	tr    *T.Trace

	clockSrc clock.Clock

	clk clock.Clock

	speedCh chan float64

	streamOut StreamHandle

	edgeRow int32

	nodeRowFor func(id string) (int32, bool)

	buildFrame func(tick uint32, sx, sy, sz, ex, ey, ez float32, label string, events []rowevent.RowEvent) []byte
}

func New(edgeID, srcID, dstID, srcHandle, dstHandle string, srcGeom, dstGeom nodegeom.NodeGeom, tr *T.Trace, clockSrc clock.Clock) *EdgeMover {
	return &EdgeMover{
		edgeID:   edgeID,
		srcID:    srcID,
		dstID:    dstID,
		srcH:     srcHandle,
		dstH:     dstHandle,
		srcGeom:  srcGeom,
		dstGeom:  dstGeom,
		extIn:    make(chan movemsg.Msg, InboxDepth),
		srcIn:    make(chan movemsg.Msg, InboxDepth),
		dstIn:    make(chan movemsg.Msg, InboxDepth),
		stepsIn:  make(chan int, 1),
		tr:       tr,
		clockSrc: clockSrc,
		clk:      clock.NewRealClock(),
		edgeRow:  -1,
	}
}

const InboxDepth = 8

func (m *EdgeMover) SrcID() string     { return m.srcID }
func (m *EdgeMover) DstID() string     { return m.dstID }
func (m *EdgeMover) SrcHandle() string { return m.srcH }
func (m *EdgeMover) DstHandle() string { return m.dstH }

func (m *EdgeMover) SetOut(out *outport.Out) { m.out = out }

func (m *EdgeMover) SetDest(dest *wire.PacedWire) { m.dest = dest }

func (m *EdgeMover) Dest() *wire.PacedWire { return m.dest }

func (m *EdgeMover) SetSpeedCh(ch chan float64) { m.speedCh = ch }

func (m *EdgeMover) SetStream(h StreamHandle, edgeRow int32, nodeRowFor func(id string) (int32, bool), buildFrame func(tick uint32, sx, sy, sz, ex, ey, ez float32, label string, events []rowevent.RowEvent) []byte) {
	m.streamOut = h
	m.edgeRow = edgeRow
	m.nodeRowFor = nodeRowFor
	m.buildFrame = buildFrame
}

func (m *EdgeMover) Select(ctx context.Context, on bool) {
	msg := movemsg.Msg{Kind: movemsg.KindSelect, Bool: on}
	if ctx == nil {
		m.extIn <- msg
		return
	}
	select {
	case m.extIn <- msg:
	case <-ctx.Done():
	}
}

func (m *EdgeMover) TrySendFromSrc(msg movemsg.Msg) bool {
	select {
	case m.srcIn <- msg:
		return true
	default:
		return false
	}
}

func (m *EdgeMover) TrySendFromDst(msg movemsg.Msg) bool {
	select {
	case m.dstIn <- msg:
		return true
	default:
		return false
	}
}

func (m *EdgeMover) SendSteps(steps int) {
	if m.stepsIn == nil {
		return
	}
	select {
	case m.stepsIn <- steps:
	default:
		select {
		case <-m.stepsIn:
		default:
		}
		select {
		case m.stepsIn <- steps:
		default:
		}
	}
}
