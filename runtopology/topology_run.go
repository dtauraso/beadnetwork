package runtopology

import (
	"context"
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/nodes/wire/clock"

	SF "github.com/dtauraso/wirefold/Buffer/streamframe"
	T "github.com/dtauraso/wirefold/Trace"
	W "github.com/dtauraso/wirefold/nodes/Wiring"
	"github.com/dtauraso/wirefold/nodes/Wiring/scene"
)

// RunTopology loads and runs the topology under ctx, blocking until ctx is
// cancelled or all nodes exit. Shared by main's Run and RunTest.
//
// clk is the single monotonic clock every wire reads to time its own delivery
// (MODEL.md). Both callers (Run, RunTest) pass a real clock; it is always non-nil.
// The clock is free-running (no play/pause gate).
func RunTopology(ctx context.Context, cancel context.CancelFunc, topologyPath string, clk clock.Clock) {
	// The VIEW stream (camera+overlay+scene, one singleton row) — per-owner buffer rows
	// (memory/feedback_no_single_writer_bridge.md, memory/feedback_no_single_writer_bridge.md): WIREFOLD_STREAM_FDS
	// is now MANDATORY (the old central accumulator + fallback packer were deleted along
	// with this migration's final step — see Buffer/streamframe/stream_fds.go). The gesture/stdin-reader
	// goroutine (nodes/Wiring's MoveDispatch, wired below once it exists) is the sole WRITER of
	// this stream.
	streamFDs := SF.ParseStreamFDs(os.Getenv("WIREFOLD_STREAM_FDS"))
	viewFile, viewStreamWired := streamFDs.Open(SF.StreamKindView, 0)
	// Trace is now just the breadcrumb writer (the central event channel/drain and the
	// -trace JSONL dump were deleted — memory/feedback_no_single_writer_bridge.md's final step: every
	// emitting goroutine packs its own frame directly; see Trace/Trace.go's doc comment).
	tr := T.New()
	// DEBUG BREADCRUMB channel: each Breadcrumb() call site emits a structured
	// Kind==KindBreadcrumb EVENT row on its own owning per-owner stream (node/edge/
	// interior/VIEW) — see Trace.go's Breadcrumb/Trace-struct doc comments and each
	// call site's writeStreamFrame/writeEvents/EmitBreadcrumb. There is no longer a
	// separate production stdout sink here; probe-merge.sh --debug decodes these
	// buffer-carried breadcrumb rows (filtered by the Debug flag) instead of parsing
	// a JSON stdout line.

	// The clock is free-running (no play/pause gate): it starts ticking at construction
	// and never halts. Startup geometry is NOT emitted here — each node's own goroutine
	// emits its geometry once at startup (below, after this function's node-goroutine
	// launch loop); see the row-seeding comment there for why the buffer's row tables do
	// not depend on that emit order.
	// SCENE TABS (nodes/Wiring/scene/scene_tabs.go). topologyPath is the ANCHOR — the fixed path
	// the extension host launched against; the scene actually LOADED is the selected tab's
	// sibling directory. An untabbed anchor (any tree that is not the tab-0 directory —
	// every test fixture, every one-off run) resolves to itself and streams an empty strip.
	sceneTabNames := scene.SceneTabNames(topologyPath)
	sceneTabSelected := scene.SelectedSceneIndex(topologyPath)
	scenePath := scene.ResolveScenePath(topologyPath)
	nodes, slotReg, md, speedSinks, err := W.LoadTopology(ctx, scenePath, tr, clk)
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

	// Arm tab switching. The ANCHOR (not scenePath) is what the selection is persisted
	// against — it is the one path that is the same whichever tab is showing. cancel ends
	// this run; the extension host's runner is looping, so it respawns against the same
	// anchor and this function re-resolves the newly selected scene above.
	md.EnableSceneSwitch(topologyPath, cancel)

	// Launch the per-node and per-edge move-handler goroutines (decentralized
	// node-move: each node/edge drains its own inbox and recomputes its own geometry).
	// moverWG covers every nodeMover/edgeMover goroutine Start launched (see its doc
	// comment). Waiting on it is what lets Close() run with nothing still emitting —
	// the reason Trace needs no mutex.
	moverWG := md.Start(ctx)

	stdinWG, gestureWG := startStdinReader(ctx, cancel, slotReg, md, tr, speedSinks)
	wg := launchNodeGoroutines(ctx, nodes)
	joinAll(wg, moverWG, stdinWG, gestureWG)
}
