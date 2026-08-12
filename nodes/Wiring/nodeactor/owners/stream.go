package owners

import (
	"encoding/binary"

	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/nodeframe"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/streamclaim"
	"github.com/dtauraso/wirefold/nodes/rowevent"
)

type Stream struct {
	streamOut streamclaim.StreamHandle
	nodeRow   int32
	kindID    uint8

	buildFrame nodeframe.NodeFrameBuilder

	// selfEvents carries row events raised by a SELF-DRIVEN kind's own goroutine
	// (breadcrumbs, its opening geometry emit) to the geometry peer, which is the one
	// goroutine allowed to write this stream. Writing it from both would race.
	selfEvents chan []rowevent.RowEvent
}

// PostSelfEvents hands row events from a self-driven kind's own goroutine to the geometry
// peer, non-blocking: a breadcrumb is diagnostics, and dropping one is strictly better than
// stalling the kind's loop on a peer.
func (s *Stream) PostSelfEvents(events []rowevent.RowEvent) {
	if s.selfEvents == nil {
		return
	}
	select {
	case s.selfEvents <- events:
	default:
	}
}

// DrainSelfEvents collects whatever a self-driven kind posted since the last pass, for the
// geometry peer to carry on its next frame.
func (s *Stream) DrainSelfEvents() []rowevent.RowEvent {
	var out []rowevent.RowEvent
	for {
		select {
		case ev := <-s.selfEvents:
			out = append(out, ev...)
		default:
			return out
		}
	}
}

func (s *Stream) NodeRow() int32 { return s.nodeRow }

func (s *Stream) KindID() uint8 { return s.kindID }

func (s *Stream) Ready() bool { return s.streamOut.Ok() && s.buildFrame != nil }

func (s *Stream) SetStream(streamOut streamclaim.StreamHandle, row int32, kindID uint8, buildFrame nodeframe.NodeFrameBuilder) {
	if s.selfEvents == nil {
		s.selfEvents = make(chan []rowevent.RowEvent, selfEventDepth)
	}
	s.streamOut = streamOut
	s.nodeRow = row
	s.kindID = kindID
	s.buildFrame = buildFrame
}

func (s *Stream) WriteFrame(input nodeframe.NodeFrameInput) {
	if !s.Ready() {
		return
	}
	frame := s.buildFrame(input)
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], uint32(len(frame)))
	_, _ = s.streamOut.Write(hdr[:])
	_, _ = s.streamOut.Write(frame)
}
