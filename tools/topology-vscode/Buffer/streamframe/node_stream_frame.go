package streamframe

import (
	"encoding/binary"
	"fmt"

	B "github.com/dtauraso/wirefold/tools/topology-vscode/Buffer"
)

type NodeStreamFrame struct {
	Tick uint32

	NodeRow int32

	NodeID int32

	IndexR, IndexPhi, IndexTheta int32
	HasPos                       uint8

	Radius float32

	LabelAnchorX, LabelAnchorY, LabelAnchorZ float32

	PolePhi, PoleTheta float32

	RingMatrix [16]float32

	TopTiltVectorLen float32

	TopTiltVectorIdx int32

	Selected, KindID, Hovered, LatchedSel uint8

	LatticePoints uint8

	RoundsToParallel, MsgsToParallel int32

	DragRLocked, DragPhiLocked uint8
	DragThetaMax               float32
	DragActive                 uint8
	HasKindRule                uint8
	KindRuleActive             uint8
	PoleRingR                  float32

	SelfRLocked, SelfPhiLocked uint8
	SelfThetaMax               float32
	SelfActive                 uint8

	RuleGroupID, RuleGroupSize int32

	Label string

	TiltArrows []TiltArrow

	ChannelVectors []ChannelVector

	Events []StreamEvent
}

type TiltArrow struct {
	Received uint8
	Shaft    [16]float32
	Head     [16]float32
}

type ChannelVector struct {
	Shaft [16]float32
	Head  [16]float32
}

func BuildNodeStreamFrame(f NodeStreamFrame) []byte {
	labelBytes := []byte(f.Label)

	arrowsSize := len(f.TiltArrows) * B.BufTiltArrowStride
	channelsSize := len(f.ChannelVectors) * B.BufChannelVectorStride
	size := B.BufNodeStreamFrameHeaderSize + B.BufNodeStride + len(labelBytes) + arrowsSize + channelsSize
	buf := make([]byte, size)
	off := 0
	binary.LittleEndian.PutUint32(buf[off:], f.Tick)
	off += 4
	binary.LittleEndian.PutUint32(buf[off:], uint32(len(labelBytes)))
	off += 4
	binary.LittleEndian.PutUint32(buf[off:], uint32(len(f.TiltArrows)))
	off += 4
	binary.LittleEndian.PutUint32(buf[off:], uint32(len(f.ChannelVectors)))
	off += 4

	m := f.RingMatrix
	B.SetNodeRow(buf[off:off+B.BufNodeStride], 0, f.NodeID, f.IndexR, f.IndexPhi, f.IndexTheta, f.HasPos, f.Radius,
		f.LabelAnchorX, f.LabelAnchorY, f.LabelAnchorZ,
		f.PolePhi, f.PoleTheta,
		m[0], m[1], m[2], m[3], m[4], m[5], m[6], m[7], m[8], m[9], m[10], m[11], m[12], m[13], m[14], m[15],
		f.TopTiltVectorLen, f.TopTiltVectorIdx,
		f.Selected, f.KindID, 0, uint32(len(labelBytes)), f.Hovered, f.LatchedSel, f.LatticePoints, f.RoundsToParallel, f.MsgsToParallel,
		f.DragRLocked, f.DragPhiLocked, f.DragThetaMax, f.DragActive, f.HasKindRule, f.KindRuleActive,
		f.PoleRingR,
		f.SelfRLocked, f.SelfPhiLocked, f.SelfThetaMax, f.SelfActive,
		f.RuleGroupID, f.RuleGroupSize)
	off += B.BufNodeStride

	copy(buf[off:off+len(labelBytes)], labelBytes)
	off += len(labelBytes)

	for _, a := range f.TiltArrows {
		s, h := a.Shaft, a.Head
		B.SetTiltArrowRow(buf[off:off+B.BufTiltArrowStride], 0, a.Received,
			s[0], s[1], s[2], s[3], s[4], s[5], s[6], s[7], s[8], s[9], s[10], s[11], s[12], s[13], s[14], s[15],
			h[0], h[1], h[2], h[3], h[4], h[5], h[6], h[7], h[8], h[9], h[10], h[11], h[12], h[13], h[14], h[15])
		off += B.BufTiltArrowStride
	}

	for _, c := range f.ChannelVectors {
		s, h := c.Shaft, c.Head
		B.SetChannelVectorRow(buf[off:off+B.BufChannelVectorStride], 0,
			s[0], s[1], s[2], s[3], s[4], s[5], s[6], s[7], s[8], s[9], s[10], s[11], s[12], s[13], s[14], s[15],
			h[0], h[1], h[2], h[3], h[4], h[5], h[6], h[7], h[8], h[9], h[10], h[11], h[12], h[13], h[14], h[15])
		off += B.BufChannelVectorStride
	}

	if off != size {
		panic(fmt.Sprintf(
			"BuildNodeStreamFrame: packed %d bytes for node row %d but allocated %d — the section walk and the size formula disagree; a section was added, reordered, or resized in one of the two and not the other",
			off, f.NodeRow, size))
	}

	return append(buf, BuildEventsSection(f.Events)...)
}

func BuildInteriorStreamFrame(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []StreamEvent) []byte {
	n := len(present)

	for _, s := range []struct {
		name string
		n    int
	}{{"value", len(value)}, {"ox", len(ox)}, {"oy", len(oy)}, {"oz", len(oz)}} {
		if s.n != n {
			panic(fmt.Sprintf(
				"BuildInteriorStreamFrame: %d present slots but %s has %d entries — the interior slot slices are parallel, one entry per slot",
				n, s.name, s.n))
		}
	}
	buf := make([]byte, B.BufInteriorStreamFrameHeaderSize+n*B.BufInteriorStride)
	binary.LittleEndian.PutUint32(buf[0:], tick)
	interiorBuf := buf[4:]
	for i := 0; i < n; i++ {
		B.SetInteriorRow(interiorBuf, i, present[i], value[i], ox[i], oy[i], oz[i])
	}
	return append(buf, BuildEventsSection(events)...)
}
