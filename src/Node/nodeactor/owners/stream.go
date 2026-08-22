package owners

import (
	T "github.com/dtauraso/wirefold/src/Trace"

	"github.com/dtauraso/wirefold/src/Node/nodeactor/nodeframe"
)

type Stream struct {
	nodeRow int32
	kindID  uint8

	writeValues nodeframe.NodeFrameBuilder

	selfEvents chan []T.RowEvent
}

func (s *Stream) PostSelfEvents(events []T.RowEvent) {
	if s.selfEvents == nil {
		return
	}
	select {
	case s.selfEvents <- events:
	default:
	}
}

func (s *Stream) DrainSelfEvents() []T.RowEvent {
	var out []T.RowEvent
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

func (s *Stream) Ready() bool { return s.writeValues != nil }

func (s *Stream) SetStream(row int32, kindID uint8, writeValues nodeframe.NodeFrameBuilder) {
	if s.selfEvents == nil {
		s.selfEvents = make(chan []T.RowEvent, selfEventDepth)
	}
	s.nodeRow = row
	s.kindID = kindID
	s.writeValues = writeValues
}

func (s *Stream) WriteFrame(input nodeframe.NodeFrameInput) {
	if !s.Ready() {
		return
	}
	s.writeValues(input)
}
