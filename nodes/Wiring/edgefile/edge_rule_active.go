package edgefile

import "github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"

type ruleActiveFields struct {
	DragActive *bool `json:"dragActive,omitempty"`
}

func LoadEdgeRuleActive(root, src, label string) bool {
	var f ruleActiveFields
	if !jsonpersist.ReadJSONIfExists(edgeFilePath(root, src, label), &f) {
		return true
	}
	if f.DragActive == nil {
		return true
	}
	return *f.DragActive
}

func WriteEdgeRuleActive(root, src, label string, active bool) error {
	return jsonpersist.ReadModifyWriteJSON(edgeFilePath(root, src, label), func(m map[string]any) {
		m["dragActive"] = active
	})
}
