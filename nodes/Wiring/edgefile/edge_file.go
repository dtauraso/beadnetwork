package edgefile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
)

type edgeFile struct {
	SourceHandle string `json:"sourceHandle"`
	Target       string `json:"target"`
	TargetHandle string `json:"targetHandle"`
	Kind         string `json:"kind"`
	Label        string `json:"label"`
}

func edgeFilePath(root, src, label string) string {
	return filepath.Join(root, "nodes", src, "edges", label+".json")
}

func edgesDirPath(root, id string) string {
	return filepath.Join(root, "nodes", id, "edges")
}

func SourceHandleFor(root, src, srcPort string, broadcast bool) (string, string, bool) {
	used := countHandlesOn(root, src, srcPort)
	if broadcast {
		return fmt.Sprintf("%s%d", srcPort, used), "", true
	}
	if used > 0 {
		return "", fmt.Sprintf("%s's %s is already connected, and it carries one edge", src, srcPort), false
	}
	return srcPort, "", true
}

func countHandlesOn(root, src, srcPort string) int {
	entries, err := os.ReadDir(edgesDirPath(root, src))
	if err != nil {
		return 0
	}
	used := 0
	for _, e := range entries {
		if e.IsDir() { // path-resolution-ok: skipping a stray directory, not resolving a scene path
			continue
		}
		var ef edgeFile
		jsonpersist.ReadJSONBestEffort(filepath.Join(edgesDirPath(root, src), e.Name()), &ef)
		if handleIsOn(ef.SourceHandle, srcPort) {
			used++
		}
	}
	return used
}

func handleIsOn(handle, srcPort string) bool {
	if handle == srcPort {
		return true
	}
	rest, cut := strings.CutPrefix(handle, srcPort)
	if !cut || rest == "" {
		return false
	}
	for _, r := range rest {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func WriteEdgeFile(root, src, srcPort, target, targetPort string) error {
	label := src + "To" + target
	path := edgeFilePath(root, src, label)
	return jsonpersist.ReadModifyWriteJSON(path, func(m map[string]any) {
		m["sourceHandle"] = srcPort
		m["target"] = target
		m["targetHandle"] = targetPort
		m["kind"] = "chain"
		m["label"] = label
	})
}

func RemoveEdgesTo(root, id string, nodeIDs []string) error {
	for _, n := range nodeIDs {
		entries, err := os.ReadDir(edgesDirPath(root, n))
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() { // path-resolution-ok: skipping a stray directory, not resolving a scene path
				continue
			}
			path := filepath.Join(edgesDirPath(root, n), e.Name())
			var ef edgeFile
			jsonpersist.ReadJSONBestEffort(path, &ef)
			if ef.Target != id {
				continue
			}
			if err := os.Remove(path); err != nil {
				return err
			}
			if err := os.Remove(edgeDragPath(root, n, ef.Label)); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}
	return nil
}
