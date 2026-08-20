package bufferlayout

const (
	BreadcrumbTopologyLoaded uint8 = iota
	BreadcrumbRowSeedCountMismatch
	BreadcrumbPoleToggleGo
	BreadcrumbWindowClear
	BreadcrumbWindowOpen
	BreadcrumbDwellStart
	BreadcrumbBeadPlaceBufferFull
	BreadcrumbDragCommit
	BreadcrumbBeadBreadcrumbsDropped
	BreadcrumbPairSeedUnknown
	BreadcrumbPairLatticeAdopt
	BreadcrumbViewport
)

var BreadcrumbLabels = []string{
	"topology-loaded",
	"row-seed-count-mismatch",
	"pole-toggle-go",
	"window_clear",
	"window_open",
	"dwell_start",
	"bead-place-buffer-full",
	"drag.commit",
	"bead-breadcrumbs-dropped",
	"pair-seed-unknown",
	"pair-lattice-adopt",
	"viewport",
}

func BreadcrumbLabelID(name string) (uint8, bool) {
	for i, n := range BreadcrumbLabels {
		if n == name {
			return uint8(i), true
		}
	}
	return 0, false
}
