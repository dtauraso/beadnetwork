package Node

import (
	"path/filepath"

	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

func NodeDir(root, id string) string { return nodeDirPath(root, id) }

func PlaceDirs(nodeDir string) (drag, base string) {
	return filepath.Join(nodeDir, "drag"), BaseDir(nodeDir)
}

func WritePlace(root, id string, p polarindex.Place) error {
	dir := dragDir(root, id)
	for name, v := range map[string]float64{FileAtR: p.R, FileAtPhi: p.Phi, FileAtTheta: p.Theta} {
		if err := WriteAtomicIfChanged(filepath.Join(dir, name), v); err != nil {
			return err
		}
	}
	return nil
}

func WriteNodeS(root, id string, s int) error {
	return WriteAtomicIfChanged(filepath.Join(dragDir(root, id), FileS), s)
}

func WriteNodeRef(root, id string, phi, theta, s int) error {
	for _, f := range []struct {
		name string
		v    int
	}{{FileRefPhi, phi}, {FileRefTheta, theta}, {FileRefS, s}} {
		if err := WriteAtomicIfChanged(filepath.Join(dragDir(root, id), f.name), f.v); err != nil {
			return err
		}
	}
	return nil
}

func (m *NodeGeometry) PersistRoot() string { return m.persistRoot }
