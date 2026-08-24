package edgefile

import (
	"fmt"
	"github.com/dtauraso/beadnetwork/Categories/Node/Edge/edgetable"
	"os"
	"path/filepath"
	"strings"
)

func edgeDirPath(root, src, label string) string {
	return filepath.Join(root, "nodes", src, "edges", label)
}

func readEdgeString(root, src, label, name string) string {
	var v string
	ReadIfExists(filepath.Join(edgeDirPath(root, src, label), name), &v)
	return v
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
		if !e.IsDir() { // path-resolution-ok: an edge is a directory now; skip strays
			continue
		}
		if handleIsOn(readEdgeString(root, src, e.Name(), FileSourceHandle), srcPort) {
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
	dir := edgeDirPath(root, src, edgetable.ChannelName(src, target))
	for name, value := range map[string]string{
		FileSourceHandle: srcPort,
		FileTarget:       target,
		FileTargetHandle: targetPort,
		FileKind:         "chain",
	} {
		if err := WriteAtomic(filepath.Join(dir, name), value); err != nil {
			return err
		}
	}
	return nil
}

func RemoveEdgesTo(root, id string, nodeIDs []string) error {
	for _, n := range nodeIDs {
		entries, err := os.ReadDir(edgesDirPath(root, n))
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() { // path-resolution-ok: an edge is a directory now; skip strays
				continue
			}
			label := e.Name()
			if readEdgeString(root, n, label, FileTarget) != id {
				continue
			}
			if err := os.RemoveAll(edgeDirPath(root, n, label)); err != nil {
				return err
			}
			if err := os.RemoveAll(edgeDragDir(root, n, label)); err != nil {
				return err
			}
		}
	}
	return nil
}
