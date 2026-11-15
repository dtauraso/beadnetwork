package Dispatch

import (
	"slices"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills"
	Flags "github.com/dtauraso/beadnetwork/Categories/Scene/View/Flags"
)

func applyOverlayAct(md *MoveDispatch, a Sheet.Act) {
	if a.Op == Pills.OpOverlayBoth {
		flags := strings.Split(a.K, ",")
		allOn := !slices.ContainsFunc(flags, func(f string) bool {
			read, ok := Flags.OverlayFlagRead[f]
			return !ok || !read(&md.UI.OV)
		})
		Flags.SetCount(&md.UI.OV, &md.UI, flags, !allOn)
	} else {
		Flags.ToggleFlag(&md.UI.OV, &md.UI, a.K)
	}
	md.UI.PersistOverlays(md.UI.OV)
}
