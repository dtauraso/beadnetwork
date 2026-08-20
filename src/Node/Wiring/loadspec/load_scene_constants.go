package loadspec

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/dtauraso/wirefold/src/Node/Wiring/polarindex"
	lattice "github.com/dtauraso/wirefold/src/Node/wire/lattice"
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
	if math.Abs(c.ConstantR-lattice.SlotR) > 1e-9 {
		return polarindex.SceneConstants{}, fmt.Errorf("loadTree: constants.json: %s constantR=%v is not lattice.SlotR=%v — one radial index IS one slot, the step a bead takes per wake, so the scene grid and the bead lattice are the same grid (a bead width is %v slots)", path, c.ConstantR, lattice.SlotR, lattice.SlotsPerBead)
	}
	return c, nil
}
