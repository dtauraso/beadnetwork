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

	PoleTheta, PolePhi float32

	RingAxisTheta, RingAxisPhi float32

	TopTiltVectorLen float32

	TopTiltVectorTheta float32

	BottomTiltVectorTheta float32

	CoplanarNormalTheta float32

	ReceivedVectorLen float32

	ReceivedVectorTheta float32

	Selected, KindID, Hovered, LatchedSel uint8

	LatticePoints uint8

	RoundsToParallel, MsgsToParallel int32

	Label string

	ChainBeadOX, ChainBeadOY, ChainBeadOZ []float32
	ChainBeadLit                          []uint8
	ChainBeadLitValue                     []int32

	// OutPole* is one unit direction per OUTGOING neighbour — the pole of that
	// edge's frame, pointing along the path vector the node stores for it.
	OutPoleDX, OutPoleDY, OutPoleDZ []float32

	Events []StreamEvent
}

func BuildNodeStreamFrame(f NodeStreamFrame) []byte {
	labelBytes := []byte(f.Label)
	chainBeadCount := len(f.ChainBeadOX)

	for _, s := range []struct {
		name string
		n    int
	}{{"ChainBeadOY", len(f.ChainBeadOY)}, {"ChainBeadOZ", len(f.ChainBeadOZ)}, {"ChainBeadLit", len(f.ChainBeadLit)}, {"ChainBeadLitValue", len(f.ChainBeadLitValue)}} {
		if s.n != chainBeadCount {
			panic(fmt.Sprintf(
				"BuildNodeStreamFrame: node row %d has %d chain-bead OX entries but %s has %d — the chain-bead slices are parallel, one entry per bead",
				f.NodeRow, chainBeadCount, s.name, s.n))
		}
	}

	outPoleCount := len(f.OutPoleDX)
	for _, s := range []struct {
		name string
		n    int
	}{{"OutPoleDY", len(f.OutPoleDY)}, {"OutPoleDZ", len(f.OutPoleDZ)}} {
		if s.n != outPoleCount {
			panic(fmt.Sprintf(
				"BuildNodeStreamFrame: node row %d has %d out-pole DX entries but %s has %d — the out-pole slices are parallel, one entry per outgoing neighbour",
				f.NodeRow, outPoleCount, s.name, s.n))
		}
	}

	size := B.BufNodeStreamFrameHeaderSize + B.BufNodeStride + len(labelBytes) +
		chainBeadCount*B.BufChainBeadStride + outPoleCount*B.BufOutPoleStride
	buf := make([]byte, size)
	off := 0
	binary.LittleEndian.PutUint32(buf[off:], f.Tick)
	off += 4
	binary.LittleEndian.PutUint32(buf[off:], uint32(len(labelBytes)))
	off += 4
	binary.LittleEndian.PutUint32(buf[off:], uint32(chainBeadCount))
	off += 4
	binary.LittleEndian.PutUint32(buf[off:], uint32(outPoleCount))
	off += 4

	B.SetNodeRow(buf[off:off+B.BufNodeStride], 0, f.NodeID, f.CX, f.CY, f.CZ, f.Radius, f.SphereR, f.VRX, f.VRY, f.VRZ, f.FRX, f.FRY, f.FRZ,
		f.PoleTheta, f.PolePhi, f.RingAxisTheta, f.RingAxisPhi, f.TopTiltVectorLen, f.TopTiltVectorTheta, f.BottomTiltVectorTheta, f.CoplanarNormalTheta, f.ReceivedVectorLen, f.ReceivedVectorTheta, f.Selected, f.KindID, 0, uint32(len(labelBytes)), f.Hovered, f.LatchedSel, f.LatticePoints, f.RoundsToParallel, f.MsgsToParallel)
	off += B.BufNodeStride

	copy(buf[off:off+len(labelBytes)], labelBytes)
	off += len(labelBytes)

	for i := 0; i < chainBeadCount; i++ {
		rowOff := off + i*B.BufChainBeadStride
		B.SetChainBeadRow(buf[rowOff:rowOff+B.BufChainBeadStride], 0, f.ChainBeadOX[i], f.ChainBeadOY[i], f.ChainBeadOZ[i], f.ChainBeadLit[i], f.ChainBeadLitValue[i])
	}
	off += chainBeadCount * B.BufChainBeadStride

	for i := 0; i < outPoleCount; i++ {
		rowOff := off + i*B.BufOutPoleStride
		B.SetOutPoleRow(buf[rowOff:rowOff+B.BufOutPoleStride], 0, f.OutPoleDX[i], f.OutPoleDY[i], f.OutPoleDZ[i])
	}
	off += outPoleCount * B.BufOutPoleStride

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
