package Overlay

import (
	"path/filepath"
	"sort"

	"github.com/dtauraso/wirefold/src/valuefile"
)

func OverlayFlagFile(overlaysDir, flag string) string {
	return filepath.Join(overlaysDir, flag+valuefile.Ext)
}

func WriteSceneOverlays(overlaysDir string, ov OverlayState) error {
	flags := make([]string, 0, len(OverlayFlagRead))
	for flag := range OverlayFlagRead {
		flags = append(flags, flag)
	}
	sort.Strings(flags)
	for _, flag := range flags {
		if err := valuefile.WriteAtomic(OverlayFlagFile(overlaysDir, flag), OverlayFlagRead[flag](&ov)); err != nil {
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
		if valuefile.ReadIfExists(OverlayFlagFile(overlaysDir, flag), &v) {
			set(&ov, v)
			found = true
		}
	}
	return ov, found
}
