package nodeactor

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/nodeframe"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/streamclaim"
)

func (s *nodeStream) NodeRow() int32 { return s.nodeRow }

func (s *nodeStream) SetStream(streamOut streamclaim.StreamHandle, row int32, kindID uint8, buildFrame nodeframe.NodeFrameBuilder) {
	s.streamOut = streamOut
	s.nodeRow = row
	s.kindID = kindID
	s.buildFrame = buildFrame
}
