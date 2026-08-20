package nodefile

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/src/Node/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/src/Chrome/PolarRulesPanel"
)

const (
	FileType       = "type.json"
	FileGate       = "gate.json"
	FileIndexPhi   = "index-phi.json"
	FileIndexTheta = "index-theta.json"
	FileIndexR     = "index-r.json"
	FileTiltIdx    = "top-tilt-vector-phi-idx.json"

	DirDragRule = "drag-rule"
	DirSelfRule = "self-rule"

	FileRuleR        = "r.json"
	FileRulePhi      = "phi.json"
	FileRuleMaxTheta = "max-theta.json"
)

func BaseDir(nodeDir string) string { return filepath.Join(nodeDir, "base") }

func ReadDragRule(dir string) *PolarRulesPanel.DragRule {
	var rule PolarRulesPanel.DragRule
	any := false
	read := func(name string, dst **float64) {
		var v float64
		if jsonpersist.ReadJSONIfExists(filepath.Join(dir, name), &v) {
			*dst = &v
			any = true
		}
	}
	read(FileRuleR, &rule.R)
	read(FileRulePhi, &rule.Phi)
	read(FileRuleMaxTheta, &rule.MaxTheta)
	if !any {
		return nil
	}
	return &rule
}

func WriteDragRule(dir string, rule *PolarRulesPanel.DragRule) error {
	for _, name := range []string{FileRuleR, FileRulePhi, FileRuleMaxTheta} {
		if err := jsonpersist.RemoveIfPresent(filepath.Join(dir, name)); err != nil {
			return err
		}
	}
	if rule == nil {
		return nil
	}
	write := func(name string, v *float64) error {
		if v == nil {
			return nil
		}
		return jsonpersist.WriteJSONAtomic(filepath.Join(dir, name), *v)
	}
	if err := write(FileRuleR, rule.R); err != nil {
		return err
	}
	if err := write(FileRulePhi, rule.Phi); err != nil {
		return err
	}
	return write(FileRuleMaxTheta, rule.MaxTheta)
}
