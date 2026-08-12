package runtopology

import (
	"context"
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/nodes/wire/clock"

	SF "github.com/dtauraso/wirefold/Buffer/streamframe"
	T "github.com/dtauraso/wirefold/Trace"
	Bld "github.com/dtauraso/wirefold/nodes/Wiring/build"
	"github.com/dtauraso/wirefold/nodes/Wiring/scene"
)

func RunTopology(ctx context.Context, cancel context.CancelFunc, topologyPath string, clk clock.Clock) {

	streamFDs := SF.ParseStreamFDs(os.Getenv("WIREFOLD_STREAM_FDS"))
	viewFile, viewStreamWired := streamFDs.Open(SF.StreamKindView, 0)

	tr := T.New()

	sceneTabNames := scene.SceneTabNames(topologyPath)
	sceneTabSelected := scene.SelectedSceneIndex(topologyPath)
	scenePath := scene.ResolveScenePath(topologyPath)
	nodes, slotReg, md, speedSinks, err := Bld.LoadTopology(ctx, scenePath, tr, clk)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load topology: %v\n", err)
		os.Exit(1)
	}
	wireEdgeStreams(streamFDs, md)
	wireNodeStreams(streamFDs, md)
	wireViewStream(md, viewFile, viewStreamWired, sceneTabNames, sceneTabSelected)
	emitStartupBreadcrumbs(tr, md, scenePath, len(nodes))
	checkRowSeedCount(tr, md, len(nodes))
	loadSceneState(scenePath, md, tr, speedSinks)

	md.Scenes.AnchorPath = topologyPath
	md.Scenes.Quit = cancel

	moverWG := md.Start(ctx)

	stdinWG, gestureWG := startStdinReader(ctx, cancel, slotReg, md, tr, speedSinks)
	wg := launchNodeGoroutines(ctx, nodes)
	joinAll(wg, moverWG, stdinWG, gestureWG)
}
