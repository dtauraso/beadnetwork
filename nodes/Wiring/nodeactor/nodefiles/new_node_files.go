package nodefiles

import (
	"os"
	"path/filepath"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/gitskip"
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

func entityReadModifyWrite(path string, mutate func(map[string]any)) error {
	if err := jsonpersist.ReadModifyWriteJSON(path, mutate); err != nil {
		return err
	}
	gitskip.Mark(path)
	return nil
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
	return positionfile.Write(root, id, positionfile.JSON{
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
