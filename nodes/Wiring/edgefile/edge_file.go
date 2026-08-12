package edgefile

import (
	"os"
	"path/filepath"

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

func WriteEdgeFile(root, src, srcPort, target, targetPort string) error {
	label := src + "To" + target
	return jsonpersist.WriteJSONAtomic(edgeFilePath(root, src, label), edgeFile{
		SourceHandle: srcPort, Target: target, TargetHandle: targetPort, Kind: "chain", Label: label,
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
		}
	}
	return nil
}
