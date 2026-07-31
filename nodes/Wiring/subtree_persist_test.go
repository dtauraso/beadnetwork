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

// quantizedDragTarget returns the position a drag to target actually COMMITS to under
// the scene lattice — walkBeadPath from the node's CURRENT (pre-drag) center toward
// target, in whole bead-length strides (quantized_move.go's "one bead of arc in every
// direction" model, docs/bead-lattice.md) — the raw target unchanged when quantizedLayout
// is off. Test callers that assert convergence (pollDragConverged) must poll for THIS
// point, not the raw target, now that a committed drag is snapped rather than continuous
// (docs/which-lattice-a-node-lives-on.md). MUST be called BEFORE the drag commits (reads
// the node's pre-drag center as the walk's starting point) — this replaced an earlier
// version that derived the target from a fixed-1-degree-angular-tick quantized triple
// (measureScalar/offsetScenePolar), independent of the node's current position; the walk
// model is NOT independent of it, so this function is no longer stable to call after the
// drag has already moved the node.
func quantizedDragTarget(md *MoveDispatch, nodeID string, target vec3) vec3 {
	if !md.lq.quantizedLayout {
		return target
	}
	prev, ok := md.centerOfNode(nodeID)
	if !ok {
		return target
	}
	return walkBeadPath(prev, target)
}

// pollDragConverged waits until the named node's committed center matches the point a
// drag to target actually commits to (want, from quantizedDragTarget) — a drag now
// always runs asynchronously on the node's OWN mover goroutine (moveMsgKindDrag,
// node6-drag-decentralized.md generalized to every node), so RootMove returning true only
// means the message was ENQUEUED, not that commitLocal (and its quantOffsetPersist.schedule
// call) has run yet. Tests that read persisted state right after RootMove must wait for
// this convergence first, exactly as the node_move_test.go cascade tests already do.
//
// want MUST be computed by the CALLER, via quantizedDragTarget, BEFORE calling RootMove
// — not recomputed here. quantizedDragTarget now walks from the node's CURRENT (pre-drag)
// center (walkBeadPath), so calling it AFTER RootMove has already been issued races the
// dragged node's own mover goroutine: if the commit has landed by the time this function
// reads the center, "current" is already the POST-drag position, and a fresh
// quantizedDragTarget call would walk another stride past the point this poll is
// actually waiting for. Passing want in avoids that race entirely.
func pollDragConverged(t *testing.T, md *MoveDispatch, nodeID string, want vec3) {
	t.Helper()
	const eps = 1e-6
	deadline := time.Now().Add(2 * time.Second)
	for {
		// centerOfNode drains each nodeMover's centerOut channel into the dispatch's own
		// centerMirror (drainCenterMirror) and reads that — same path production's
		// RunStdinReader uses.
		c, ok := md.centerOfNode(nodeID)
		if ok && math.Abs(c.X-want.X) <= eps && math.Abs(c.Y-want.Y) <= eps && math.Abs(c.Z-want.Z) <= eps {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("node %s drag never converged to quantized target %+v", nodeID, want)
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

// pollLocalPolarsFileWritten waits until <root>/nodes/<id>/local-polars.json exists on
// disk. Sibling of pollPositionFileWritten above, and it exists because the in-memory
// sync point CANNOT cover the disk write.
//
// neighborSetCRequantize emits its "abc-drag" breadcrumb BEFORE it calls
// persistLocalPolars (quantized_move.go — the breadcrumb sits about eighteen lines above
// the persist), so waitForAbcDrag establishes happens-before for the LayoutHolder read
// and nothing more. Borrowing it for a DISK assertion, as this test used to, reads the
// file in the window between the breadcrumb and the write: it passed almost always and
// failed under load with "local-polars.json was never written". Verified by widening that
// window with a temporary sleep before the persist — without this poll the test fails
// with exactly that message, with it the test passes.
//
// Polling rather than moving the breadcrumb after the persist: production ordering should
// not be rearranged to serve a test, and every other waiter on that breadcrumb would then
// be blocked behind disk I/O it does not care about.
func pollLocalPolarsFileWritten(t *testing.T, root, nodeID string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		if readJSONIfExists(localPolarsFilePath(root, nodeID), &localPolarsFileJSON{}) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("drag never wrote local-polars.json for %s, so the requantize did not persist", nodeID)
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
	dstWant := quantizedDragTarget(md, "dst", dstTarget)
	if !md.RootMove("dst", dstTarget) {
		t.Fatal("RootMove(dst) returned false")
	}
	pollDragConverged(t, md, "dst", dstWant)

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
	// QuantIR is stored verbatim (SnapQuantIR no longer exists — there is one
	// lattice now, bead_lattice.go's BeadStepCells doc comment), so the raw
	// round(rr/sr) IS the stored value, with nothing left to snap.
	wantR := int(math.Round(rr / sr))
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
	if d := gotDstCenter.Sub(dstWant).Length(); d > 1e-6 {
		t.Fatalf("dst's persisted scenePolar should equal the quantized drag target: persisted=%+v want=%+v raw-target=%+v (off by %g)", gotDstCenter, dstWant, dstTarget, d)
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
	pollLocalPolarsFileWritten(t, root, "src")
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
// persisting, and RELOADING from disk must place the node at EXACTLY the point the drag
// COMMITTED to — the scene-lattice-snapped point (quantizedDragTarget), now that a
// commit draws/persists the quantized position rather than the raw continuous target
// (docs/which-lattice-a-node-lives-on.md). The exact scene-polar position persisted
// (scenePolarR/Theta/Phi) is still the LOSSLESS round-trip source of truth for THAT
// committed point — it is not re-rounded a second time on reload, which is what this
// test still guards: a reload must not drift further away from the committed point,
// only the drag itself is allowed to snap.
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
	want := quantizedDragTarget(md, "dst", target)
	if !md.RootMove("dst", target) {
		t.Fatal("RootMove(dst) returned false")
	}
	pollDragConverged(t, md, "dst", want)
	pollPositionFileWritten(t, root, "dst")

	// Reload from disk into a fresh MoveDispatch and read dst's center back.
	md2 := loadTreeMD(t, root)
	got, ok := md2.centerOfNode("dst")
	if !ok {
		t.Fatal("dst missing after reload")
	}
	const eps = 1e-6
	if d := got.Sub(want).Length(); d > eps {
		t.Fatalf("dst did not round-trip: dragged to %+v (committed to %+v), reloaded at %+v (off by %g)", target, want, got, d)
	}
}
