package viewpersist

import (
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/Categories/Polar/polar"
	Scene "github.com/dtauraso/wirefold/Categories/Scene"
)

// The sphere is WRITTEN here, as it changes, and READ in Startup, when a scene
// opens. Those are two different moments with two different callers, so they
// are two different places — splitting on what happens when, not on the fact
// that both touch a file.
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

// LogPersistErr is this concern's own: a persist failure names the thing that
// failed to keep, and says so where it happened.
func LogPersistErr(label, path string, err error) {
	fmt.Fprintf(os.Stderr, "%s: persist %s: %v\n", label, path, err)
}
