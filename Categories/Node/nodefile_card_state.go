package Node

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func cardStateDir(root, id string) string {
	return filepath.Join(dragDir(root, id), "state")
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

func ReadCardState(root, id string) map[string]int {
	dir := cardStateDir(root, id)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	out := map[string]int{}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".bin") {
			continue
		}
		var v int
		if ReadIfExists(filepath.Join(dir, name), &v) {
			out[strings.TrimSuffix(name, ".bin")] = v
		}
	}
	return out
}

func (m *NodeGeometry) PersistRoot() string { return m.persistRoot }
