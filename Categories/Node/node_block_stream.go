package Node

import (
	"github.com/dtauraso/wirefold/Categories/Node/nodeframe"
)

type Stream struct {
	nodeRow int32
	kindID  uint8

	writeValues nodeframe.NodeFrameBuilder
}

func (s *Stream) NodeRow() int32 { return s.nodeRow }

func (s *Stream) KindID() uint8 { return s.kindID }

func (s *Stream) Ready() bool { return s.writeValues != nil }

func (s *Stream) SetStream(row int32, kindID uint8, writeValues nodeframe.NodeFrameBuilder) {
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
