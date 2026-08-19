package main

import (
	"encoding/hex"

	streamframe "github.com/dtauraso/wirefold/src/Buffer/streamframe"
)

func buildNodeFrame() nodeFrameFixture {
	f := nodeFrameFixture{
		Tick: 4242, NodeRow: 7, NodeId: 8,
		IndexR: 3, IndexPhi: 5, IndexTheta: 7, HasPos: 1, Radius: 14.0625,
		PolePhi: 2.1, PoleTheta: -1.3,
		TopTiltVectorLen: 9.5, TopTiltVectorIdx: 2,
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
		Tick:             f.Tick,
		NodeRow:          f.NodeRow,
		NodeID:           f.NodeId,
		IndexR:           f.IndexR,
		IndexPhi:         f.IndexPhi,
		IndexTheta:       f.IndexTheta,
		HasPos:           f.HasPos,
		Radius:           f.Radius,
		PolePhi:          f.PolePhi,
		PoleTheta:        f.PoleTheta,
		TopTiltVectorLen: f.TopTiltVectorLen,
		TopTiltVectorIdx: f.TopTiltVectorIdx,
		Selected:         f.Selected,
		KindID:           f.KindID,
		Hovered:          f.Hovered,
		LatchedSel:       f.LatchedSel,
		LatticePoints:    f.LatticePoints,
		RoundsToParallel: f.RoundsToParallel,
		MsgsToParallel:   f.MsgsToParallel,
		Label:            f.Label,
		Events:           nil,
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
	raw := streamframe.BuildEdgeStreamFrame(f.Tick, nil)
	f.Hex = hex.EncodeToString(raw)
	return f
}

func buildBeadFrame() beadFrameFixture {
	f := beadFrameFixture{
		Tick:    8383,
		NodeRow: 3,
		Beads: []edgeBeadFixture{
			{X: 1.5, Y: 2.25, Z: 3.125, Value: 1, EdgeRow: 0},
			{X: -4.5, Y: 5.25, Z: -6.125, Value: 2, EdgeRow: 2},
		},
	}
	f.Hex = hex.EncodeToString(streamframe.BuildBeadStreamFrame(f.Tick, f.NodeRow, nil))
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
	_ = present
	raw := streamframe.BuildInteriorStreamFrame(f.Tick, nil)
	f.Hex = hex.EncodeToString(raw)
	return f
}
