// gen-stream-fixture is the Go SIDE of the Go->TS stream-frame cross-language fixture
// (mirrors tools/topology-vscode/scripts/gen-input-fixture-src.ts, which is the TS side of
// the TS->Go input-record fixture — see nodes/Wiring/inputcodec/input_fixture_test.go's header for the
// gap this closes).
//
// It builds real per-owner stream-frame bytes with the REAL production packers
// (streamframe.BuildNodeStreamFrame / BuildEdgeStreamFrame / BuildInteriorStreamFrame), using
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
// tools/topology-vscode/test/buffer/stream-fixture.test.ts decodes the fixture's hex with the REAL
// TS decoders (decodeNodeStreamFrame/decodeEdgeStreamFrame/decodeInteriorStreamFrame in
// buffer-decode-node.ts/buffer-decode-edge.ts/buffer-decode-interior.ts) and asserts every
// field — the actual cross-language byte-level
// agreement check. It also regenerates this fixture live (via `go run`) and diffs it
// against the committed copy, so a stale fixture fails loudly instead of silently testing
// its own past self (same freshness shape as TestInputFixtureFreshness, mirrored onto the
// opposite direction).
//
// The fixture's own JSON shapes are in types.go; each frame's builder (pure construction +
// a call to the real packer) is in build_frames.go. This file is only the CLI entry point:
// resolve the output path, build the fixture, and write it.
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

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
