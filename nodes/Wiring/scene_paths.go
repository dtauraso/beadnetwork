package Wiring

// scene_paths.go — resolves ONLY the scene-level paths under a topology tree: the
// one-file-per-writer view/camera.json, view/overlays.json, view/sphere.json.
// topologyPath is always the tree root directory — LoadTopology rejects anything else
// (topo_spec.go's "a topology is a directory tree and nothing else") — so every function
// here is a plain filepath.Join, no os.Stat/IsDir resolution needed.
//
// The pre-split shared view/scene.json sidecar (cameraPolar/overlay-flags/sceneSphere all
// in one document) and its best-effort read fallback (sceneJSONPath/sceneCameraPath) were
// REMOVED — no such file exists anywhere in this repo's tree, and nothing wrote it once the
// one-file-per-writer split landed. A topology directory holding ONLY that legacy sidecar
// now loses its camera pose, overlay flags and scene sphere on load (falls back to
// defaultViewpoint / defaultOverlayState / a content-fit sphere) instead of migrating them
// forward. This is a compatibility removal, not a cleanup — see the git log for this file.
//
// Node/port/edge path construction does NOT live here: it lives with the goroutine that
// owns those files — node_mover.go (a node's own meta/position/data/local-polars/
// cascade-edges/inputs/outputs) and edge_mover.go (an edge's own
// nodes/<source>/edges/<label>.json). See MODEL.md / docs/planning/decentralized-
// persistence.md "The model": the owner writes the file AND owns the path.
//
// Guard: tools/check-scene-path-resolution.sh enforces the split by path pattern — see the
// guard's own header for the current rule.

import "path/filepath"

// sceneViewFilePath resolves <topologyPath>/view/<name>. Backs the one-file-per-writer
// split: camera.json, overlays.json and sphere.json each replace one of the three writers
// that used to share scene.json, and each resolves its path through this one shared helper.
func sceneViewFilePath(topologyPath, name string) string {
	return filepath.Join(topologyPath, "view", name)
}

// cameraFilePath is the WRITE-side location of the persisted camera pose — the sole
// successor to scene.json's cameraPolar key. writeSceneCameraPolar is its only writer.
func cameraFilePath(topologyPath string) string {
	return sceneViewFilePath(topologyPath, "camera.json")
}

// overlaysFilePath is the WRITE-side location of the persisted overlay-visibility flags —
// the sole successor to scene.json's overlay keys. writeSceneOverlays is its only writer.
func overlaysFilePath(topologyPath string) string {
	return sceneViewFilePath(topologyPath, "overlays.json")
}

// sphereFilePath is the WRITE-side location of the persisted scene sphere — the sole
// successor to scene.json's sceneSphere key. writeSceneSphere is its only writer.
func sphereFilePath(topologyPath string) string {
	return sceneViewFilePath(topologyPath, "sphere.json")
}
