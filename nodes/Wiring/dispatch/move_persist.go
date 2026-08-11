package dispatch

// EnableViewpointPersist arms gesture-driven camera persistence: every subsequent
// EmitViewpoint (orbit/zoom/pan/home) writes the current viewpoint to
// `<topologyPath>/view/camera.json` (nodes/Wiring/camerapersist, via md.Persist —
// nodes/Wiring/viewpersist). Call AFTER SeedInitialViewpoint so the seed's own emit does
// not write the loaded/default pose back. Go owns this write (MODEL.md); the old path
// persists the camera via its own TS scene-save.
func (md *MoveDispatch) EnableViewpointPersist(topologyPath string) {
	p := md.Persist.ArmViewpoint(topologyPath)
	md.UI.VP.Persist = p.Schedule
}

// EnableEditPersist arms disk persistence for the FSM-applied topology edits:
//   - node-drag (RootMove) → the moved node's own position.json + local-polars.json,
//     written by that node's OWN mover (nm.persistRoot, set below on every mover — see
//     quant_offset_persist.go's persistQuantOffset/persistLocalPolars)
//   - ring-move (applyRingAnchor) → the moved port's own node's OWN mover writes the
//     port's anchorId back to the port json file (scene_anchor_persist.go's
//     persistPortAnchor), same nm.persistRoot as above
//   - overlays (applyUpdate toggle/set) → overlay-visibility keys in view/overlays.json
//   - clock speed (applyUpdate clock/speed) → view/speed.json (scene_speed_persist.go)
//   - pair lattice point count (applyUpdate scene/latticePoints) → view/lattice.json
//     (scene_lattice_persist.go)
//
// topologyPath is always the tree root directory — LoadTopology rejects anything else
// (topo_spec.go) — so it is used directly as root for the per-node/per-port persisters.
// Call AFTER SeedInitialViewpoint so the seed emits do not write the loaded state back.
func (md *MoveDispatch) EnableEditPersist(topologyPath string) {
	root := topologyPath
	// The LOADED scene's own root, kept so a structural edit (scene_structure.go's node
	// create/delete) can write into the tree that is actually showing. Every other persister
	// already closes over a path derived from it; this is the one operation that needs
	// the root itself, because it creates and removes whole node directories rather than
	// rewriting one known file.
	md.Scenes.TreeRoot = root
	md.Persist.ArmEdit(topologyPath)
	// Every node's own mover writes its own position.json/local-polars.json/port-anchor
	// files — set the tree root on each nodeMover directly rather than routing writes
	// through a shared MoveDispatch-owned persister (docs/planning/decentralized-
	// persistence.md "The model"). A plain field write on each mover, done here before
	// any mover goroutine starts (Start runs after EnableEditPersist in every real call
	// path), so no synchronization is needed.
	for _, nm := range md.MR.NodeGeoms() {
		nm.SetPersistRoot(root)
	}
}
