package nodefile

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/src/valuefile"
	"github.com/dtauraso/wirefold/src/Chrome/Panels/PolarRulesPanel"
)

const (
	FileType       = "type.bin"
	FileGate       = "gate.bin"
	FileIndexPhi   = "index-phi.bin"
	FileIndexTheta = "index-theta.bin"
	FileIndexR     = "index-r.bin"
	FileTiltIdx    = "top-tilt-vector-phi-idx.bin"

	DirDragRule = "drag-rule"
	DirSelfRule = "self-rule"

	FileRuleR        = "r.bin"
	FileRulePhi      = "phi.bin"
	FileRuleMaxTheta = "max-theta.bin"
)

func BaseDir(nodeDir string) string { return filepath.Join(nodeDir, "base") }

func ReadDragRule(dir string) *PolarRulesPanel.DragRule {
	var rule PolarRulesPanel.DragRule
	any := false
	read := func(name string, dst **float64) {
		var v float64
		if valuefile.ReadIfExists(filepath.Join(dir, name), &v) {
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
		if err := valuefile.RemoveIfPresent(filepath.Join(dir, name)); err != nil {
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
		return valuefile.WriteAtomic(filepath.Join(dir, name), *v)
	}
	if err := write(FileRuleR, rule.R); err != nil {
		return err
	}
	if err := write(FileRulePhi, rule.Phi); err != nil {
		return err
	}
	return write(FileRuleMaxTheta, rule.MaxTheta)
}
