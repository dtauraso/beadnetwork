package Scene

import (
	"fmt"
	"os"
	"time"
)

func WriteSpawnIdentity(sceneRoot string) {
	if err := WriteAtomicIfChanged(SceneValuePath(sceneRoot, "spawn"), time.Now().UnixMilli()); err != nil {
		fmt.Fprintf(os.Stderr, "write spawn identity: %v\n", err)
	}
}
