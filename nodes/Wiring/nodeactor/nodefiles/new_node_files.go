package nodefiles

import (
	"os"
	"path/filepath"

	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/positionfile"
)

func nodeMetaFilePath(root, id string) string {
	return filepath.Join(root, "nodes", id, "meta.json")
}

func nodeDirPath(root, id string) string {
	return filepath.Join(root, "nodes", id)
}

type newNodeMeta struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type newNodePosition struct {
	ScenePolarR     float64 `json:"scenePolarR"`
	ScenePolarPhi   float64 `json:"scenePolarPhi"`
	ScenePolarTheta float64 `json:"scenePolarTheta"`
}

func WriteNewNodeFiles(root, id, kind string, scenePolarR, phi, theta float64) error {
	dir := nodeDirPath(root, id)
	if err := os.MkdirAll(filepath.Join(dir, "edges"), 0o755); err != nil {
		return err
	}
	if err := jsonpersist.WriteJSONAtomic(nodeMetaFilePath(root, id), newNodeMeta{ID: id, Type: kind}); err != nil {
		return err
	}
	return jsonpersist.WriteJSONAtomic(positionfile.FilePath(root, id), newNodePosition{
		ScenePolarR: scenePolarR, ScenePolarPhi: phi, ScenePolarTheta: theta,
	})
}

func RemoveNodeDir(root, id string) error {
	return os.RemoveAll(nodeDirPath(root, id))
}
