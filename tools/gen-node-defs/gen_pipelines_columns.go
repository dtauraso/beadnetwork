package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dtauraso/wirefold/tools/gen-node-defs/buflayout"
)

func generateColumnStreams(repoRoot string) {
	schemaDir := filepath.Join(repoRoot, "tools", "topology-vscode", "Buffer", "bufschema")
	schema, err := buflayout.ParseBufferLayoutDir(schemaDir)
	if err != nil {
		fatalf("parse buffer layout for column streams: %v", err)
	}

	goPath := filepath.Join(repoRoot, "tools", "topology-vscode", "Buffer", "column_streams_gen.go")
	if err := buflayout.WriteColumnStreamsGo(goPath, schema); err != nil {
		fatalf("write %s: %v", goPath, err)
	}
	tsPath := filepath.Join(repoRoot, "tools", "topology-vscode", "Buffer", "column-streams-gen.ts")
	if err := buflayout.WriteColumnStreamsTS(tsPath, schema); err != nil {
		fatalf("write %s: %v", tsPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s and %s\n", goPath, tsPath)
}
