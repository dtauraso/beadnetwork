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

func Without(ov OverlayState, flags []string) OverlayState {
	return setAll(ov, flags, false)
}

func With(ov OverlayState, flags []string) OverlayState {
	return setAll(ov, flags, true)
}

func setAll(ov OverlayState, flags []string, v bool) OverlayState {
	for _, flag := range flags {
		OverlayFlagWrite[flag](&ov, v)
	}
	return ov
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
