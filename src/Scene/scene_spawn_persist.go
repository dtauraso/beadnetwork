package Scene

import (
	"fmt"
	"os"
	"time"

	"github.com/dtauraso/wirefold/src/valuefile"
)

func WriteSpawnIdentity(sceneRoot string) {
	if err := valuefile.WriteAtomicIfChanged(SceneValuePath(sceneRoot, "spawn"), time.Now().UnixMilli()); err != nil {
		fmt.Fprintf(os.Stderr, "write spawn identity: %v\n", err)
	}
}
