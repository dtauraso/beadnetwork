package viewpersist

import (
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/Categories/Polar/polar"
	Scene "github.com/dtauraso/wirefold/Categories/Scene"
)

func WriteSceneSphere(sceneRoot string, s polar.SceneSphere) error {
	for name, value := range map[string]float64{
		"cx": s.Center.X, "cy": s.Center.Y, "cz": s.Center.Z,
		"radius": s.Radius,
	} {
		if err := Scene.WriteAtomicIfChanged(Scene.SceneValuePath(sceneRoot, name), value); err != nil {
			return err
		}
	}
	return nil
}

func LogPersistErr(label, path string, err error) {
	fmt.Fprintf(os.Stderr, "%s: persist %s: %v\n", label, path, err)
}
