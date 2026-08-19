package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dtauraso/wirefold/tools/gen-node-defs/buflayout"
)

func generateColumnStreams(repoRoot string) {
	schemaDir := repoRoot
	schema, err := buflayout.ParseBufferLayoutTree(schemaDir)
	if err != nil {
		fatalf("parse buffer layout for column streams: %v", err)
	}

	goPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "Buffer", "column_streams_gen.go")
	if err := buflayout.WriteColumnStreamsGo(goPath, schema); err != nil {
		fatalf("write %s: %v", goPath, err)
	}
	tsPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "Buffer", "column-streams-gen.ts")
	if err := buflayout.WriteColumnStreamsTS(tsPath, schema); err != nil {
		fatalf("write %s: %v", tsPath, err)
	}
	perBlock, err := buflayout.WriteBlockColumnsTS(schema, tsPath)
	if err != nil {
		fatalf("write per-block columns: %v", err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s and %s, plus %d per-block column file(s)\n",
		goPath, tsPath, len(perBlock))
}
