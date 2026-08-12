package nodeactor

import (
	"encoding/binary"

	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/nodeframe"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/streamclaim"
)

func (s *nodeStream) NodeRow() int32 { return s.nodeRow }

func (s *nodeStream) KindID() uint8 { return s.kindID }

func (s *nodeStream) Ready() bool { return s.streamOut.Ok() && s.buildFrame != nil }

func (s *nodeStream) SetStream(streamOut streamclaim.StreamHandle, row int32, kindID uint8, buildFrame nodeframe.NodeFrameBuilder) {
	s.streamOut = streamOut
	s.nodeRow = row
	s.kindID = kindID
	s.buildFrame = buildFrame
}

func (s *nodeStream) WriteFrame(input nodeframe.NodeFrameInput) {
	if !s.Ready() {
		return
	}
	frame := s.buildFrame(input)
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], uint32(len(frame)))
	_, _ = s.streamOut.Write(hdr[:])
	_, _ = s.streamOut.Write(frame)
}
