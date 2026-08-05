// gen-stream-fixture is the Go SIDE of the Go->TS stream-frame cross-language fixture
// (mirrors tools/topology-vscode/scripts/gen-input-fixture-src.ts, which is the TS side of
// the TS->Go input-record fixture — see nodes/Wiring/input_fixture_test.go's header for the
// gap this closes).
//
// It builds real per-owner stream-frame bytes with the REAL production packers
// (Buffer.BuildNodeStreamFrame / BuildEdgeStreamFrame / BuildInteriorStreamFrame), using
// distinctive, all-different field values, and emits a JSON fixture:
//
//	{"nodeFrame": {...fields..., "hex": "..."},
//	 "edgeFrame": {...fields..., "hex": "..."},
//	 "interiorFrame": {...fields..., "hex": "..."}}
//
// The committed copy lives at tools/topology-vscode/test/fixtures/stream_fixture.json,
// regenerated via `go run ./tools/gen-stream-fixture <outPath>` (outPath defaults to that
// path, resolved from the repo root this binary is invoked from).
//
// tools/topology-vscode/test/stream-fixture.test.ts decodes the fixture's hex with the REAL
// TS decoders (decodeNodeStreamFrame/decodeEdgeStreamFrame/decodeInteriorStreamFrame in
// buffer-decode.ts) and asserts every field — the actual cross-language byte-level
// agreement check. It also regenerates this fixture live (via `go run`) and diffs it
// against the committed copy, so a stale fixture fails loudly instead of silently testing
// its own past self (same freshness shape as TestInputFixtureFreshness, mirrored onto the
// opposite direction).
package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/Buffer"
)

// chainBeadFixture is ONE node-local chain-bead offset row (Buffer bufLayoutChainBead).
type chainBeadFixture struct {
	OX       float32 `json:"ox"`
	OY       float32 `json:"oy"`
	OZ       float32 `json:"oz"`
	Lit      uint8   `json:"lit"`
	LitValue int32   `json:"litValue"`
}

type nodeFrameFixture struct {
	Tick      uint32  `json:"tick"`
	NodeRow   int32   `json:"nodeRow"`
	NodeId    int32   `json:"nodeId"`
	CX        float32 `json:"cx"`
	CY        float32 `json:"cy"`
	CZ        float32 `json:"cz"`
	Radius    float32 `json:"radius"`
	SphereR   float32 `json:"sphereR"`
	VRX       float32 `json:"vrx"`
	VRY       float32 `json:"vry"`
	VRZ       float32 `json:"vrz"`
	FRX       float32 `json:"frx"`
	FRY       float32 `json:"fry"`
	FRZ       float32 `json:"frz"`
	PoleTheta float32 `json:"poleTheta"`
	PolePhi   float32 `json:"polePhi"`
	// The DRAWN ring's axis, distinct from the navigation pole above — see
	// Buffer/layout.go's RingAxisTheta/RingAxisPhi for why they are two values.
	RingAxisTheta float32 `json:"ringAxisTheta"`
	RingAxisPhi   float32 `json:"ringAxisPhi"`
	// The node's own drawn vector length along that axis; 0 means it draws none.
	TiltVectorLen float32 `json:"tiltVectorLen"`
	// The vector's OWN direction — separate from RingAxisTheta/Phi above (Buffer/layout.go's
	// TiltVectorTheta/TiltVectorPhi).
	TiltVectorTheta float32 `json:"tiltVectorTheta"`
	TiltVectorPhi   float32 `json:"tiltVectorPhi"`
	// The SECOND vector's direction — a quarter turn from the first, in the ring's plane.
	CoplanarNormalTheta float32 `json:"coplanarNormalTheta"`
	CoplanarNormalPhi   float32 `json:"coplanarNormalPhi"`
	// The THIRD vector: the direction last received on this node's tilt-vector channel
	// (Buffer/layout.go's ReceivedVectorLen/Theta/Phi); 0 length means nothing received yet.
	ReceivedVectorLen   float32            `json:"receivedVectorLen"`
	ReceivedVectorTheta float32            `json:"receivedVectorTheta"`
	ReceivedVectorPhi   float32            `json:"receivedVectorPhi"`
	Selected            uint8              `json:"selected"`
	KindID              uint8              `json:"kindId"`
	Hovered             uint8              `json:"hovered"`
	LatchedSel          uint8              `json:"latchedSel"`
	ChainBeads          []chainBeadFixture `json:"chainBeads"`
	Label               string             `json:"label"`
	Hex                 string             `json:"hex"`
}

type edgeFrameFixture struct {
	Tick     uint32  `json:"tick"`
	SX       float32 `json:"sx"`
	SY       float32 `json:"sy"`
	SZ       float32 `json:"sz"`
	EX       float32 `json:"ex"`
	EY       float32 `json:"ey"`
	EZ       float32 `json:"ez"`
	Selected uint8   `json:"selected"`
	Label    string  `json:"label"`
	Hex      string  `json:"hex"`
}

type interiorFrameFixture struct {
	Tick uint32 `json:"tick"`
	// Present is []int, not []uint8: Go's encoding/json marshals []uint8 as a base64
	// string (it treats it as []byte), which would silently strip the fixture's own
	// human-readable expected values. []int marshals as a plain JSON number array.
	Present []int     `json:"present"`
	Value   []int32   `json:"value"`
	OX      []float32 `json:"ox"`
	OY      []float32 `json:"oy"`
	OZ      []float32 `json:"oz"`
	Hex     string    `json:"hex"`
}

type streamFixture struct {
	NodeFrame     nodeFrameFixture     `json:"nodeFrame"`
	EdgeFrame     edgeFrameFixture     `json:"edgeFrame"`
	InteriorFrame interiorFrameFixture `json:"interiorFrame"`
}

func buildNodeFrame() nodeFrameFixture {
	f := nodeFrameFixture{
		Tick: 4242, NodeRow: 7, NodeId: 8,
		CX: 11.5, CY: -12.25, CZ: 13.125, Radius: 14.0625, SphereR: 200.5,
		VRX: 21.5, VRY: 22.25, VRZ: 23.125, FRX: 24.0625, FRY: 25.5, FRZ: 26.25,
		PoleTheta: 2.1, PolePhi: -1.3,
		RingAxisTheta: 1.4, RingAxisPhi: 0.7,
		TiltVectorLen: 9.5, TiltVectorTheta: 0.5, TiltVectorPhi: -0.9,
		CoplanarNormalTheta: 0.55, CoplanarNormalPhi: -0.35,
		ReceivedVectorLen: 8.75, ReceivedVectorTheta: 0.25, ReceivedVectorPhi: -0.15,
		Selected: 1, KindID: 3, Hovered: 1, LatchedSel: 0,
		Label: "widgetNode",
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

	raw := Buffer.BuildNodeStreamFrame(
		f.Tick, f.NodeRow, f.NodeId,
		f.CX, f.CY, f.CZ, f.Radius, f.SphereR,
		f.VRX, f.VRY, f.VRZ, f.FRX, f.FRY, f.FRZ,
		f.PoleTheta, f.PolePhi, f.RingAxisTheta, f.RingAxisPhi, f.TiltVectorLen, f.TiltVectorTheta, f.TiltVectorPhi, f.CoplanarNormalTheta, f.CoplanarNormalPhi,
		f.ReceivedVectorLen, f.ReceivedVectorTheta, f.ReceivedVectorPhi,
		f.Selected, f.KindID, f.Hovered, f.LatchedSel,
		f.Label,
		chainOX, chainOY, chainOZ, chainLit, chainLitVal,
		nil,
	)
	f.Hex = hex.EncodeToString(raw)
	return f
}

func buildEdgeFrame() edgeFrameFixture {
	f := edgeFrameFixture{
		Tick: 8181, SX: 12.5, SY: -13.25, SZ: 14.125, EX: 34.5, EY: -35.25, EZ: 36.125,
		Selected: 1, Label: "edgeLabel",
	}
	raw := Buffer.BuildEdgeStreamFrame(f.Tick, f.SX, f.SY, f.SZ, f.EX, f.EY, f.EZ, f.Selected, f.Label, nil)
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
	raw := Buffer.BuildInteriorStreamFrame(f.Tick, present, f.Value, f.OX, f.OY, f.OZ, nil)
	f.Hex = hex.EncodeToString(raw)
	return f
}

func main() {
	outPath := "tools/topology-vscode/test/fixtures/stream_fixture.json"
	if len(os.Args) > 1 {
		outPath = os.Args[1]
	}

	fx := streamFixture{
		NodeFrame:     buildNodeFrame(),
		EdgeFrame:     buildEdgeFrame(),
		InteriorFrame: buildInteriorFrame(),
	}

	out, err := json.MarshalIndent(fx, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "gen-stream-fixture: marshal failed: %v\n", err)
		os.Exit(1)
	}
	out = append(out, '\n')
	if err := os.WriteFile(outPath, out, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "gen-stream-fixture: writing %s failed: %v\n", outPath, err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "wrote %s\n", outPath)
}
