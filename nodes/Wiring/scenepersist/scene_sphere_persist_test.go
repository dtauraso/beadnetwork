package scenepersist

// scene_sphere_persist_test.go — round-trip test for the pure WriteSceneSphere/
// LoadSceneSphere pair, moved out of nodes/Wiring/scene_sphere_persist_test.go: it drove
// only these two functions and t.TempDir(), no MoveDispatch/loadTreeMD.

import (
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// TestSceneSphereRoundTrip: WriteSceneSphere then LoadSceneSphere returns the same sphere.
func TestSceneSphereRoundTrip(t *testing.T) {
	dir := t.TempDir()

	want := geom.SceneSphere{Center: wire.Vec3{X: 10, Y: -20, Z: 30}, Radius: 250}
	if err := WriteSceneSphere(scenepaths.SphereFilePath(dir), want); err != nil {
		t.Fatalf("WriteSceneSphere: %v", err)
	}
	got, ok := LoadSceneSphere(dir)
	if !ok {
		t.Fatal("LoadSceneSphere: ok=false after write")
	}
	if got != want {
		t.Fatalf("round-trip: got %+v want %+v", got, want)
	}
}
