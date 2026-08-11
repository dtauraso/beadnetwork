// build_move_dispatch.go — the MoveDispatch-construction phase of buildFromSpec:
// builds the mover graph from initial geometry and edge endpoints, then seeds every
// per-node field (quantized offset, scene flags, self kind, neighbor kinds, outgoing
// chain targets) that a mover needs before any node goroutine exists.

package dispatch

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/scene"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
)

// buildMoveDispatch builds the MoveDispatch from initial geometry and edge
// endpoints. It creates one nodeMover per node and one edgeMover per edge; each
// owns its geometry and recomputes itself on a node-move (no central
// coordinator). The trace lets each mover stream its own node/edge geometry on a
// move. Outs + dest wires are bound later (bindDispatch) once node construction
// has populated them. Also declares the double-link movement graph (links.go;
// polar locks ride on it in a later step — the lock system and the central polar
// position store have been removed, so node positions live in the movers' held
// geometry) and installs the aimed-port registry for drag-time aiming.
func (b *buildCtx) buildMoveDispatch() error {
	// SPEC order (b.spec.Nodes/Edges — the deterministic directory-sorted order parseSpec
	// read the topology in), NOT map iteration order, so the buffer's row seed
	// (md.GS.NodeSeedsFn/EdgeSeedsFn) gives every node/edge a deterministic row.
	nodeOrder := make([]string, len(b.spec.Nodes))
	for i, n := range b.spec.Nodes {
		nodeOrder[i] = n.ID
	}
	edgeOrder := make([]string, len(b.spec.Edges))
	for i, e := range b.spec.Edges {
		edgeOrder[i] = e.Label
	}
	md, err := newMoveDispatch(b.nodeGeoms, b.edgeEndpoints, b.tr, nodeOrder, edgeOrder, b.clk, &b.speedSinks, b.spec.RowCount)
	if err != nil {
		return fmt.Errorf("buildMoveDispatch: %w", err)
	}
	// Seed the scene's lattice point count from view/lattice.json BEFORE buildNodes runs,
	// so BuildArgs.LatticePointsSeed (called from each PairNode's own build func) hands back
	// the loaded count rather than the compile-time default.
	scenepersist.LoadLatticePoints(&md.UI, b.scenePath)
	if b.hasScene {
		// Persisted scene sphere: install it now so md.UI.SceneSphere is consistent straight out
		// of LoadTopology (a fresh/legacy scene has none — main.go's LoadSceneSphere then
		// content-fits it from the loaded node centers).
		md.UI.SceneSphere = b.sphere
	}

	// The quantized layout is authoritative by default — b.quantizedOffsets was already
	// resolved (stored offset, or measured from the pre-quantized center) by
	// computeQuantizedLayout, which also overwrote b.nodeGeoms so the nodeMovers newMoveDispatch
	// just built above are already seeded from the composed centers. Seed each node's OWN
	// mover field (nodeMover.quantOffset) from it here — there is no shared md.quantizedOffsets
	// map anymore (that map, read/written by multiple mover goroutines for different keys,
	// was the "concurrent map read and map write" fatal fixed by node6-drag-decentralized.md's
	// per-node ownership). A node missing an entry in b.quantizedOffsets keeps its
	// nodeMover's zero-value quantOffset, matching the old map's zero-value-on-miss read.
	// PER SCENE, not always-on (scene/scene_tabs.go's QuantizedDrag): a bead-distance step is
	// invisible in a scene that is large against it and dominant in one that is not.
	md.lq.QuantizedLayout = scene.SceneUsesQuantizedDrag(b.scenePath)
	// Per scene as well (scene/scene_tabs.go's CoplanarEdges): each node's own copy, set here on
	// the single-threaded build path, read afterwards only by that node's own goroutine.
	coplanarEdges := scene.SceneWantsCoplanarEdges(b.scenePath)
	upAxis := scene.SceneWantsUpAxis(b.scenePath)
	if coplanarEdges || upAxis {
		for _, nm := range md.mr.NodeGeoms() {
			nm.SetSceneFlags(coplanarEdges, upAxis)
		}
	}
	for id, off := range b.quantizedOffsets {
		if nm, ok := md.mr.NodeGeoms()[id]; ok {
			nm.SetQuantOffset(off)
		}
	}
	// Seed each node's OWN selfKind (specNode.Type), set once at construction.
	for _, n := range b.spec.Nodes {
		nm, ok := md.mr.NodeGeoms()[n.ID]
		if !ok {
			continue
		}
		nm.SetSelfKind(n.Type)
		if n.TopTiltVectorThetaIdx != nil {
			nm.SetTopTiltVectorThetaIdx(*n.TopTiltVectorThetaIdx)
		}
	}
	// Seed each node's OWN neighborKinds map — every DIRECT domain-adjacent neighbor id
	// mapped to that neighbor's own kind name, derived straight from the loaded spec's
	// node list + edges (no separate persisted file: adjacency is already known from
	// b.spec.Edges, and a node's kind is already known from b.spec.Nodes, so keeping a
	// second stored copy in sync would only be a second place for the two to drift).
	// UNDIRECTED: both endpoints of every edge learn the other's kind.
	kindByID := make(map[string]string, len(b.spec.Nodes))
	for _, n := range b.spec.Nodes {
		kindByID[n.ID] = n.Type
	}
	linkNeighborKind := func(fromID, toID string) {
		nm, ok := md.mr.NodeGeoms()[fromID]
		if !ok {
			return
		}
		nm.AddNeighborKind(toID, kindByID[toID])
	}
	for _, e := range b.spec.Edges {
		linkNeighborKind(e.Source, e.Target)
		linkNeighborKind(e.Target, e.Source)
	}
	// Seed each node's OWN outgoing-edge targets (nodeMover.outTargets) — the chains it
	// owns (chain_beads.go). A chain belongs to exactly one endpoint: the source, matching
	// where the edge is stored on disk.
	for _, e := range b.spec.Edges {
		nm, ok := md.mr.NodeGeoms()[e.Source]
		if !ok {
			continue
		}
		nm.AddOutTarget(e.Target)
	}
	b.md = md
	return nil
}
