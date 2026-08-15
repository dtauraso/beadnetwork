package nodefiles

import (
	"os"
	"path/filepath"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
)

func nodeBaseFilePath(root, id string) string {
	return filepath.Join(root, "nodes", id, "base.json")
}

func nodeRuleActiveFilePath(root, id string) string {
	return filepath.Join(root, "nodes", id, "rule-active.json")
}

func nodeDirPath(root, id string) string {
	return filepath.Join(root, "nodes", id)
}

func entityReadModifyWrite(path string, mutate func(map[string]any)) error {
	return jsonpersist.ReadModifyWriteJSON(path, mutate)
}

func WriteNewNodeFiles(root, id, kind string, scenePolarR, phi, theta float64) error {
	dir := nodeDirPath(root, id)
	if err := os.MkdirAll(filepath.Join(dir, "edges"), 0o755); err != nil {
		return err
	}
	return entityReadModifyWrite(nodeBaseFilePath(root, id), func(m map[string]any) {
		m["id"] = id
		m["type"] = kind
		m["scenePolarR"] = scenePolarR
		m["scenePolarPhi"] = phi
		m["scenePolarTheta"] = theta
	})
}

func WriteDragRule(root, id string, rule *polar.DragRule) error {
	return entityReadModifyWrite(nodeBaseFilePath(root, id), func(m map[string]any) {
		if rule == nil {
			delete(m, "drag")
			return
		}
		drag := map[string]any{}
		if rule.Phi != nil {
			drag["phi"] = *rule.Phi
		}
		if rule.MaxTheta != nil {
			drag["maxTheta"] = *rule.MaxTheta
		}
		m["drag"] = drag
	})
}

type ruleActiveFile struct {
	Active bool `json:"active"`
}

func WriteDragActive(root, id string, active bool) error {
	return jsonpersist.WriteJSONAtomic(nodeRuleActiveFilePath(root, id), ruleActiveFile{Active: active})
}

func LoadDragActive(root, id string) bool {
	var f ruleActiveFile
	if !jsonpersist.ReadJSONIfExists(nodeRuleActiveFilePath(root, id), &f) {
		return true
	}
	return f.Active
}

func RemoveNodeDir(root, id string) error {
	return os.RemoveAll(nodeDirPath(root, id))
}
