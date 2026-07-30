package main

//go:generate go run ./tools/gen-node-defs

import (
	"context"
	"flag"
	"fmt"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	B "github.com/dtauraso/wirefold/Buffer"
	T "github.com/dtauraso/wirefold/Trace"
	W "github.com/dtauraso/wirefold/nodes/Wiring"
)

// toStreamEvents converts a nodeMover/edgeMover/interiorStream goroutine's own
// row-resolved events (Wiring.RowEvent, string kind — kept Buffer-independent there) into
// Buffer.StreamEvent (numeric kind, via Buffer.KindID) for packing into that SAME
// goroutine's own frame's trailing EVENTS section (memory/feedback_no_single_writer_bridge.md).
// Pure value conversion — no shared state, safe to call from any owner goroutine.
func toStreamEvents(events []wire.RowEvent) []B.StreamEvent {
	if len(events) == 0 {
		return nil
	}
	out := make([]B.StreamEvent, len(events))
	for i, e := range events {
		out[i] = B.StreamEvent{
			Kind:          B.KindID(e.Kind),
			NodeRow:       e.NodeRow,
			PortRow:       e.PortRow,
			TargetRow:     e.TargetRow,
			TargetPortRow: e.TargetPortRow,
			EdgeRow:       e.EdgeRow,
			Slot:          e.Slot,
			Value:         e.Value,
			Bead:          uint32(e.Bead),
			BeadSteps:     float32(e.BeadSteps),
			SimLatencyMs:  float32(e.SimLatencyMs),
			X:             float32(e.X),
			Y:             float32(e.Y),
			Z:             float32(e.Z),
			F:             float32(e.F),
			Label:         e.Label,
			Debug:         e.Debug,
			Text:          e.Text,
		}
	}
	return out
}

// runTopology loads and runs the topology under ctx, blocking until ctx is
// cancelled or all nodes exit. Shared by Run and RunTest.
//
// clk is the single monotonic clock every wire reads to time its own delivery
// (MODEL.md). Both callers (Run, RunTest) pass a real clock; it is always non-nil.
// The clock is free-running (no play/pause gate).
func runTopology(ctx context.Context, cancel context.CancelFunc, topologyPath string, clk wire.Clock) {
	// The VIEW stream (camera+overlay+scene, one singleton row) — per-owner buffer rows
	// (memory/feedback_no_single_writer_bridge.md, memory/feedback_no_single_writer_bridge.md): WIREFOLD_STREAM_FDS
	// is now MANDATORY (the old central accumulator + fallback packer were deleted along
	// with this migration's final step — see Buffer/stream_fds.go). The gesture/stdin-reader
	// goroutine (nodes/Wiring's MoveDispatch, wired below once it exists) is the sole WRITER of
	// this stream.
	streamFDs := B.ParseStreamFDs(os.Getenv("WIREFOLD_STREAM_FDS"))
	viewFile, viewStreamWired := streamFDs.Open(B.StreamKindView, 0)
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
	nodes, slotReg, md, speedSinks, err := W.LoadTopology(ctx, topologyPath, tr, clk)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load topology: %v\n", err)
		os.Exit(1)
	}
	// The per-edge dedicated stream (memory/feedback_no_single_writer_bridge.md): when
	// WIREFOLD_STREAM_FDS carries an "edge" entry, wire every edgeMover to its OWN fd
	// (fd = baseFd + edgeRow, edgeRow = the stable seed order — see
	// MoveDispatch.SetEdgeStreams). No edges (edgeBase absent) leaves every edgeMover's
	// streamOut at its zero value (nil) — there is nothing to stream.
	//
	// REPORT the asymmetry rather than skipping it silently. A graph that loaded N edges
	// but received no edge fds streams nothing, so the editor draws no edges — and without
	// this message that is indistinguishable from a broken edge path, sending the reader
	// through the code (disk layout, bundle freshness, recent merges) when the cause is
	// operational: a VS Code extension host left running across a change, holding older fd
	// plumbing. Reopening a file does not restart it; only "Developer: Reload Window" does
	// (memory/feedback_two_process_editor_reload.md).
	//
	// This is the same class runCommand.ts's MAX_EDGE_STREAMS overflow was fixed for in
	// 93d2e9b6, and its reasoning applies verbatim: silently disabling every dedicated
	// per-edge stream is "the quietest possible failure for the loudest consequence."
	// That fix covered the count-too-large case on the TS side; this covers the fds-absent
	// case on the Go side, which is the one an operator actually hits.
	//
	// Not fatal: a deliberate no-fd launch (headless runs, tools with no dedicated pipes)
	// is legitimate input, exactly like a large topology. Loud, not dead.
	if _, edgeFDsWired := streamFDs[B.StreamKindEdge]; !edgeFDsWired {
		if n := len(md.EdgeSeeds()); n > 0 {
			fmt.Fprintf(os.Stderr,
				"stream-fd mismatch: topology loaded %d edges but WIREFOLD_STREAM_FDS carries no %q entry; "+
					"every edgeMover's stream stays nil, so NO EDGES will be drawn. If the editor was open "+
					"across a rebuild, run \"Developer: Reload Window\" — reopening the file restarts only the "+
					"webview, not the extension host that allocates these fds.\n",
				n, B.StreamKindEdge)
		}
	}
	if edgeBase, ok := streamFDs[B.StreamKindEdge]; ok {
		// Edge selection is no longer an injected lookup: each edgeMover owns its OWN
		// selected bit, set via a moveMsgKindSelect message the gesture goroutine sends
		// on select/deselect (MoveDispatch.sendEdgeSelect).
		md.SetEdgeStreams(edgeBase, md.PortRowFor, md.NodeRowFor,
			func(tick uint32, srcPortRow, dstPortRow int32, selected uint8, label string, events []wire.RowEvent) []byte {
				return B.BuildEdgeStreamFrame(tick, srcPortRow, dstPortRow, selected, label, toStreamEvents(events))
			})
	}
	// The two per-node dedicated streams (memory/feedback_no_single_writer_bridge.md):
	// NODE (geometry+ports+label, written by each nodeMover) and INTERIOR (interior
	// beads, written by each node's OWN Update goroutine — the SECOND emitting goroutine
	// per node). Both require the SAME "node" AND "interior" WIREFOLD_STREAM_FDS entries
	// (a node stream with no interior counterpart, or vice versa, would leave one of the
	// two goroutines with nowhere fresh to write while the other has one — so both are
	// required together).
	//
	// The "both required together" rule above is enforced by a silent skip: one entry
	// present without the other leaves BOTH streams unwired and says nothing. Same class
	// as the edge case, and harder to spot because the half that IS wired looks healthy.
	_, nodeFDsWired := streamFDs[B.StreamKindNode]
	_, interiorFDsWired := streamFDs[B.StreamKindInterior]
	if nodeFDsWired != interiorFDsWired {
		fmt.Fprintf(os.Stderr,
			"stream-fd mismatch: WIREFOLD_STREAM_FDS carries %q=%t but %q=%t; they are required "+
				"together, so BOTH per-node streams stay unwired and node geometry/interior beads "+
				"will not be drawn.\n",
			B.StreamKindNode, nodeFDsWired, B.StreamKindInterior, interiorFDsWired)
	}
	if nodeBase, ok := streamFDs[B.StreamKindNode]; ok {
		if interiorBase, ok2 := streamFDs[B.StreamKindInterior]; ok2 {
			// Selection/hover/abc-drag/kind are no longer injected lookups: each
			// nodeMover owns its OWN selected/hovered/latchedSel/gotDragMsg/dragDelta*
			// bits, set via moveMsgKindSelect/Hover/Latched/AbcReset messages the gesture
			// goroutine sends (or, for kindID, resolved once here at construction).
			// kindIDFor resolves a node's static load-time kind string to its NODE_DEFS
			// index (Buffer.NodeKindID) — injected so Wiring stays Buffer-independent.
			md.SetNodeStreams(nodeBase, interiorBase,
				md.NodeRowFor,
				func(tick uint32, nodeRow int32, cx, cy, cz, radius, sphereR float32, vrx, vry, vrz, frx, fry, frz float32, selected, kindID, hovered, latchedSel, gotDragMsg uint8, dragDeltaA, dragDeltaB, dragDeltaC, dragRequantCount int32, gotForwardMsg uint8, forwardDeltaA, forwardDeltaB, forwardDeltaC, forwardFromRow int32, cascadeRelay uint8, label string, portNames []string, portDX, portDY, portDZ, portPX, portPY, portPZ []float32, portIsInput, portHovered []uint8, dstNodeRows []int32, chainBeadOX, chainBeadOY, chainBeadOZ []float32, chainBeadLit []uint8, chainBeadLitValue []int32, events []wire.RowEvent) []byte {
					return B.BuildNodeStreamFrame(tick, nodeRow, cx, cy, cz, radius, sphereR, vrx, vry, vrz, frx, fry, frz,
						selected, kindID, hovered, latchedSel, gotDragMsg, dragDeltaA, dragDeltaB, dragDeltaC, dragRequantCount,
						gotForwardMsg, forwardDeltaA, forwardDeltaB, forwardDeltaC, forwardFromRow, cascadeRelay,
						label, portNames, portDX, portDY, portDZ, portPX, portPY, portPZ, portIsInput, portHovered,
						dstNodeRows, chainBeadOX, chainBeadOY, chainBeadOZ, chainBeadLit, chainBeadLitValue, toStreamEvents(events))
				},
				func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []wire.RowEvent) []byte {
					return B.BuildInteriorStreamFrame(tick, present, value, ox, oy, oz, toStreamEvents(events))
				},
				B.NodeKindID)
		}
	}
	// The VIEW stream's write side (Step C, memory/feedback_no_single_writer_bridge.md): wire md as the
	// stream's owner/writer BEFORE anything that can change camera/overlay/scene-sphere/
	// selection/hover reaches it (SeedInitialViewpoint/LoadOverlays/LoadSceneSphere below,
	// then the launched movers/stdin reader) — mirrors SetEdgeStreams/SetNodeStreams'
	// "wire before it can fire" ordering above. Only when the dedicated fd is actually
	// wired (viewStreamWired) — left uncalled otherwise (no WIREFOLD_STREAM_FDS "view"
	// entry, e.g. a non-extension launch with no dedicated pipes at all).
	if viewStreamWired {
		md.SetViewStream(viewFile,
			func(tick uint32,
				camPX, camPY, camPZ, camR, camPosTheta, camPosPhi, camUpTheta, camUpPhi float32,
				sceneTori, scenePoles, nodePoles, selSpherePoles, handholds, labelsGlobal, overlaysVis, cascadeLinks uint8,
				dragNodeRow int32,
				groupLenTime, groupLenInput, groupLenGate float32,
				sceneCX, sceneCY, sceneCZ, sceneRadius float32,
				events []wire.RowEvent,
			) []byte {
				return B.BuildViewStreamFrame(tick,
					camPX, camPY, camPZ, camR, camPosTheta, camPosPhi, camUpTheta, camUpPhi,
					B.OverlayRow{
						SceneTori: sceneTori, ScenePoles: scenePoles, NodePoles: nodePoles,
						SelSpherePoles: selSpherePoles, Handholds: handholds, LabelsGlobal: labelsGlobal,
						OverlaysVis: overlaysVis, CascadeLinks: cascadeLinks,
						DragNodeRow:  dragNodeRow,
						GroupLenTime: groupLenTime, GroupLenInput: groupLenInput, GroupLenGate: groupLenGate,
					},
					sceneCX, sceneCY, sceneCZ, sceneRadius,
					toStreamEvents(events))
			})
		// LayoutLink is load-time-once — emit each pair once, now that the view stream is
		// wired, so the .probe log's per-kind count still matches the -trace reference.
		// tr.LayoutLink itself (loader.go's emitLayoutLinks, already run inside
		// LoadTopology above) is UNCHANGED — this is purely the EVENT-block/.probe-log
		// representation; the render-path LayoutLink section is node_mover.go's own
		// layoutLinkTos, carried on each node's own stream frame.
		for _, pair := range md.LayoutLinkPairs() {
			nodeRow := int32(-1)
			if r, ok := md.NodeRowFor(pair[0]); ok {
				nodeRow = r
			}
			targetRow := int32(-1)
			if r, ok := md.NodeRowFor(pair[1]); ok {
				targetRow = r
			}
			md.EmitLayoutLinkViewEvent(nodeRow, targetRow)
		}
	}
	// One example startup breadcrumb — proves the debug channel end-to-end and is genuinely
	// useful (which topology loaded, how many nodes). Sparse: once per run.
	tr.Breadcrumb("topology-loaded", topologyPath, "", fmt.Sprintf("nodes=%d", len(nodes)))
	// Structured buffer counterpart: rides the VIEW stream (no per-node stream exists
	// yet for a startup-only event, and this runs on the main goroutine before any
	// per-node/edge/interior goroutine exists). topologyPath is genuinely free-form
	// (a filesystem path), so it rides the sanctioned Text column; nodes count is
	// the typed Value column.
	md.EmitBreadcrumb(wire.RowEvent{
		Label: T.BreadcrumbTopologyLoaded, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Value: int32(len(nodes)), Text: topologyPath,
	})

	// Sparse, one-time startup sanity check (CLAUDE.md DEBUG BREADCRUMB channel): every
	// node LoadTopology returned should have a row-seed entry (md.NodeSeeds(), the SAME
	// spec-order row table nodes/Wiring's own move-dispatch/stream wiring above already
	// uses). A mismatch means md.NodeSeeds() (spec order) and LoadTopology's node list
	// diverged — a real topology bug — and must be visible.
	if len(md.NodeSeeds()) != len(nodes) {
		tr.Breadcrumb("row-seed-count-mismatch", "", "", fmt.Sprintf("NodeSeeds=%d nodes=%d", len(md.NodeSeeds()), len(nodes)))
		// Structured buffer counterpart, VIEW stream (same reasoning as
		// topology-loaded above). Value=NodeSeeds count, X=nodes count — both
		// small typed ints, no free-form text needed.
		md.EmitBreadcrumb(wire.RowEvent{
			Label: T.BreadcrumbRowSeedCountMismatch, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
			Value: int32(len(md.NodeSeeds())), X: float64(len(nodes)),
		})
	}

	// Initial camera viewpoint = FILE DATA. Go reads the saved camera from
	// <topologyPath>/view/camera.json itself and installs it into the gesture-FSM viewpoint,
	// so the buffer camera columns carry a real, non-degenerate saved pose from the first
	// frame (pan works immediately). Absent/malformed file → a fixed non-degenerate default.
	//
	// The buffer's node/edge/port row-identity tables now live ON md itself (built once at
	// load, in newMoveDispatch's buildRowTables call, from the same spec-order nodeSeeds/
	// edgeSeeds each per-owner stream frame uses below) — a node/edge/port hit (which
	// carries only a numeric buffer row index) resolves back to its identity via
	// md.LookupNodeRow/LookupEdgeRow/LookupPortRow with no separate resolver wiring.
	// Initial camera viewpoint = FILE DATA: Go reads the saved camera from
	// <topologyPath>/view/camera.json and installs it into the gesture-FSM viewpoint.
	W.SeedInitialViewpoint(topologyPath, md, tr)
	// Restore persisted overlay visibility: seed md.ov from overlays.json and emit each flag
	// so the buffer streams the saved overlay state from the first frame. Seed BEFORE
	// EnableEditPersist so the seed's own emit does not write the loaded state back.
	md.LoadOverlays(topologyPath, tr)
	// Arm the WRITE side AFTER the seeds: from here, every gesture that changes the FSM
	// viewpoint (orbit/zoom/pan/home) debounces a write of the current pose back to
	// <topologyPath>/view/camera.json, so navigate-then-reload round-trips.
	// Arming after the seed keeps the seed's own emit from persisting the loaded/default pose.
	md.EnableViewpointPersist(topologyPath)
	// Arm disk persistence for the FSM-applied edits (node-drag position, ring-move
	// anchor) — debounced Go-side read-modify-writes, armed after the seeds so their
	// own emits do not write loaded state back.
	md.EnableEditPersist(topologyPath)

	// Install the scene sphere (persisted, or a content-fit centroid for a fresh
	// scene) BEFORE launching the movers and the stdin reader. It only needs the
	// movers to be BUILT (their seeded centers, available since LoadTopology), not
	// running; installing it after Start left md.sceneSphere written unsynchronized
	// while the mover/gesture goroutines could already read it on the drag path.
	md.LoadSceneSphere(topologyPath)

	// Launch the per-node and per-edge move-handler goroutines (decentralized
	// node-move: each node/edge drains its own inbox and recomputes its own geometry).
	// moverWG covers every nodeMover/edgeMover goroutine Start launched (see its doc
	// comment). Waiting on it is what lets Close() run with nothing still emitting —
	// the reason Trace needs no mutex.
	moverWG := md.Start(ctx)

	// Read the editor→Go bridge: "edit" JSON lines (op = create/update/delete)
	// from stdin. When stdin reaches EOF (extension host disconnect), cancel the context.
	//
	// stdinWG tracks ONLY this dispatch-loop goroutine, not RunStdinReader's internal
	// frame-reader goroutine. That inner goroutine blocks in io.ReadFull(os.Stdin),
	// which does NOT select on ctx — it is unblocked only by closing the fd (which
	// RunStdinReader itself arranges when r is an io.Closer and ctx is done). On a
	// non-pollable fd that close could still leave the read parked, so waiting on it
	// here would turn a leak into a hang. RunStdinReader's dispatch loop, in contrast,
	// selects on ctx.Done() and returns immediately on cancel regardless of the frame
	// reader's state — that promptness is what stdinWG actually certifies. The frame
	// reader goroutine is deliberately left un-waited (detached); in production it
	// outlives the process only as long as it takes the OS to tear down the closed fd,
	// which is bounded by process exit, not by this WaitGroup.
	stdinWG := new(sync.WaitGroup)
	stdinWG.Add(1)
	go func() {
		defer stdinWG.Done()
		W.RunStdinReader(ctx, os.Stdin, slotReg, md, tr, speedSinks)
		cancel()
	}()

	wg := new(sync.WaitGroup)
	wg.Add(len(nodes))
	for _, node := range nodes {
		go func() {
			defer wg.Done()
			node.Update(ctx)
		}()
	}

	// Wait for every tracked goroutine to exit — node Update loops, nodeMover/
	// edgeMover goroutines, and the stdin dispatch loop — before closing the trace.
	// No grace timeout: every one of these goroutines' only blocking call is
	// SleepCycle, which selects on ctx.Done(), so cancel-to-return is bounded by one
	// clock tick (~16ms), not by an arbitrary grace window. If a goroutine ever fails
	// to exit, wg.Wait() below hangs visibly instead of silently proceeding past a
	// still-running goroutine — a hang names the bug; a grace timeout hides it.
	done := make(chan struct{})
	go func() {
		wg.Wait()
		moverWG.Wait()
		stdinWG.Wait()
		close(done)
	}()
	<-done
}

// Run wires the topology and blocks until SIGTERM/SIGINT or stdin EOF.
// This is the live-run path used by the extension host. It uses a production
// free-running RealClock (no play/pause gate).
func Run(topologyPath string) {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	runTopology(ctx, cancel, topologyPath, wire.NewRealClock())
}

// RunTest wires the topology and lets it run for dur before cancelling, using a
// production RealClock. Used by automated tests that need a self-terminating run.
func RunTest(dur time.Duration, topologyPath string) {
	ctx, cancel := context.WithTimeout(context.Background(), dur)
	defer cancel()
	runTopology(ctx, cancel, topologyPath, wire.NewRealClock())
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
