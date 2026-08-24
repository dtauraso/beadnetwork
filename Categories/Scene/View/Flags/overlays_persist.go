package Flags

import (
	"path/filepath"
)

func BlockPath(sceneRoot string) string {
	return filepath.Join(sceneRoot, filepath.FromSlash(BlockRelPath))
}

func WriteSceneOverlays(sceneRoot string, ov OverlayState) error {
	w := NewBlobWriter(BlockPath(sceneRoot), FlagNames)
	w.Begin()
	for _, flag := range FlagNames {
		w.Bool(flag, OverlayFlagRead[flag](&ov))
	}
	return w.Flush()
}

func LoadSceneOverlays(sceneRoot string) (OverlayState, bool) {
	ov := DefaultOverlayState()
	r, ok := ReadBlob(BlockPath(sceneRoot), FlagNames)
	if !ok {
		return ov, false
	}
	found := false
	for _, flag := range FlagNames {
		v, got := r.Bool(flag)
		if !got {
			continue
		}
		OverlayFlagWrite[flag](&ov, v)
		found = true
	}
	return ov, found
}
