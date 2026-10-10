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

	"github.com/dtauraso/beadnetwork/Categories/Scene/Startup"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Tabs"
	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	NodeKind "github.com/dtauraso/beadnetwork/Categories/Node"
	EdgeB "github.com/dtauraso/beadnetwork/Categories/Node/Edge"
	bead "github.com/dtauraso/beadnetwork/Categories/Ring/Bead"
	NodeShape "github.com/dtauraso/beadnetwork/Categories/Ring/NodeShape"
	SceneB "github.com/dtauraso/beadnetwork/Categories/Scene"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"
)

func run(ctx context.Context, cancel context.CancelFunc, topologyPath string, clk clock.Clock) {
	router := Startup.NewSceneRouter()
	open := func(idx int) (Startup.OpenedScene, error) {
		return openScene(ctx, router, topologyPath, idx, clk)
	}
	if err := router.Run(ctx, cancel, Scenes.SelectedIndex(topologyPath), open); err != nil {
		fmt.Fprintf(os.Stderr, "load topology: %v\n", err)
		os.Exit(1)
	}
}

func openScene(processCtx context.Context, router *Startup.SceneRouter, topologyPath string, idx int, clk clock.Clock) (Startup.OpenedScene, error) {
	scenePath := Scenes.PathFor(topologyPath, idx)
	SceneB.WriteSpawnIdentity(scenePath)

	ctx, cancel := context.WithCancel(processCtx)
	sc, err := Startup.Load(ctx, scenePath, clk)
	if err != nil {
		cancel()
		return Startup.OpenedScene{}, err
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
	md.UI.TabStrip.Names = Tabs.TabNames()
	md.UI.TabStrip.Selected = idx
	md.UI.SetSceneRoot(scenePath)
	md.UI.WriteRingSurfaces(NodeShape.CanonicalRingSurfacePointsFlat(), bead.CanonicalRingSurfacePointsFlat())

	Startup.EmitStartupBreadcrumbs(md, scenePath, len(sc.Nodes))
	Startup.CheckRowSeedCount(md, len(sc.Nodes))
	Startup.LoadSceneState(scenePath, md, speedSinks)

	md.Scenes.AnchorPath = topologyPath
	md.Scenes.Reload = func() { router.Reload(processCtx, idx) }
	md.Scenes.Open = func(i int) { router.Open(processCtx, i) }
	md.Scenes.Loaded = idx

	moverWG := md.Start(ctx)
	inbox, gestureWG := Startup.StartGestureActor(ctx, md, speedSinks, clk, Scenes.InputDirPath(scenePath))
	nodesWG := launchNodes(ctx, sc.Nodes)
	wait := func() { joinAll(nodesWG, moverWG, gestureWG) }
	return Startup.OpenedScene{
		Inbox: inbox,
		Stop:  func() { cancel(); wait() },
		Wait:  wait,
	}, nil
}

func launchNodes(ctx context.Context, nodes []Startup.BuiltNode) *sync.WaitGroup {
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
