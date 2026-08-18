package nodefiles

import (
	"github.com/dtauraso/wirefold/tools/topology-vscode/PolarRulesPanel"
	"os"
	"path/filepath"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/polarindex"
)

func nodeBaseFilePath(root, id string) string {
	return filepath.Join(root, "nodes", id, "base.json")
}

func nodeDirPath(root, id string) string {
	return filepath.Join(root, "nodes", id)
}

func entityReadModifyWrite(path string, mutate func(map[string]any)) error {
	return jsonpersist.ReadModifyWriteJSON(path, mutate)
}

func WriteNewNodeFiles(root, id, kind string, p polar.Polar, sc polarindex.SceneConstants) error {
	dir := nodeDirPath(root, id)
	if err := os.MkdirAll(filepath.Join(dir, "edges"), 0o755); err != nil {
		return err
	}
	idx := polarindex.MeasureIndex(p, sc)
	return entityReadModifyWrite(nodeBaseFilePath(root, id), func(m map[string]any) {
		m["id"] = id
		m["type"] = kind
		m["indexPhi"] = idx.Phi
		m["indexTheta"] = idx.Theta
		m["indexR"] = idx.R
	})
}

func WriteDragRule(root, id string, rule *PolarRulesPanel.DragRule) error {
	return entityReadModifyWrite(nodeBaseFilePath(root, id), func(m map[string]any) {
		if rule == nil {
			delete(m, "drag")
			return
		}
		drag := map[string]any{}
		if rule.R != nil {
			drag["r"] = *rule.R
		}
		if rule.Phi != nil {
			drag["phi"] = *rule.Phi
		}
		if rule.MaxTheta != nil {
			drag["maxTheta"] = *rule.MaxTheta
		}
		m["drag"] = drag
	})
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
	return entityReadModifyWrite(nodeBaseFilePath(root, id), func(m map[string]any) {
		if rule == nil {
			delete(m, "selfDrag")
			return
		}
		drag := map[string]any{}
		if rule.R != nil {
			drag["r"] = *rule.R
		}
		if rule.Phi != nil {
			drag["phi"] = *rule.Phi
		}
		if rule.MaxTheta != nil {
			drag["maxTheta"] = *rule.MaxTheta
		}
		m["selfDrag"] = drag
	})
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
