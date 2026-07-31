package Wiring

// drag_persist_e2e_test.go — higher-fidelity headless validation of the current
// drag/neighbor model (MODEL.md "Node positions & movement locks"): dragging a node X
// moves only X; each direct domain neighbor STAYS PUT and re-quantizes its OWN stored
// local polar to X from the live offset (moveMsgKindNeighborSetC ->
// neighborSetCRequantize -> requantizePoleTraced). Per
// memory/feedback_headless_repro_verifies_persistence.md, this drives the REAL loader +
// mover goroutines + debounced persisters against a 3-node star topology and asserts on
// the RE-PERSISTED meta.json BYTES on disk (not just in-memory state), because green
// unit tests have hidden live persistence failures before.

import (
	"context"
	"encoding/json"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeStar3 lays down a minimal directory-tree topology with THREE nodes — a hub "A"
// and two leaves "B" and "C", each with one edge to A — so a drag on A exercises TWO
// independent neighbor requantizations at once (not just the one this package's other
// fixture, writeTree's 2-node src/dst, can exercise).
func writeStar3(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mk := func(rel, body string) { writeTreeFile(t, root, rel, body) }
	// Scene-polar (r, theta, phi) triples — arbitrary but distinct, spread the three
	// nodes apart so a drag on A demonstrably changes the quantized bearing to both
	// leaves, not just their distance.
	mk("nodes/A/meta.json", `{"id":"A","type":"SrcNode","r":100,"scenePolarR":150,"scenePolarTheta":1.2,"scenePolarPhi":0.3}`)
	mk("nodes/A/outputs/Out.json", `{"name":"Out"}`)
	mk("nodes/B/meta.json", `{"id":"B","type":"SinkNode","r":100,"scenePolarR":100,"scenePolarTheta":1.0,"scenePolarPhi":1.2}`)
	mk("nodes/B/inputs/In.json", `{"name":"In"}`)
	mk("nodes/C/meta.json", `{"id":"C","type":"SinkNode","r":100,"scenePolarR":90,"scenePolarTheta":0.9,"scenePolarPhi":-1.0}`)
	mk("nodes/C/inputs/In.json", `{"name":"In"}`)
	// Both edges share A's single "Out" output port — a fan-OUT, symmetric with the
	// fan-IN this package's srcNode/sinkNode kinds already exercise (two edges sharing
	// one destination "In" port). validate.go only checks the port NAME is declared on
	// the kind, not that a handle is used by at most one edge.
	mk("nodes/A/edges/eAB.json", `{"label":"eAB","kind":"data","sourceHandle":"Out","target":"B","targetHandle":"In"}`)
	mk("nodes/A/edges/eAC.json", `{"label":"eAC","kind":"data","sourceHandle":"Out","target":"C","targetHandle":"In"}`)
	writeCascadeEdgesFromEdges(t, root, map[string]string{"A": "SrcNode", "B": "SinkNode", "C": "SinkNode"},
		[][2]string{{"A", "B"}, {"A", "C"}})
	return root
}

// persistedMeta reads and merges <root>/nodes/<id>/{meta,position,local-polars}.json into
// one raw map, for asserting on the RE-PERSISTED bytes rather than in-memory state. Since
// the one-file-per-writer split the
// position/local-polars fields live in their OWN files, not meta.json; merging here keeps
// every existing by-key assertion in this test unchanged.
func persistedMeta(t *testing.T, root, id string) map[string]json.RawMessage {
	t.Helper()
	m := map[string]json.RawMessage{}
	raw, err := os.ReadFile(filepath.Join(root, "nodes", id, "meta.json"))
	if err != nil {
		t.Fatalf("read %s meta.json: %v", id, err)
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal %s meta.json: %v (raw=%s)", id, err, raw)
	}
	if praw, err := os.ReadFile(positionFilePath(root, id)); err == nil {
		var pm map[string]json.RawMessage
		if err := json.Unmarshal(praw, &pm); err != nil {
			t.Fatalf("unmarshal %s position.json: %v (raw=%s)", id, err, praw)
		}
		for k, v := range pm {
			m[k] = v
		}
	}
	if lraw, err := os.ReadFile(localPolarsFilePath(root, id)); err == nil {
		var lm map[string]json.RawMessage
		if err := json.Unmarshal(lraw, &lm); err != nil {
			t.Fatalf("unmarshal %s local-polars.json: %v (raw=%s)", id, err, lraw)
		}
		for k, v := range lm {
			m[k] = v
		}
	}
	return m
}

// persistedScenePolarCenter reads a persisted meta.json's exact scenePolar triple and
// converts it back to a world center about sceneCenter — the same reconstruction the
// loader performs on reload.
func persistedScenePolarCenter(t *testing.T, m map[string]json.RawMessage, sceneCenter vec3) vec3 {
	t.Helper()
	var p polar
	for key, dst := range map[string]*float64{
		"scenePolarR": &p.R, "scenePolarTheta": &p.Theta, "scenePolarPhi": &p.Phi,
	} {
		raw, ok := m[key]
		if !ok {
			t.Fatalf("meta.json missing %s: keys=%v", key, m)
		}
		if err := json.Unmarshal(raw, dst); err != nil {
			t.Fatalf("unmarshal %s: %v", key, err)
		}
	}
	return sceneCenter.Add(polar2cart(p))
}

// persistedLocalPolarTo reads a persisted meta.json's localPolars entry to a given
// neighbor id, and its persisted local pole.
func persistedLocalPolarTo(t *testing.T, m map[string]json.RawMessage, to string) (quantITheta, quantIPhi, quantIR int, stepTheta, stepPhi, stepR float64, found bool) {
	t.Helper()
	raw, ok := m["localPolars"]
	if !ok {
		return 0, 0, 0, 0, 0, 0, false
	}
	type localPolarJSON struct {
		To          string  `json:"to"`
		QuantITheta int     `json:"quantITheta"`
		QuantIPhi   int     `json:"quantIPhi"`
		QuantIR     int     `json:"quantIR"`
		StepTheta   float64 `json:"stepTheta"`
		StepPhi     float64 `json:"stepPhi"`
		StepR       float64 `json:"stepR"`
	}
	var list []localPolarJSON
	if err := json.Unmarshal(raw, &list); err != nil {
		t.Fatalf("unmarshal localPolars: %v", err)
	}
	for _, lp := range list {
		if lp.To == to {
			return lp.QuantITheta, lp.QuantIPhi, lp.QuantIR, lp.StepTheta, lp.StepPhi, lp.StepR, true
		}
	}
	return 0, 0, 0, 0, 0, 0, false
}

// TestDragPersistsOnlyDraggedNodeAndRequantizesNeighborsOnDisk drives a real drag of the
// hub node A on a 3-node star topology, through the real loader + mover goroutines +
// debounced persisters, and asserts on the RE-PERSISTED meta.json bytes:
//
//  1. A's own persisted scenePolar/center changed to the drag target.
//  2. B's and C's persisted scenePolar/center are UNCHANGED (neighbors stay put).
//  3. B's and C's persisted LocalPolar to A always matches a FRESH quantization of live
//     geometry (re-quantized, never held) — not necessarily CHANGED from before, since
//     the union drag-cell model moves A by exactly one bead in only ONE neighbour's own
//     frame per RootMove call (see checkNeighbor's doc comment below).
//  4. No degenerate step (StepR) was written for any node — it must be one of the two
//     known sane grid constants (wire.DefaultLocalStepR for a per-node local-polar entry, stepR for
//     the scene-level quantized cache), never a near-zero value.
func TestDragPersistsOnlyDraggedNodeAndRequantizesNeighborsOnDisk(t *testing.T) {
	root := writeStar3(t)
	md := loadTreeMD(t, root)
	md.EnableEditPersist(root)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := md.Start(ctx)
	// Registered AFTER t.TempDir(), so t.Cleanup LIFO runs it BEFORE the
	// dir RemoveAll: cancel() only SIGNALS, and a goroutine still closing
	// its stream fds races the cleanup ("bad file descriptor").
	t.Cleanup(func() { cancel(); wg.Wait() })

	lhB, ok := md.lq.layoutHolders["B"]
	if !ok {
		t.Fatal("no LayoutHolder for B")
	}
	lhC, ok := md.lq.layoutHolders["C"]
	if !ok {
		t.Fatal("no LayoutHolder for C")
	}

	centerBefore := map[string]vec3{}
	for _, id := range []string{"A", "B", "C"} {
		c, ok := md.centerOfNode(id)
		if !ok {
			t.Fatalf("no center for %s before drag", id)
		}
		centerBefore[id] = c
	}

	var lpBBefore, lpCBefore wire.LocalPolar
	for _, lp := range lhB.LocalPolarsSnapshot() {
		if lp.To == "A" {
			lpBBefore = lp
		}
	}
	for _, lp := range lhC.LocalPolarsSnapshot() {
		if lp.To == "A" {
			lpCBefore = lp
		}
	}
	if lpBBefore.To != "A" {
		t.Fatal("B has no pre-drag LocalPolar entry for A")
	}
	if lpCBefore.To != "A" {
		t.Fatal("C has no pre-drag LocalPolar entry for A")
	}

	// Sync point for the post-drag lhB/lhC reads below: B and C each write their own
	// requantized LocalPolar entry (SetLocalPolar/SetPole, on their OWN goroutine,
	// inside neighborSetCRequantize) strictly BEFORE logging their own "abc-drag"
	// breadcrumb in that same call — waiting for both breadcrumbs (rather than polling
	// lhB/lhC directly, a data race against B's/C's own mover goroutines) establishes
	// the happens-before edge. See time_node_abc_drag_breadcrumb_test.go.
	var dbg syncBuffer
	md.tr.SetSink(&dbg)

	// Drag A off both leaves' prior bearings. Under the UNION drag-cell model
	// (quantized_move.go's dragCandidateCells) a single RootMove call moves A by exactly
	// ONE bead-index-step in whichever usable neighbour's frame produced the nearest
	// candidate — never a multi-bead walk toward a far target — so this raw displacement
	// is deliberately not large; it only needs to be off-axis enough that the nearest
	// candidate is plausibly an angular (not purely radial) one for the generating
	// neighbour. See checkNeighbor below for what is and isn't guaranteed for B and C.
	target := centerBefore["A"].Add(vec3{X: 5, Y: -3, Z: 2}.Scale(wire.BeadStepR))
	// wantA is the point the drag actually COMMITS to — the scene-lattice-snapped
	// target, not the raw one (docs/which-lattice-a-node-lives-on.md; commitNodeMoveLocal
	// now draws/persists the quantized position, never the continuous raw target).
	wantA := quantizedDragTarget(md, "A", target)
	if !md.RootMove("A", target) {
		t.Fatal("RootMove(A) returned false")
	}
	pollDragConverged(t, md, "A", wantA)

	// Wait for BOTH B's and C's own "abc-drag" breadcrumbs before reading their
	// LocalPolar entries to A.
	waitForAbcDrag(t, &dbg, "B")
	waitForAbcDrag(t, &dbg, "C")

	// The poll above only proves QuantIR eventually changed; it is not proof that no
	// FURTHER unwanted cascade message is in flight toward B or C. neighborSetCRequantize
	// logs exactly one "abc-drag" breadcrumb per moveMsgKindNeighborSetC it handles (see
	// its doc comment), so each of B's/C's abc-drag counts IS its neighborSetC receipt
	// count — a genuine production outcome, not an intercepted in-flight message. The old
	// cascade kinds ("equalize"/"trigger"/"gatePlace"/"requantize") don't exist anywhere
	// in moveMsgKind's vocabulary any more (grep node_move.go's moveMsgKind* constants:
	// only anchor/center/centers/drag/neighborSetC/dragStart remain) — a message of those
	// kinds is unrepresentable by construction, so there is nothing left to tap for. The
	// sleep widens the window for any further (unwanted) breadcrumb to land before the
	// count check below.
	time.Sleep(cascadeSettle)
	for _, id := range []string{"B", "C"} {
		if n := len(abcDragDeltasFor(t, &dbg, id)); n != 1 {
			t.Fatalf("expected exactly one abc-drag (== one moveMsgKindNeighborSetC) delivered to %s; got %d", id, n)
		}
	}

	// In-memory sanity before touching disk: B and C must NOT have moved.
	for _, id := range []string{"B", "C"} {
		c, ok := md.centerOfNode(id)
		if !ok {
			t.Fatalf("no center for %s after drag", id)
		}
		if d := c.Sub(centerBefore[id]).Length(); d > 1e-9 {
			t.Fatalf("%s must stay put on an A drag: before=%+v after=%+v (moved by %g)", id, centerBefore[id], c, d)
		}
	}

	// nodeMover.persistQuantOffset writes synchronously now (no debounce), but it runs
	// a few statements after A's center converges (pollDragConverged above), on that same
	// node-mover goroutine (quantized_move.go commitNodeMoveLocal) — poll the read-back so
	// this does not race ahead of that write landing (same shape as pollDragConverged).
	pollPositionFileWritten(t, root, "A")

	// ---- Read the RE-PERSISTED bytes on disk. ----
	metaA := persistedMeta(t, root, "A")
	metaB := persistedMeta(t, root, "B")
	metaC := persistedMeta(t, root, "C")

	// (a) A's own persisted scenePolar/center changed to the QUANTIZED drag target — the
	// lattice point the commit snapped to, not the raw continuous target.
	gotA := persistedScenePolarCenter(t, metaA, md.ui.sceneSphere.Center)
	if d := gotA.Sub(wantA).Length(); d > 1e-6 {
		t.Fatalf("(a) A's persisted center should equal the quantized drag target: persisted=%+v want=%+v raw-target=%+v (off by %g)", gotA, wantA, target, d)
	}

	// (b) B's and C's persisted scenePolar/center are UNCHANGED.
	gotB := persistedScenePolarCenter(t, metaB, md.ui.sceneSphere.Center)
	if d := gotB.Sub(centerBefore["B"]).Length(); d > 1e-6 {
		t.Fatalf("(b) B's persisted center must stay put on an A drag: pre-drag=%+v persisted=%+v (off by %g)", centerBefore["B"], gotB, d)
	}
	gotC := persistedScenePolarCenter(t, metaC, md.ui.sceneSphere.Center)
	if d := gotC.Sub(centerBefore["C"]).Length(); d > 1e-6 {
		t.Fatalf("(b) C's persisted center must stay put on an A drag: pre-drag=%+v persisted=%+v (off by %g)", centerBefore["C"], gotC, d)
	}

	// (c) B's and C's persisted LocalPolar to A must match a FRESH quantization of the
	// live post-drag geometry — re-quantized, never held stale — computed the same way
	// requantizePoleTraced does at the cart<->polar boundary (same recipe this package's
	// neighbor_setc_test.go / subtree_persist_test.go already use in-memory; here read
	// back from the PERSISTED bytes).
	//
	// UNCHANGED VALUES ARE NOT AN ERROR here, deliberately, unlike the older walkBeadPath
	// model this replaced: under the union drag-cell model a single RootMove moves A by
	// exactly ONE bead-index-step in whichever neighbour's frame produced the nearest
	// candidate (dragCandidateCells' doc comment), so ONLY that generating neighbour is
	// guaranteed a change at all — see TestTangencyAfterDrag_LastBeadExactlyReachesTargetSurface
	// (drag_cell_test.go) for the exactness that neighbour gets. The OTHER neighbour's
	// distance/bearing to A may quantize to the SAME index it already had (a move that
	// happens to land within the same rounding cell for a neighbour it wasn't generated
	// from) — that is accepted and expected, not a regression, which is why this check
	// only pins "matches a fresh quantization", not "changed from before".
	checkNeighbor := func(id string, meta map[string]json.RawMessage, lh *wire.LayoutHolder, before wire.LocalPolar) {
		t.Helper()
		qTheta, qPhi, qR, stTheta, stPhi, stR, found := persistedLocalPolarTo(t, meta, "A")
		if !found {
			t.Fatalf("%s's persisted meta.json has no localPolars entry for A: keys=%v", id, meta)
		}
		_ = before // kept as a parameter for checkNeighbor's call-site symmetry; no longer compared
		// Recompute the expected fresh quantization from live post-drag geometry and
		// compare, exactly as neighbor_setc_test.go's in-memory assertion (2) does.
		selfCenter, ok := md.centerOfNode(id)
		if !ok {
			t.Fatalf("no center for %s to recompute expected quantization", id)
		}
		aCenter, ok := md.centerOfNode("A")
		if !ok {
			t.Fatal("no center for A to recompute expected quantization")
		}
		offset := aCenter.Sub(selfCenter)
		d, r := dirFromOffset(offset)
		pole := lh.Pole()
		c, psi := azimuthFrom(pole, d)
		st, sp, sr := lh.LocalPolarSteps("A")
		wantTheta := int(math.Round(c / st))
		wantPhi := int(math.Round(psi / sp))
		// QuantIR is stored verbatim (SnapQuantIR no longer exists — there is one
		// lattice now, bead_lattice.go), so the persisted value is exactly the raw
		// fresh quantization.
		wantR := int(math.Round(r / sr))
		if qTheta != wantTheta || qPhi != wantPhi || qR != wantR {
			t.Fatalf("(c) %s's persisted local polar to A should match a fresh quantization of the live offset: got=(theta=%d,phi=%d,r=%d) want=(theta=%d,phi=%d,r=%d)",
				id, qTheta, qPhi, qR, wantTheta, wantPhi, wantR)
		}

		// (d) No degenerate step got written — StepR must be the sane local-polar grid
		// constant (wire.DefaultLocalStepR), never a near-zero value like 1e-06. Checking equality to
		// wire.DefaultLocalStepR (a known-positive constant) already subsumes any "is it near zero"
		// check, so there is no separate degenerate-threshold branch here.
		if stR != wire.DefaultLocalStepR {
			t.Fatalf("(d) %s's persisted local-polar StepR should be the sane grid constant wire.DefaultLocalStepR=%g, got %g", id, wire.DefaultLocalStepR, stR)
		}
		// StepTheta/StepPhi are no longer a fixed literal (wire.DefaultLocalStepTheta/Phi)
		// — they're r-DEPENDENT (wire.AngularStepsForR, layout_holder.go), one bead of ARC
		// at THIS entry's own radius, so the sane-value check recomputes the SAME oracle
		// from the fresh (raw, unrounded) radius/colatitude — r/c, already computed above
		// via dirFromOffset/azimuthFrom — rather than comparing to the old fixed constant.
		// Using the RAW r/c (not wantR*sr, which re-rounds through the index first) keeps
		// this oracle from disagreeing with production over nothing but that rounding.
		wantStepTheta, _ := wire.AngularStepsForR(r, 0)
		_, wantStepPhi := wire.AngularStepsForR(r, c)
		if stTheta != wantStepTheta || stPhi != wantStepPhi {
			t.Fatalf("(d) %s's persisted local-polar StepTheta/StepPhi should be the r-derived grid constants (%g,%g), got (%g,%g)",
				id, wantStepTheta, wantStepPhi, stTheta, stPhi)
		}
	}
	checkNeighbor("B", metaB, lhB, lpBBefore)
	checkNeighbor("C", metaC, lhC, lpCBefore)

	// (d, continued) A's own persisted scene-level quantized-cache step is also sane,
	// never degenerate.
	const degenerate = 1e-6
	for _, key := range []string{"stepTheta", "stepPhi", "stepR"} {
		raw, ok := metaA[key]
		if !ok {
			t.Fatalf("A's persisted meta.json missing %s: keys=%v", key, metaA)
		}
		var v float64
		if err := json.Unmarshal(raw, &v); err != nil {
			t.Fatalf("unmarshal A %s: %v", key, err)
		}
		if v <= degenerate {
			t.Fatalf("A's persisted %s is degenerate: %g", key, v)
		}
	}

	// Reload from disk into a completely fresh MoveDispatch (a second, independent
	// process-restart-equivalent load) and re-derive the same three invariants purely
	// from what LoadTopology reconstructs — the strongest form of "read persisted bytes":
	// actually feeding them back through the real loader.
	md2 := loadTreeMD(t, root)
	gotA2, ok := md2.centerOfNode("A")
	if !ok {
		t.Fatal("A missing after reload")
	}
	if d := gotA2.Sub(wantA).Length(); d > 1e-6 {
		t.Fatalf("reload: A did not round-trip to the quantized drag target: got=%+v want=%+v raw-target=%+v", gotA2, wantA, target)
	}
	gotB2, ok := md2.centerOfNode("B")
	if !ok {
		t.Fatal("B missing after reload")
	}
	if d := gotB2.Sub(centerBefore["B"]).Length(); d > 1e-6 {
		t.Fatalf("reload: B should still be at its pre-drag center: got=%+v want=%+v", gotB2, centerBefore["B"])
	}
	gotC2, ok := md2.centerOfNode("C")
	if !ok {
		t.Fatal("C missing after reload")
	}
	if d := gotC2.Sub(centerBefore["C"]).Length(); d > 1e-6 {
		t.Fatalf("reload: C should still be at its pre-drag center: got=%+v want=%+v", gotC2, centerBefore["C"])
	}
}
