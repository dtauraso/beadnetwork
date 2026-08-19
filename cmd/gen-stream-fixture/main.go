package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	outPath := "src/test/fixtures/stream_fixture.json"
	if len(os.Args) > 1 {
		outPath = os.Args[1]
	}

	fx := streamFixture{
		NodeFrame:     buildNodeFrame(),
		EdgeFrame:     buildEdgeFrame(),
		BeadFrame:     buildBeadFrame(),
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
