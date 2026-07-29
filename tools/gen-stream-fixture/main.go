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

type portFixture struct {
	NodeRow  int32   `json:"nodeRow"`
	DX       float32 `json:"dx"`
	DY       float32 `json:"dy"`
	DZ       float32 `json:"dz"`
	PX       float32 `json:"px"`
	PY       float32 `json:"py"`
	PZ       float32 `json:"pz"`
	IsInput  uint8   `json:"isInput"`
	Hovered  uint8   `json:"hovered"`
	NameLen  uint32  `json:"nameLen"`
	NameText string  `json:"name"`
}

type layoutLinkFixture struct {
	DstNodeRow int32 `json:"dstNodeRow"`
}

type nodeFrameFixture struct {
	Tick             uint32              `json:"tick"`
	NodeRow          int32               `json:"nodeRow"`
	CX               float32             `json:"cx"`
	CY               float32             `json:"cy"`
	CZ               float32             `json:"cz"`
	Radius           float32             `json:"radius"`
	SphereR          float32             `json:"sphereR"`
	VRX              float32             `json:"vrx"`
	VRY              float32             `json:"vry"`
	VRZ              float32             `json:"vrz"`
	FRX              float32             `json:"frx"`
	FRY              float32             `json:"fry"`
	FRZ              float32             `json:"frz"`
	Selected         uint8               `json:"selected"`
	KindID           uint8               `json:"kindId"`
	Hovered          uint8               `json:"hovered"`
	LatchedSel       uint8               `json:"latchedSel"`
	GotDragMsg       uint8               `json:"gotDragMsg"`
	DragDeltaA       int32               `json:"dragDeltaA"`
	DragDeltaB       int32               `json:"dragDeltaB"`
	DragDeltaC       int32               `json:"dragDeltaC"`
	DragRequantCount int32               `json:"dragRequantCount"`
	GotForwardMsg    uint8               `json:"gotForwardMsg"`
	ForwardDeltaA    int32               `json:"forwardDeltaA"`
	ForwardDeltaB    int32               `json:"forwardDeltaB"`
	ForwardDeltaC    int32               `json:"forwardDeltaC"`
	ForwardFromRow   int32               `json:"forwardFromRow"`
	CascadeRelay     uint8               `json:"cascadeRelay"`
	Label            string              `json:"label"`
	Ports            []portFixture       `json:"ports"`
	LayoutLinks      []layoutLinkFixture `json:"layoutLinks"`
	Hex              string              `json:"hex"`
}

type edgeFrameFixture struct {
	Tick       uint32    `json:"tick"`
	SrcPortRow int32     `json:"srcPortRow"`
	DstPortRow int32     `json:"dstPortRow"`
	Selected   uint8     `json:"selected"`
	Label      string    `json:"label"`
	BeadVal    []int32   `json:"beadVal"`
	BeadX      []float32 `json:"beadX"`
	BeadY      []float32 `json:"beadY"`
	BeadZ      []float32 `json:"beadZ"`
	Hex        string    `json:"hex"`
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
		Tick: 4242, NodeRow: 7,
		CX: 11.5, CY: -12.25, CZ: 13.125, Radius: 14.0625, SphereR: 200.5,
		VRX: 21.5, VRY: 22.25, VRZ: 23.125, FRX: 24.0625, FRY: 25.5, FRZ: 26.25,
		Selected: 1, KindID: 3, Hovered: 1, LatchedSel: 0, GotDragMsg: 1,
		DragDeltaA: -101, DragDeltaB: 102, DragDeltaC: -103, DragRequantCount: 9,
		GotForwardMsg: 1, ForwardDeltaA: -201, ForwardDeltaB: 202, ForwardDeltaC: -203, ForwardFromRow: 17, CascadeRelay: 1,
		Label: "widgetNode",
		Ports: []portFixture{
			{NodeRow: 7, DX: 1.5, DY: -2.25, DZ: 3.125, PX: 31.5, PY: -32.25, PZ: 33.125, IsInput: 1, Hovered: 0, NameText: "in"},
			{NodeRow: 7, DX: -4.5, DY: 5.25, DZ: -6.125, PX: 41.5, PY: 42.25, PZ: -43.125, IsInput: 0, Hovered: 1, NameText: "out"},
			{NodeRow: 7, DX: 7.5, DY: 8.25, DZ: 9.125, PX: 51.5, PY: -52.25, PZ: 53.125, IsInput: 0, Hovered: 0, NameText: "ctrl"},
		},
		LayoutLinks: []layoutLinkFixture{{DstNodeRow: 2}, {DstNodeRow: 9}},
	}

	portNames := make([]string, len(f.Ports))
	portDX := make([]float32, len(f.Ports))
	portDY := make([]float32, len(f.Ports))
	portDZ := make([]float32, len(f.Ports))
	portPX := make([]float32, len(f.Ports))
	portPY := make([]float32, len(f.Ports))
	portPZ := make([]float32, len(f.Ports))
	portIsInput := make([]uint8, len(f.Ports))
	portHovered := make([]uint8, len(f.Ports))
	for i, p := range f.Ports {
		f.Ports[i].NameLen = uint32(len(p.NameText))
		portNames[i] = p.NameText
		portDX[i] = p.DX
		portDY[i] = p.DY
		portDZ[i] = p.DZ
		portPX[i] = p.PX
		portPY[i] = p.PY
		portPZ[i] = p.PZ
		portIsInput[i] = p.IsInput
		portHovered[i] = p.Hovered
	}
	dstNodeRows := make([]int32, len(f.LayoutLinks))
	for i, ll := range f.LayoutLinks {
		dstNodeRows[i] = ll.DstNodeRow
	}

	raw := Buffer.BuildNodeStreamFrame(
		f.Tick, f.NodeRow,
		f.CX, f.CY, f.CZ, f.Radius, f.SphereR,
		f.VRX, f.VRY, f.VRZ, f.FRX, f.FRY, f.FRZ,
		f.Selected, f.KindID, f.Hovered, f.LatchedSel, f.GotDragMsg,
		f.DragDeltaA, f.DragDeltaB, f.DragDeltaC, f.DragRequantCount,
		f.GotForwardMsg,
		f.ForwardDeltaA, f.ForwardDeltaB, f.ForwardDeltaC, f.ForwardFromRow,
		f.CascadeRelay,
		f.Label,
		portNames,
		portDX, portDY, portDZ, portPX, portPY, portPZ,
		portIsInput, portHovered,
		dstNodeRows,
		nil,
	)
	f.Hex = hex.EncodeToString(raw)
	return f
}

func buildEdgeFrame() edgeFrameFixture {
	f := edgeFrameFixture{
		Tick: 8181, SrcPortRow: 12, DstPortRow: 34, Selected: 1, Label: "edgeLabel",
		BeadVal: []int32{5, -6, 7},
		BeadX:   []float32{61.5, -62.25, 63.125},
		BeadY:   []float32{71.5, 72.25, -73.125},
		BeadZ:   []float32{-81.5, 82.25, 83.125},
	}
	raw := Buffer.BuildEdgeStreamFrame(f.Tick, f.SrcPortRow, f.DstPortRow, f.Selected, f.Label, f.BeadVal, f.BeadX, f.BeadY, f.BeadZ, nil)
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
