package loadspec

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/dtauraso/wirefold/nodes/Wiring/polarindex"
	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

func loadSceneConstants(root string) (polarindex.SceneConstants, error) {
	path := filepath.Join(root, "constants.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return polarindex.SceneConstants{}, fmt.Errorf("loadTree: constants.json: read %s: %w", path, err)
	}
	var c polarindex.SceneConstants
	if err := json.Unmarshal(raw, &c); err != nil {
		return polarindex.SceneConstants{}, fmt.Errorf("loadTree: constants.json: parse %s: %w", path, err)
	}
	if c.ConstantR == 0 || c.ConstantPhi == 0 || c.ConstantTheta == 0 {
		return polarindex.SceneConstants{}, fmt.Errorf("loadTree: constants.json: %s is missing constantR/constantPhi/constantTheta", path)
	}
	if math.Abs(c.ConstantR-lattice.BeadStepR) > 1e-9 {
		return polarindex.SceneConstants{}, fmt.Errorf("loadTree: constants.json: %s constantR=%v disagrees with lattice.BeadStepR=%v — the radial grid is required to match the bead lattice spacing", path, c.ConstantR, lattice.BeadStepR)
	}
	return c, nil
}
