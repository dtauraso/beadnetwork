package nodefiles

import (
	"os"
	"path/filepath"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/polarindex"
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
	Active     bool            `json:"active"`
	KindActive *bool           `json:"kindActive,omitempty"`
	EdgeActive map[string]bool `json:"edgeActive,omitempty"`
}

func WriteKindRuleActive(root, id string, active bool) error {
	return jsonpersist.ReadModifyWriteJSON(nodeRuleActiveFilePath(root, id), func(m map[string]any) {
		m["kindActive"] = active
	})
}

func WriteEdgeRuleActive(root, id, target string, active bool) error {
	return jsonpersist.ReadModifyWriteJSON(nodeRuleActiveFilePath(root, id), func(m map[string]any) {
		edges, _ := m["edgeActive"].(map[string]any)
		if edges == nil {
			edges = map[string]any{}
		}
		edges[target] = active
		m["edgeActive"] = edges
	})
}

func LoadEdgeRuleActive(root, id, target string) bool {
	var f ruleActiveFile
	if !jsonpersist.ReadJSONIfExists(nodeRuleActiveFilePath(root, id), &f) {
		return true
	}
	active, stored := f.EdgeActive[target]
	return !stored || active
}

func LoadKindRuleActive(root, id string) bool {
	var f ruleActiveFile
	if !jsonpersist.ReadJSONIfExists(nodeRuleActiveFilePath(root, id), &f) {
		return true
	}
	if f.KindActive == nil {
		return true
	}
	return *f.KindActive
}

func WriteDragActive(root, id string, active bool) error {
	return jsonpersist.ReadModifyWriteJSON(nodeRuleActiveFilePath(root, id), func(m map[string]any) {
		m["active"] = active
	})
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
