package Scene

import (
	"fmt"
	"os"
	"time"

	"github.com/dtauraso/beadnetwork/Categories/Polar/polar"
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

func LogPersistErr(label, path string, err error) {
	fmt.Fprintf(os.Stderr, "%s: persist %s: %v\n", label, path, err)
}
