package main

import (
	"encoding/hex"

	streamframe "github.com/dtauraso/wirefold/Buffer/streamframe"
)

func buildNodeFrame() nodeFrameFixture {
	f := nodeFrameFixture{
		Tick: 4242, NodeRow: 7, NodeId: 8,
		CX: 11.5, CY: -12.25, CZ: 13.125, Radius: 14.0625, SphereR: 200.5,
		VRX: 21.5, VRY: 22.25, VRZ: 23.125, FRX: 24.0625, FRY: 25.5, FRZ: 26.25,
		PolePhi: 2.1, PoleTheta: -1.3,
		RingAxisPhi: 1.4, RingAxisTheta: 0.7,
		TopTiltVectorLen: 9.5, TopTiltVectorIdx: 2, TopTiltVectorPhi: 0.5,
		CoplanarNormalPhi:   0.55,
		BottomTiltVectorPhi: 2.9,
		ReceivedVectorLen:   8.75, ReceivedVectorPhi: 0.25,
		Selected: 1, KindID: 3, Hovered: 1, LatchedSel: 0,
		LatticePoints:    12,
		RoundsToParallel: 3,
		MsgsToParallel:   6,
		Label:            "widgetNode",
		ChainBeads: []chainBeadFixture{
			{OX: 61.5, OY: -62.25, OZ: 63.125, Lit: 1, LitValue: 1},
			{OX: -64.5, OY: 65.25, OZ: -66.125},
		},
	}

	chainOX := make([]float32, len(f.ChainBeads))
	chainOY := make([]float32, len(f.ChainBeads))
	chainOZ := make([]float32, len(f.ChainBeads))
	chainLit := make([]uint8, len(f.ChainBeads))
	chainLitVal := make([]int32, len(f.ChainBeads))
	for i, cb := range f.ChainBeads {
		chainOX[i], chainOY[i], chainOZ[i], chainLit[i], chainLitVal[i] = cb.OX, cb.OY, cb.OZ, cb.Lit, cb.LitValue
	}

	raw := streamframe.BuildNodeStreamFrame(streamframe.NodeStreamFrame{
		Tick:                f.Tick,
		NodeRow:             f.NodeRow,
		NodeID:              f.NodeId,
		CX:                  f.CX,
		CY:                  f.CY,
		CZ:                  f.CZ,
		Radius:              f.Radius,
		SphereR:             f.SphereR,
		VRX:                 f.VRX,
		VRY:                 f.VRY,
		VRZ:                 f.VRZ,
		FRX:                 f.FRX,
		FRY:                 f.FRY,
		FRZ:                 f.FRZ,
		PolePhi:             f.PolePhi,
		PoleTheta:           f.PoleTheta,
		RingAxisPhi:         f.RingAxisPhi,
		RingAxisTheta:       f.RingAxisTheta,
		TopTiltVectorLen:    f.TopTiltVectorLen,
		TopTiltVectorIdx:    f.TopTiltVectorIdx,
		TopTiltVectorPhi:    f.TopTiltVectorPhi,
		BottomTiltVectorPhi: f.BottomTiltVectorPhi,
		CoplanarNormalPhi:   f.CoplanarNormalPhi,
		ReceivedVectorLen:   f.ReceivedVectorLen,
		ReceivedVectorPhi:   f.ReceivedVectorPhi,
		Selected:            f.Selected,
		KindID:              f.KindID,
		Hovered:             f.Hovered,
		LatchedSel:          f.LatchedSel,
		LatticePoints:       f.LatticePoints,
		RoundsToParallel:    f.RoundsToParallel,
		MsgsToParallel:      f.MsgsToParallel,
		Label:               f.Label,
		Events:              nil,
	})
	f.Hex = hex.EncodeToString(raw)
	return f
}

func buildEdgeFrame() edgeFrameFixture {
	f := edgeFrameFixture{
		Tick: 8181, SX: 12.5, SY: -13.25, SZ: 14.125, EX: 34.5, EY: -35.25, EZ: 36.125,
		SrcNodeRow: 3,
		Label:      "edgeLabel",
	}
	raw := streamframe.BuildEdgeStreamFrame(f.Tick, f.SX, f.SY, f.SZ, f.EX, f.EY, f.EZ, f.SrcNodeRow, f.Label,
		[]streamframe.EdgeBead{{X: 1.5, Y: 2.25, Z: 3.125, Value: 1}}, nil)
	f.Hex = hex.EncodeToString(raw)
	return f
}

func buildInteriorFrame() interiorFrameFixture {
	f := interiorFrameFixture{
		Tick:    5151,
		Present: []int{1, 0, 1, 1},
		Value:   []int32{-11, 0, 22, -33},
		OX:      []float32{1.5, 0, -3.125, 4.0625},
		OY:      []float32{-1.5, 0, 3.125, -4.0625},
		OZ:      []float32{2.5, 0, -5.125, 6.0625},
	}
	present := make([]uint8, len(f.Present))
	for i, p := range f.Present {
		present[i] = uint8(p)
	}
	raw := streamframe.BuildInteriorStreamFrame(f.Tick, present, f.Value, f.OX, f.OY, f.OZ, nil)
	f.Hex = hex.EncodeToString(raw)
	return f
}
