package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dtauraso/wirefold/Categories/Scene/scenebuild"

	"github.com/dtauraso/wirefold/Categories/Chrome/Tabs"
	clock "github.com/dtauraso/wirefold/Categories/Clock"
	NodeKind "github.com/dtauraso/wirefold/Categories/Node"
	EdgeB "github.com/dtauraso/wirefold/Categories/Node/Edge"
	"github.com/dtauraso/wirefold/Categories/NodeKinds/portwiring"
	bead "github.com/dtauraso/wirefold/Categories/Ring/Bead"
	NodeShape "github.com/dtauraso/wirefold/Categories/Ring/NodeShape"
	SceneB "github.com/dtauraso/wirefold/Categories/Scene"
	"github.com/dtauraso/wirefold/Categories/Scene/inputactor"
	"github.com/dtauraso/wirefold/Categories/Scene/scene"
	"github.com/dtauraso/wirefold/Categories/Scene/scenepaths"
)

func run(ctx context.Context, cancel context.CancelFunc, topologyPath string, clk clock.Clock) {
	scenePath := scene.ResolvePath(topologyPath)
	SceneB.WriteSpawnIdentity(scenePath)

	sc, err := scenebuild.Load(ctx, scenePath, clk)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load topology: %v\n", err)
		os.Exit(1)
	}
	md, speedSinks := sc.Dispatch, sc.SpeedSinks

	md.Sw.SetEdgeStreams(md.GS.EdgeSeeds, md.MR.Edges(), md.MR.NodeGeoms(), md.RT.NodeRowFor,
		EdgeB.NewEdgeSink(scenePath, len(md.RT.EdgeRowTable)))
	sinks := NodeKind.NewSinks(scenePath, len(md.RT.NodeRowTable))
	md.Sw.SetNodeStreams(md.GS.NodeSeeds, md.MR.NodeGeoms(), scenePath,
		sinks.Beads, md.RT.NodeRowFor, sinks.Node, NodeKind.NodeKindID)
	md.UI.WriteOwnTrace()

	md.UI.OwnerCounts.Nodes = int32(len(md.RT.NodeRowTable))
	md.UI.OwnerCounts.Edges = int32(len(md.RT.EdgeRowTable))
	md.UI.SceneTabNames = Tabs.TabNames()
	md.UI.SceneTabSelected = Tabs.SelectedIndex(topologyPath)
	md.UI.SetSceneRoot(scenePath)
	md.UI.WriteRingSurfaces(NodeShape.CanonicalRingSurfacePointsFlat(), bead.CanonicalRingSurfacePointsFlat())

	scenebuild.EmitStartupBreadcrumbs(md, scenePath, len(sc.Nodes))
	scenebuild.CheckRowSeedCount(md, len(sc.Nodes))
	scenebuild.LoadSceneState(scenePath, md, speedSinks)

	md.Scenes.AnchorPath = topologyPath
	md.Scenes.Quit = cancel
	md.Scenes.Loaded = md.UI.SceneTabSelected

	moverWG := md.Start(ctx)
	stdinWG, gestureWG := inputactor.StartStdinReader(ctx, cancel, md, speedSinks, clk, scenepaths.InputDirPath(scenePath))
	joinAll(launchNodes(ctx, sc.Nodes), moverWG, stdinWG, gestureWG)
}

func launchNodes(ctx context.Context, nodes []portwiring.Node) *sync.WaitGroup {
	wg := new(sync.WaitGroup)
	wg.Add(len(nodes))
	for _, node := range nodes {
		go func() {
			defer wg.Done()
			node.Update(ctx)
		}()
	}
	return wg
}

func joinAll(groups ...*sync.WaitGroup) {
	for _, g := range groups {
		g.Wait()
	}
}

func Run(topologyPath string) {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	run(ctx, cancel, topologyPath, clock.NewRealClock())
}

func RunTest(dur time.Duration, topologyPath string) {
	ctx, cancel := context.WithTimeout(context.Background(), dur)
	defer cancel()
	run(ctx, cancel, topologyPath, clock.NewRealClock())
}

func main() {
	dur := flag.Duration("duration", 0, "if non-zero, run for this duration then exit (test mode)")
	topologyPath := flag.String("topology", "topology", "path to topology JSON spec")
	flag.Parse()
	if *dur > 0 {
		RunTest(*dur, *topologyPath)
	} else {
		Run(*topologyPath)
	}
}
