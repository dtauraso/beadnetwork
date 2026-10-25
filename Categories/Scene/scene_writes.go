package Scene

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
)

func WriteSpawnIdentity(sceneRoot string) {
	if err := WriteAtomicIfChanged(SceneValuePath(sceneRoot, "spawn"), time.Now().UnixMilli()); err != nil {
		fmt.Fprintf(os.Stderr, "write spawn identity: %v\n", err)
	}
}

func WriteSceneSphere(sceneRoot string, s polar.SceneSphere) error {
	for name, value := range map[string]float64{
		"cx": s.Center.X, "cy": s.Center.Y, "cz": s.Center.Z,
		"radius": s.Radius,
	} {
		if err := WriteAtomicIfChanged(SceneValuePath(sceneRoot, name), value); err != nil {
			return err
		}
	}
	return nil
}

func WriteFrame(sceneRoot string, tPhi, tTheta int, cr, rho float64) error {
	for name, value := range map[string]any{"maxIndexPhi": tPhi, "maxIndexTheta": tTheta, "constantR": cr} {
		if err := WriteAtomicIfChanged(SceneValuePath(sceneRoot, name), value); err != nil {
			return err
		}
	}
	return WriteAtomicIfChanged(filepath.Join(ConstantsDir(sceneRoot), FileRho), rho)
}

func LogPersistErr(label, path string, err error) {
	fmt.Fprintf(os.Stderr, "%s: persist %s: %v\n", label, path, err)
}
