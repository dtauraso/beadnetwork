package streamframe

import (
	"encoding/binary"
	"fmt"

	B "github.com/dtauraso/wirefold/Buffer"
)

type NodeStreamFrame struct {
	Tick uint32

	NodeRow int32

	NodeID int32

	CX, CY, CZ float32

	Radius, SphereR float32

	VRX, VRY, VRZ float32
	FRX, FRY, FRZ float32

	PolePhi, PoleTheta float32

	RingAxisPhi, RingAxisTheta float32

	RingTubeRadius float32

	RingPoints []float32

	TopTiltVectorLen float32

	TopTiltVectorIdx int32

	TopTiltVectorPhi float32

	BottomTiltVectorPhi float32

	CoplanarNormalPhi float32

	ReceivedVectorLen float32

	ReceivedVectorPhi float32

	Selected, KindID, Hovered, LatchedSel uint8

	LatticePoints uint8

	RoundsToParallel, MsgsToParallel int32

	Label string

	Events []StreamEvent
}

func BuildNodeStreamFrame(f NodeStreamFrame) []byte {
	labelBytes := []byte(f.Label)

	if len(f.RingPoints)%3 != 0 {
		panic(fmt.Sprintf(
			"BuildNodeStreamFrame: RingPoints has %d floats for node row %d — must be a whole number of XYZ triples",
			len(f.RingPoints), f.NodeRow))
	}
	ringPointCount := len(f.RingPoints) / 3

	size := B.BufNodeStreamFrameHeaderSize + B.BufNodeStride + len(labelBytes) + ringPointCount*B.BufRingPointStride
	buf := make([]byte, size)
	off := 0
	binary.LittleEndian.PutUint32(buf[off:], f.Tick)
	off += 4
	binary.LittleEndian.PutUint32(buf[off:], uint32(len(labelBytes)))
	off += 4
	binary.LittleEndian.PutUint32(buf[off:], uint32(ringPointCount))
	off += 4

	B.SetNodeRow(buf[off:off+B.BufNodeStride], 0, f.NodeID, f.CX, f.CY, f.CZ, f.Radius, f.SphereR, f.VRX, f.VRY, f.VRZ, f.FRX, f.FRY, f.FRZ,
		f.PolePhi, f.PoleTheta, f.RingAxisPhi, f.RingAxisTheta, f.RingTubeRadius, f.TopTiltVectorLen, f.TopTiltVectorIdx,
		f.TopTiltVectorPhi, f.BottomTiltVectorPhi,
		f.CoplanarNormalPhi, f.ReceivedVectorLen, f.ReceivedVectorPhi,
		f.Selected, f.KindID, 0, uint32(len(labelBytes)), f.Hovered, f.LatchedSel, f.LatticePoints, f.RoundsToParallel, f.MsgsToParallel)
	off += B.BufNodeStride

	copy(buf[off:off+len(labelBytes)], labelBytes)
	off += len(labelBytes)

	for i := 0; i < ringPointCount; i++ {
		rowOff := off + i*B.BufRingPointStride
		B.SetRingPointRow(buf[rowOff:rowOff+B.BufRingPointStride], 0, f.RingPoints[i*3], f.RingPoints[i*3+1], f.RingPoints[i*3+2])
	}
	off += ringPointCount * B.BufRingPointStride

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
