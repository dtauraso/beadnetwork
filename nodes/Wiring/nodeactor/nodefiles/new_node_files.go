package nodefiles

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/positionfile"
)

func nodeMetaFilePath(root, id string) string {
	return filepath.Join(root, "nodes", id, "meta.json")
}

func nodeRuleActiveFilePath(root, id string) string {
	return filepath.Join(root, "nodes", id, "rule-active.json")
}

func nodeDirPath(root, id string) string {
	return filepath.Join(root, "nodes", id)
}

type newNodePosition struct {
	ScenePolarR     float64 `json:"scenePolarR"`
	ScenePolarPhi   float64 `json:"scenePolarPhi"`
	ScenePolarTheta float64 `json:"scenePolarTheta"`
}

func entityReadModifyWrite(path string, mutate func(map[string]any)) error {
	m := map[string]any{}
	raw, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("entityReadModifyWrite: read %s: %w", path, err)
	}
	if err == nil {
		if err := json.Unmarshal(raw, &m); err != nil {
			return fmt.Errorf(
				"entityReadModifyWrite: %s exists but does not parse as JSON (%w) — refusing to "+
					"overwrite it, since that would discard every key this write does not set",
				path, err)
		}
	}
	mutate(m)
	return jsonpersist.WriteJSONAtomic(path, m)
}

func WriteNewNodeFiles(root, id, kind string, scenePolarR, phi, theta float64) error {
	dir := nodeDirPath(root, id)
	if err := os.MkdirAll(filepath.Join(dir, "edges"), 0o755); err != nil {
		return err
	}
	if err := entityReadModifyWrite(nodeMetaFilePath(root, id), func(m map[string]any) {
		m["id"] = id
		m["type"] = kind
	}); err != nil {
		return err
	}
	return jsonpersist.WriteJSONAtomic(positionfile.FilePath(root, id), newNodePosition{
		ScenePolarR: scenePolarR, ScenePolarPhi: phi, ScenePolarTheta: theta,
	})
}

func WriteOrbitRule(root, id string, rule *polar.OrbitRule) error {
	return entityReadModifyWrite(nodeMetaFilePath(root, id), func(m map[string]any) {
		if rule == nil {
			delete(m, "orbit")
			return
		}
		orbit := map[string]any{}
		if rule.Phi != nil {
			orbit["phi"] = *rule.Phi
		}
		if rule.MaxTheta != nil {
			orbit["maxTheta"] = *rule.MaxTheta
		}
		m["orbit"] = orbit
	})
}

type ruleActiveFile struct {
	Active bool `json:"active"`
}

func WriteOrbitActive(root, id string, active bool) error {
	return jsonpersist.WriteJSONAtomic(nodeRuleActiveFilePath(root, id), ruleActiveFile{Active: active})
}

func LoadOrbitActive(root, id string) bool {
	var f ruleActiveFile
	if !jsonpersist.ReadJSONIfExists(nodeRuleActiveFilePath(root, id), &f) {
		return true
	}
	return f.Active
}

func RemoveNodeDir(root, id string) error {
	return os.RemoveAll(nodeDirPath(root, id))
}
