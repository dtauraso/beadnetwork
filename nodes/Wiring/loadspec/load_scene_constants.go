package loadspec

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/dtauraso/wirefold/nodes/Wiring/polarindex"
	lattice "github.com/dtauraso/wirefold/nodes/bead/lattice"
)

func LoadSceneConstants(root string) (polarindex.SceneConstants, error) {
	return loadSceneConstants(root)
}

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
	if c.ConstantR == 0 || c.MaxIndexPhi <= 0 || c.MaxIndexTheta <= 0 {
		return polarindex.SceneConstants{}, fmt.Errorf("loadTree: constants.json: %s is missing constantR, or maxIndexPhi/maxIndexTheta is not a positive step count", path)
	}
	if c.MaxIndexPhi%2 != 0 || c.MaxIndexTheta%2 != 0 {
		return polarindex.SceneConstants{}, fmt.Errorf("loadTree: constants.json: %s maxIndexPhi=%d maxIndexTheta=%d must be EVEN — a half turn is maxIndex/2 steps, and an odd ring has no exact half", path, c.MaxIndexPhi, c.MaxIndexTheta)
	}
	slotsPerBead := lattice.BeadStepR / c.ConstantR
	if math.Abs(slotsPerBead-math.Round(slotsPerBead)) > 1e-9 {
		return polarindex.SceneConstants{}, fmt.Errorf("loadTree: constants.json: %s constantR=%v does not divide lattice.BeadStepR=%v a whole number of times (got %v) — one radial index is one slot a bead steps onto, so a bead width has to be a whole number of them or beads land off the bead lattice", path, c.ConstantR, lattice.BeadStepR, slotsPerBead)
	}
	return c, nil
}
