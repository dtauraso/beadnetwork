package loadspec

import (
	"fmt"
	"math"
	"path/filepath"

	lattice "github.com/dtauraso/wirefold/src/Node/BeadAnimation/lattice"
	"github.com/dtauraso/wirefold/src/Polar/polarindex"
	"github.com/dtauraso/wirefold/src/valuefile"
)

const (
	ConstantsDirName  = "constants"
	FileConstantR     = "constant-r.bin"
	FileMaxIndexPhi   = "max-index-phi.bin"
	FileMaxIndexTheta = "max-index-theta.bin"
)

func ConstantsDir(root string) string { return filepath.Join(root, ConstantsDirName) }

func LoadSceneConstants(root string) (polarindex.SceneConstants, error) {
	return loadSceneConstants(root)
}

func loadSceneConstants(root string) (polarindex.SceneConstants, error) {
	dir := ConstantsDir(root)
	var c polarindex.SceneConstants
	missing := []string{}
	if !valuefile.ReadIfExists(filepath.Join(dir, FileConstantR), &c.ConstantR) {
		missing = append(missing, FileConstantR)
	}
	if !valuefile.ReadIfExists(filepath.Join(dir, FileMaxIndexPhi), &c.MaxIndexPhi) {
		missing = append(missing, FileMaxIndexPhi)
	}
	if !valuefile.ReadIfExists(filepath.Join(dir, FileMaxIndexTheta), &c.MaxIndexTheta) {
		missing = append(missing, FileMaxIndexTheta)
	}
	if len(missing) > 0 {
		return polarindex.SceneConstants{}, fmt.Errorf("loadTree: %s: missing %v — each scene constant is one file holding one value", dir, missing)
	}
	if c.ConstantR == 0 || c.MaxIndexPhi <= 0 || c.MaxIndexTheta <= 0 {
		return polarindex.SceneConstants{}, fmt.Errorf("loadTree: %s: constant-r is zero, or max-index-phi/max-index-theta is not a positive step count", dir)
	}
	if c.MaxIndexPhi%2 != 0 || c.MaxIndexTheta%2 != 0 {
		return polarindex.SceneConstants{}, fmt.Errorf("loadTree: %s: maxIndexPhi=%d maxIndexTheta=%d must be EVEN — a half turn is maxIndex/2 steps, and an odd ring has no exact half", dir, c.MaxIndexPhi, c.MaxIndexTheta)
	}
	if math.Abs(c.ConstantR-lattice.SlotR) > 1e-9 {
		return polarindex.SceneConstants{}, fmt.Errorf("loadTree: %s: constantR=%v is not lattice.SlotR=%v — one radial index IS one slot, the step a bead takes per wake, so the scene grid and the bead lattice are the same grid (a bead width is %v slots)", dir, c.ConstantR, lattice.SlotR, lattice.SlotsPerBead)
	}
	return c, nil
}
