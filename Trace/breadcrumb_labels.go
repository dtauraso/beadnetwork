package Trace

const (
	BreadcrumbTopologyLoaded uint8 = iota
	BreadcrumbRowSeedCountMismatch
	BreadcrumbPoleToggleGo
	BreadcrumbWindowClear
	BreadcrumbWindowOpen
	BreadcrumbDwellStart
	BreadcrumbAbcDrag
	BreadcrumbWireSendBufferFull

	BreadcrumbDragCommit

	BreadcrumbWireBreadcrumbsDropped

	BreadcrumbChainAim

	BreadcrumbNeighborCenterRecv

	BreadcrumbNeighborSetCRecv

	BreadcrumbBeadCrud

	BreadcrumbPairSeedUnknown

	BreadcrumbPairLatticeAdopt

	BreadcrumbOutAngleFix

	BreadcrumbEdgeGeom

	BreadcrumbEdgeBeads

	BreadcrumbDragActivePersist
)

var BreadcrumbLabels = []string{
	"topology-loaded",
	"row-seed-count-mismatch",
	"pole-toggle-go",
	"window_clear",
	"window_open",
	"dwell_start",
	"abc-drag",
	"wire-send-buffer-full",
	"drag.commit",
	"wire-breadcrumbs-dropped",
	"chain-aim",
	"neighbor-center-recv",
	"neighbor-setc-recv",
	"bead-crud",
	"pair-seed-unknown",
	"pair-lattice-adopt",
	"out-angle-fix",
	"edge-geom",
	"edge-beads",
	"drag-active-persist",
}

func BreadcrumbLabelID(name string) (uint8, bool) {
	for i, n := range BreadcrumbLabels {
		if n == name {
			return uint8(i), true
		}
	}
	return 0, false
}
