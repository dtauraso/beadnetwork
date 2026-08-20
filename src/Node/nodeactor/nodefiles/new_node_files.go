package nodefiles

import (
	"github.com/dtauraso/wirefold/src/Chrome/Panels/PolarRulesPanel"
	"os"
	"path/filepath"

	"github.com/dtauraso/wirefold/src/Polar/polar"
	"github.com/dtauraso/wirefold/src/jsonpersist"
	"github.com/dtauraso/wirefold/src/Node/nodefile"
	"github.com/dtauraso/wirefold/src/Polar/polarindex"
)

func nodeDirPath(root, id string) string {
	return filepath.Join(root, "nodes", id)
}

func nodeBaseDir(root, id string) string {
	return nodefile.BaseDir(nodeDirPath(root, id))
}

func WriteNewNodeFiles(root, id, kind string, p polar.Polar, sc polarindex.SceneConstants) error {
	dir := nodeDirPath(root, id)
	if err := os.MkdirAll(filepath.Join(dir, "edges"), 0o755); err != nil {
		return err
	}
	idx := polarindex.MeasureIndex(p, sc)
	base := nodeBaseDir(root, id)
	if err := jsonpersist.WriteJSONAtomic(filepath.Join(base, nodefile.FileType), kind); err != nil {
		return err
	}
	for name, value := range map[string]int{
		nodefile.FileIndexPhi:   idx.Phi,
		nodefile.FileIndexTheta: idx.Theta,
		nodefile.FileIndexR:     idx.R,
	} {
		if err := jsonpersist.WriteJSONAtomic(filepath.Join(base, name), value); err != nil {
			return err
		}
	}
	return nil
}

func WriteDragRule(root, id string, rule *PolarRulesPanel.DragRule) error {
	return nodefile.WriteDragRule(filepath.Join(nodeBaseDir(root, id), nodefile.DirDragRule), rule)
}

const (
	FileDragActive = "drag.json"
	FileKindActive = "kind.json"
	FileSelfActive = "self.json"
	DirEdgeActive  = "edges"
)

func ruleActiveFilePath(root, id, name string) string {
	return filepath.Join(root, "nodes", id, "rule-active", name)
}

func edgeActiveFilePath(root, id, target string) string {
	return filepath.Join(root, "nodes", id, "rule-active", DirEdgeActive, target+".json")
}

func readActive(path string) bool {
	var v bool
	if !jsonpersist.ReadJSONIfExists(path, &v) {
		return true
	}
	return v
}

func WriteKindRuleActive(root, id string, active bool) error {
	return jsonpersist.WriteJSONAtomic(ruleActiveFilePath(root, id, FileKindActive), active)
}

func LoadKindRuleActive(root, id string) bool {
	return readActive(ruleActiveFilePath(root, id, FileKindActive))
}

func WriteEdgeRuleActive(root, id, target string, active bool) error {
	return jsonpersist.WriteJSONAtomic(edgeActiveFilePath(root, id, target), active)
}

func LoadEdgeRuleActive(root, id, target string) bool {
	return readActive(edgeActiveFilePath(root, id, target))
}

func WriteSelfDragRule(root, id string, rule *PolarRulesPanel.DragRule) error {
	return nodefile.WriteDragRule(filepath.Join(nodeBaseDir(root, id), nodefile.DirSelfRule), rule)
}

func WriteSelfRuleActive(root, id string, active bool) error {
	return jsonpersist.WriteJSONAtomic(ruleActiveFilePath(root, id, FileSelfActive), active)
}

func LoadSelfRuleActive(root, id string) bool {
	return readActive(ruleActiveFilePath(root, id, FileSelfActive))
}

func WriteDragActive(root, id string, active bool) error {
	return jsonpersist.WriteJSONAtomic(ruleActiveFilePath(root, id, FileDragActive), active)
}

func LoadDragActive(root, id string) bool {
	return readActive(ruleActiveFilePath(root, id, FileDragActive))
}

func RemoveNodeDir(root, id string) error {
	return os.RemoveAll(nodeDirPath(root, id))
}
