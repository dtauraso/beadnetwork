package Node

import (
	"fmt"
	"path/filepath"
	"strings"
)

func cardStateDir(root, id string) string {
	return filepath.Join(nodeDirPath(root, id), "data", "state")
}

func WriteCardState(root, id, key string, value int) error {
	if id == "" || id == "." || id == ".." || strings.ContainsAny(id, `/\`) {
		return fmt.Errorf("unsafe node id %q", id)
	}
	if key == "" || strings.ContainsAny(key, `/\.`) {
		return fmt.Errorf("unsafe card state key %q", key)
	}
	return WriteAtomicIfChanged(filepath.Join(cardStateDir(root, id), key+".bin"), value)
}

func (m *NodeGeometry) PersistRoot() string { return m.persistRoot }
