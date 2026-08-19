package OverlaysDropdown

import (
	"path/filepath"
	"sort"

	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
)

func OverlayFlagFile(overlaysDir, flag string) string {
	return filepath.Join(overlaysDir, flag+".json")
}

func WriteSceneOverlays(overlaysDir string, ov OverlayState) error {
	flags := make([]string, 0, len(OverlayFlagRead))
	for flag := range OverlayFlagRead {
		flags = append(flags, flag)
	}
	sort.Strings(flags)
	for _, flag := range flags {
		if err := jsonpersist.WriteJSONAtomic(OverlayFlagFile(overlaysDir, flag), OverlayFlagRead[flag](&ov)); err != nil {
			return err
		}
	}
	return nil
}

func LoadSceneOverlays(overlaysDir string) (OverlayState, bool) {
	ov := DefaultOverlayState()
	found := false
	for flag, set := range OverlayFlagWrite {
		var v bool
		if jsonpersist.ReadJSONIfExists(OverlayFlagFile(overlaysDir, flag), &v) {
			set(&ov, v)
			found = true
		}
	}
	return ov, found
}
