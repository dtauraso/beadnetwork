package Node

import (
	"os"
	"path/filepath"

	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

func NodesDir(root string) string { return filepath.Join(root, "nodes") }

func nodeDirPath(root, id string) string {
	return filepath.Join(NodesDir(root), id)
}

func nodeBaseDir(root, id string) string {
	return BaseDir(nodeDirPath(root, id))
}

func WriteNewNodeFiles(root, id, kind string, p polar.Polar, sc polarindex.SceneConstants) error {
	dir := nodeDirPath(root, id)
	if err := os.MkdirAll(filepath.Join(dir, "edges"), 0o755); err != nil {
		return err
	}
	idx := polarindex.MeasureIndex(p, sc)
	base := nodeBaseDir(root, id)
	if err := WriteAtomic(filepath.Join(base, FileType), kind); err != nil {
		return err
	}
	for name, value := range map[string]int{
		FileIndexPhi:   idx.Phi,
		FileIndexTheta: idx.Theta,
		FileIndexR:     idx.R,
	} {
		if err := WriteAtomic(filepath.Join(base, name), value); err != nil {
			return err
		}
	}
	return nil
}

func WriteNewPlaceNode(root, id, kind string, p polarindex.Place) error {
	base := nodeBaseDir(root, id)
	for name, v := range map[string]any{FileType: kind, FileAtR: p.R, FileAtPhi: p.Phi, FileAtTheta: p.Theta} {
		if err := WriteAtomic(filepath.Join(base, name), v); err != nil {
			return err
		}
	}
	return nil
}

func RemoveNodeDir(root, id string) error {
	return os.RemoveAll(nodeDirPath(root, id))
}
