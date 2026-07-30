package Wiring

import (
	"context"
	"encoding/json"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"math"
	"os"
	"testing"
	"time"
)

// pollDragConverged waits until the named node's committed center matches target — a
// drag now always runs asynchronously on the node's OWN mover goroutine (moveMsgKindDrag,
// node6-drag-decentralized.md generalized to every node), so RootMove returning true only
// means the message was ENQUEUED, not that commitLocal (and its quantOffsetPersist.schedule
// call) has run yet. Tests that read persisted state right after RootMove must wait for
// this convergence first, exactly as the node_move_test.go cascade tests already do.
func pollDragConverged(t *testing.T, md *MoveDispatch, nodeID string, target vec3) {
	t.Helper()
	const eps = 1e-6
	deadline := time.Now().Add(2 * time.Second)
	for {
		// centerOfNode drains each nodeMover's centerOut channel into the dispatch's own
		// centerMirror (drainCenterMirror) and reads that — same path production's
		// RunStdinReader uses.
		c, ok := md.centerOfNode(nodeID)
		if ok && math.Abs(c.X-target.X) <= eps && math.Abs(c.Y-target.Y) <= eps && math.Abs(c.Z-target.Z) <= eps {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("node %s drag never converged to target %+v", nodeID, target)
		}
		time.Sleep(time.Millisecond)
	}
}

// pollPositionFileWritten waits until <root>/nodes/<id>/position.json exists on disk.
// nodeMover.persistQuantOffset now writes synchronously (no debounce — see
// scene_persist.go's header comment), but pollDragConverged only guarantees the dragged
// node's CENTER has published (applyCenter's atomic snap store); commitNodeMoveLocal's
// nm.persistQuantOffset() call runs a few statements LATER on that SAME node-mover
// goroutine (quantized_move.go commitNodeMoveLocal), so reading disk immediately after
// convergence can still race ahead of that write landing. Polling the read-back under a
// deadline (same shape as pollDragConverged) makes the wait for "the write has run"
// observable without reaching into mover internals.
func pollPositionFileWritten(t *testing.T, root, nodeID string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		if readJSONIfExists(positionFilePath(root, nodeID), &positionFileJSON{}) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("drag never wrote the new position.json for %s", nodeID)
		}
		time.Sleep(time.Millisecond)
	}
}

// Individual snapping: dragging a node moves and persists ONLY that node (its grid-snapped
// scalar triple, quantITheta/quantIPhi/quantIR — the sole persisted position source under
// the plain-polar model), leaving every other node untouched — no subtree cascade.
func TestIndividualSnap_OnlyDraggedNodePersists(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)
	md.EnableEditPersist(root)
	// Every drag (moveMsgKindDrag, node6-drag-decentralized.md generalized to every
	// node) commits on the dragged node's OWN mover goroutine — Start the movers so
	// something drains dst's inbox.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := md.Start(ctx)
	// Registered AFTER t.TempDir(), so t.Cleanup LIFO runs it BEFORE the
	// dir RemoveAll: cancel() only SIGNALS, and a goroutine still closing
	// its stream fds races the cleanup ("bad file descriptor").
	t.Cleanup(func() { cancel(); wg.Wait() })

	lhSrc, ok := md.lq.layoutHolders["src"]
	if !ok {
		t.Fatal("no LayoutHolder for src")
	}
	var lpBefore wire.LocalPolar
	for _, lp := range lhSrc.LocalPolarsSnapshot() {
		if lp.To == "dst" {
			lpBefore = lp
		}
	}
	srcCenterBefore, ok := md.centerOfNode("src")
	if !ok {
		t.Fatal("no center for src before drag")
	}

	// Sync point for the post-drag lhSrc read below: src (the neighbor, NOT the dragged
	// node) writes its own requantized LocalPolar entry (SetLocalPolar/SetPole, on src's
	// OWN goroutine, inside neighborSetCRequantize) strictly BEFORE it logs its
	// "abc-drag" breadcrumb in that same call — waiting for the breadcrumb (rather than
	// polling lhSrc directly, a data race against src's own mover goroutine) establishes
	// the happens-before edge. See time_node_abc_drag_breadcrumb_test.go.
	var dbg syncBuffer
	md.tr.SetSink(&dbg)

	dstTarget := vec3{X: 60, Y: 20, Z: -10}
	if !md.RootMove("dst", dstTarget) {
		t.Fatal("RootMove(dst) returned false")
	}
	pollDragConverged(t, md, "dst", dstTarget)

	// src is dst's plain (role-free) direct neighbor: per the single-assignment set-c
	// REQUANTIZE model (node_move.go moveMsgKindNeighborSetC / neighborSetCRequantize),
	// src STAYS PUT — only dst moved — and re-quantizes its OWN stored local polar
	// (QuantITheta/QuantIPhi/QuantIR) to dst fresh from the live offset. Wait for src's
	// own "abc-drag" breadcrumb (see the sync-point comment above) before reading.
	waitForAbcDrag(t, &dbg, "src")
	var lpAfter wire.LocalPolar
	for _, lp := range lhSrc.LocalPolarsSnapshot() {
		if lp.To == "dst" {
			lpAfter = lp
		}
	}
	if lpAfter.QuantIR == lpBefore.QuantIR {
		t.Fatalf("src's local polar to dst never picked up the requantize: before=%+v after=%+v", lpBefore, lpAfter)
	}

	// src's own world center must NOT have moved — only dst moved.
	srcCenter, ok := md.centerOfNode("src")
	if !ok {
		t.Fatal("no center for src after drag")
	}
	if d := srcCenter.Sub(srcCenterBefore).Length(); d > 1e-9 {
		t.Fatalf("src must stay put on a dst drag: before=%+v after=%+v (moved by %g)", srcCenterBefore, srcCenter, d)
	}

	// src's requantized local polar to dst must match a fresh quantization of the live
	// offset (dst_newcenter - src_center) about src's own pole.
	dstCenter, ok := md.centerOfNode("dst")
	if !ok {
		t.Fatal("no center for dst after drag")
	}
	offset := dstCenter.Sub(srcCenter)
	dd, rr := dirFromOffset(offset)
	cc, psi := azimuthFrom(lhSrc.Pole(), dd)
	st, sp, sr := lpAfter.EffectiveSteps()
	wantTheta := int(math.Round(cc / st))
	wantPhi := int(math.Round(psi / sp))
	// QuantIR is snapped to the bead lattice at every write (wire.SnapQuantIR, called from
	// SetLocalPolar) — the raw round(rr/sr) is not itself the stored value, so this must
	// snap too or a raw round landing on the wrong side of a bead-step boundary would
	// mismatch by 1. Previously invisible at the old LocalStepR=2.0 (this fixture's numbers
	// happened not to straddle a boundary); the flip to the bead-authored LocalStepR=2.24
	// (docs/bead-lattice.md "The lattice is derived, not the bead") moved the boundary and
	// exposed the missing snap.
	wantR := wire.SnapQuantIR(int(math.Round(rr / sr)))
	if lpAfter.QuantITheta != wantTheta || lpAfter.QuantIPhi != wantPhi || lpAfter.QuantIR != wantR {
		t.Fatalf("src's requantized local polar to dst should match a fresh quantization of the live offset: got=(theta=%d,phi=%d,r=%d) want=(theta=%d,phi=%d,r=%d)",
			lpAfter.QuantITheta, lpAfter.QuantIPhi, lpAfter.QuantIR, wantTheta, wantPhi, wantR)
	}

	pollPositionFileWritten(t, root, "dst")

	// dst's meta got its EXACT scene-polar position (the lossless source of truth loaded
	// verbatim on reload) plus the quantized scalar triple as a self-describing cache; src
	// is byte-for-byte unchanged.
	dstRaw, err := os.ReadFile(positionFilePath(root, "dst"))
	if err != nil {
		t.Fatalf("read dst position.json: %v", err)
	}
	var dst map[string]json.RawMessage
	_ = json.Unmarshal(dstRaw, &dst)
	// These were once PRESENCE checks (`if _, ok := dst[k]; !ok`), which a mutation audit
	// showed could not fail: `ok` is true for ANY value, so corrupting the written
	// quantIR by +12345 still passed. A persisted position is only meaningful by VALUE,
	// so decode and compare — the same standard the src block below already applied.
	for _, k := range []string{"scenePolarR", "scenePolarTheta", "scenePolarPhi", "quantITheta", "quantIPhi", "quantIR"} {
		if _, ok := dst[k]; !ok {
			t.Fatalf("dst %s not persisted (exact position is the source of truth): %s", k, dstRaw)
		}
	}
	gotDstCenter := persistedScenePolarCenter(t, dst, md.ui.sceneSphere.Center)
	if d := gotDstCenter.Sub(dstTarget).Length(); d > 1e-6 {
		t.Fatalf("dst's persisted scenePolar should equal the drag target: persisted=%+v target=%+v (off by %g)", gotDstCenter, dstTarget, d)
	}
	// The quantized triple is a self-describing CACHE of that same position: it must be a
	// fresh quantization of what was persisted, not a stale or arbitrary value.
	var gotQT, gotQP, gotQR int
	for k, into := range map[string]*int{"quantITheta": &gotQT, "quantIPhi": &gotQP, "quantIR": &gotQR} {
		if err := json.Unmarshal(dst[k], into); err != nil {
			t.Fatalf("unmarshal dst %s: %v", k, err)
		}
	}
	// Oracle is deliberately INDEPENDENT of the production quantization formula: index ×
	// its own persisted step must reconstruct the persisted polar to within one quantum.
	// Re-running the production quantizer here would be circular — it would agree with
	// itself no matter what either of them computed.
	var stepT, stepP, stepR float64
	for k, into := range map[string]*float64{"stepTheta": &stepT, "stepPhi": &stepP, "stepR": &stepR} {
		if err := json.Unmarshal(dst[k], into); err != nil {
			t.Fatalf("unmarshal dst %s: %v", k, err)
		}
	}
	var dstPolar polar
	for k, into := range map[string]*float64{"scenePolarR": &dstPolar.R, "scenePolarTheta": &dstPolar.Theta, "scenePolarPhi": &dstPolar.Phi} {
		if err := json.Unmarshal(dst[k], into); err != nil {
			t.Fatalf("unmarshal dst %s: %v", k, err)
		}
	}
	for _, c := range []struct {
		name     string
		idx      int
		step, of float64
	}{
		{"quantITheta", gotQT, stepT, dstPolar.Theta},
		{"quantIPhi", gotQP, stepP, dstPolar.Phi},
		{"quantIR", gotQR, stepR, dstPolar.R},
	} {
		if c.step <= 0 {
			t.Fatalf("dst persisted a non-positive step for %s: %g", c.name, c.step)
		}
		if d := math.Abs(float64(c.idx)*c.step - c.of); d > c.step {
			t.Fatalf("dst's persisted %s=%d × step %g = %g, which is %g away from the persisted value %g — the quant cache does not describe the position it is cached against",
				c.name, c.idx, c.step, float64(c.idx)*c.step, d, c.of)
		}
	}

	// src's requantized local polar to dst must be ON DISK, not merely in memory. The
	// assertion above (lpAfter, from lhSrc.LocalPolarsSnapshot) reads the LIVE
	// LayoutHolder — the audit confirmed it passes even with WriteLocalPolars persisting
	// nothing at all, so it says nothing about persistence despite this test's name.
	var srcLP localPolarsFileJSON
	if !readJSONIfExists(localPolarsFilePath(root, "src"), &srcLP) {
		t.Fatalf("src's local-polars.json was never written, so the requantize did not persist")
	}
	var diskLP localPolarJSON
	var foundLP bool
	for _, lp := range srcLP.LocalPolars {
		if lp.To == "dst" {
			diskLP, foundLP = lp, true
		}
	}
	if !foundLP {
		t.Fatalf("src's local-polars.json has no entry for dst: %+v", srcLP.LocalPolars)
	}
	if diskLP.QuantITheta != lpAfter.QuantITheta || diskLP.QuantIPhi != lpAfter.QuantIPhi || diskLP.QuantIR != lpAfter.QuantIR {
		t.Fatalf("src's PERSISTED local polar to dst should match its requantized live value: disk=(theta=%d,phi=%d,r=%d) live=(theta=%d,phi=%d,r=%d)",
			diskLP.QuantITheta, diskLP.QuantIPhi, diskLP.QuantIR, lpAfter.QuantITheta, lpAfter.QuantIPhi, lpAfter.QuantIR)
	}

	// src's persisted position must reflect its UNCHANGED scene-polar position: under the
	// single-assignment set-c REQUANTIZE model, a dst drag never moves src — only src's
	// local-polar edge to dst (its own requantized bearing/distance) changes, not its own
	// scene position, so src's own position.json is never (re)written; its scenePolar still
	// lives in meta.json (the legacy inline location) — persistedMeta reads both and merges,
	// exactly as loadTree does. Assert src's persisted scenePolar still matches its pre-drag
	// world center.
	srcA := persistedMeta(t, root, "src")
	for _, k := range []string{"scenePolarR", "scenePolarTheta", "scenePolarPhi"} {
		if _, ok := srcA[k]; !ok {
			t.Fatalf("src %s not persisted: %v", k, srcA)
		}
	}
	var gotP polar
	if err := json.Unmarshal(srcA["scenePolarR"], &gotP.R); err != nil {
		t.Fatalf("unmarshal src scenePolarR: %v", err)
	}
	if err := json.Unmarshal(srcA["scenePolarTheta"], &gotP.Theta); err != nil {
		t.Fatalf("unmarshal src scenePolarTheta: %v", err)
	}
	if err := json.Unmarshal(srcA["scenePolarPhi"], &gotP.Phi); err != nil {
		t.Fatalf("unmarshal src scenePolarPhi: %v", err)
	}
	gotCenter := md.ui.sceneSphere.Center.Add(polar2cart(gotP))
	if d := gotCenter.Sub(srcCenterBefore).Length(); d > 1e-6 {
		t.Fatalf("src's persisted scenePolar should still match its pre-drag (unmoved) world center: persisted=%+v pre-drag=%+v", gotCenter, srcCenterBefore)
	}
}

// TestDragPositionRoundTripsExactly: dragging a node to an arbitrary continuous target,
// persisting, and RELOADING from disk must place the node at EXACTLY that target — the
// exact scene-polar position is the lossless source of truth (not the coarse quantized
// triple, which would round the drag away).
func TestDragPositionRoundTripsExactly(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)
	md.EnableEditPersist(root)
	// Every drag commits on the dragged node's OWN mover goroutine — Start the movers
	// so something drains dst's inbox (node6-drag-decentralized.md, generalized).
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := md.Start(ctx)
	// Registered AFTER t.TempDir(), so t.Cleanup LIFO runs it BEFORE the
	// dir RemoveAll: cancel() only SIGNALS, and a goroutine still closing
	// its stream fds races the cleanup ("bad file descriptor").
	t.Cleanup(func() { cancel(); wg.Wait() })

	target := vec3{X: 63.7, Y: -21.3, Z: 44.9}
	if !md.RootMove("dst", target) {
		t.Fatal("RootMove(dst) returned false")
	}
	pollDragConverged(t, md, "dst", target)
	pollPositionFileWritten(t, root, "dst")

	// Reload from disk into a fresh MoveDispatch and read dst's center back.
	md2 := loadTreeMD(t, root)
	got, ok := md2.centerOfNode("dst")
	if !ok {
		t.Fatal("dst missing after reload")
	}
	const eps = 1e-6
	if d := got.Sub(target).Length(); d > eps {
		t.Fatalf("dst did not round-trip: dragged to %+v, reloaded at %+v (off by %g)", target, got, d)
	}
}
