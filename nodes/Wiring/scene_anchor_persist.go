package Wiring

// scene_anchor_persist.go — the WRITE side of port-anchor-as-file-data.
//
// The read side (loader_tree.go readPorts → specPort.AnchorId) loads each port's ring-anchor
// index from `<root>/nodes/<id>/{inputs,outputs}/<port>.json`'s `anchorId` field. This file
// is the mirror: when the gesture FSM commits a ring-move (applyRingAnchor snaps the port to
// a ring-anchor index and routes a moveMsgKindAnchor to the node's own mover), the node's OWN
// nodeMover persists that index back to the same port file — on its own goroutine, in its
// handle() (node_mover.go) — PRESERVING the other fields (name), so ring-move-then-reload
// round-trips.
//
// Same shape as the node-position writers: SYNCHRONOUS (persistPortAnchor writes
// immediately, inline on the node's own mover goroutine as it processes the anchor message —
// see scene_persist.go's header comment for why the prior debounce was removed),
// READ-MODIFY-WRITE (only `anchorId` is replaced), FIRE-AND-FORGET. nm.persistRoot == ""
// (unarmed) disables it.
//
// Path construction (nodePortFilePath) lives in node_mover.go, not here — a port's
// path belongs to its owning node (docs/planning/decentralized-persistence.md
// "The model").

import (
	"encoding/json"
	"fmt"
)

// persistPortAnchor writes THIS node's own port-anchor change to its port file,
// synchronously, on THIS node's own mover goroutine (handle's moveMsgKindAnchor case calls
// it right after mutating m.geom's held AnchorId). nm.persistRoot == "" (unarmed) is a
// no-op.
func (nm *nodeMover) persistPortAnchor(port string, isInput bool, anchorID int) {
	if nm.persistRoot == "" {
		return
	}
	if err := writePortAnchor(nm.persistRoot, nm.id, port, isInput, anchorID); err != nil {
		logPersistErr("scene_anchor_persist", nm.id+"/"+port, err)
	}
}

// writePortAnchor sets ONLY the anchorId field of the port file, preserving the other fields
// (name). The port file must already exist (a placed port always has one).
func writePortAnchor(root, node, port string, isInput bool, anchorID int) error {
	if !safeTreePathComponent(node) || !safeTreePathComponent(port) {
		return fmt.Errorf("unsafe node/port %q/%q", node, port)
	}
	// nodePortFilePath (node_mover.go) is the sole path constructor for a port's
	// geometry file — a node owns its own port paths.
	path := nodePortFilePath(root, node, port, isInput)
	return entityReadModifyWrite(path, func(obj map[string]json.RawMessage) {
		b, _ := json.Marshal(anchorID)
		obj["anchorId"] = b
	})
}
