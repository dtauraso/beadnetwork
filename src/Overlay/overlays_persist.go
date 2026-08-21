package Overlay

import (
	"path/filepath"
	"sort"

	"github.com/dtauraso/wirefold/src/valuefile"
)

// OverlayFlagFile resolves a flag against the SCENE ROOT using the generated
// declaration — the same one src/Overlay/paths/ hands the renderer. Building
// the name here instead is how Go came to write <flag>.json while the renderer
// read <flag>.bin.
func OverlayFlagFile(sceneRoot, flag string) string {
	rel, ok := FlagPath[flag]
	if !ok {
		panic("Overlay.OverlayFlagFile: flag " + flag + " has no entry in the generated FlagPath, so Go and the renderer disagree about where it lives; run go generate ./...")
	}
	return filepath.Join(sceneRoot, rel)
}

func WriteSceneOverlays(sceneRoot string, ov OverlayState) error {
	flags := make([]string, 0, len(OverlayFlagRead))
	for flag := range OverlayFlagRead {
		flags = append(flags, flag)
	}
	sort.Strings(flags)
	for _, flag := range flags {
		if err := valuefile.WriteAtomic(OverlayFlagFile(sceneRoot, flag), OverlayFlagRead[flag](&ov)); err != nil {
			return err
		}
	}
	return nil
}

func LoadSceneOverlays(sceneRoot string) (OverlayState, bool) {
	ov := DefaultOverlayState()
	found := false
	for flag, set := range OverlayFlagWrite {
		var v bool
		if valuefile.ReadIfExists(OverlayFlagFile(sceneRoot, flag), &v) {
			set(&ov, v)
			found = true
		}
	}
	return ov, found
}
