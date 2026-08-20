package main

//go:generate go run .

import (
	"os"
	"path/filepath"

	"github.com/dtauraso/wirefold/scripts/genpaths"
	"github.com/dtauraso/wirefold/src/Buffer/gen/buflayout"
	"github.com/dtauraso/wirefold/src/Buffer/gen/params"
	"github.com/dtauraso/wirefold/src/Buffer/gen/tracekinds"
)

func main() {
	genpaths.Name = "Buffer/gen"
	repoRoot, srcRoot := genpaths.Roots()

	generateCurveParams(repoRoot, srcRoot)
	generateShadingParams(repoRoot, srcRoot)
	generateColumnStreams(repoRoot, srcRoot)
	generateBufferLayout(repoRoot, srcRoot)
	generateFrameTags(srcRoot)
	generateTraceKinds(srcRoot)
}

func generateTraceKinds(srcRoot string) {
	dir := filepath.Join(srcRoot, "Buffer")
	kinds, err := tracekinds.ParseTraceKinds(dir)
	if err != nil {
		genpaths.Fatalf("parse trace kinds: %v", err)
	}
	labels, err := tracekinds.ParseBreadcrumbLabels(dir)
	if err != nil {
		genpaths.Fatalf("parse breadcrumb labels: %v", err)
	}
	outPath := filepath.Join(dir, "trace-kinds.ts")
	if err := tracekinds.WriteTraceKinds(outPath, kinds, labels); err != nil {
		genpaths.Fatalf("write %s: %v", outPath, err)
	}
	genpaths.Announce(outPath, len(kinds), "kinds")
}

func generateCurveParams(repoRoot, srcRoot string) {
	goPath := filepath.Join(genpaths.NetworkDir(repoRoot), "nodegeom", "curve_params.go")
	curveParams, err := params.ParseCurveParams(goPath)
	if err != nil {
		genpaths.Fatalf("parse curve params: %v", err)
	}
	tsPath := filepath.Join(srcRoot, "Buffer", "curve-params.ts")
	if err := params.WriteCurveParams(tsPath, curveParams); err != nil {
		genpaths.Fatalf("write %s: %v", tsPath, err)
	}
	genpaths.Announce(tsPath, len(curveParams), "constants")
}

func generateShadingParams(repoRoot, srcRoot string) {
	goPath := filepath.Join(genpaths.NetworkDir(repoRoot), "nodegeom", "shading_params.go")
	shadingParams, err := params.ParseShadingParams(repoRoot, goPath)
	if err != nil {
		genpaths.Fatalf("parse shading params: %v", err)
	}
	tsPath := filepath.Join(srcRoot, "Buffer", "shading-params.ts")
	if err := params.WriteShadingParams(tsPath, shadingParams); err != nil {
		genpaths.Fatalf("write %s: %v", tsPath, err)
	}
	genpaths.Announce(tsPath, len(shadingParams), "constants")
}

func generateColumnStreams(repoRoot, srcRoot string) {
	schema, err := buflayout.ParseBufferLayoutTree(repoRoot)
	if err != nil {
		genpaths.Fatalf("parse buffer layout for column streams: %v", err)
	}

	goPath := filepath.Join(srcRoot, "Buffer", "column_streams_gen.go")
	if err := buflayout.WriteColumnStreamsGo(goPath, schema); err != nil {
		genpaths.Fatalf("write %s: %v", goPath, err)
	}
	tsPath := filepath.Join(srcRoot, "Buffer", "column-streams-gen.ts")
	if err := buflayout.WriteColumnStreamsTS(tsPath, schema); err != nil {
		genpaths.Fatalf("write %s: %v", tsPath, err)
	}
	perBlock, err := buflayout.WriteBlockColumnsTS(schema, tsPath)
	if err != nil {
		genpaths.Fatalf("write per-block columns: %v", err)
	}
	genpaths.Announce(goPath, 1, "column stream file")
	genpaths.Announce(tsPath, len(perBlock), "per-block column files")
}

func generateBufferLayout(repoRoot, srcRoot string) {
	bufSchema, err := buflayout.ParseBufferLayoutTree(repoRoot)
	if err != nil {
		genpaths.Fatalf("parse buffer layout: %v", err)
	}
	bufferDir := filepath.Join(srcRoot, "Buffer")
	goPath := filepath.Join(bufferDir, "buffer_layout_gen.go")
	goRowsPath := filepath.Join(bufferDir, "buffer_layout_gen_rows.go")
	goRows2Path := filepath.Join(bufferDir, "buffer_layout_gen_rows2.go")
	goSingletonsPath := filepath.Join(bufferDir, "buffer_layout_gen_singletons.go")
	if err := buflayout.WriteBufferLayoutGo(goPath, goRowsPath, goRows2Path, goSingletonsPath, bufSchema); err != nil {
		genpaths.Fatalf("write buffer layout go: %v", err)
	}
	announceBlocks(bufSchema.Blocks, goPath, goRowsPath, goRows2Path, goSingletonsPath)

	tsPath := filepath.Join(bufferDir, "buffer-layout.ts")
	tsRowsPath := filepath.Join(bufferDir, "buffer-layout-rows-gen.ts")
	tsRows2Path := filepath.Join(bufferDir, "buffer-layout-rows2-gen.ts")
	tsSingletonsPath := filepath.Join(bufferDir, "buffer-layout-singletons-gen.ts")
	if err := buflayout.WriteBufferLayoutTS(tsPath, tsRowsPath, tsRows2Path, tsSingletonsPath, bufSchema); err != nil {
		genpaths.Fatalf("write buffer layout ts: %v", err)
	}
	announceBlocks(bufSchema.Blocks, tsPath, tsRowsPath, tsRows2Path, tsSingletonsPath)
}

func announceBlocks[T any](blocks []T, paths ...string) {
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		genpaths.Announce(path, len(blocks), "blocks")
	}
}

func generateFrameTags(srcRoot string) {
	goPath := filepath.Join(srcRoot, "Buffer", "frame_tags.go")
	header, consts, err := parseFrameTags(goPath)
	if err != nil {
		genpaths.Fatalf("parse frame tags: %v", err)
	}
	tsPath := filepath.Join(srcRoot, "Buffer", "frame-tags.ts")
	if err := writeFrameTags(tsPath, header, consts); err != nil {
		genpaths.Fatalf("write %s: %v", tsPath, err)
	}
	genpaths.Announce(tsPath, len(consts), "constants")
}
